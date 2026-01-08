package cmd

import (
	"fmt"
	"strings"

	"github.com/salmonumbrella/fastmail-cli/internal/tracking"
	"github.com/spf13/cobra"
)

func newEmailTrackStatusCmd(_ *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show tracking configuration status",
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := tracking.LoadConfig()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			path, _ := tracking.ConfigPath()
			if path != "" {
				fmt.Printf("config_path\t%s\n", path)
			}

			if !cfg.IsConfigured() {
				fmt.Printf("configured\tfalse\n")
				return nil
			}

			fmt.Printf("configured\ttrue\n")
			fmt.Printf("worker_url\t%s\n", cfg.WorkerURL)
			fmt.Printf("admin_configured\t%t\n", strings.TrimSpace(cfg.AdminKey) != "")

			return nil
		},
	}

	return cmd
}
