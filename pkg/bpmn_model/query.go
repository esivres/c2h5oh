package bpmn_model

import "fmt"

// Query provides a type-safe way to work with collections of BPMN elements.
type Query[T any] struct {
	items []T
}

// NewQuery creates a new Query from a slice of items.
func NewQuery[T any](items []T) Query[T] {
	return Query[T]{items: items}
}

// List returns all items in the query.
func (q Query[T]) List() []T {
	return q.items
}

// Count returns the number of items.
func (q Query[T]) Count() int {
	return len(q.items)
}

// SingleResult returns the single item or an error if count != 1.
func (q Query[T]) SingleResult() (T, error) {
	if len(q.items) == 0 {
		var zero T
		return zero, fmt.Errorf("query returned no results")
	}
	if len(q.items) > 1 {
		var zero T
		return zero, fmt.Errorf("query returned %d results, expected 1", len(q.items))
	}
	return q.items[0], nil
}

// Filter returns a new Query with only items matching the predicate.
func (q Query[T]) Filter(predicate func(T) bool) Query[T] {
	var result []T
	for _, item := range q.items {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return Query[T]{items: result}
}

// FilterByType returns a new Query containing only items that implement type U.
func FilterByType[U any, T any](q Query[T]) Query[U] {
	var result []U
	for _, item := range q.items {
		if typed, ok := any(item).(U); ok {
			result = append(result, typed)
		}
	}
	return Query[U]{items: result}
}

// --- Navigation functions ---

// SucceedingNodes returns a Query of all nodes directly following the given node via outgoing sequence flows.
func SucceedingNodes(node FlowNode) Query[FlowNode] {
	var result []FlowNode
	for _, sf := range node.GetOutgoingSequenceFlows() {
		target := sf.GetTarget()
		if target != nil {
			result = append(result, target)
		}
	}
	return NewQuery(result)
}

// PreviousNodes returns a Query of all nodes directly preceding the given node via incoming sequence flows.
func PreviousNodes(node FlowNode) Query[FlowNode] {
	var result []FlowNode
	for _, sf := range node.GetIncomingSequenceFlows() {
		source := sf.GetSource()
		if source != nil {
			result = append(result, source)
		}
	}
	return NewQuery(result)
}

// --- Model-level queries ---

// QueryProcesses returns all processes in the model.
func QueryProcesses(bmi *BpmnModelInstance) Query[Process] {
	return NewQuery(GetTypedElements[Process](bmi.ModelInstance))
}

// QueryFlowElements returns all flow elements in a process.
func QueryFlowElements(proc Process) Query[FlowElement] {
	return NewQuery(proc.GetFlowElements())
}
