package coverage

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// intent records whether a phase claims to create a file (NEW) or edit an
// existing one (UPDATED / anything else).
type intent int

const (
	edit intent = iota
	create
)

type fileRef struct {
	path   string
	intent intent
}

// phase is one "## Phase N" slice of a drafted plan.
type phase struct {
	num      int       // the N in "## Phase N"
	title    string    // text after the colon, if any
	files    []fileRef // **File**: entries declared within this phase
	criteria []string  // success-criterion checkbox texts within this phase
	staged   []string  // files listed under a **Stage**: line within this phase
}

// ForwardRef is a phase editing a file that only a later phase creates.
type ForwardRef struct {
	File        string `json:"file"`
	EditPhase   int    `json:"editPhase"`
	CreatePhase int    `json:"createPhase"`
}

// Coverage groups the coverage-mapping findings.
type Coverage struct {
	// OrphanedCriteria are success criteria in a phase that emits no files.
	OrphanedCriteria []string `json:"orphanedCriteria"`
	// UncoveredFiles are files staged for commit that no **File**: entry produces.
	UncoveredFiles []string `json:"uncoveredFiles"`
	// UnjustifiedFiles are **File**: entries that are never staged for commit.
	UnjustifiedFiles []string `json:"unjustifiedFiles"`
}

// Ordering groups the phase-ordering findings.
type Ordering struct {
	ForwardRefs []ForwardRef `json:"forwardRefs"`
	Cycles      [][]string   `json:"cycles"`
}

// Existence groups the file-existence findings.
type Existence struct {
	// MissingEditTargets are files a phase claims to edit that neither exist on
	// disk nor are created by any phase.
	MissingEditTargets []string `json:"missingEditTargets"`
	// DoubleCreated are files created by more than one phase.
	DoubleCreated []string `json:"doubleCreated"`
}

// Result is the full pre-lock verdict, JSON-shaped to mirror the rpi_verify_*
// family.
type Result struct {
	Coverage  Coverage  `json:"coverage"`
	Ordering  Ordering  `json:"ordering"`
	Existence Existence `json:"existence"`
	// HardFailure is true iff a coverage gap (orphaned criterion, uncovered
	// file, or unjustified file) exists. This is the signal --ff honors as
	// blocking; ordering/existence findings never set it.
	HardFailure bool `json:"hardFailure"`
}

var (
	phaseHeadingRe = regexp.MustCompile(`^## Phase (\d+):?\s*(.*)`)
	backtickRe     = regexp.MustCompile("`" + `([^` + "`" + `]+)` + "`")
)

// parsePlan splits a drafted plan into its phases, capturing each phase's
// **File**: entries (with create/edit intent), success-criterion checkboxes,
// and **Stage**: file lists.
func parsePlan(content string) []phase {
	var phases []phase

	for _, line := range strings.Split(content, "\n") {
		if m := phaseHeadingRe.FindStringSubmatch(line); m != nil {
			num, _ := strconv.Atoi(m[1])
			phases = append(phases, phase{num: num, title: strings.TrimSpace(m[2])})
			continue
		}
		if len(phases) == 0 {
			continue
		}
		// cur is recomputed each iteration; we never append to phases inside
		// this block, so the pointer stays valid for the iteration.
		cur := &phases[len(phases)-1]

		if m := planFileRe.FindStringSubmatch(line); m != nil {
			it := edit
			if strings.Contains(strings.ToUpper(m[2]), "NEW") {
				it = create
			}
			cur.files = append(cur.files, fileRef{path: m[1], intent: it})
		}

		if strings.Contains(line, "**Stage**:") {
			for _, sm := range backtickRe.FindAllStringSubmatch(line, -1) {
				cur.staged = append(cur.staged, sm[1])
			}
		}

		if m := uncheckedRe.FindStringSubmatch(line); m != nil {
			cur.criteria = append(cur.criteria, m[2])
		} else if m := checkedRe.FindStringSubmatch(line); m != nil {
			cur.criteria = append(cur.criteria, m[2])
		}
	}

	return phases
}

