package export

import "time"

// Collector accumulates events during a transaction.
// After successful commit, the collected events are sent to the Exporter.
//
// Usage within the partition loop:
//
//	collector := export.NewCollector()
//	// pass collector into storage transaction context
//	// repository implementations call collector.Record(...)
//	// after commit:
//	exporter.Export(ctx, collector.Events())
type Collector struct {
	events   []Event
	position uint64
}

// NewCollector creates a new event collector with a starting position.
func NewCollector(startPosition uint64) *Collector {
	return &Collector{position: startPosition}
}

// Record adds an event to the collector.
// Position is assigned automatically.
func (c *Collector) Record(
	valueType ValueType,
	recordType RecordType,
	key uint64,
	processInstanceKey uint64,
	processDefinitionKey uint64,
	elementId string,
	elementType string,
	value []byte,
) {
	c.position++
	c.events = append(c.events, Event{
		Position:             c.position,
		ValueType:            valueType,
		RecordType:           recordType,
		Key:                  key,
		ProcessInstanceKey:   processInstanceKey,
		ProcessDefinitionKey: processDefinitionKey,
		ElementId:            elementId,
		ElementType:          elementType,
		Timestamp:            time.Now(),
		Value:                value,
	})
}

// Events returns all collected events.
func (c *Collector) Events() []Event {
	return c.events
}

// Len returns the number of collected events.
func (c *Collector) Len() int {
	return len(c.events)
}

// LastPosition returns the last assigned position.
func (c *Collector) LastPosition() uint64 {
	return c.position
}

// Reset clears collected events but preserves the position counter.
func (c *Collector) Reset() {
	c.events = c.events[:0]
}
