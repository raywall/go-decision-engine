package engine

// ArgType defines supported types for rule inputs.
type ArgType string

const (
	StringType ArgType = "string"
	IntType    ArgType = "int"
	BoolType   ArgType = "bool"
	FloatType  ArgType = "float"
)

// RuleResult represents the result of a rule evaluation.
type RuleResult struct {
	Name   string
	Result bool
	Error  error
}
