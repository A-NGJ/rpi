package main

import (
	"testing"
)

func TestUpgradeCmd_Registered(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "upgrade" {
			if cmd.Short == "" {
				t.Error("upgrade command has empty Short description")
			}
			return
		}
	}
	t.Error("upgrade command not registered on rootCmd")
}
