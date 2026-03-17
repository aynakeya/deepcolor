package pipeline

import (
	"fmt"
)

type OpEffect string

const (
	EffectPure       OpEffect = "pure"
	EffectSideEffect OpEffect = "side_effect"
)

type Extractor func(ctx *Context, st *State, arg string) (any, error)
type Mapper func(ctx *Context, in any, args []any) (any, error)
type Predicate func(ctx *Context, in any, args []any) (bool, error)

type mapperEntry struct {
	fn     Mapper
	effect OpEffect
}

type predicateEntry struct {
	fn     Predicate
	effect OpEffect
}

type Registry struct {
	extractors map[string]Extractor
	mappers    map[string]mapperEntry
	predicates map[string]predicateEntry
}

func NewRegistry() *Registry {
	return &Registry{
		extractors: map[string]Extractor{},
		mappers:    map[string]mapperEntry{},
		predicates: map[string]predicateEntry{},
	}
}

func NewRegistryWithBuiltins() *Registry {
	r := NewRegistry()
	registerBuiltins(r)
	return r
}

func (r *Registry) RegisterExtractor(name string, fn Extractor) {
	r.extractors[name] = fn
}

func (r *Registry) RegisterMapper(name string, effect OpEffect, fn Mapper) {
	r.mappers[name] = mapperEntry{fn: fn, effect: effect}
}

func (r *Registry) RegisterPredicate(name string, effect OpEffect, fn Predicate) {
	r.predicates[name] = predicateEntry{fn: fn, effect: effect}
}

func (r *Registry) HasExtractor(name string) bool {
	_, ok := r.extractors[name]
	return ok
}

func (r *Registry) HasMapper(name string) bool {
	_, ok := r.mappers[name]
	return ok
}

func (r *Registry) HasPredicate(name string) bool {
	_, ok := r.predicates[name]
	return ok
}

func (r *Registry) getExtractor(name string) (Extractor, error) {
	fn, ok := r.extractors[name]
	if !ok {
		return nil, &Error{Kind: ErrKindRegistry, Op: "extractor", Message: fmt.Sprintf("extractor %s not registered", name)}
	}
	return fn, nil
}

func (r *Registry) getMapper(name string) (mapperEntry, error) {
	fn, ok := r.mappers[name]
	if !ok {
		return mapperEntry{}, &Error{Kind: ErrKindRegistry, Op: "mapper", Message: fmt.Sprintf("mapper %s not registered", name)}
	}
	return fn, nil
}

func (r *Registry) getPredicate(name string) (predicateEntry, error) {
	fn, ok := r.predicates[name]
	if !ok {
		return predicateEntry{}, &Error{Kind: ErrKindRegistry, Op: "predicate", Message: fmt.Sprintf("predicate %s not registered", name)}
	}
	return fn, nil
}
