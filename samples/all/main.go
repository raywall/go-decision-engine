package main

import (
	"context"
	"fmt"
	"time"

	"github.com/raywall/go-decision-engine/decision"
	"github.com/raywall/go-decision-engine/decision/engine"
)

func main() {
	fmt.Println("Go Decision Engine")

	cache := engine.NewRuleCache()

	// Shared schema (like your original DecisionArgs idea, but cleaner)
	schema := map[string]engine.ArgType{
		"age":     engine.IntType,
		"country": engine.StringType,
		"score":   engine.IntType,
	}

	// =========================================================
	// Scenario 1 - Single Rule (Original Behavior)
	// =========================================================
	fmt.Println("\n=== Scenario 1: Single Rule Validation ===")

	ruleAdult := must(decision.NewDecisionRule(
		"Adult",
		"age >= 18",
		schema,
		cache,
	))

	input1 := map[string]any{
		"age":     20,
		"country": "BR",
		"score":   500,
		"extra":   "ignored",
	}

	ctx1 := context.Background()

	res, err := ruleAdult.Evaluate(ctx1, input1)
	fmt.Printf("Adult -> %v (err=%v)\n", res, err)

	// =========================================================
	// 🧪 Scenario 2 - Multiple Rules (RuleSet)
	// =========================================================
	fmt.Println("\n=== Scenario 2: Rule Set (Parallel Execution) ===")

	rules := []*decision.DecisionRule{
		ruleAdult, // reused (cache hit)
		must(decision.NewDecisionRule("Brazilian", "country == 'BR'", schema, cache)),
		must(decision.NewDecisionRule("HighScore", "score > 700", schema, cache)),
	}

	rs := decision.DecisionRuleSet{Rules: rules}

	input2 := map[string]any{
		"age":     25,
		"country": "BR",
		"score":   720,
	}

	results := rs.EvaluateAll(context.Background(), input2)

	for _, r := range results {
		fmt.Printf("%s -> %v (err=%v)\n", r.Name, r.Result, r.Error)
	}

	// =========================================================
	// 🧪 Scenario 3 - Cache Demonstration
	// =========================================================
	fmt.Println("\n=== Scenario 3: Cache Reuse ===")

	// Same expression → should hit cache
	ruleCached := must(decision.NewDecisionRule(
		"Adult-Again",
		"age >= 18",
		schema,
		cache,
	))

	res, err = ruleCached.Evaluate(context.Background(), input2)
	fmt.Printf("Adult-Again -> %v (err=%v)\n", res, err)

	// =========================================================
	// 🧪 Scenario 4 - Timeout Handling
	// =========================================================
	fmt.Println("\n=== Scenario 4: Context Timeout ===")

	ctxTimeout, cancelTimeout := context.WithTimeout(context.Background(), time.Nanosecond)
	defer cancelTimeout()

	_, err = ruleAdult.Evaluate(ctxTimeout, input2)
	fmt.Printf("Timeout result -> err=%v\n", err)

	// =========================================================
	// 🧪 Scenario 5 - Rule Set with Timeout
	// =========================================================
	fmt.Println("\n=== Scenario 5: Rule Set with Timeout ===")

	ctxSetTimeout, cancelSetTimeout := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancelSetTimeout()

	results = rs.EvaluateAll(ctxSetTimeout, input2)

	for _, r := range results {
		fmt.Printf("%s -> %v (err=%v)\n", r.Name, r.Result, r.Error)
	}

	// =========================================================
	// 🧪 Scenario 6 - Manual Cancel (Fail-Fast Simulation)
	// =========================================================
	fmt.Println("\n=== Scenario 6: Manual Cancel (Fail-Fast Simulation) ===")

	ctxCancel, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(1 * time.Millisecond)
		cancel()
	}()

	results = rs.EvaluateAll(ctxCancel, input2)

	for _, r := range results {
		fmt.Printf("%s -> %v (err=%v)\n", r.Name, r.Result, r.Error)
	}

	// =========================================================
	// 🧪 Scenario 7 - Invalid Input (Type Safety)
	// =========================================================
	fmt.Println("\n=== Scenario 7: Invalid Input (Type Validation) ===")

	invalidInput := map[string]any{
		"age":     "twenty", // wrong type
		"country": "BR",
		"score":   700,
	}

	res, err = ruleAdult.Evaluate(context.Background(), invalidInput)
	fmt.Printf("Invalid input -> %v (err=%v)\n", res, err)

	// =========================================================
	// 🧪 Scenario 8 - Complex Rule
	// =========================================================
	fmt.Println("\n=== Scenario 8: Complex Rule ===")

	complexRule := must(decision.NewDecisionRule(
		"Elite",
		"age > 25 && country == 'BR' && score > 800",
		schema,
		cache,
	))

	inputComplex := map[string]any{
		"age":     30,
		"country": "BR",
		"score":   850,
	}

	res, err = complexRule.Evaluate(context.Background(), inputComplex)
	fmt.Printf("Elite -> %v (err=%v)\n", res, err)
}

// Helper
func must(r *decision.DecisionRule, err error) *decision.DecisionRule {
	if err != nil {
		panic(err)
	}
	return r
}
