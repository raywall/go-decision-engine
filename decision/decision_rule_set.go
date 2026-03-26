package decision

import (
	"context"
	"sync"

	"github.com/raywall/go-decision-engine/decision/engine"
)

// DecisionRuleSet represents a collection of rules.
type DecisionRuleSet struct {
	Rules []*DecisionRule
}

// EvaluateAll executes all rules concurrently.
func (rs *DecisionRuleSet) EvaluateAll(ctx context.Context, input map[string]any) []engine.RuleResult {
	var wg sync.WaitGroup
	results := make(chan engine.RuleResult, len(rs.Rules))

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, rule := range rs.Rules {
		wg.Add(1)

		go func(r *DecisionRule) {
			defer wg.Done()

			res, err := r.Evaluate(ctx, input)

			select {
			case results <- engine.RuleResult{
				Name:   r.Name,
				Result: res,
				Error:  err,
			}:
			case <-ctx.Done():
			}
		}(rule)
	}

	wg.Wait()
	close(results)

	var output []engine.RuleResult
	for r := range results {
		output = append(output, r)
	}

	return output
}
