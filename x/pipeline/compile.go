package pipeline

import "fmt"

func ValidateSchema(plan Plan, reg *Registry) error {
	if plan.SchemaVersion == "" {
		return &Error{Kind: ErrKindValidate, Op: "schema", Message: "schemaVersion required"}
	}
	if plan.SchemaVersion != VersionV1 {
		return &Error{Kind: ErrKindValidate, Op: "schema", Message: "unsupported schemaVersion"}
	}
	if reg == nil {
		return &Error{Kind: ErrKindValidate, Op: "schema", Message: "registry is nil"}
	}
	for i, step := range plan.Steps {
		if step == nil {
			return &Error{Kind: ErrKindValidate, Op: "schema", Path: fmt.Sprintf("steps[%d]", i), Message: "nil step"}
		}
		switch s := step.(type) {
		case ExtractStep:
			for j, a := range s.Assignments {
				if a.To == "" || a.Expr == "" {
					return &Error{Kind: ErrKindValidate, Op: "extract", Path: fmt.Sprintf("steps[%d].assignments[%d]", i, j), Message: "to and expr required"}
				}
				if _, err := CompileAccessor(a.To); err != nil {
					return err
				}
				k, _, err := parseExpr(a.Expr)
				if err != nil {
					return err
				}
				if !reg.HasExtractor(k) {
					return &Error{Kind: ErrKindValidate, Op: "extract", Path: a.Expr, Message: "extractor not found"}
				}
			}
		case MapStep:
			for j, t := range s.Transforms {
				if _, err := CompileAccessor(t.Path); err != nil {
					return err
				}
				for k, op := range t.Ops {
					if op.Name == "" || !reg.HasMapper(op.Name) {
						return &Error{Kind: ErrKindValidate, Op: "map", Path: fmt.Sprintf("steps[%d].transforms[%d].ops[%d]", i, j, k), Message: "unknown mapper"}
					}
				}
			}
		case FilterStep:
			for j, c := range s.Conditions {
				if _, err := CompileAccessor(c.Path); err != nil {
					return err
				}
				if c.Op.Name == "" || !reg.HasPredicate(c.Op.Name) {
					return &Error{Kind: ErrKindValidate, Op: "filter", Path: fmt.Sprintf("steps[%d].conditions[%d]", i, j), Message: "unknown predicate"}
				}
			}
		case ProjectStep:
			if err := validateNode(s.Root, reg); err != nil {
				return err
			}
		default:
			return &Error{Kind: ErrKindValidate, Op: "schema", Path: fmt.Sprintf("steps[%d]", i), Message: "unsupported step type"}
		}
	}
	if plan.Output != nil {
		if err := validateNode(plan.Output, reg); err != nil {
			return err
		}
	}
	return nil
}

func CompilePipeline(plan Plan, reg *Registry) (*CompiledPipeline, error) {
	if err := ValidateSchema(plan, reg); err != nil {
		return nil, err
	}
	cp := &CompiledPipeline{plan: plan, registry: reg, exec: DefaultExecutor(), steps: make([]compiledStep, 0, len(plan.Steps)+1)}

	for _, step := range plan.Steps {
		s, err := compileStep(step)
		if err != nil {
			return nil, err
		}
		cp.steps = append(cp.steps, s)
	}
	if plan.Output != nil {
		n, err := compileNode(plan.Output)
		if err != nil {
			return nil, err
		}
		cp.steps = append(cp.steps, &compiledProjectStep{root: n})
	}
	return cp, nil
}

func (p *CompiledPipeline) WithExecutor(ex Executor) *CompiledPipeline {
	p.exec = ex
	return p
}

