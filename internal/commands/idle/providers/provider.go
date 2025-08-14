package providers

import (
	"fmt"
	"sync"
	"time"
)

// Cookie represents an inhibition cookie returned by a provider
type Cookie interface {
	// String returns a string representation of the cookie
	String() string
}

// IdleProvider defines the interface for idle prevention providers
type IdleProvider interface {
	// Name returns the provider name
	Name() string

	// Available checks if this provider can be used in the current environment
	Available() bool

	// Priority returns the provider priority (higher = preferred)
	Priority() int

	// Inhibit creates an idle inhibition with the given reason
	Inhibit(reason string) (Cookie, error)

	// Uninhibit releases an idle inhibition using its cookie
	Uninhibit(cookie Cookie) error

	// Status returns whether an inhibition is currently active
	Status() (bool, error)
}

// Registry manages available idle providers
type Registry struct {
	mu        sync.RWMutex
	providers map[string]IdleProvider
	ordered   []IdleProvider // Ordered by priority
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]IdleProvider),
		ordered:   make([]IdleProvider, 0),
	}
}

// Register adds a provider to the registry
func (r *Registry) Register(provider IdleProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := provider.Name()
	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	r.providers[name] = provider

	// Insert provider in priority order
	inserted := false
	for i, p := range r.ordered {
		if provider.Priority() > p.Priority() {
			r.ordered = append(r.ordered[:i], append([]IdleProvider{provider}, r.ordered[i:]...)...)
			inserted = true
			break
		}
	}
	if !inserted {
		r.ordered = append(r.ordered, provider)
	}

	return nil
}

// Get returns a provider by name
func (r *Registry) Get(name string) (IdleProvider, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]
	return provider, exists
}

// GetAvailable returns all available providers in priority order
func (r *Registry) GetAvailable() []IdleProvider {
	r.mu.RLock()
	defer r.mu.RUnlock()

	available := make([]IdleProvider, 0)
	for _, provider := range r.ordered {
		if provider.Available() {
			available = append(available, provider)
		}
	}

	return available
}

// GetBest returns the highest priority available provider
func (r *Registry) GetBest() (IdleProvider, error) {
	available := r.GetAvailable()
	if len(available) == 0 {
		return nil, fmt.Errorf("no available idle providers found")
	}
	return available[0], nil
}

// StringCookie is a simple string-based cookie implementation
type StringCookie struct {
	Value string
}

func (c StringCookie) String() string {
	return c.Value
}

// TimedCookie represents a cookie with an expiration time
type TimedCookie struct {
	Cookie     Cookie
	Expiration time.Time
}

func (c TimedCookie) String() string {
	return fmt.Sprintf("%s (expires: %s)", c.Cookie.String(), c.Expiration.Format(time.RFC3339))
}

// IsExpired checks if the cookie has expired
func (c TimedCookie) IsExpired() bool {
	return time.Now().After(c.Expiration)
}

// TimeRemaining returns the time remaining until expiration
func (c TimedCookie) TimeRemaining() time.Duration {
	remaining := time.Until(c.Expiration)
	if remaining < 0 {
		return 0
	}
	return remaining
}
