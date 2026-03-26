package engine

import (
	"sync"

	"github.com/google/cel-go/cel"
)

// RuleCache stores compiled CEL programs to avoid recompilation.
type RuleCache struct {
	mu    sync.RWMutex
	cache map[string]cel.Program
}

// NewRuleCache creates a new cache instance.
func NewRuleCache() *RuleCache {
	return &RuleCache{
		cache: make(map[string]cel.Program),
	}
}

// GetOrCompile retrieves a cached program or compiles it if not present.
func (c *RuleCache) GetOrCompile(env *cel.Env, expr string) (cel.Program, error) {
	c.mu.RLock()
	if prog, ok := c.cache[expr]; ok {
		c.mu.RUnlock()
		return prog, nil
	}
	c.mu.RUnlock()

	ast, issues := env.Parse(expr)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}

	prg, err := env.Program(ast)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cache[expr] = prg
	c.mu.Unlock()

	return prg, nil
}
