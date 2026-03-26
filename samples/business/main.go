package main

import (
	"context"
	"fmt"
	"time"

	"github.com/raywall/go-decision-engine/decision"
	"github.com/raywall/go-decision-engine/decision/engine"
)

type Contract struct {
	ClientID       string  `json:"client_id" decision:"-"`
	ContractNumber string  `json:"contract_number" decision:"-"`
	Product        string  `json:"product" decision:"product"`
	Value          float64 `json:"contract_value" decision:"value"`
}

func main() {

	cache := engine.NewRuleCache()

	schema := map[string]engine.ArgType{
		"product": engine.StringType,
		"value":   engine.FloatType,
	}

	ruleSet := decision.DecisionRuleSet{
		Rules: []*decision.DecisionRule{
			must(decision.NewDecisionRule(
				"Corporate Contract",
				"product == 'PJ' && value > 1000",
				schema,
				cache,
			)),
			must(decision.NewDecisionRule(
				"Business Contract",
				"product == 'PJ' && value < 1000",
				schema,
				cache,
			)),
		},
	}

	user := Contract{
		ClientID:       "03f9d380-e8fc-4bca-894d-0da09ad80d5e",
		ContractNumber: "PJ123456789BR",
		Product:        "PJ",
		Value:          900,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	res, err := ruleSet.EvaluateAll(ctx, user)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Company Contract -> %v (err=%v)\n", res, err)
}

func must(r *decision.DecisionRule, err error) *decision.DecisionRule {
	if err != nil {
		panic(err)
	}
	return r
}
