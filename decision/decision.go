// Package decision provides a flexible, expression-driven validation engine
// built on top of CEL (Common Expression Language).
package decision

import (
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"
)

// DecisionArgs is the central validation unit.
//
//   - Args       – field paths to extract ("active", "client.user.active").
//   - Data       – filtered map produced by ParseToData; mirrors source nesting.
//   - Expression – CEL expression evaluated against Data; must return bool.
type DecisionArgs struct {
	Args       []string
	Data       map[string]any
	Expression string
}

// New creates a DecisionArgs with the given args and CEL expression.
func New(expression string, args ...string) *DecisionArgs {
	return &DecisionArgs{
		Args:       args,
		Expression: expression,
	}
}

// ParseToData converts source into the internal Data map, keeping only the
// fields referenced in Args.
//
// source can be map[string]any, a struct, *struct, or []byte (JSON).
//
// Nested paths ("client.user.active") are resolved and stored as nested maps
// so that CEL field-selection syntax works identically.
func (d *DecisionArgs) ParseToData(source any) error {
	sourceMap, err := anyToMap(source)
	if err != nil {
		return fmt.Errorf("decision.ParseToData: cannot convert source: %w", err)
	}

	d.Data = make(map[string]any)

	for _, arg := range d.Args {
		parts := strings.Split(arg, ".")
		value, err := extractNested(sourceMap, parts)
		if err != nil {
			return fmt.Errorf("decision.ParseToData: arg %q — %w", arg, err)
		}
		insertNested(d.Data, parts, value)
	}

	return nil
}

// Validate compiles and evaluates the CEL Expression against Data.
//
//   - (true,  nil)   – expression passed
//   - (false, nil)   – expression failed
//   - (false, error) – compile/eval error, or non-bool result
//
// All top-level Data keys are declared as dyn CEL variables, enabling
// full nested map and slice access inside expressions.
func (d *DecisionArgs) Validate() (bool, error) {
	if d.Data == nil {
		return false, fmt.Errorf("decision.Validate: Data is nil — call ParseToData first")
	}

	env, err := d.buildEnv()
	if err != nil {
		return false, err
	}

	ast, issues := env.Compile(d.Expression)
	if issues != nil && issues.Err() != nil {
		return false, fmt.Errorf("decision.Validate: compile error: %w", issues.Err())
	}

	prg, err := env.Program(ast)
	if err != nil {
		return false, fmt.Errorf("decision.Validate: program error: %w", err)
	}

	out, _, err := prg.Eval(d.Data)
	if err != nil {
		return false, fmt.Errorf("decision.Validate: eval error: %w", err)
	}

	result, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("decision.Validate: expected bool result, got %T", out.Value())
	}

	return result, nil
}

// ValidateWith is a shorthand for ParseToData + Validate in one call.
func (d *DecisionArgs) ValidateWith(source any) (bool, error) {
	if err := d.ParseToData(source); err != nil {
		return false, err
	}
	return d.Validate()
}

func (d *DecisionArgs) buildEnv() (*cel.Env, error) {
	opts := make([]cel.EnvOption, 0, len(d.Data))
	for key := range d.Data {
		opts = append(opts, cel.Variable(key, cel.DynType))
	}
	env, err := cel.NewEnv(opts...)
	if err != nil {
		return nil, fmt.Errorf("decision.Validate: CEL env error: %w", err)
	}
	return env, nil
}
