package pipeline

type Context struct {
	TraceID       string
	SourceName    string
	Strict        bool
	MaxDepth      int
	MaxIterations int
}

func DefaultContext() Context {
	return Context{
		Strict:        true,
		MaxDepth:      32,
		MaxIterations: 100000,
	}
}

type ErrorMode string

const (
	ErrorModeFailFast ErrorMode = "fail_fast"
	ErrorModeContinue ErrorMode = "continue"
	ErrorModeCollect  ErrorMode = "collect"
)

type Executor struct {
	ErrorMode        ErrorMode
	AllowSideEffects bool
}

func DefaultExecutor() Executor {
	return Executor{
		ErrorMode:        ErrorModeFailFast,
		AllowSideEffects: false,
	}
}
