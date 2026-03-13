package export

import "context"

// Exporter sends events to external consumers.
// Implementations may write to a message queue, database, file, etc.
type Exporter interface {
	// Export sends a batch of events.
	// Events within a batch are from a single transaction and must be
	// delivered atomically (all or none) and in order.
	Export(ctx context.Context, events []Event) error
}

// ExporterFunc is an adapter to use ordinary functions as Exporters.
type ExporterFunc func(ctx context.Context, events []Event) error

func (f ExporterFunc) Export(ctx context.Context, events []Event) error {
	return f(ctx, events)
}

// MultiExporter fans out events to multiple exporters.
type MultiExporter struct {
	exporters []Exporter
}

// NewMultiExporter creates an exporter that sends to all provided exporters.
func NewMultiExporter(exporters ...Exporter) *MultiExporter {
	return &MultiExporter{exporters: exporters}
}

func (m *MultiExporter) Export(ctx context.Context, events []Event) error {
	for _, e := range m.exporters {
		if err := e.Export(ctx, events); err != nil {
			return err
		}
	}
	return nil
}
