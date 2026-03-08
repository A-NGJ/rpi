package main

import (
	"testing"
)

func TestParseCheckboxes(t *testing.T) {
	content := `## Phase 1
### Tasks
- [x] First item
- [ ] Second item
- [x] Third item
## Phase 2
### More Tasks
- [ ] Fourth item`

	result := parseCheckboxes(content)
	if result.Total != 4 {
		t.Errorf("expected 4 total, got %d", result.Total)
	}
	if result.Checked != 2 {
		t.Errorf("expected 2 checked, got %d", result.Checked)
	}
	if result.Unchecked != 2 {
		t.Errorf("expected 2 unchecked, got %d", result.Unchecked)
	}
}

func TestParseCheckboxesAllComplete(t *testing.T) {
	content := `- [x] Done one
- [x] Done two
- [x] Done three`

	result := parseCheckboxes(content)
	if result.Unchecked != 0 {
		t.Errorf("expected 0 unchecked, got %d", result.Unchecked)
	}
	if result.Checked != 3 {
		t.Errorf("expected 3 checked, got %d", result.Checked)
	}
}

func TestParseCheckboxesWithContext(t *testing.T) {
	content := `## Phase 1: Setup
### Build
- [ ] Run make build
### Test
- [ ] Run make test
## Phase 2: Deploy
### Release
- [ ] Tag version`

	result := parseCheckboxes(content)
	if len(result.UncheckedItems) != 3 {
		t.Fatalf("expected 3 unchecked items, got %d", len(result.UncheckedItems))
	}

	item := result.UncheckedItems[0]
	if item.Phase != "Phase 1: Setup" {
		t.Errorf("expected phase 'Phase 1: Setup', got '%s'", item.Phase)
	}
	if item.Section != "Build" {
		t.Errorf("expected section 'Build', got '%s'", item.Section)
	}
	if item.Text != "Run make build" {
		t.Errorf("expected text 'Run make build', got '%s'", item.Text)
	}

	item2 := result.UncheckedItems[2]
	if item2.Phase != "Phase 2: Deploy" {
		t.Errorf("expected phase 'Phase 2: Deploy', got '%s'", item2.Phase)
	}
}

func TestExtractPlanFiles(t *testing.T) {
	content := `### Task 1
**File**: ` + "`internal/git/context.go`" + `
Some description here.
### Task 2
**File**: ` + "`cmd/rpi/verify.go`" + `
More text.`

	files := extractPlanFiles(content)
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(files), files)
	}
	if files[0] != "internal/git/context.go" {
		t.Errorf("expected internal/git/context.go, got %s", files[0])
	}
	if files[1] != "cmd/rpi/verify.go" {
		t.Errorf("expected cmd/rpi/verify.go, got %s", files[1])
	}
}

func TestExtractPlanFilesCodeBlocks(t *testing.T) {
	content := "**File**: `src/main.go`\n```go\npackage main\n```\n**File**: `src/util.go`"

	files := extractPlanFiles(content)
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(files), files)
	}
}

func TestComparePlanVsGit(t *testing.T) {
	planFiles := []string{"a.go", "b.go", "c.go"}
	gitFiles := []string{"a.go", "c.go", "d.go"}

	result := comparePlanVsGit(planFiles, gitFiles)

	if len(result.MissingFromGit) != 1 || result.MissingFromGit[0] != "b.go" {
		t.Errorf("expected missing=[b.go], got %v", result.MissingFromGit)
	}
	if len(result.UnexpectedInGit) != 1 || result.UnexpectedInGit[0] != "d.go" {
		t.Errorf("expected unexpected=[d.go], got %v", result.UnexpectedInGit)
	}
}

func TestScanMarkers(t *testing.T) {
	content := `package main

// TODO: implement this
func foo() {} // FIXME: broken
`

	markers := scanMarkers("main.go", content)
	if len(markers) != 2 {
		t.Fatalf("expected 2 markers, got %d: %v", len(markers), markers)
	}
	if markers[0].Type != "TODO" {
		t.Errorf("expected TODO, got %s", markers[0].Type)
	}
	if markers[0].Line != 3 {
		t.Errorf("expected line 3, got %d", markers[0].Line)
	}
	if markers[1].Type != "FIXME" {
		t.Errorf("expected FIXME, got %s", markers[1].Type)
	}
}

func TestScanMarkersCleanFile(t *testing.T) {
	content := `package main

func main() {
	fmt.Println("hello")
}
`
	markers := scanMarkers("main.go", content)
	if len(markers) != 0 {
		t.Errorf("expected 0 markers, got %d: %v", len(markers), markers)
	}
}

func TestScanMarkersMultipleTypes(t *testing.T) {
	content := `// TODO: first
// FIXME: second
// HACK: third
// TODO: fourth
`

	markers := scanMarkers("test.go", content)
	if len(markers) != 4 {
		t.Fatalf("expected 4 markers, got %d", len(markers))
	}

	counts := map[string]int{}
	for _, m := range markers {
		counts[m.Type]++
	}
	if counts["TODO"] != 2 {
		t.Errorf("expected 2 TODO, got %d", counts["TODO"])
	}
	if counts["FIXME"] != 1 {
		t.Errorf("expected 1 FIXME, got %d", counts["FIXME"])
	}
	if counts["HACK"] != 1 {
		t.Errorf("expected 1 HACK, got %d", counts["HACK"])
	}
}
