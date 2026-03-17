package dynmarshaller

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type DynamicMarshallable struct {
	Type string
}

type Recoverable interface {
	Recover() error
}

type DynamicUnmarshaller[T any] struct {
	registry map[string]T
}

// NewDynamicUnmarshaller creates a new DynamicUnmarshaller
func NewDynamicUnmarshaller[T any](registry map[string]T) *DynamicUnmarshaller[T] {
	return &DynamicUnmarshaller[T]{
		registry: registry,
	}
}

// Register overwrite the existing entry if the key already exists
// otherwise it will add a new entry
// must be called in program initialization stage
func (d *DynamicUnmarshaller[T]) Register(name string, t T) {
	d.registry[name] = t
}

// Unmarshal unmarshals the data into the type specified by the Type field
func (d *DynamicUnmarshaller[T]) Unmarshal(data []byte) (T, error) {
	var f DynamicMarshallable
	if err := json.Unmarshal(data, &f); err != nil {
		return *new(T), err
	}
	t, ok := d.registry[f.Type]
	if !ok {
		return *new(T), errors.New(fmt.Sprintf("type %s haven't register yet", f.Type))
	}
	v := reflect.New(reflect.TypeOf(t).Elem()).Interface()
	err := json.Unmarshal(data, v)
	if err != nil {
		return *new(T), err
	}
	if i, ok := v.(Recoverable); ok {
		return v.(T), i.Recover()
	}
	return v.(T), nil
}

// GetNames returns the names of all registered types
func (d *DynamicUnmarshaller[T]) GetNames() []string {
	names := make([]string, 0)
	for key, _ := range d.registry {
		names = append(names, key)
	}
	return names
}
