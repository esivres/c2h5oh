// Package engine implements the stream processor — the partition loop
// that dequeues intents, executes behaviors, and publishes events.
package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/esivres/c2h5oh/pkg/processing/behavior"
	"github.com/esivres/c2h5oh/pkg/processing/export"
	"github.com/esivres/c2h5oh/pkg/processing/intent"
	"github.com/esivres/c2h5oh/pkg/processing/storage"
)

// IntentObserver is called after an intent has been successfully processed.
type IntentObserver func(i intent.Intent)

// Processor is the main processing loop for a single partition.
// It dequeues intents, looks up behaviors, executes them transactionally,
// and publishes events. All processing is sequential (single-writer per partition).
type Processor struct {
	store       storage.Store
	exporter    export.Exporter
	registry    *behavior.Registry
	queue       *Queue
	keyGen      *KeyGenerator
	collector   *export.Collector
	logger      *slog.Logger
	retryPolicy RetryPolicy
	observers   []IntentObserver
	partitionId uint8
}

// Config holds configuration for a Processor.
type Config struct {
	Store         storage.Store
	Exporter      export.Exporter
	Registry      *behavior.Registry
	RetryPolicy   *RetryPolicy
	TimerChecker  *TimerCheckerConfig
	Logger        *slog.Logger
	Observers     []IntentObserver
	QueueSize     int
	StartSequence uint64
	StartPosition uint64
	PartitionId   uint8
}

// NewProcessor creates a new partition processor.
func NewProcessor(cfg Config) *Processor {
	if cfg.QueueSize == 0 {
		cfg.QueueSize = 256
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}
	if cfg.Exporter == nil {
		cfg.Exporter = export.ExporterFunc(func(_ context.Context, _ []export.Event) error { return nil })
	}

	rp := DefaultRetryPolicy()
	if cfg.RetryPolicy != nil {
		rp = *cfg.RetryPolicy
	}

	return &Processor{
		partitionId: cfg.PartitionId,
		store:       cfg.Store,
		registry:    cfg.Registry,
		exporter:    cfg.Exporter,
		queue:       NewQueue(cfg.QueueSize, nil),
		keyGen:      NewKeyGenerator(cfg.PartitionId, cfg.StartSequence),
		collector:   export.NewCollector(cfg.StartPosition),
		retryPolicy: rp,
		observers:   cfg.Observers,
		logger:      cfg.Logger.With("partition", cfg.PartitionId),
	}
}

// Submit adds an intent to the processing queue.
// The intent key is assigned automatically.
func (p *Processor) Submit(i intent.Intent) {
	p.assignKey(i)
	p.queue.Enqueue(i)
}

// TrySubmit adds an intent without blocking. Returns false if queue is full.
func (p *Processor) TrySubmit(i intent.Intent) bool {
	p.assignKey(i)
	return p.queue.TryEnqueue(i)
}

// Run starts the processing loop. Blocks until ctx is canceled.
func (p *Processor) Run(ctx context.Context) error {
	return p.RunWithTimerChecker(ctx, nil)
}

// RunWithTimerChecker starts the processing loop with an optional timer checker.
func (p *Processor) RunWithTimerChecker(ctx context.Context, timerCfg *TimerCheckerConfig) error {
	p.logger.Info("processor started")
	defer p.logger.Info("processor stopped")

	if timerCfg != nil {
		go p.runTimerChecker(ctx, *timerCfg)
	}

	done := ctx.Done()
	for {
		i, ok := p.queue.Dequeue(done)
		if !ok {
			return ctx.Err()
		}

		if err := p.processIntent(ctx, i); err != nil {
			p.logger.Error("intent processing failed",
				"intent", i.IntentType(),
				"error", err,
			)
			p.handleError(ctx, i, err)
		}
	}
}

