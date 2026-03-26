package decision

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// anyToMap converts any value into map[string]any.
// Accepts: map[string]any, struct/*struct, []byte (JSON).
func anyToMap(source any) (map[string]any, error) {
	if source == nil {
		return nil, fmt.Errorf("source is nil")
	}
	if m, ok := source.(map[string]any); ok {
		return m, nil
	}
	if b, ok := source.([]byte); ok {
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, fmt.Errorf("cannot unmarshal JSON bytes: %w", err)
		}
		return m, nil
	}

	rv := reflect.ValueOf(source)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, fmt.Errorf("source pointer is nil")
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Struct, reflect.Map:
		b, err := json.Marshal(source)
		if err != nil {
			return nil, fmt.Errorf("marshal error: %w", err)
		}
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			return nil, fmt.Errorf("unmarshal error: %w", err)
		}
		return m, nil
	default:
		return nil, fmt.Errorf("unsupported source type %T", source)
	}
}

// extractNested resolves a dotted path inside a nested map.
// e.g. ["client","user","active"] → source["client"]["user"]["active"]
func extractNested(source map[string]any, parts []string) (any, error) {
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty path")
	}
	val, ok := source[parts[0]]
	if !ok {
		return nil, fmt.Errorf("key %q not found", parts[0])
	}
	if len(parts) == 1 {
		return val, nil
	}
	nested, ok := val.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("expected map at %q for traversal, got %T", parts[0], val)
	}
	return extractNested(nested, parts[1:])
}

// insertNested writes value at a dotted path into target,
// creating intermediate maps as needed.
func insertNested(target map[string]any, parts []string, value any) {
	if len(parts) == 1 {
		target[parts[0]] = value
		return
	}
	child, ok := target[parts[0]].(map[string]any)
	if !ok {
		child = make(map[string]any)
		target[parts[0]] = child
	}
	insertNested(child, parts[1:], value)
}
