package decision_test

import (
	"context"
	"testing"

	"github.com/raywall/go-decision-engine/decision"
	"github.com/raywall/go-decision-engine/decision/engine"
	"github.com/raywall/go-decision-engine/decision/tags"
)

func TestRuleSet(t *testing.T) {
	cache := engine.NewRuleCache()

	schema := map[string]engine.ArgType{
		"age":     engine.IntType,
		"country": engine.StringType,
	}

	r1, _ := decision.NewDecisionRule("Adult", "age >= 18", schema, cache)
	r2, _ := decision.NewDecisionRule("BR", "country == 'BR'", schema, cache)

	rs := decision.DecisionRuleSet{
		Rules: []*decision.DecisionRule{r1, r2},
	}

	input := map[string]any{
		"age":     25,
		"country": "BR",
	}

	results := rs.EvaluateAll(context.Background(), input)

	if len(results) != 2 {
		t.Fatal("expected 2 results")
	}
}

func TestRuleSetWithStructInput(t *testing.T) {
	cache := engine.NewRuleCache()

	schema := map[string]engine.ArgType{
		"age": engine.IntType,
	}

	r1, _ := decision.NewDecisionRule("Adult", "age >= 18", schema, cache)

	rs := decision.DecisionRuleSet{
		Rules: []*decision.DecisionRule{r1},
	}

	type User struct {
		Age int `decision:"age"`
	}

	user := User{Age: 25}

	data, _ := tags.ParseToMap(user, map[string]any{"age": nil})

	results := rs.EvaluateAll(context.Background(), data)

	if !results[0].Result {
		t.Fatal("expected true")
	}
}