// Analyze runs the deterministic pre-lock checks over a drafted plan. root is
// the directory against which edit-target existence is resolved (the repo root
// for CLI use; a test dir in unit tests). The same content + same tree always
// produces the same verdict.
func Analyze(content, root string) Result {
	phases := parsePlan(content)

	// Index create/edit declarations by phase position (for ordering) and
	// gather the produced/staged file sets (for coverage).
	creators := map[string][]int{}
	editors := map[string][]int{}
	produced := map[string]bool{}
	stagedSet := map[string]bool{}

	for i := range phases {
		for _, f := range phases[i].files {
			produced[f.path] = true
			if f.intent == create {
				creators[f.path] = append(creators[f.path], i)
			} else {
				editors[f.path] = append(editors[f.path], i)
			}
		}
		for _, s := range phases[i].staged {
			stagedSet[s] = true
		}
	}

	result := Result{
		Coverage: Coverage{
			OrphanedCriteria: []string{},
			UncoveredFiles:   []string{},
			UnjustifiedFiles: []string{},
		},
		Ordering: Ordering{
			ForwardRefs: []ForwardRef{},
			Cycles:      [][]string{},
		},
		Existence: Existence{
			MissingEditTargets: []string{},
			DoubleCreated:      []string{},
		},
	}

	// coverage: staged files that no **File**: entry produces.
	seenUncovered := map[string]bool{}
	for i := range phases {
		for _, s := range phases[i].staged {
			if !produced[s] && !seenUncovered[s] {
				result.Coverage.UncoveredFiles = append(result.Coverage.UncoveredFiles, s)
				seenUncovered[s] = true
			}
		}
	}

	// coverage: **File**: entries that are never staged for commit.
	seenUnjustified := map[string]bool{}
	for i := range phases {
		for _, f := range phases[i].files {
			if !stagedSet[f.path] && !seenUnjustified[f.path] {
				result.Coverage.UnjustifiedFiles = append(result.Coverage.UnjustifiedFiles, f.path)
				seenUnjustified[f.path] = true
			}
		}
	}

	// coverage: criteria in a phase that emits no files have no work to deliver them.
	for i := range phases {
		if len(phases[i].files) == 0 {
			result.Coverage.OrphanedCriteria = append(result.Coverage.OrphanedCriteria, phases[i].criteria...)
		}
	}

	// ordering: a phase editing a file that a later phase creates is a forward reference.
	for path, eds := range editors {
		for _, ci := range creators[path] {
			for _, ei := range eds {
				if ci > ei {
					result.Ordering.ForwardRefs = append(result.Ordering.ForwardRefs, ForwardRef{
						File:        path,
						EditPhase:   phases[ei].num,
						CreatePhase: phases[ci].num,
					})
				}
			}
		}
	}
	sort.Slice(result.Ordering.ForwardRefs, func(a, b int) bool {
		x, y := result.Ordering.ForwardRefs[a], result.Ordering.ForwardRefs[b]
		if x.File != y.File {
			return x.File < y.File
		}
		if x.EditPhase != y.EditPhase {
			return x.EditPhase < y.EditPhase
		}
		return x.CreatePhase < y.CreatePhase
	})

	result.Ordering.Cycles = detectCycles(phases, creators, editors)

	// existence: a file created by more than one phase.
	for path, crs := range creators {
		if len(crs) >= 2 {
			result.Existence.DoubleCreated = append(result.Existence.DoubleCreated, path)
		}
	}
	sort.Strings(result.Existence.DoubleCreated)

	// existence: an edit target that neither a phase creates nor exists on disk.
	for path := range editors {
		if len(creators[path]) > 0 {
			continue
		}
		if fileExists(root, path) {
			continue
		}
		result.Existence.MissingEditTargets = append(result.Existence.MissingEditTargets, path)
	}
	sort.Strings(result.Existence.MissingEditTargets)

	result.HardFailure = len(result.Coverage.OrphanedCriteria) > 0 ||
		len(result.Coverage.UncoveredFiles) > 0 ||
		len(result.Coverage.UnjustifiedFiles) > 0

	return result
}

func fileExists(root, path string) bool {
	full := path
	if root != "" {
		full = filepath.Join(root, path)
	}
	_, err := os.Stat(full)
	return err == nil
}

// detectCycles finds cycles in the create→edit phase-dependency graph (a
// creator phase must precede every editor of the same file). Returns each
// cycle as an ordered list of "Phase N" labels. Deterministic: adjacency lists
// and the cycle list are sorted.
func detectCycles(phases []phase, creators, editors map[string][]int) [][]string {
	adj := map[int][]int{}
	for path, crs := range creators {
		for _, c := range crs {
			for _, e := range editors[path] {
				if c != e {
					adj[c] = append(adj[c], e)
				}
			}
		}
	}
	for k := range adj {
		sort.Ints(adj[k])
	}

	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := map[int]int{}
	var stack []int
	var cycles [][]string

	label := func(i int) string { return "Phase " + strconv.Itoa(phases[i].num) }

	var dfs func(u int)
	dfs = func(u int) {
		color[u] = gray
		stack = append(stack, u)
		for _, v := range adj[u] {
			switch color[v] {
			case gray:
				var cyc []string
				started := false
				for _, n := range stack {
					if n == v {
						started = true
					}
					if started {
						cyc = append(cyc, label(n))
					}
				}
				cyc = append(cyc, label(v))
				cycles = append(cycles, cyc)
			case white:
				dfs(v)
			}
		}
		stack = stack[:len(stack)-1]
		color[u] = black
	}

	for i := range phases {
		if color[i] == white {
			dfs(i)
		}
	}

	sort.Slice(cycles, func(a, b int) bool {
		return strings.Join(cycles[a], ">") < strings.Join(cycles[b], ">")
	})
	if cycles == nil {
		return [][]string{}
	}
	return cycles
}
