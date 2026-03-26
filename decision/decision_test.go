package decision_test

import (
	"encoding/json"
	"testing"

	"github.com/raywall/go-decision-engine/decision"
)

// --------------------------------------------------------------------------
// Helpers
// --------------------------------------------------------------------------

func mustParseAndValidate(t *testing.T, d *decision.DecisionArgs, source any) bool {
	t.Helper()
	ok, err := d.ValidateWith(source)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return ok
}

// --------------------------------------------------------------------------
// ParseToData tests
// --------------------------------------------------------------------------

func TestParseToData_SimpleKey(t *testing.T) {
	d := decision.New("active == true", "active")
	source := map[string]any{"username": "teste", "password": "1234", "active": true}

	if err := d.ParseToData(source); err != nil {
		t.Fatalf("ParseToData failed: %v", err)
	}

	if len(d.Data) != 1 {
		t.Errorf("expected 1 key in Data, got %d", len(d.Data))
	}
	if d.Data["active"] != true {
		t.Errorf("expected active=true, got %v", d.Data["active"])
	}
	// Sensitive fields must NOT leak into Data.
	if _, present := d.Data["password"]; present {
		t.Error("password must not be present in Data")
	}
}

func TestParseToData_NestedPath(t *testing.T) {
	d := decision.New("client.user.active == true", "client.user.active")
	source := map[string]any{
		"client": map[string]any{
			"user": map[string]any{
				"active": true,
				"secret": "hidden",
			},
		},
	}

	if err := d.ParseToData(source); err != nil {
		t.Fatalf("ParseToData failed: %v", err)
	}

	// client.user.active must be present.
	client, ok := d.Data["client"].(map[string]any)
	if !ok {
		t.Fatal("expected client to be map")
	}
	user, ok := client["user"].(map[string]any)
	if !ok {
		t.Fatal("expected client.user to be map")
	}
	if user["active"] != true {
		t.Errorf("expected active=true, got %v", user["active"])
	}
	// secret must not leak.
	if _, present := user["secret"]; present {
		t.Error("secret must not be present in Data")
	}
}

func TestParseToData_SliceArg(t *testing.T) {
	d := decision.New(`"admin" in roles`, "roles")
	source := map[string]any{"roles": []any{"admin", "editor"}, "other": "drop"}

	if err := d.ParseToData(source); err != nil {
		t.Fatalf("ParseToData failed: %v", err)
	}

	roles, ok := d.Data["roles"].([]any)
	if !ok {
		t.Fatalf("expected roles to be []any, got %T", d.Data["roles"])
	}
	if len(roles) != 2 {
		t.Errorf("expected 2 roles, got %d", len(roles))
	}
}

func TestParseToData_FromStruct(t *testing.T) {
	type User struct {
		Username string `json:"username"`
		Active   bool   `json:"active"`
		Secret   string `json:"secret"`
	}

	d := decision.New("active == true", "active")
	if err := d.ParseToData(User{Username: "alice", Active: true, Secret: "x"}); err != nil {
		t.Fatalf("ParseToData from struct failed: %v", err)
	}
	if d.Data["active"] != true {
		t.Error("expected active=true")
	}
}

func TestParseToData_FromJSONBytes(t *testing.T) {
	raw := []byte(`{"username":"bob","active":false}`)
	d := decision.New("active == false", "active")
	if err := d.ParseToData(raw); err != nil {
		t.Fatalf("ParseToData from bytes failed: %v", err)
	}
	if d.Data["active"] != false {
		t.Errorf("expected active=false, got %v", d.Data["active"])
	}
}

func TestParseToData_MissingArg_ReturnsError(t *testing.T) {
	d := decision.New("active == true", "active", "missing_field")
	source := map[string]any{"active": true}
	if err := d.ParseToData(source); err == nil {
		t.Error("expected error for missing arg, got nil")
	}
}

// --------------------------------------------------------------------------
// Validate tests
// --------------------------------------------------------------------------

func TestValidate_SimpleBoolean_True(t *testing.T) {
	d := decision.New("active == true", "active")
	source := map[string]any{"username": "teste", "password": "1234", "active": true}
	if !mustParseAndValidate(t, d, source) {
		t.Error("expected validation to pass")
	}
}

func TestValidate_SimpleBoolean_False(t *testing.T) {
	d := decision.New("active == true", "active")
	source := map[string]any{"active": false}
	if mustParseAndValidate(t, d, source) {
		t.Error("expected validation to fail for inactive user")
	}
}

