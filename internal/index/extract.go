package index

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var (
	goPackageRe     = regexp.MustCompile(`^package\s+(\w+)`)
	pyClassIndentRe = regexp.MustCompile(`^class\s+`)
)

// ExtractSymbols scans a file line-by-line and returns all matched symbols plus the detected package name.
func ExtractSymbols(filePath string, cfg *LangConfig) ([]Symbol, string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	if isBinary(f) {
		return nil, "", nil
	}
	// Reset after binary check.
	if _, err := f.Seek(0, 0); err != nil {
		return nil, "", err
	}

	lang := cfg.Name
	var pkg string
	var symbols []Symbol
	var currentScope string

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Package detection.
		if pkg == "" {
			pkg = detectPackage(lang, line, filePath)
		}

		// Scope tracking.
		currentScope = updateScope(lang, line, currentScope)

		// Pattern matching — first match wins.
		for _, pat := range cfg.Patterns {
			m := pat.Re.FindStringSubmatch(line)
			if m == nil || pat.NameGroup >= len(m) {
				continue
			}
			name := m[pat.NameGroup]
			kind := pat.Kind

			// For Python, indented def inside a class is a method.
			if lang == "python" && kind == "function" && len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
				kind = "method"
			}

			sym := Symbol{
				Name:      name,
				Kind:      kind,
				File:      filePath,
				Line:      lineNum,
				Package:   pkg,
				Scope:     scopeFor(lang, kind, currentScope),
				Signature: trimSignature(line),
				Exported:  isExported(lang, name, line),
			}
			symbols = append(symbols, sym)
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, "", err
	}
	return symbols, pkg, nil
}

func detectPackage(lang, line, filePath string) string {
	switch lang {
	case "go":
		if m := goPackageRe.FindStringSubmatch(line); m != nil {
			return m[1]
		}
	case "python":
		return filepath.Base(filepath.Dir(filePath))
	case "javascript", "typescript":
		return filepath.Base(filepath.Dir(filePath))
	case "rust":
		return filepath.Base(filepath.Dir(filePath))
	}
	return ""
}

func updateScope(lang, line, current string) string {
	switch lang {
	case "go":
		// Go methods have receivers — scope is tracked per-symbol via receiver syntax, not via block nesting.
		// Structs/interfaces don't nest functions, so no scope tracking needed.
		return ""
	case "python":
		// Top-level class declaration resets scope.
		if pyClassIndentRe.MatchString(line) {
			m := Languages["python"].Patterns[0].Re.FindStringSubmatch(line)
			if m != nil {
				return m[1]
			}
		}
		// Non-indented non-class line resets scope.
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' && line[0] != '#' {
			return ""
		}
		return current
	case "javascript", "typescript":
		for _, pat := range Languages[lang].Patterns {
			if pat.Kind == "class" {
				if m := pat.Re.FindStringSubmatch(line); m != nil {
					return m[1]
				}
			}
		}
		// Closing brace at column 0 resets scope.
		if strings.TrimSpace(line) == "}" {
			return ""
		}
		return current
	case "rust":
		// impl block tracking.
		implRe := regexp.MustCompile(`^impl\s+(?:\w+\s+for\s+)?(\w+)`)
		if m := implRe.FindStringSubmatch(line); m != nil {
			return m[1]
		}
		if strings.TrimSpace(line) == "}" && current != "" {
			return ""
		}
		return current
	}
	return current
}

func scopeFor(lang, kind, currentScope string) string {
	switch lang {
	case "go":
		// Go scope is extracted from method receiver, not from block nesting.
		return ""
	default:
		if kind == "method" {
			return currentScope
		}
		return ""
	}
}

func trimSignature(line string) string {
	s := strings.TrimSpace(line)
	// Strip trailing opening brace.
	s = strings.TrimRight(s, " {")
	if len(s) > 120 {
		s = s[:120]
	}
	return s
}

func isExported(lang, name, line string) bool {
	switch lang {
	case "go":
		if len(name) == 0 {
			return false
		}
		return unicode.IsUpper(rune(name[0]))
	case "python":
		return !strings.HasPrefix(name, "_")
	case "javascript", "typescript":
		return strings.HasPrefix(strings.TrimSpace(line), "export")
	case "rust":
		trimmed := strings.TrimSpace(line)
		return strings.HasPrefix(trimmed, "pub ") || strings.HasPrefix(trimmed, "pub(")
	}
	return true
}

// Import extraction regexes.
var (
	// Go
	goSingleImportRe = regexp.MustCompile(`^import\s+(?:(\w+)\s+)?"([^"]+)"`)
	goBlockOpenRe    = regexp.MustCompile(`^import\s*\(`)
	goBlockEntryRe   = regexp.MustCompile(`^\s+(?:(\w+)\s+)?"([^"]+)"`)

	// Python
	pyImportRe     = regexp.MustCompile(`^import\s+(\S+)`)
	pyFromImportRe = regexp.MustCompile(`^from\s+(\S+)\s+import`)

	// JS/TS
	jsImportFromRe  = regexp.MustCompile(`^import\s+.*?\s+from\s+['"]([^'"]+)['"]`)
	jsImportPlainRe = regexp.MustCompile(`^import\s+['"]([^'"]+)['"]`)
	jsRequireRe     = regexp.MustCompile(`require\(['"]([^'"]+)['"]\)`)
	jsImportMultiRe = regexp.MustCompile(`^import\s+\{`)

	// Rust
	rustUseSingleRe = regexp.MustCompile(`^(?:pub\s+)?use\s+(.+);`)
	rustUseBlockRe  = regexp.MustCompile(`^(?:pub\s+)?use\s+(\S+?)\{`)
	rustModRe       = regexp.MustCompile(`^(?:pub\s+)?mod\s+(\w+);`)
)

