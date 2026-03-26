# Go Decision Engine

A lightweight, high-performance rule engine written in Go using CEL (Common Expression Language).

## 🚀 Features

- Strongly typed rule inputs
- CEL-based expression evaluation
- Rule compilation cache (high performance)
- Concurrent rule execution (goroutines)
- Context-aware execution (timeouts & cancellation)
- Simple and extensible design

---

## 🧠 Core Concepts

### DecisionRule

Represents a single rule:

- Schema (typed inputs)
- CEL expression
- Compiled program

### DecisionRuleSet

Executes multiple rules concurrently.

### RuleCache

Avoids recompiling expressions, improving performance significantly.

---

## 🧪 Example

```go
schema := map[string]ArgType{
  "age": IntType,
}

rule, _ := NewDecisionRule("Adult", "age >= 18", schema, cache)

result, _ := rule.Evaluate(ctx, map[string]any{"age": 20})
```

⚡ Performance
Rules are compiled once and reused
Parallel execution using goroutines
Minimal memory overhead
🧵 Concurrency Model
One goroutine per rule
Fan-out / fan-in pattern
Context cancellation support
⏱ Timeout Example
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()
🔮 Future Improvements
Rule prioritization
Fail-fast execution
Observability (metrics, tracing)
Dynamic rule loading (DB / API)
DSL for non-technical users
📄 License

MIT


---

# 🧠 Final Thoughts

This version is:

- **Much cleaner** (no artificial separation like TargetData)
- **More idiomatic Go**
- **Production-ready baseline**
- **Aligned with scalable architectures (like rule engines / policy engines)**

---

If you want next step, I’d strongly recommend:

👉 Adding **Datadog tracing + rule-level spans**  
👉 Turning this into a **microservice with REST/gRPC**  
👉 Or evolving into a **policy engine like OPA-lite (but faster in Go)**