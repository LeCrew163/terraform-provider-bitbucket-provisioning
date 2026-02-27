package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ── isValidProjectKey ────────────────────────────────────────────────────────

func TestIsValidProjectKey(t *testing.T) {
	tests := []struct {
		key   string
		valid bool
	}{
		// ── Valid keys ────────────────────────────────────────────────────────
		{"AB", true},                         // minimum length (2)
		{"ABC", true},                        // simple letters
		{"MYPROJECT", true},                  // common pattern
		{"MY_PROJECT", true},                 // with underscore
		{"PROJ123", true},                    // letters + digits
		{"A1", true},                         // letter + digit
		{"A_1_2_3", true},                    // multiple underscores
		{"Z", false},                         // only 1 char (too short)
		{"A" + strings.Repeat("B", 127), true}, // exactly 128 chars (max)

		// ── Invalid: too short ────────────────────────────────────────────────
		{"", false},  // empty
		{"A", false}, // single char

		// ── Invalid: too long ─────────────────────────────────────────────────
		{"A" + strings.Repeat("B", 128), false}, // 129 chars (over max)

		// ── Invalid: wrong case ───────────────────────────────────────────────
		{"abc", false},   // all lowercase
		{"Abc", false},   // mixed case
		{"aBC", false},   // starts lowercase

		// ── Invalid: illegal first character ──────────────────────────────────
		{"_ABC", false}, // starts with underscore
		{"1ABC", false}, // starts with digit

		// ── Invalid: illegal characters ───────────────────────────────────────
		{"ABC-DEF", false},  // hyphen
		{"ABC DEF", false},  // space
		{"ABC.DEF", false},  // dot
		{"ABC/DEF", false},  // slash
		{"ABC@DEF", false},  // at sign
		{"ABC!DEF", false},  // exclamation mark
	}

	for _, tc := range tests {
		t.Run(tc.key+"_"+boolStr(tc.valid), func(t *testing.T) {
			got := isValidProjectKey(tc.key)
			if got != tc.valid {
				t.Errorf("isValidProjectKey(%q) = %v, want %v", tc.key, got, tc.valid)
			}
		})
	}
}

func boolStr(b bool) string {
	if b {
		return "valid"
	}
	return "invalid"
}

// ── projectKeyValidator ──────────────────────────────────────────────────────

func TestProjectKeyValidator_Description(t *testing.T) {
	v := &projectKeyValidator{}
	ctx := context.Background()
	want := "Project key must be uppercase alphanumeric with underscores, 2-128 characters"

	if got := v.Description(ctx); got != want {
		t.Errorf("Description() = %q, want %q", got, want)
	}

	if got := v.MarkdownDescription(ctx); got != want {
		t.Errorf("MarkdownDescription() = %q, want %q", got, want)
	}
}

func TestProjectKeyValidator_ValidateString_ValidKeys(t *testing.T) {
	ctx := context.Background()
	v := &projectKeyValidator{}

	validKeys := []string{"AB", "MYPROJECT", "MY_PROJECT", "PROJ123", "A1"}

	for _, key := range validKeys {
		t.Run(key, func(t *testing.T) {
			req := validator.StringRequest{ConfigValue: types.StringValue(key)}
			resp := &validator.StringResponse{}
			v.ValidateString(ctx, req, resp)
			if resp.Diagnostics.HasError() {
				t.Errorf("ValidateString(%q) produced unexpected error: %s", key, resp.Diagnostics)
			}
		})
	}
}

func TestProjectKeyValidator_ValidateString_InvalidKeys(t *testing.T) {
	ctx := context.Background()
	v := &projectKeyValidator{}

	invalidKeys := []string{"a", "A", "_ABC", "1ABC", "abc", "AB-CD", "AB CD"}

	for _, key := range invalidKeys {
		t.Run(key, func(t *testing.T) {
			req := validator.StringRequest{ConfigValue: types.StringValue(key)}
			resp := &validator.StringResponse{}
			v.ValidateString(ctx, req, resp)
			if !resp.Diagnostics.HasError() {
				t.Errorf("ValidateString(%q) expected an error but produced none", key)
			}
		})
	}
}

func TestProjectKeyValidator_ValidateString_NullValue(t *testing.T) {
	ctx := context.Background()
	v := &projectKeyValidator{}

	req := validator.StringRequest{ConfigValue: types.StringNull()}
	resp := &validator.StringResponse{}
	v.ValidateString(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("ValidateString(null) should not produce an error, got: %s", resp.Diagnostics)
	}
}

func TestProjectKeyValidator_ValidateString_UnknownValue(t *testing.T) {
	ctx := context.Background()
	v := &projectKeyValidator{}

	req := validator.StringRequest{ConfigValue: types.StringUnknown()}
	resp := &validator.StringResponse{}
	v.ValidateString(ctx, req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("ValidateString(unknown) should not produce an error, got: %s", resp.Diagnostics)
	}
}
