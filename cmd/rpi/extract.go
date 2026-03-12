package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/frontmatter"
	"github.com/spf13/cobra"
)

var sectionFlag string

var extractCmd = &cobra.Command{
	Use:   "extract <path>",
	Short: "Extract a section from a markdown file",
	Long:  "Parse a markdown file and extract the content of a specific ## heading by case-insensitive prefix match.",
	Args:  cobra.ExactArgs(1),
	RunE:  runExtract,
}

func init() {
	addFormatFlag(extractCmd)
	extractCmd.Flags().StringVar(&sectionFlag, "section", "", "Section heading to extract (required)")
	extractCmd.MarkFlagRequired("section")
	rootCmd.AddCommand(extractCmd)
}

func runExtract(cmd *cobra.Command, args []string) error {
	doc, err := frontmatter.Parse(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	content, ok := frontmatter.ExtractSection(doc.Body, sectionFlag)
	if !ok {
		fmt.Fprintf(os.Stderr, "section %q not found\n", sectionFlag)
		os.Exit(1)
	}

	format := formatFlag
	if format == "" {
		format = "text"
	}

	switch format {
	case "text", "md":
		fmt.Print(content)
	case "json":
		out := map[string]string{
			"path":    args[0],
			"section": sectionFlag,
			"content": content,
		}
		data, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s (expected json, md, or text)", format)
	}

	return nil
}
