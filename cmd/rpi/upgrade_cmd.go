package main

import (
	"github.com/A-NGJ/rpi/internal/upgrade"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade rpi binary to the latest release",
	Long: `Check GitHub Releases for a newer version of rpi and, if found,
download, verify, and replace the current binary.

The command compares the installed version against the latest GitHub release.
If a newer version is available, it downloads the correct binary for your
platform, verifies its SHA256 checksum, and replaces the current binary.`,
	Example: "  rpi upgrade",
	RunE: func(cmd *cobra.Command, args []string) error {
		return upgrade.Upgrade(cmd.OutOrStdout(), "https://api.github.com", "A-NGJ/rpi", version)
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
