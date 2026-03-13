package export

import (
	"context"
	"fmt"
	"log/slog"
)

// LogExporter writes events to slog so the user can see process execution in the console.
type LogExporter struct {
	logger *slog.Logger
}

// NewLogExporter creates an exporter that logs events using the provided logger.
func NewLogExporter(logger *slog.Logger) *LogExporter {
	return &LogExporter{logger: logger}
}

func (e *LogExporter) Export(_ context.Context, events []Event) error {
	for _, ev := range events {
		attrs := []slog.Attr{
			slog.Uint64("key", ev.Key),
			slog.String("type", string(ev.ValueType)),
			slog.String("record", string(ev.RecordType)),
		}
		if ev.ProcessInstanceKey != 0 {
			attrs = append(attrs, slog.Uint64("pi", ev.ProcessInstanceKey))
		}
		if ev.ElementId != "" {
			attrs = append(attrs, slog.String("element", ev.ElementId))
		}
		if ev.ElementType != "" {
			attrs = append(attrs, slog.String("elementType", ev.ElementType))
		}

		msg := fmt.Sprintf("%s %s", ev.ValueType, ev.RecordType)
		args := make([]any, len(attrs))
		for i, a := range attrs {
			args[i] = a
		}
		e.logger.Info(msg, args...)
	}
	return nil
}