func TestValidate_StringComparison(t *testing.T) {
	d := decision.New(`role == "admin"`, "role")
	source := map[string]any{"role": "admin", "other": "ignored"}
	if !mustParseAndValidate(t, d, source) {
		t.Error("expected role==admin to pass")
	}
}

func TestValidate_NumericComparison(t *testing.T) {
	d := decision.New("age >= 18", "age")
	source := map[string]any{"age": float64(21)} // JSON numbers are float64
	if !mustParseAndValidate(t, d, source) {
		t.Error("expected age>=18 to pass for 21")
	}
}

func TestValidate_NestedPath(t *testing.T) {
	d := decision.New("client.user.active == true", "client.user.active")
	source := map[string]any{
		"client": map[string]any{
			"user": map[string]any{
				"active": true,
			},
		},
	}
	if !mustParseAndValidate(t, d, source) {
		t.Error("expected nested active==true to pass")
	}
}

func TestValidate_SliceContains(t *testing.T) {
	d := decision.New(`"admin" in roles`, "roles")
	source := map[string]any{"roles": []any{"viewer", "admin"}}
	if !mustParseAndValidate(t, d, source) {
		t.Error("expected 'admin' in roles to pass")
	}
}

func TestValidate_SliceNotContains(t *testing.T) {
	d := decision.New(`"admin" in roles`, "roles")
	source := map[string]any{"roles": []any{"viewer"}}
	if mustParseAndValidate(t, d, source) {
		t.Error("expected 'admin' in roles to fail when admin absent")
	}
}

func TestValidate_CompoundExpression(t *testing.T) {
	d := decision.New(`active == true && age >= 18`, "active", "age")
	source := map[string]any{"active": true, "age": float64(20), "extra": "ignored"}
	if !mustParseAndValidate(t, d, source) {
		t.Error("expected compound expression to pass")
	}
}

func TestValidate_CompoundExpression_FailsWhenInactive(t *testing.T) {
	d := decision.New(`active == true && age >= 18`, "active", "age")
	source := map[string]any{"active": false, "age": float64(20)}
	if mustParseAndValidate(t, d, source) {
		t.Error("expected compound expression to fail for inactive")
	}
}

func TestValidate_WithoutParseToData_ReturnsError(t *testing.T) {
	d := decision.New("active == true", "active")
	_, err := d.Validate()
	if err == nil {
		t.Error("expected error when Data is nil")
	}
}

func TestValidate_BadExpression_ReturnsError(t *testing.T) {
	d := decision.New("this is not valid CEL !!!", "active")
	source := map[string]any{"active": true}
	_ = d.ParseToData(source)
	_, err := d.Validate()
	if err == nil {
		t.Error("expected compile error for invalid CEL expression")
	}
}

func TestValidate_NonBoolExpression_ReturnsError(t *testing.T) {
	// d := decision.New("active", "active") // returns value, not bool
	// active is a bool so this actually returns bool — use a string instead
	d2 := decision.New("username", "username")
	source := map[string]any{"username": "alice"}
	_ = d2.ParseToData(source)
	_, err := d2.Validate()
	if err == nil {
		t.Error("expected error for non-bool expression")
	}
}

// --------------------------------------------------------------------------
// JSON round-trip (simulates data coming from an API or DB response)
// --------------------------------------------------------------------------

func TestParseToData_FromAPIResponse(t *testing.T) {
	apiJSON := `{
		"id": 42,
		"username": "alice",
		"active": true,
		"subscription": {
			"plan": "pro",
			"trial": false
		},
		"roles": ["editor", "moderator"],
		"secret_token": "do-not-expose"
	}`

	var apiData map[string]any
	if err := json.Unmarshal([]byte(apiJSON), &apiData); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	d := decision.New(
		`active == true && subscription.plan == "pro" && "editor" in roles`,
		"active",
		"subscription.plan",
		"roles",
	)

	ok, err := d.ValidateWith(apiData)
	if err != nil {
		t.Fatalf("ValidateWith: %v", err)
	}
	if !ok {
		t.Error("expected validation to pass for pro active editor")
	}

	// Sensitive fields must not be present.
	if _, present := d.Data["secret_token"]; present {
		t.Error("secret_token must not leak into Data")
	}
	if _, present := d.Data["id"]; present {
		t.Error("id must not leak into Data")
	}
}
