package main

import (
	"context"
	"fmt"
	"time"

	"github.com/raywall/go-decision-engine/decision"
	"github.com/raywall/go-decision-engine/decision/engine"
)

type User struct {
	Age     int    `decision:"age"`
	Country string `decision:"country"`
	Score   int    `decision:"score"`
}

func main() {

	cache := engine.NewRuleCache()

	schema := map[string]engine.ArgType{
		"age":     engine.IntType,
		"country": engine.StringType,
		"score":   engine.IntType,
	}

	rule := must(decision.NewDecisionRule(
		"Elite",
		"age > 25 && country == 'BR' && score > 800",
		schema,
		cache,
	))

	user := User{
		Age:     30,
		Country: "BR",
		Score:   900,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ok, err := rule.EvaluateFrom(ctx, user)

	fmt.Printf("Elite -> %v (err=%v)\n", ok, err)
}

func must(r *decision.DecisionRule, err error) *decision.DecisionRule {
	if err != nil {
		panic(err)
	}
	return r
}