// processIntent executes a single intent transactionally.
func (p *Processor) processIntent(ctx context.Context, i intent.Intent) error {
	// Unwrap retry envelope to get the actual intent for behavior lookup
	actual := unwrapIntent(i)

	// Secondary dispatch by element type if the intent supports it
	var b behavior.Behavior
	if et, ok := actual.(behavior.ElementTyped); ok && et.GetElementType() != "" {
		b = p.registry.LookupWithElementType(actual.IntentType(), et.GetElementType())
	} else {
		b = p.registry.Lookup(actual.IntentType())
	}
	if b == nil {
		return fmt.Errorf("no behavior registered for intent %q", actual.IntentType())
	}

	p.logger.Debug("processing intent",
		"intent", actual.IntentType(),
		"processInstance", actual.GetProcessInstanceKey(),
	)

	// Execute behavior within a storage transaction
	handler := behavior.AsHandler(b, actual)
	followUp, err := p.store.Execute(ctx, handler)
	if err != nil {
		return err
	}

	// Emit event for the processed intent
	p.emitIntentEvent(actual)

	// Notify observers
	for _, obs := range p.observers {
		obs(actual)
	}

	// Publish collected events after successful commit
	if p.collector.Len() > 0 {
		if expErr := p.exporter.Export(ctx, p.collector.Events()); expErr != nil {
			p.logger.Error("event export failed", "error", expErr)
		}
		p.collector.Reset()
	}

	// Enqueue follow-up intents
	for _, fi := range followUp {
		p.assignKey(fi)
		p.queue.Enqueue(fi)
	}

	return nil
}

// handleError applies error handling strategy based on intent origin.
func (p *Processor) handleError(ctx context.Context, i intent.Intent, err error) {
	switch i.GetOrigin() {
	case intent.Internal:
		// Internal intent failure → incident immediately
		p.createIncident(ctx, unwrapIntent(i), err)

	case intent.External:
		// External intent failure → retry with backoff, then incident
		attempt := 0
		if env, ok := i.(*retryEnvelope); ok {
			attempt = env.attempt
		}

		if attempt < p.retryPolicy.MaxRetries {
			delay := p.retryPolicy.interval(attempt)
			p.logger.Warn("retrying external intent",
				"intent", i.IntentType(),
				"attempt", attempt+1,
				"delay", delay,
				"error", err,
			)
			envelope := &retryEnvelope{
				wrapped: unwrapIntent(i),
				attempt: attempt + 1,
			}
			time.AfterFunc(delay, func() {
				if !p.queue.TryEnqueue(envelope) {
					p.logger.Warn("retry enqueue failed (queue full or shutting down)",
						"intent", envelope.IntentType())
				}
			})
		} else {
			p.logger.Error("external intent retries exhausted",
				"intent", i.IntentType(),
				"attempts", attempt,
				"error", err,
			)
			p.createIncident(ctx, unwrapIntent(i), err)
		}
	}
}

// createIncident submits a CreateIncidentIntent for a failed intent.
// Does not create incident for failed incident intents (prevents infinite loop).
func (p *Processor) createIncident(_ context.Context, failed intent.Intent, err error) {
	if failed.IntentType() == intent.CreateIncident || failed.IntentType() == intent.ResolveIncident {
		p.logger.Error("incident intent failed — dropping to prevent infinite loop",
			"intent", failed.IntentType(),
			"error", err,
		)
		return
	}

	incident := &intent.CreateIncidentIntent{
		Header: intent.Header{
			Origin:             intent.Internal,
			ProcessInstanceKey: failed.GetProcessInstanceKey(),
		},
		ErrorType:    "INTENT_PROCESSING",
		ErrorMessage: err.Error(),
	}

	p.Submit(incident)
}

