package pipeline

type SchemaVersion string

const VersionV1 SchemaVersion = "v1"

type Plan struct {
	SchemaVersion SchemaVersion
	Source        SourceSpec
	Steps         []Step
	Output        Node
}

type SourceSpec struct {
	Name string
}

type StepType string

const (
	StepExtract StepType = "extract"
	StepMap     StepType = "map"
	StepFilter  StepType = "filter"
	StepProject StepType = "project"
)

type Step interface {
	stepType() StepType
}

type ExtractStep struct {
	Assignments []ExtractAssignment
}

type ExtractAssignment struct {
	To   string
	Expr string
}

func (ExtractStep) stepType() StepType { return StepExtract }

type MapStep struct {
	Transforms []MapTransform
}

type MapTransform struct {
	Path string
	Ops  []OpSpec
}

func (MapStep) stepType() StepType { return StepMap }

type FilterStep struct {
	Conditions []Condition
	Mode       FilterMode
}

type FilterMode string

const (
	FilterAll FilterMode = "all"
	FilterAny FilterMode = "any"
)

type Condition struct {
	Path   string
	Op     OpSpec
	Negate bool
}

func (FilterStep) stepType() StepType { return StepFilter }

type ProjectStep struct {
	Root Node
}

func (ProjectStep) stepType() StepType { return StepProject }

type OpSpec struct {
	Name string
	Args []any
}

type NodeKind string

const (
	NodeField  NodeKind = "field"
	NodeValue  NodeKind = "value"
	NodeObject NodeKind = "object"
	NodeArray  NodeKind = "array"
	NodeOp     NodeKind = "op"
)

type Node interface {
	nodeKind() NodeKind
}

type FieldNode struct {
	Path string
}

func (FieldNode) nodeKind() NodeKind { return NodeField }

type ValueNode struct {
	Value any
}

func (ValueNode) nodeKind() NodeKind { return NodeValue }

type ObjectNode struct {
	Fields map[string]Node
}

func (ObjectNode) nodeKind() NodeKind { return NodeObject }

type ArrayNode struct {
	Items []Node
}

func (ArrayNode) nodeKind() NodeKind { return NodeArray }

type OpNode struct {
	Input Node
	Op    OpSpec
}

func (OpNode) nodeKind() NodeKind { return NodeOp }
