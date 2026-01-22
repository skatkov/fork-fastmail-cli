package tracking

import (
	"errors"
	"fmt"
	"os"

	keyringlib "github.com/99designs/keyring"
	"github.com/salmonumbrella/fastmail-cli/internal/keyring"
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

func openKeyring() (keyringlib.Keyring, error) {
	configDir, err := ConfigDir()
	if err != nil {
		return nil, err
	}

	ring, err := keyringlib.Open(keyringlib.Config{
		ServiceName: keyringService,
		AllowedBackends: []keyringlib.BackendType{
			keyringlib.KeychainBackend,
			keyringlib.WinCredBackend,
			keyringlib.SecretServiceBackend,
			keyringlib.FileBackend,
		},
		FileDir:          configDir,
		FilePasswordFunc: keyringlib.TerminalPrompt,
	})
	if err != nil {
		return nil, err
	}
	return keyring.Wrap(ring, keyring.DefaultTimeout), nil
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

	if err := ring.Set(keyringlib.Item{
		Key:  trackingKeySecretKey,
		Data: []byte(trackingKey),
	}); err != nil {
		return fmt.Errorf("store tracking key: %w", err)
	}

	if err := ring.Set(keyringlib.Item{
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
		if errors.Is(err, keyringlib.ErrKeyNotFound) {
			return "", "", nil
		}
		return "", "", fmt.Errorf("read tracking key: %w", err)
	}

	akItem, err := ring.Get(adminKeySecretKey)
	if err != nil {
		if errors.Is(err, keyringlib.ErrKeyNotFound) {
			return string(tkItem.Data), "", nil
		}
		return "", "", fmt.Errorf("read admin key: %w", err)
	}

	return string(tkItem.Data), string(akItem.Data), nil
}
