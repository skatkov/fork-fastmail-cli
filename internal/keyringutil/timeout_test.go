package keyringutil

import (
	"errors"
	"testing"
	"time"

	keyringlib "github.com/99designs/keyring"
	cerrors "github.com/salmonumbrella/fastmail-cli/internal/errors"
)

type blockingKeyring struct {
	block chan struct{}
}

func (k *blockingKeyring) Get(key string) (keyringlib.Item, error) {
	<-k.block
	return keyringlib.Item{Key: key}, nil
}

func (k *blockingKeyring) GetMetadata(key string) (keyringlib.Metadata, error) {
	<-k.block
	return keyringlib.Metadata{}, nil
}

func (k *blockingKeyring) Set(item keyringlib.Item) error {
	<-k.block
	return nil
}

func (k *blockingKeyring) Remove(key string) error {
	<-k.block
	return nil
}

func (k *blockingKeyring) Keys() ([]string, error) {
	<-k.block
	return []string{"alpha"}, nil
}

type testKeyring struct {
	keys []string
}

func (k *testKeyring) Get(key string) (keyringlib.Item, error) {
	return keyringlib.Item{Key: key}, nil
}

func (k *testKeyring) GetMetadata(key string) (keyringlib.Metadata, error) {
	return keyringlib.Metadata{}, nil
}

func (k *testKeyring) Set(item keyringlib.Item) error {
	return nil
}

func (k *testKeyring) Remove(key string) error {
	return nil
}

func (k *testKeyring) Keys() ([]string, error) {
	return k.keys, nil
}

func TestTimeoutKeyring_KeysTimeout(t *testing.T) {
	block := make(chan struct{})
	ring := &blockingKeyring{block: block}
	wrapped := Wrap(ring, 10*time.Millisecond)

	start := time.Now()
	_, err := wrapped.Keys()
	close(block)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	var timeoutErr *TimeoutError
	if !errors.As(err, &timeoutErr) {
		t.Fatalf("expected TimeoutError, got %T", err)
	}
	if !cerrors.ContainsSuggestion(err) {
		t.Fatalf("expected suggestion on timeout error")
	}
	if time.Since(start) > 100*time.Millisecond {
		t.Fatalf("timeout took too long: %s", time.Since(start))
	}
}

func TestTimeoutKeyring_KeysSuccess(t *testing.T) {
	ring := &testKeyring{keys: []string{"alpha", "bravo"}}
	wrapped := Wrap(ring, 10*time.Millisecond)

	keys, err := wrapped.Keys()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(keys) != 2 || keys[0] != "alpha" || keys[1] != "bravo" {
		t.Fatalf("unexpected keys: %v", keys)
	}
}
