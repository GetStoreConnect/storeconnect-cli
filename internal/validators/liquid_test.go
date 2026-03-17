package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLiquidValidator(t *testing.T) {
	validator := NewLiquidValidator()
	assert.NotNil(t, validator)
}

func TestLiquidValidator_Validate(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantErr     bool
		wantErrMsgs []string
	}{
		{
			name:    "valid template with if/endif",
			content: `<h1>Welcome</h1>{% if user %}Hello {{ user.name }}{% endif %}`,
			wantErr: false,
		},
		{
			name:    "valid template with for/endfor",
			content: `{% for item in items %}{{ item }}{% endfor %}`,
			wantErr: false,
		},
		{
			name:    "valid template with multiple tags",
			content: `<!DOCTYPE html><html>{{ content_for_layout }}</html>`,
			wantErr: false,
		},
		{
			name:        "unclosed if tag",
			content:     `{% if user %}Hello`,
			wantErr:     true,
			wantErrMsgs: []string{"Unclosed tag"},
		},
		{
			name:        "unclosed for tag",
			content:     `{% for item in items %}{{ item }}`,
			wantErr:     true,
			wantErrMsgs: []string{"Unclosed tag"},
		},
		{
			name:        "unclosed unless tag",
			content:     `{% unless condition %}Content`,
			wantErr:     true,
			wantErrMsgs: []string{"Unclosed tag"},
		},
		{
			name:        "unclosed case tag",
			content:     `{% case status %}{% when 'active' %}Active`,
			wantErr:     true,
			wantErrMsgs: []string{"Unclosed tag"},
		},
		{
			name:        "unclosed capture tag",
			content:     `{% capture my_var %}Some text`,
			wantErr:     true,
			wantErrMsgs: []string{"Unclosed tag"},
		},
		{
			name:        "unclosed block tag",
			content:     `{% block content %}Block content`,
			wantErr:     true,
			wantErrMsgs: []string{"Unclosed tag"},
		},
		{
			name:        "unclosed tablerow tag",
			content:     `{% tablerow product in products %}{{ product.name }}`,
			wantErr:     true,
			wantErrMsgs: []string{"Unclosed tag"},
		},
		{
			name:        "unclosed comment tag",
			content:     `{% comment %}This is a comment`,
			wantErr:     true,
			wantErrMsgs: []string{"Unclosed tag"},
		},
		{
			name:        "mismatched tags - if/endunless",
			content:     `{% if user %}Hello{% endunless %}`,
			wantErr:     true,
			wantErrMsgs: []string{"Mismatched tags"},
		},
		{
			name:        "mismatched tags - for/endcase",
			content:     `{% for item in items %}{{ item }}{% endcase %}`,
			wantErr:     true,
			wantErrMsgs: []string{"Mismatched tags"},
		},
		{
			name:        "unmatched closing tag",
			content:     `Hello {% endif %}`,
			wantErr:     true,
			wantErrMsgs: []string{"Unexpected closing tag"},
		},
		{
			name:        "unmatched variable braces",
			content:     `{{ user.name }`,
			wantErr:     true,
			wantErrMsgs: []string{"Unmatched"},
		},
		{
			name:    "nested tags - properly closed",
			content: `{% if user %}{% for item in items %}{{ item }}{% endfor %}{% endif %}`,
			wantErr: false,
		},
		{
			name:        "nested tags - inner not closed",
			content:     `{% if user %}{% for item in items %}{{ item }}{% endif %}`,
			wantErr:     true,
			wantErrMsgs: []string{"Mismatched tags"},
		},
		{
			name:        "nested tags - outer not closed",
			content:     `{% if user %}{% for item in items %}{{ item }}{% endfor %}`,
			wantErr:     true,
			wantErrMsgs: []string{"Unclosed tag"},
		},
		{
			name:        "multiple errors in single template",
			content:     `{% if user %}Hello {{ name }{% for item in items %}`,
			wantErr:     true,
			wantErrMsgs: []string{"Unmatched"},
		},
		{
			name: "all opening tags",
			content: `
				{% if true %}OK{% endif %}
				{% unless false %}OK{% endunless %}
				{% for i in array %}OK{% endfor %}
				{% case var %}{% when 'a' %}OK{% endcase %}
				{% capture x %}OK{% endcapture %}
				{% block main %}OK{% endblock %}
				{% tablerow i in array %}OK{% endtablerow %}
				{% comment %}OK{% endcomment %}
			`,
			wantErr: false,
		},
		{
			name:    "self-closing tags don't need closing",
			content: `{% assign x = 5 %}{% include 'partial' %}{% break %}{% continue %}`,
			wantErr: false,
		},
		{
			name:    "empty template",
			content: ``,
			wantErr: false,
		},
		{
			name:    "template with only HTML",
			content: `<html><body>No Liquid tags</body></html>`,
			wantErr: false,
		},
		{
			name:    "template with variables only",
			content: `{{ var1 }} {{ var2 }} {{ var3 }}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewLiquidValidator()
			errs := validator.Validate(tt.content)

			if tt.wantErr {
				require.NotEmpty(t, errs, "expected errors but got none")
				// Check that all expected error messages appear somewhere in the errors
				for _, expectedMsg := range tt.wantErrMsgs {
					found := false
					for _, err := range errs {
						if contains(err, expectedMsg) {
							found = true
							break
						}
					}
					assert.True(t, found, "expected error message '%s' not found in: %v", expectedMsg, errs)
				}
			} else {
				assert.Empty(t, errs, "expected no errors but got: %v", errs)
			}
		})
	}
}

func TestIsOpeningTag(t *testing.T) {
	tests := []struct {
		tag  string
		want bool
	}{
		{"if", true},
		{"unless", true},
		{"for", true},
		{"case", true},
		{"capture", true},
		{"block", true},
		{"tablerow", true},
		{"comment", true},
		{"assign", false},
		{"include", false},
		{"endif", false},
		{"endfor", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			got := isOpeningTag(tt.tag)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsClosingTag(t *testing.T) {
	tests := []struct {
		tag  string
		want bool
	}{
		{"endif", true},
		{"endunless", true},
		{"endfor", true},
		{"endcase", true},
		{"endcapture", true},
		{"endblock", true},
		{"endtablerow", true},
		{"endcomment", true},
		{"if", false},
		{"for", false},
		{"assign", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			got := isClosingTag(tt.tag)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetOpeningTag(t *testing.T) {
	tests := []struct {
		closingTag string
		want       string
	}{
		{"endif", "if"},
		{"endunless", "unless"},
		{"endfor", "for"},
		{"endcase", "case"},
		{"endcapture", "capture"},
		{"endblock", "block"},
		{"endtablerow", "tablerow"},
		{"endcomment", "comment"},
		{"unknown", "unknown"}, // No "end" prefix, so returns as-is
	}

	for _, tt := range tests {
		t.Run(tt.closingTag, func(t *testing.T) {
			got := getOpeningTag(tt.closingTag)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
