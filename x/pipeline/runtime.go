package pipeline

import (
	"strings"
)

type State struct {
	Input  any
	Doc    map[string]any
	Output any
}

type CompiledPipeline struct {
	plan     Plan
	registry *Registry
	exec     Executor
	steps    []compiledStep
}

type compiledStep interface {
	execStep(ctx *Context, st *State, c *CompiledPipeline) error
}

type compiledExtractStep struct {
	assignments []compiledExtractAssign
}

type compiledExtractAssign struct {
	to        *Accessor
	extractor string
	arg       string
}

func (s *compiledExtractStep) execStep(ctx *Context, st *State, c *CompiledPipeline) error {
	for _, a := range s.assignments {
		ex, err := c.registry.getExtractor(a.extractor)
		if err != nil {
			return err
		}
		v, err := ex(ctx, st, a.arg)
		if err != nil {
			if c.exec.ErrorMode == ErrorModeContinue {
				continue
			}
			return err
		}
		if err = a.to.Set(&st.Doc, v); err != nil {
			if c.exec.ErrorMode == ErrorModeContinue {
				continue
			}
			return err
		}
	}
	return nil
}

type compiledMapStep struct {
	transforms []compiledMapTransform
}

type compiledMapTransform struct {
	path *Accessor
	ops  []OpSpec
}

func (s *compiledMapStep) execStep(ctx *Context, st *State, c *CompiledPipeline) error {
	for _, t := range s.transforms {
		v, err := t.path.Get(st.Doc)
		if err != nil {
			if c.exec.ErrorMode == ErrorModeContinue {
				continue
			}
			return err
		}
		for _, op := range t.ops {
			entry, e := c.registry.getMapper(op.Name)
			if e != nil {
				return e
			}
			if entry.effect == EffectSideEffect && !c.exec.AllowSideEffects {
				return &Error{Kind: ErrKindRuntime, Op: op.Name, Message: "side effects disabled"}
			}
			v, err = entry.fn(ctx, v, op.Args)
			if err != nil {
				if c.exec.ErrorMode == ErrorModeContinue {
					break
				}
				return err
			}
		}
		if err = t.path.Set(&st.Doc, v); err != nil {
			if c.exec.ErrorMode == ErrorModeContinue {
				continue
			}
			return err
		}
	}
	return nil
}

type compiledFilterStep struct {
	mode       FilterMode
	conditions []compiledCondition
}

type compiledCondition struct {
	path   *Accessor
	op     OpSpec
	negate bool
}

func (s *compiledFilterStep) execStep(ctx *Context, st *State, c *CompiledPipeline) error {
	if len(s.conditions) == 0 {
		return nil
	}
	result := s.mode == FilterAll
	for _, cond := range s.conditions {
		v, err := cond.path.Get(st.Doc)
		if err != nil {
			return err
		}
		pred, err := c.registry.getPredicate(cond.op.Name)
		if err != nil {
			return err
		}
		if pred.effect == EffectSideEffect && !c.exec.AllowSideEffects {
			return &Error{Kind: ErrKindRuntime, Op: cond.op.Name, Message: "side effects disabled"}
		}
		ok, err := pred.fn(ctx, v, cond.op.Args)
		if err != nil {
			return err
		}
		if cond.negate {
			ok = !ok
		}
		if s.mode == FilterAll {
			result = result && ok
		} else {
			result = result || ok
		}
	}
	if !result {
		st.Output = nil
		st.Doc["$filtered"] = true
	}
	return nil
}

type compiledProjectStep struct {
	root compiledNode
}

func (s *compiledProjectStep) execStep(ctx *Context, st *State, c *CompiledPipeline) error {
	if filtered, _ := st.Doc["$filtered"].(bool); filtered {
		return nil
	}
	v, err := s.root.eval(ctx, st, c)
	if err != nil {
		return err
	}
	st.Output = v
	return nil
}

type compiledNode interface {
	eval(ctx *Context, st *State, c *CompiledPipeline) (any, error)
}

type compiledFieldNode struct{ path *Accessor }

func (n compiledFieldNode) eval(_ *Context, st *State, _ *CompiledPipeline) (any, error) {
	return n.path.Get(st.Doc)
}

type compiledValueNode struct{ value any }

func (n compiledValueNode) eval(_ *Context, _ *State, _ *CompiledPipeline) (any, error) {
	return n.value, nil
}

type compiledObjectNode struct{ fields map[string]compiledNode }

func (n compiledObjectNode) eval(ctx *Context, st *State, c *CompiledPipeline) (any, error) {
	out := map[string]any{}
	for k, child := range n.fields {
		v, err := child.eval(ctx, st, c)
		if err != nil {
			return nil, err
		}
		out[k] = v
	}
	return out, nil
}

type compiledArrayNode struct{ items []compiledNode }

func (n compiledArrayNode) eval(ctx *Context, st *State, c *CompiledPipeline) (any, error) {
	out := make([]any, 0, len(n.items))
	for _, child := range n.items {
		v, err := child.eval(ctx, st, c)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

type compiledOpNode struct {
	in compiledNode
	op OpSpec
}

func (n compiledOpNode) eval(ctx *Context, st *State, c *CompiledPipeline) (any, error) {
	v, err := n.in.eval(ctx, st, c)
	if err != nil {
		return nil, err
	}
	entry, err := c.registry.getMapper(n.op.Name)
	if err != nil {
		return nil, err
	}
	if entry.effect == EffectSideEffect && !c.exec.AllowSideEffects {
		return nil, &Error{Kind: ErrKindRuntime, Op: n.op.Name, Message: "side effects disabled"}
	}
	return entry.fn(ctx, v, n.op.Args)
}

func (p *CompiledPipeline) Run(ctx Context, input any) (any, error) {
	st := &State{Input: input, Doc: map[string]any{}, Output: nil}
	if len(p.steps) == 0 {
		return nil, &Error{Kind: ErrKindRuntime, Op: "run", Message: "empty pipeline"}
	}
	for _, step := range p.steps {
		if err := step.execStep(&ctx, st, p); err != nil {
			return nil, err
		}
	}
	if st.Output != nil {
		return st.Output, nil
	}
	return st.Doc, nil
}

func parseExpr(expr string) (string, string, error) {
	parts := strings.SplitN(expr, ":", 2)
	if len(parts) != 2 || parts[0] == "" {
		return "", "", &Error{Kind: ErrKindValidate, Op: "expr", Message: "expression must be kind:arg"}
	}
	return parts[0], parts[1], nil
}
