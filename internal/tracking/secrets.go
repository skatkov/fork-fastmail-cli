package tracking

import (
	"errors"
	"fmt"
	"os"

	"github.com/99designs/keyring"
	"github.com/salmonumbrella/fastmail-cli/internal/keyringutil"
)

const keyringService = "email-tracking"

var (
	errMissingTrackingKey = errors.New("missing tracking key")
	errMissingAdminKey    = errors.New("missing admin key")
)

const (
	trackingKeySecretKey = "tracking_key"
	adminKeySecretKey    = "admin_key"
)

func openKeyring() (keyring.Keyring, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return nil, err
	}

	ring, err := keyring.Open(keyring.Config{
		ServiceName: keyringService,
		AllowedBackends: []keyring.BackendType{
			keyring.KeychainBackend,
			keyring.WinCredBackend,
			keyring.SecretServiceBackend,
			keyring.FileBackend,
		},
		FileDir:          configDir,
		FilePasswordFunc: keyring.TerminalPrompt,
	})
	if err != nil {
		return nil, err
	}
	return keyringutil.Wrap(ring), nil
}

// SaveSecrets stores tracking keys in the keyring
func SaveSecrets(trackingKey, adminKey string) error {
	if trackingKey == "" {
		return errMissingTrackingKey
	}

	if adminKey == "" {
		return errMissingAdminKey
	}

	ring, err := openKeyring()
	if err != nil {
		return fmt.Errorf("open keyring: %w", err)
	}

	if err := ring.Set(keyring.Item{
		Key:  trackingKeySecretKey,
		Data: []byte(trackingKey),
	}); err != nil {
		return fmt.Errorf("store tracking key: %w", err)
	}

	if err := ring.Set(keyring.Item{
		Key:  adminKeySecretKey,
		Data: []byte(adminKey),
	}); err != nil {
		return fmt.Errorf("store admin key: %w", err)
	}

	return nil
}

// LoadSecrets retrieves tracking keys from the keyring
func LoadSecrets() (trackingKey, adminKey string, err error) {
	ring, err := openKeyring()
	if err != nil {
		// If keyring unavailable, return empty (config might have keys inline)
		if os.IsNotExist(err) {
			return "", "", nil
		}
		return "", "", fmt.Errorf("open keyring: %w", err)
	}

	tkItem, err := ring.Get(trackingKeySecretKey)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return "", "", nil
		}
		return "", "", fmt.Errorf("read tracking key: %w", err)
	}

	akItem, err := ring.Get(adminKeySecretKey)
	if err != nil {
		if errors.Is(err, keyring.ErrKeyNotFound) {
			return string(tkItem.Data), "", nil
		}
		return "", "", fmt.Errorf("read admin key: %w", err)
	}

	return string(tkItem.Data), string(akItem.Data), nil
}
