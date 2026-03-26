package decision_test

import (
	"context"
	"testing"

	"github.com/raywall/go-decision-engine/decision"
	"github.com/raywall/go-decision-engine/decision/engine"
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

	results, err := rs.EvaluateAll(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
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

	results, err := rs.EvaluateAll(context.Background(), user)
	if err != nil {
		t.Fatal(err)
	}
	if !results[0].Result {
		t.Fatal("expected true")
	}
}