// ExtractImports scans a file and returns all import statements.
func ExtractImports(filePath string, lang string) ([]Import, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var imports []Import
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)
	lineNum := 0

	// State machine for multi-line imports.
	type blockState struct {
		active    bool
		startLine int
		lang      string
		prefix    string // for Rust use blocks: the path prefix before {
	}
	var block blockState

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Handle active multi-line block.
		if block.active {
			switch block.lang {
			case "go":
				trimmed := strings.TrimSpace(line)
				if trimmed == ")" || trimmed == ");" {
					block.active = false
					continue
				}
				if m := goBlockEntryRe.FindStringSubmatch(line); m != nil {
					imp := Import{File: filePath, ImportPath: m[2], Alias: m[1], Line: lineNum}
					imports = append(imports, imp)
				}
			case "python":
				trimmed := strings.TrimSpace(line)
				if strings.Contains(trimmed, ")") {
					block.active = false
				}
				// We already recorded the from-import on the opening line.
			case "javascript", "typescript":
				// Look for the closing "from 'path'" line.
				if m := regexp.MustCompile(`from\s+['"]([^'"]+)['"]`).FindStringSubmatch(line); m != nil {
					imp := Import{File: filePath, ImportPath: m[1], Line: block.startLine}
					imports = append(imports, imp)
					block.active = false
				}
			case "rust":
				trimmed := strings.TrimSpace(line)
				if strings.Contains(trimmed, "}") {
					block.active = false
				}
				// We already recorded the use-block on the opening line.
			}
			continue
		}

		switch lang {
		case "go":
			if goBlockOpenRe.MatchString(line) {
				block = blockState{active: true, startLine: lineNum, lang: "go"}
				continue
			}
			if m := goSingleImportRe.FindStringSubmatch(line); m != nil {
				imp := Import{File: filePath, ImportPath: m[2], Alias: m[1], Line: lineNum}
				imports = append(imports, imp)
			}

		case "python":
			if m := pyFromImportRe.FindStringSubmatch(line); m != nil {
				imp := Import{File: filePath, ImportPath: m[1], Line: lineNum}
				imports = append(imports, imp)
				if strings.Contains(line, "(") && !strings.Contains(line, ")") {
					block = blockState{active: true, startLine: lineNum, lang: "python"}
				}
				continue
			}
			if m := pyImportRe.FindStringSubmatch(line); m != nil {
				imp := Import{File: filePath, ImportPath: m[1], Line: lineNum}
				imports = append(imports, imp)
			}

		case "javascript", "typescript":
			// import ... from 'path' (single line with from)
			if m := jsImportFromRe.FindStringSubmatch(line); m != nil {
				imp := Import{File: filePath, ImportPath: m[1], Line: lineNum}
				imports = append(imports, imp)
				continue
			}
			// import 'path' (side-effect import)
			if m := jsImportPlainRe.FindStringSubmatch(line); m != nil {
				imp := Import{File: filePath, ImportPath: m[1], Line: lineNum}
				imports = append(imports, imp)
				continue
			}
			// Multi-line: import { ... (no from on this line)
			if jsImportMultiRe.MatchString(line) && !strings.Contains(line, "from") {
				block = blockState{active: true, startLine: lineNum, lang: lang}
				continue
			}
			// require('path')
			if m := jsRequireRe.FindStringSubmatch(line); m != nil {
				imp := Import{File: filePath, ImportPath: m[1], Line: lineNum}
				imports = append(imports, imp)
			}

		case "rust":
			// mod x;
			if m := rustModRe.FindStringSubmatch(line); m != nil {
				imp := Import{File: filePath, ImportPath: m[1], Line: lineNum}
				imports = append(imports, imp)
				continue
			}
			// Multi-line use block: use path::{  (no closing } on same line)
			if m := rustUseBlockRe.FindStringSubmatch(line); m != nil && !strings.Contains(line, "}") {
				// Collect the full block content.
				prefix := m[1]
				// Start collecting items from opening line.
				content := strings.TrimSpace(line)
				blockStartLine := lineNum
				for scanner.Scan() {
					lineNum++
					trimmed := strings.TrimSpace(scanner.Text())
					content += " " + trimmed
					if strings.Contains(trimmed, "}") {
						break
					}
				}
				// Extract the full use path from collected content.
				if idx := strings.Index(content, "{"); idx >= 0 {
					// Get everything between { and }
					inner := content[idx+1:]
					if ci := strings.Index(inner, "}"); ci >= 0 {
						inner = inner[:ci]
					}
					// Split by comma, trim each item, filter empties.
					parts := strings.Split(inner, ",")
					var cleaned []string
					for _, p := range parts {
						p = strings.TrimSpace(p)
						if p != "" {
							cleaned = append(cleaned, p)
						}
					}
					importPath := prefix + "{" + strings.Join(cleaned, ", ") + "}"
					imp := Import{File: filePath, ImportPath: importPath, Line: blockStartLine}
					imports = append(imports, imp)
				}
				continue
			}
			// Single-line use
			if m := rustUseSingleRe.FindStringSubmatch(line); m != nil {
				importPath := strings.TrimSpace(m[1])
				imp := Import{File: filePath, ImportPath: importPath, Line: lineNum}
				imports = append(imports, imp)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return imports, nil
}

// isBinary checks the first 512 bytes for null bytes.
func isBinary(f *os.File) bool {
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil || n == 0 {
		return false
	}
	for _, b := range buf[:n] {
		if b == 0 {
			return true
		}
	}
	return false
}
