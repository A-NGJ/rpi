package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/A-NGJ/rpi/internal/workflow"
)

const (
	contractBeginPrefix = "<!-- rpi:contract:begin"
	contractEndMarker   = "<!-- rpi:contract:end -->"
)

// writeContractBlock refreshes the RPI Skill Contract block inside rulesPath.
// Behavior:
//   - rulesPath does not exist: no-op (fresh init writes the rendered block
//     directly via the template).
//   - both begin and end markers present: replace the fenced region in place.
//   - neither marker present: append the block to the end of the file, with
//     one blank line separating it from prior content.
//   - exactly one marker present: warn and leave the file untouched.
//
// Writes are skipped when the resulting bytes equal the existing bytes, so
// repeated calls are idempotent.
func writeContractBlock(w io.Writer, rulesPath string) error {
	blockBytes, err := workflow.ReadAsset("contracts/rpi-contract.md.template")
	if err != nil {
		return fmt.Errorf("read contract asset: %w", err)
	}
	block := strings.TrimRight(string(blockBytes), "\n")

	existing, err := os.ReadFile(rulesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s: %w", rulesPath, err)
	}

	beginIdx := indexOfMarkerLine(existing, contractBeginPrefix)
	endIdx, endLen := indexOfEndMarker(existing)

	var newContent []byte
	var action string

	switch {
	case beginIdx >= 0 && endIdx >= 0:
		if endIdx < beginIdx {
			logWarning(w, fmt.Sprintf("%s contains a malformed contract block (end marker precedes begin); leaving file untouched", rulesPath))
			return nil
		}
		// Replace bytes from start of begin-marker line through end of end-marker line
		// (including the trailing newline that follows the end marker line, if any).
		end := endIdx + endLen
		if end < len(existing) && existing[end] == '\n' {
			end++
		}
		var buf bytes.Buffer
		buf.Write(existing[:beginIdx])
		buf.WriteString(block)
		buf.WriteByte('\n')
		buf.Write(existing[end:])
		newContent = buf.Bytes()
		action = "Updated RPI skill contract block"
	case beginIdx < 0 && endIdx < 0:
		// Ensure a single blank line between prior content and the block.
		trimmed := bytes.TrimRight(existing, "\n")
		var buf bytes.Buffer
		buf.Write(trimmed)
		if len(trimmed) > 0 {
			buf.WriteString("\n\n")
		}
		buf.WriteString(block)
		buf.WriteByte('\n')
		newContent = buf.Bytes()
		action = "Inserted RPI skill contract block"
	case beginIdx >= 0:
		logWarning(w, fmt.Sprintf("%s contains a malformed contract block (missing %s marker); leaving file untouched", rulesPath, contractEndMarker))
		return nil
	default:
		logWarning(w, fmt.Sprintf("%s contains a malformed contract block (missing %s marker); leaving file untouched", rulesPath, contractBeginPrefix))
		return nil
	}

	if bytes.Equal(newContent, existing) {
		return nil
	}

	info, statErr := os.Stat(rulesPath)
	mode := os.FileMode(0644)
	if statErr == nil {
		mode = info.Mode().Perm()
	}
	if err := os.WriteFile(rulesPath, newContent, mode); err != nil {
		return fmt.Errorf("write %s: %w", rulesPath, err)
	}
	logSuccess(w, action)
	return nil
}

// indexOfMarkerLine returns the byte offset where a line starting with marker
// begins. Returns -1 if no such line exists.
func indexOfMarkerLine(content []byte, marker string) int {
	m := []byte(marker)
	pos := 0
	for pos <= len(content) {
		idx := bytes.Index(content[pos:], m)
		if idx < 0 {
			return -1
		}
		absolute := pos + idx
		if absolute == 0 || content[absolute-1] == '\n' {
			return absolute
		}
		pos = absolute + 1
	}
	return -1
}

// indexOfEndMarker returns the byte offset and length of the contract end
// marker line. Returns -1, 0 if not found. The marker must appear at the
// start of a line.
func indexOfEndMarker(content []byte) (int, int) {
	idx := indexOfMarkerLine(content, contractEndMarker)
	if idx < 0 {
		return -1, 0
	}
	return idx, len(contractEndMarker)
}
