package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/raywall/go-decision-engine/decision"
)

func main() {
	// 1. Simple boolean — active user gate
	userFromDB := map[string]any{
		"username": "teste",
		"password": "1234", // sensitive — must NOT leak into Data
		"active":   true,
	}
	d1 := decision.New("active == true", "active")
	ok, err := d1.ValidateWith(userFromDB)
	must(err)
	fmt.Printf("[active user]   pass=%v  data=%v\n", ok, d1.Data) // data has no password

	// 2. Nested path
	apiData := map[string]any{
		"id": 42,
		"subscription": map[string]any{
			"plan":  "pro",
			"trial": false,
		},
		"secret_token": "never-exposed",
	}
	d2 := decision.New(
		`subscription.plan == "pro" && subscription.trial == false`,
		"subscription.plan",
		"subscription.trial",
	)
	ok, err = d2.ValidateWith(apiData)
	must(err)
	fmt.Printf("[subscription]  pass=%v  data=%v\n", ok, d2.Data)

	// 3. Slice membership
	d3 := decision.New(`"admin" in roles`, "roles")
	ok, err = d3.ValidateWith(map[string]any{"roles": []any{"viewer", "admin"}})
	must(err)
	fmt.Printf("[roles admin]   pass=%v\n", ok)

	// 4. Compound — age + active + role
	d4 := decision.New(`active == true && age >= 18 && role == "member"`, "active", "age", "role")
	ok, err = d4.ValidateWith(map[string]any{"active": true, "age": float64(25), "role": "member", "ssn": "secret"})
	must(err)
	fmt.Printf("[compound]      pass=%v  data=%v\n", ok, d4.Data)

	// 5. Struct source
	type DBUser struct {
		Username string `json:"username"`
		Active   bool   `json:"active"`
		Plan     string `json:"plan"`
		Internal string `json:"internal"`
	}
	d5 := decision.New(`active == true && plan == "enterprise"`, "active", "plan")
	ok, err = d5.ValidateWith(DBUser{Active: true, Plan: "enterprise", Internal: "secret"})
	must(err)
	fmt.Printf("[struct]        pass=%v  data=%v\n", ok, d5.Data)

	// 6. Raw JSON bytes (HTTP response simulation)
	rawJSON, _ := json.Marshal(map[string]any{
		"client": map[string]any{"user": map[string]any{"active": true}},
		"noise":  "ignored",
	})
	d6 := decision.New("client.user.active == true", "client.user.active")
	ok, err = d6.ValidateWith(rawJSON)
	must(err)
	fmt.Printf("[json bytes]    pass=%v  data=%v\n", ok, d6.Data)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