// emitIntentEvent generates an export event based on the processed intent.
func (p *Processor) emitIntentEvent(i intent.Intent) {
	type keyGetter interface {
		GetKey() uint64
	}
	var key uint64
	if kg, ok := i.(keyGetter); ok {
		key = kg.GetKey()
	}

	piKey := i.GetProcessInstanceKey()

	switch v := i.(type) {
	case *intent.DeployProcessIntent:
		p.collector.Record(export.ValueProcessDefinition, export.RecordDeployed,
			key, 0, key, "", "", nil)
		_ = v

	case *intent.CreateProcessInstanceIntent:
		p.collector.Record(export.ValueProcessInstance, export.RecordCreated,
			key, key, v.ProcessDefinitionKey, "", "process", nil)

	case *intent.CompleteProcessInstanceIntent:
		p.collector.Record(export.ValueProcessInstance, export.RecordCompleted,
			key, piKey, 0, "", "process", nil)

	case *intent.CancelProcessInstanceIntent:
		p.collector.Record(export.ValueProcessInstance, export.RecordTerminated,
			key, piKey, 0, "", "process", nil)

	case *intent.ActivateElementIntent:
		p.collector.Record(export.ValueElementInstance, export.RecordActivated,
			key, piKey, v.ProcessDefinitionKey, v.ElementId, v.ElementType, nil)

	case *intent.ElementActivatedIntent:
		// Phase 2 of activation — no additional export event needed
		// (the ActivateElementIntent already emitted the Activated event)

	case *intent.CompleteElementIntent:
		p.collector.Record(export.ValueElementInstance, export.RecordCompleted,
			key, piKey, 0, "", "", nil)

	case *intent.ElementCompletedIntent:
		// Phase 2 of completion — no additional export event needed
		// (the CompleteElementIntent already emitted the Completed event)

	case *intent.TerminateElementIntent:
		p.collector.Record(export.ValueElementInstance, export.RecordTerminated,
			key, piKey, 0, "", "", nil)

	case *intent.CreateJobIntent:
		export.EmitJobEvent(p.collector, export.RecordCreated,
			key, piKey, v.ProcessDefinitionKey, "", v.Type)

	case *intent.ActivateJobIntent:
		export.EmitJobEvent(p.collector, export.RecordActivated,
			v.JobKey, piKey, 0, "", "")

	case *intent.CompleteJobIntent:
		export.EmitJobEvent(p.collector, export.RecordCompleted,
			v.JobKey, piKey, 0, "", "")

	case *intent.FailJobIntent:
		export.EmitJobEvent(p.collector, export.RecordFailed,
			v.JobKey, piKey, 0, "", "")

	case *intent.ThrowJobErrorIntent:
		export.EmitJobEvent(p.collector, export.RecordFailed,
			v.JobKey, piKey, 0, "", "")

	case *intent.TimeOutJobIntent:
		export.EmitJobEvent(p.collector, export.RecordTimedOut,
			v.JobKey, piKey, 0, "", "")

	case *intent.UpdateJobRetriesIntent:
		export.EmitJobEvent(p.collector, export.RecordUpdated,
			v.JobKey, piKey, 0, "", "")

	case *intent.UpdateJobTimeoutIntent:
		export.EmitJobEvent(p.collector, export.RecordUpdated,
			v.JobKey, piKey, 0, "", "")

	case *intent.DeleteResourceIntent:
		p.collector.Record(export.ValueProcessDefinition, export.RecordDeleted,
			v.ResourceKey, 0, v.ResourceKey, "", "", nil)

	case *intent.CreateIncidentIntent:
		p.collector.Record(export.ValueIncident, export.RecordCreated,
			key, piKey, 0, "", "", nil)

	case *intent.ResolveIncidentIntent:
		p.collector.Record(export.ValueIncident, export.RecordResolved,
			v.IncidentKey, piKey, 0, "", "", nil)

	case *intent.CreateTimerIntent:
		p.collector.Record(export.ValueTimer, export.RecordCreated,
			key, piKey, v.ProcessDefinitionKey, "", "", nil)

	case *intent.TriggerTimerIntent:
		p.collector.Record(export.ValueTimer, export.RecordTriggered,
			v.TimerKey, piKey, 0, "", "", nil)

	case *intent.CancelTimerIntent:
		p.collector.Record(export.ValueTimer, export.RecordCanceled,
			v.TimerKey, piKey, 0, "", "", nil)

	case *intent.PublishMessageIntent:
		p.collector.Record(export.ValueMessage, export.RecordCreated,
			key, 0, 0, "", "", nil)

	case *intent.OpenSubscriptionIntent:
		p.collector.Record(export.ValueMessageSubscription, export.RecordCreated,
			key, piKey, 0, "", "", nil)

	case *intent.CorrelateMessageIntent:
		p.collector.Record(export.ValueMessageSubscription, export.RecordCorrelated,
			v.SubscriptionKey, piKey, 0, "", "", nil)

	case *intent.CloseSubscriptionIntent:
		p.collector.Record(export.ValueMessageSubscription, export.RecordCanceled,
			v.SubscriptionKey, piKey, 0, "", "", nil)
	}
}

// assignKey sets the key on an intent header if it's zero.
func (p *Processor) assignKey(i intent.Intent) {
	type keyAssigner interface {
		AssignKey(uint64)
	}
	if ka, ok := i.(keyAssigner); ok {
		ka.AssignKey(p.keyGen.Next())
	}
}
