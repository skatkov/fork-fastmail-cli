package cmd

import (
	"bufio"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/salmonumbrella/fastmail-cli/internal/tracking"
	"github.com/spf13/cobra"
)

func newEmailTrackSetupCmd(_ *rootFlags) *cobra.Command {
	var workerURL, trackingKey, adminKey string

	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Set up email tracking",
		Long:  `Configure email open tracking with your Cloudflare Worker URL and keys.`,
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, err := tracking.LoadConfig()
			if err != nil {
				return fmt.Errorf("load tracking config: %w", err)
			}

			// Use existing values as defaults
			if workerURL == "" {
				workerURL = strings.TrimSpace(cfg.WorkerURL)
			}

			// Prompt for worker URL if not provided
			if workerURL == "" {
				fmt.Print("Tracking worker base URL (e.g. https://...workers.dev): ")
				reader := bufio.NewReader(os.Stdin)
				line, readErr := reader.ReadString('\n')
				if readErr == nil {
					workerURL = strings.TrimSpace(line)
				}
			}

			workerURL = strings.TrimSpace(workerURL)
			if workerURL == "" {
				return fmt.Errorf("--worker-url is required")
			}

			// Generate or use provided tracking key
			key := strings.TrimSpace(trackingKey)
			if key == "" {
				key = strings.TrimSpace(cfg.TrackingKey)
			}
			if key == "" {
				key, err = tracking.GenerateKey()
				if err != nil {
					return fmt.Errorf("generate tracking key: %w", err)
				}
			}

			// Generate or use provided admin key
			admin := strings.TrimSpace(adminKey)
			if admin == "" {
				admin = strings.TrimSpace(cfg.AdminKey)
			}
			if admin == "" {
				admin, err = generateAdminKey()
				if err != nil {
					return fmt.Errorf("generate admin key: %w", err)
				}
			}

			// Save secrets to keyring
			if err := tracking.SaveSecrets(key, admin); err != nil {
				return fmt.Errorf("save tracking secrets: %w", err)
			}

			// Save config
			cfg.Enabled = true
			cfg.WorkerURL = workerURL
			cfg.SecretsInKeyring = true
			cfg.TrackingKey = ""
			cfg.AdminKey = ""

			if err := tracking.SaveConfig(cfg); err != nil {
				return fmt.Errorf("save tracking config: %w", err)
			}

			path, _ := tracking.ConfigPath()
			fmt.Printf("configured\ttrue\n")
			if path != "" {
				fmt.Printf("config_path\t%s\n", path)
			}
			fmt.Printf("worker_url\t%s\n", cfg.WorkerURL)

			fmt.Fprintln(os.Stderr, "")
			fmt.Fprintln(os.Stderr, "Next steps (if deploying new worker):")
			fmt.Fprintln(os.Stderr, "  Use these secrets with wrangler:")
			fmt.Fprintf(os.Stderr, "    TRACKING_KEY=%s\n", key)
			fmt.Fprintf(os.Stderr, "    ADMIN_KEY=%s\n", admin)
			fmt.Fprintln(os.Stderr, "  - wrangler secret put TRACKING_KEY")
			fmt.Fprintln(os.Stderr, "  - wrangler secret put ADMIN_KEY")

			return nil
		},
	}

	cmd.Flags().StringVar(&workerURL, "worker-url", "", "Tracking worker base URL")
	cmd.Flags().StringVar(&trackingKey, "tracking-key", "", "Tracking key (base64; generates one if omitted)")
	cmd.Flags().StringVar(&adminKey, "admin-key", "", "Admin key for /opens (generates one if omitted)")

	return cmd
}

func generateAdminKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
