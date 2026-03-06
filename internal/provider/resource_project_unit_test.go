package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ── isValidProjectKey ────────────────────────────────────────────────────────
// Rules verified against Bitbucket DC 9 (local instance):
//   - Must start with a letter (a-z or A-Z)
//   - Remaining chars: letters, digits, underscores, hyphens
//   - Length: 1-128 characters

func TestIsValidProjectKey(t *testing.T) {
	tests := []struct {
		key   string
		valid bool
	}{
		// ── Valid: minimum length ─────────────────────────────────────────────
		{"A", true},   // single uppercase letter
		{"a", true},   // single lowercase letter
		{"AB", true},  // two letters

		// ── Valid: case insensitive ───────────────────────────────────────────
		{"ABC", true},       // all uppercase
		{"abc", true},       // all lowercase
		{"Abc", true},       // mixed case
		{"aBC", true},       // starts lowercase

		// ── Valid: common patterns ────────────────────────────────────────────
		{"MYPROJECT", true},  // all caps
		{"myproject", true},  // all lowercase
		{"MY_PROJECT", true}, // with underscore
		{"PROJ123", true},    // letters + digits
		{"A1", true},         // letter + digit
		{"A_1_2_3", true},    // multiple underscores
		{"ABC-DEF", true},    // hyphen
		{"A-B-C", true},      // multiple hyphens
		{"Proj-123", true},   // mixed case with hyphen

		// ── Valid: max length (128 chars) ─────────────────────────────────────
		{"A" + strings.Repeat("B", 127), true}, // exactly 128 chars

		// ── Invalid: empty ────────────────────────────────────────────────────
		{"", false},

		// ── Invalid: too long (>128 chars) ────────────────────────────────────
		{"A" + strings.Repeat("B", 128), false}, // 129 chars

		// ── Invalid: illegal first character ──────────────────────────────────
		{"_ABC", false}, // starts with underscore
		{"1ABC", false}, // starts with digit
		{"-ABC", false}, // starts with hyphen

		// ── Invalid: illegal characters ───────────────────────────────────────
		{"ABC DEF", false}, // space
		{"ABC.DEF", false}, // dot
		{"ABC/DEF", false}, // slash
		{"ABC@DEF", false}, // at sign
		{"ABC!DEF", false}, // exclamation mark
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
	want := "Project key must start with a letter and contain only letters, digits, underscores, or hyphens (1-128 characters)"

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

	validKeys := []string{"A", "AB", "abc", "MYPROJECT", "MY_PROJECT", "PROJ123", "A1", "ABC-DEF", "Proj-123"}

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

	invalidKeys := []string{"", "_ABC", "-ABC", "1ABC", "AB CD", "AB.CD"}

	for _, key := range invalidKeys {
		t.Run(key+"_invalid", func(t *testing.T) {
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
