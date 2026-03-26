package decision

import (
	"context"
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/raywall/go-decision-engine/decision/engine"
	"github.com/raywall/go-decision-engine/decision/tags"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

// DecisionRule represents a rule with schema and CEL expression.
type DecisionRule struct {
	Name       string
	Expression string
	Schema     map[string]engine.ArgType
	Program    cel.Program
}

// NewDecisionRule creates a rule with compiled CEL expression.
func NewDecisionRule(
	name string,
	expr string,
	schema map[string]engine.ArgType,
	cache *engine.RuleCache,
) (*DecisionRule, error) {

	env, err := buildEnv(schema)
	if err != nil {
		return nil, err
	}

	prg, err := cache.GetOrCompile(env, expr)
	if err != nil {
		return nil, err
	}

	return &DecisionRule{
		Name:       name,
		Expression: expr,
		Schema:     schema,
		Program:    prg,
	}, nil
}

// Evaluate validates and executes rule using map input.
func (r *DecisionRule) Evaluate(ctx context.Context, input map[string]any) (bool, error) {

	if err := r.Validate(input); err != nil {
		return false, err
	}

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	out, _, err := r.Program.Eval(input)
	if err != nil {
		return false, err
	}

	val, ok := out.Value().(bool)
	if !ok {
		return false, fmt.Errorf("rule %s did not return boolean", r.Name)
	}

	return val, nil
}

// EvaluateFrom extracts data from any struct/map using tags.
func (r *DecisionRule) EvaluateFrom(ctx context.Context, input any) (bool, error) {

	data, err := tags.ParseToMap(input, toAnySchema(r.Schema))
	if err != nil {
		return false, err
	}

	return r.Evaluate(ctx, data)
}

// Validate enforces schema typing.
func (r *DecisionRule) Validate(input map[string]any) error {
	for key, typ := range r.Schema {
		val, ok := input[key]
		if !ok {
			return fmt.Errorf("missing arg: %s", key)
		}

		switch typ {
		case engine.StringType:
			if _, ok := val.(string); !ok {
				return fmt.Errorf("%s must be string", key)
			}
		case engine.IntType:
			if _, ok := val.(int); !ok {
				return fmt.Errorf("%s must be int", key)
			}
		case engine.BoolType:
			if _, ok := val.(bool); !ok {
				return fmt.Errorf("%s must be bool", key)
			}
		case engine.FloatType:
			if _, ok := val.(float64); !ok {
				return fmt.Errorf("%s must be float64", key)
			}
		}
	}
	return nil
}

// buildEnv builds typed CEL environment.
func buildEnv(schema map[string]engine.ArgType) (*cel.Env, error) {
	var declsList []*exprpb.Decl

	for name, t := range schema {
		var celType *exprpb.Type

		switch t {
		case engine.StringType:
			celType = decls.String
		case engine.IntType:
			celType = decls.Int
		case engine.BoolType:
			celType = decls.Bool
		case engine.FloatType:
			celType = decls.Double
		default:
			return nil, fmt.Errorf("unsupported type: %s", t)
		}

		declsList = append(declsList, decls.NewVar(name, celType))
	}

	return cel.NewEnv(cel.Declarations(declsList...))
}

func toAnySchema(schema map[string]engine.ArgType) map[string]any {
	out := make(map[string]any)
	for k := range schema {
		out[k] = nil
	}
	return out
}
