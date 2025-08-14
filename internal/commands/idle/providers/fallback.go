package providers

import (
	"fmt"
	"sync"
	"time"
)

// FallbackCookie represents a fallback inhibition cookie
type FallbackCookie struct {
	id        string
	startTime time.Time
}

func (c FallbackCookie) String() string {
	return fmt.Sprintf("fallback:%s", c.id)
}

// FallbackProvider is a no-op provider used when no other providers are available
type FallbackProvider struct {
	mu     sync.Mutex
	active bool
	cookie *FallbackCookie
}

// NewFallbackProvider creates a new fallback provider
func NewFallbackProvider() *FallbackProvider {
	return &FallbackProvider{}
}

// Name returns the provider name
func (p *FallbackProvider) Name() string {
	return "fallback"
}

// Available always returns true as this is the last resort
func (p *FallbackProvider) Available() bool {
	return true
}

// Priority returns the lowest priority
func (p *FallbackProvider) Priority() int {
	return 0
}

// Inhibit creates a fake inhibition (logs warning)
func (p *FallbackProvider) Inhibit(reason string) (Cookie, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.active {
		return p.cookie, nil
	}

	p.cookie = &FallbackCookie{
		id:        fmt.Sprintf("%d", time.Now().Unix()),
		startTime: time.Now(),
	}
	p.active = true

	// This provider doesn't actually prevent idle
	// It's just a placeholder when no real providers are available
	return p.cookie, nil
}

// Uninhibit releases the fake inhibition
func (p *FallbackProvider) Uninhibit(cookie Cookie) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	fallbackCookie, ok := cookie.(*FallbackCookie)
	if !ok {
		return fmt.Errorf("invalid cookie type for fallback provider")
	}

	if p.cookie == nil || p.cookie.id != fallbackCookie.id {
		return fmt.Errorf("cookie mismatch")
	}

	p.active = false
	p.cookie = nil

	return nil
}

// Status returns whether an inhibition is currently active
func (p *FallbackProvider) Status() (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.active, nil
}
