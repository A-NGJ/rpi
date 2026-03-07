package template

import (
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Auth Refactor!", "auth-refactor"},
		{"Hello World", "hello-world"},
		{"  leading and trailing  ", "leading-and-trailing"},
		{"UPPER CASE", "upper-case"},
		{"special@#$chars", "special-chars"},
		{"multiple   spaces", "multiple-spaces"},
		{"already-slug", "already-slug"},
		{"with--double---hyphens", "with-double-hyphens"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Slugify(tt.input)
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateFilename(t *testing.T) {
	ctx := &RenderContext{
		Date:   "2026-03-08T10:00:00+01:00",
		Topic:  "templates scaffold",
		Ticket: "cli-002",
		Prefix: "cli",
		Number: 2,
	}

	tests := []struct {
		artifactType string
		ctx          *RenderContext
		want         string
	}{
		{"research", ctx, "2026-03-08-templates-scaffold.md"},
		{"design", ctx, "2026-03-08-templates-scaffold.md"},
		{"structure", ctx, "2026-03-08-templates-scaffold.md"},
		{"plan", ctx, "2026-03-08-cli-002-templates-scaffold.md"},
		{"ticket", ctx, "cli-002-templates-scaffold.md"},
		{"ticket-index", ctx, "index.md"},
		{"verify-report", ctx, "2026-03-08-verify-templates-scaffold.md"},
		{"spec", ctx, "templates-scaffold.md"},
	}

	for _, tt := range tests {
		t.Run(tt.artifactType, func(t *testing.T) {
			got := GenerateFilename(tt.artifactType, tt.ctx)
			if got != tt.want {
				t.Errorf("GenerateFilename(%q) = %q, want %q", tt.artifactType, got, tt.want)
			}
		})
	}
}

func TestGenerateFilenamePlanWithoutTicket(t *testing.T) {
	ctx := &RenderContext{
		Date:  "2026-03-08T10:00:00+01:00",
		Topic: "templates scaffold",
	}
	got := GenerateFilename("plan", ctx)
	want := "2026-03-08-templates-scaffold.md"
	if got != want {
		t.Errorf("GenerateFilename(plan, no ticket) = %q, want %q", got, want)
	}
}

func TestResolveAutoVars(t *testing.T) {
	ctx := &RenderContext{}
	err := ResolveAutoVars(ctx)
	if err != nil {
		t.Fatalf("ResolveAutoVars() error: %v", err)
	}

	if ctx.Date == "" {
		t.Error("Date should be populated")
	}
	// In a git repo, these should be real values
	if ctx.GitCommit == "" {
		t.Error("GitCommit should be populated")
	}
	if ctx.Branch == "" {
		t.Error("Branch should be populated")
	}
	if ctx.Repository == "" {
		t.Error("Repository should be populated")
	}
}

func TestRenderTemplate(t *testing.T) {
	ctx := &RenderContext{
		Date:      "2026-03-08T10:00:00+01:00",
		Topic:     "Auth Refactor",
		Ticket:    "cli-007",
		TypeLabel: "Plan",
	}

	result, err := RenderTemplate("simple", ctx, "testdata")
	if err != nil {
		t.Fatalf("RenderTemplate() error: %v", err)
	}

	if !strings.Contains(result, "topic: \"Auth Refactor\"") {
		t.Error("output should contain topic")
	}
	if !strings.Contains(result, "ticket: \"cli-007\"") {
		t.Error("output should contain ticket")
	}
	if !strings.Contains(result, "# Plan: Auth Refactor") {
		t.Error("output should contain heading with type label")
	}
}

func TestRenderTemplateConditionals(t *testing.T) {
	ctx := &RenderContext{
		Date:      "2026-03-08T10:00:00+01:00",
		Topic:     "Simple Task",
		TypeLabel: "Plan",
		// Ticket intentionally empty
	}

	result, err := RenderTemplate("simple", ctx, "testdata")
	if err != nil {
		t.Fatalf("RenderTemplate() error: %v", err)
	}

	if strings.Contains(result, "ticket:") {
		t.Error("output should NOT contain ticket when Ticket is empty")
	}
}

func TestRenderTemplateMissingFile(t *testing.T) {
	ctx := &RenderContext{}
	_, err := RenderTemplate("nonexistent", ctx, "testdata")
	if err == nil {
		t.Error("expected error for missing template file")
	}
}

func TestRenderTemplateInvalidSyntax(t *testing.T) {
	ctx := &RenderContext{}
	_, err := RenderTemplate("invalid", ctx, "testdata")
	if err == nil {
		t.Error("expected error for invalid template syntax")
	}
}
