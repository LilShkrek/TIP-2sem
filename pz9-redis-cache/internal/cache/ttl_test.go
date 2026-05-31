package cache

import (
	"testing"
	"time"
)

func TestTTLWithJitter(t *testing.T) {
	base := 120 * time.Second
	jitter := 30 * time.Second

	for i := 0; i < 100; i++ {
		ttl := TTLWithJitter(base, jitter)
		if ttl < base {
			t.Fatalf("ttl must not be less than base: got %s, want at least %s", ttl, base)
		}
		if ttl > base+jitter {
			t.Fatalf("ttl must not be greater than base+jitter: got %s, want at most %s", ttl, base+jitter)
		}
	}
}

func TestTTLWithJitterDisabled(t *testing.T) {
	base := 120 * time.Second

	ttl := TTLWithJitter(base, 0)
	if ttl != base {
		t.Fatalf("ttl mismatch: got %s, want %s", ttl, base)
	}
}
