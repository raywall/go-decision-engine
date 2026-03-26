package decision_test

import (
	"context"
	"testing"
	"time"

	"github.com/raywall/go-decision-engine/decision"
	"github.com/raywall/go-decision-engine/decision/engine"
)

func TestDecisionRule(t *testing.T) {
	cache := engine.NewRuleCache()

	rule, err := decision.NewDecisionRule(
		"Adult",
		"age >= 18",
		map[string]engine.ArgType{"age": engine.IntType},
		cache,
	)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	ok, err := rule.Evaluate(ctx, map[string]any{"age": 20})
	if err != nil || !ok {
		t.Fatalf("expected true, got %v err=%v", ok, err)
	}

	ok, err = rule.Evaluate(ctx, map[string]any{"age": 15})
	if err != nil || ok {
		t.Fatalf("expected false")
	}
}

func TestTimeout(t *testing.T) {
	cache := engine.NewRuleCache()

	rule, _ := decision.NewDecisionRule(
		"Slow",
		"true",
		map[string]engine.ArgType{},
		cache,
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancel()

	_, err := rule.Evaluate(ctx, map[string]any{})
	if err == nil {
		t.Fatal("expected timeout")
	}
}

func TestEvaluateFromStruct(t *testing.T) {
	cache := engine.NewRuleCache()

	type User struct {
		Age int `decision:"age"`
	}

	rule, _ := decision.NewDecisionRule(
		"Adult",
		"age >= 18",
		map[string]engine.ArgType{"age": engine.IntType},
		cache,
	)

	user := User{Age: 20}

	ok, err := rule.EvaluateFrom(context.Background(), user)
	if err != nil || !ok {
		t.Fatalf("expected true, got %v err=%v", ok, err)
	}
}
