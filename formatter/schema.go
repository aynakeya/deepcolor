package formatter

// Schema define the input and output of the schema
type Schema[I, T any] interface {
	Format(value I) T
}
