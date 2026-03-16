package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// generateCLIReference walks the cobra command tree and produces a markdown
// CLI reference document. Internal commands (help, completion) are excluded.
func generateCLIReference(root *cobra.Command) string {
	var buf strings.Builder
	buf.WriteString("# RPI CLI Reference\n\n")
	buf.WriteString("Available `rpi` subcommands — do NOT invent subcommands not listed here.\n\n")
	buf.WriteString("| Command | Purpose | Key Flags |\n")
	buf.WriteString("|---------|---------|----------|\n")
	walkCommands(&buf, root, root.Name())
	buf.WriteString("\n## Valid Frontmatter Statuses\n\n")
	buf.WriteString("```\n")
	buf.WriteString("draft       → active | approved | superseded\n")
	buf.WriteString("active      → complete | superseded\n")
	buf.WriteString("approved    → implemented | superseded\n")
	buf.WriteString("complete    → archived | superseded\n")
	buf.WriteString("implemented → archived | superseded\n")
	buf.WriteString("```\n\n")
	buf.WriteString("Run `rpi <command> --help` for full flag details and examples.\n")
	return buf.String()
}

// skipCommands are internal cobra commands that should not appear in the reference.
var skipCommands = map[string]bool{
	"help":       true,
	"completion": true,
}

func walkCommands(buf *strings.Builder, cmd *cobra.Command, prefix string) {
	for _, child := range cmd.Commands() {
		if child.Hidden || skipCommands[child.Name()] {
			continue
		}

		fullPath := prefix + " " + child.Name()

		subs := visibleSubcommands(child)
		if len(subs) > 0 {
			// Parent command with subcommands — recurse into children
			walkCommands(buf, child, fullPath)
		} else {
			// Leaf command — emit a table row
			flags := collectFlags(child)
			fmt.Fprintf(buf, "| `%s` | %s | %s |\n", fullPath, child.Short, flags)
		}
	}
}

func visibleSubcommands(cmd *cobra.Command) []*cobra.Command {
	var out []*cobra.Command
	for _, c := range cmd.Commands() {
		if !c.Hidden && !skipCommands[c.Name()] {
			out = append(out, c)
		}
	}
	return out
}

func collectFlags(cmd *cobra.Command) string {
	var flags []string
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		flags = append(flags, "`--"+f.Name+"`")
	})
	if len(flags) == 0 {
		return ""
	}
	return strings.Join(flags, ", ")
}