func compileStep(step Step) (compiledStep, error) {
	switch v := step.(type) {
	case ExtractStep:
		out := &compiledExtractStep{assignments: make([]compiledExtractAssign, 0, len(v.Assignments))}
		for _, a := range v.Assignments {
			ac, err := CompileAccessor(a.To)
			if err != nil {
				return nil, err
			}
			k, arg, err := parseExpr(a.Expr)
			if err != nil {
				return nil, err
			}
			out.assignments = append(out.assignments, compiledExtractAssign{to: ac, extractor: k, arg: arg})
		}
		return out, nil
	case MapStep:
		out := &compiledMapStep{transforms: make([]compiledMapTransform, 0, len(v.Transforms))}
		for _, t := range v.Transforms {
			ac, err := CompileAccessor(t.Path)
			if err != nil {
				return nil, err
			}
			out.transforms = append(out.transforms, compiledMapTransform{path: ac, ops: t.Ops})
		}
		return out, nil
	case FilterStep:
		mode := v.Mode
		if mode == "" {
			mode = FilterAll
		}
		out := &compiledFilterStep{mode: mode, conditions: make([]compiledCondition, 0, len(v.Conditions))}
		for _, c := range v.Conditions {
			ac, err := CompileAccessor(c.Path)
			if err != nil {
				return nil, err
			}
			out.conditions = append(out.conditions, compiledCondition{path: ac, op: c.Op, negate: c.Negate})
		}
		return out, nil
	case ProjectStep:
		n, err := compileNode(v.Root)
		if err != nil {
			return nil, err
		}
		return &compiledProjectStep{root: n}, nil
	default:
		return nil, &Error{Kind: ErrKindCompile, Op: "compile_step", Message: "unsupported step type"}
	}
}

func validateNode(node Node, reg *Registry) error {
	if node == nil {
		return &Error{Kind: ErrKindValidate, Op: "node", Message: "nil node"}
	}
	switch n := node.(type) {
	case FieldNode:
		_, err := CompileAccessor(n.Path)
		return err
	case ValueNode:
		return nil
	case ObjectNode:
		for k, child := range n.Fields {
			if child == nil {
				return &Error{Kind: ErrKindValidate, Op: "node", Path: k, Message: "nil child node"}
			}
			if err := validateNode(child, reg); err != nil {
				return err
			}
		}
		return nil
	case ArrayNode:
		for i, child := range n.Items {
			if child == nil {
				return &Error{Kind: ErrKindValidate, Op: "node", Path: fmt.Sprintf("array[%d]", i), Message: "nil child node"}
			}
			if err := validateNode(child, reg); err != nil {
				return err
			}
		}
		return nil
	case OpNode:
		if n.Op.Name == "" || !reg.HasMapper(n.Op.Name) {
			return &Error{Kind: ErrKindValidate, Op: "node", Path: n.Op.Name, Message: "unknown mapper in OpNode"}
		}
		return validateNode(n.Input, reg)
	default:
		return &Error{Kind: ErrKindValidate, Op: "node", Message: "unsupported node type"}
	}
}

func compileNode(node Node) (compiledNode, error) {
	switch n := node.(type) {
	case FieldNode:
		ac, err := CompileAccessor(n.Path)
		if err != nil {
			return nil, err
		}
		return compiledFieldNode{path: ac}, nil
	case ValueNode:
		return compiledValueNode{value: n.Value}, nil
	case ObjectNode:
		out := compiledObjectNode{fields: map[string]compiledNode{}}
		for k, child := range n.Fields {
			cn, err := compileNode(child)
			if err != nil {
				return nil, err
			}
			out.fields[k] = cn
		}
		return out, nil
	case ArrayNode:
		out := compiledArrayNode{items: make([]compiledNode, 0, len(n.Items))}
		for _, child := range n.Items {
			cn, err := compileNode(child)
			if err != nil {
				return nil, err
			}
			out.items = append(out.items, cn)
		}
		return out, nil
	case OpNode:
		cn, err := compileNode(n.Input)
		if err != nil {
			return nil, err
		}
		return compiledOpNode{in: cn, op: n.Op}, nil
	default:
		return nil, &Error{Kind: ErrKindCompile, Op: "compile_node", Message: "unsupported node type"}
	}
}
