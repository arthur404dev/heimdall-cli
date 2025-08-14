package providers

import (
	"fmt"
	"testing"
	"time"
)

func TestStringCookie(t *testing.T) {
	t.Run("creates string cookie correctly", func(t *testing.T) {
		value := "test-cookie-value"
		cookie := StringCookie{Value: value}

		if cookie.String() != value {
			t.Errorf("Expected %s, got %s", value, cookie.String())
		}
	})

	t.Run("handles empty string", func(t *testing.T) {
		cookie := StringCookie{Value: ""}

		if cookie.String() != "" {
			t.Errorf("Expected empty string, got %s", cookie.String())
		}
	})
}

func TestTimedCookie(t *testing.T) {
	t.Run("creates timed cookie correctly", func(t *testing.T) {
		innerCookie := StringCookie{Value: "inner"}
		expiration := time.Now().Add(1 * time.Hour)
		cookie := TimedCookie{
			Cookie:     innerCookie,
			Expiration: expiration,
		}

		result := cookie.String()
		if !containsSubstring(result, "inner") {
			t.Errorf("Expected result to contain 'inner', got %s", result)
		}
		if !containsSubstring(result, "expires") {
			t.Errorf("Expected result to contain 'expires', got %s", result)
		}
	})

	t.Run("checks expiration correctly", func(t *testing.T) {
		// Future expiration
		futureExpiration := time.Now().Add(1 * time.Hour)
		futureCookie := TimedCookie{
			Cookie:     StringCookie{Value: "future"},
			Expiration: futureExpiration,
		}

		if futureCookie.IsExpired() {
			t.Error("Future cookie should not be expired")
		}

		// Past expiration
		pastExpiration := time.Now().Add(-1 * time.Hour)
		pastCookie := TimedCookie{
			Cookie:     StringCookie{Value: "past"},
			Expiration: pastExpiration,
		}

		if !pastCookie.IsExpired() {
			t.Error("Past cookie should be expired")
		}
	})

	t.Run("calculates time remaining correctly", func(t *testing.T) {
		// Future expiration
		futureExpiration := time.Now().Add(1 * time.Hour)
		futureCookie := TimedCookie{
			Cookie:     StringCookie{Value: "future"},
			Expiration: futureExpiration,
		}

		remaining := futureCookie.TimeRemaining()
		if remaining <= 0 {
			t.Errorf("Expected positive time remaining, got %v", remaining)
		}
		if remaining > 1*time.Hour {
			t.Errorf("Time remaining should not exceed 1 hour, got %v", remaining)
		}

		// Past expiration
		pastExpiration := time.Now().Add(-1 * time.Hour)
		pastCookie := TimedCookie{
			Cookie:     StringCookie{Value: "past"},
			Expiration: pastExpiration,
		}

		remaining = pastCookie.TimeRemaining()
		if remaining != 0 {
			t.Errorf("Expected 0 time remaining for expired cookie, got %v", remaining)
		}
	})
}

func TestNewRegistry(t *testing.T) {
	t.Run("creates empty registry", func(t *testing.T) {
		registry := NewRegistry()

		if registry == nil {
			t.Fatal("Registry should not be nil")
		}
		if registry.providers == nil {
			t.Error("Providers map should be initialized")
		}
		if registry.ordered == nil {
			t.Error("Ordered slice should be initialized")
		}
		if len(registry.providers) != 0 {
			t.Errorf("Expected empty providers map, got %d items", len(registry.providers))
		}
		if len(registry.ordered) != 0 {
			t.Errorf("Expected empty ordered slice, got %d items", len(registry.ordered))
		}
	})
}

func TestRegistryRegister(t *testing.T) {
	t.Run("registers provider successfully", func(t *testing.T) {
		registry := NewRegistry()
		provider := &mockProvider{
			name:      "test",
			available: true,
			priority:  50,
		}

		err := registry.Register(provider)
		if err != nil {
			t.Errorf("Failed to register provider: %s", err.Error())
		}

		// Verify provider is registered
		retrieved, exists := registry.Get("test")
		if !exists {
			t.Error("Provider should be registered")
		}
		if retrieved != provider {
			t.Error("Retrieved provider should be the same instance")
		}
	})

	t.Run("fails to register duplicate provider", func(t *testing.T) {
		registry := NewRegistry()
		provider1 := &mockProvider{name: "test", available: true, priority: 50}
		provider2 := &mockProvider{name: "test", available: true, priority: 60}

		err := registry.Register(provider1)
		if err != nil {
			t.Fatalf("Failed to register first provider: %s", err.Error())
		}

		err = registry.Register(provider2)
		if err == nil {
			t.Error("Expected error when registering duplicate provider")
		}
		if err != nil && !containsSubstring(err.Error(), "already registered") {
			t.Errorf("Expected error to contain 'already registered', got %s", err.Error())
		}
	})

	t.Run("orders providers by priority", func(t *testing.T) {
		registry := NewRegistry()

		// Register providers in random order
		lowPriority := &mockProvider{name: "low", available: true, priority: 10}
		highPriority := &mockProvider{name: "high", available: true, priority: 90}
		mediumPriority := &mockProvider{name: "medium", available: true, priority: 50}

		registry.Register(lowPriority)
		registry.Register(highPriority)
		registry.Register(mediumPriority)

		// Check ordering (highest priority first)
		if len(registry.ordered) != 3 {
			t.Fatalf("Expected 3 providers, got %d", len(registry.ordered))
		}
		if registry.ordered[0].Name() != "high" {
			t.Errorf("Expected 'high' priority provider first, got %s", registry.ordered[0].Name())
		}
		if registry.ordered[1].Name() != "medium" {
			t.Errorf("Expected 'medium' priority provider second, got %s", registry.ordered[1].Name())
		}
		if registry.ordered[2].Name() != "low" {
			t.Errorf("Expected 'low' priority provider third, got %s", registry.ordered[2].Name())
		}
	})
}

func TestRegistryGet(t *testing.T) {
	t.Run("returns existing provider", func(t *testing.T) {
		registry := NewRegistry()
		provider := &mockProvider{name: "test", available: true, priority: 50}
		registry.Register(provider)

		retrieved, exists := registry.Get("test")
		if !exists {
			t.Error("Provider should exist")
		}
		if retrieved != provider {
			t.Error("Retrieved provider should be the same instance")
		}
	})

	t.Run("returns false for nonexistent provider", func(t *testing.T) {
		registry := NewRegistry()

		_, exists := registry.Get("nonexistent")
		if exists {
			t.Error("Nonexistent provider should not exist")
		}
	})
}

func TestRegistryGetAvailable(t *testing.T) {
	t.Run("returns only available providers", func(t *testing.T) {
		registry := NewRegistry()

		available1 := &mockProvider{name: "available1", available: true, priority: 80}
		available2 := &mockProvider{name: "available2", available: true, priority: 60}
		unavailable := &mockProvider{name: "unavailable", available: false, priority: 90}

		registry.Register(available1)
		registry.Register(unavailable)
		registry.Register(available2)

		availableProviders := registry.GetAvailable()
		if len(availableProviders) != 2 {
			t.Errorf("Expected 2 available providers, got %d", len(availableProviders))
		}

		// Should be ordered by priority
		if availableProviders[0].Name() != "available1" {
			t.Errorf("Expected 'available1' first, got %s", availableProviders[0].Name())
		}
		if availableProviders[1].Name() != "available2" {
			t.Errorf("Expected 'available2' second, got %s", availableProviders[1].Name())
		}
	})

	t.Run("returns empty slice when no providers available", func(t *testing.T) {
		registry := NewRegistry()

		unavailable1 := &mockProvider{name: "unavailable1", available: false, priority: 80}
		unavailable2 := &mockProvider{name: "unavailable2", available: false, priority: 60}

		registry.Register(unavailable1)
		registry.Register(unavailable2)

		availableProviders := registry.GetAvailable()
		if len(availableProviders) != 0 {
			t.Errorf("Expected 0 available providers, got %d", len(availableProviders))
		}
	})
}

func TestRegistryGetBest(t *testing.T) {
	t.Run("returns highest priority available provider", func(t *testing.T) {
		registry := NewRegistry()

		low := &mockProvider{name: "low", available: true, priority: 10}
		high := &mockProvider{name: "high", available: true, priority: 90}
		medium := &mockProvider{name: "medium", available: true, priority: 50}

		registry.Register(low)
		registry.Register(high)
		registry.Register(medium)

		best, err := registry.GetBest()
		if err != nil {
			t.Errorf("Failed to get best provider: %s", err.Error())
		}
		if best.Name() != "high" {
			t.Errorf("Expected 'high' priority provider, got %s", best.Name())
		}
	})

	t.Run("skips unavailable providers", func(t *testing.T) {
		registry := NewRegistry()

		unavailableHigh := &mockProvider{name: "unavailable", available: false, priority: 90}
		availableMedium := &mockProvider{name: "available", available: true, priority: 50}

		registry.Register(unavailableHigh)
		registry.Register(availableMedium)

		best, err := registry.GetBest()
		if err != nil {
			t.Errorf("Failed to get best provider: %s", err.Error())
		}
		if best.Name() != "available" {
			t.Errorf("Expected 'available' provider, got %s", best.Name())
		}
	})

	t.Run("returns error when no providers available", func(t *testing.T) {
		registry := NewRegistry()

		unavailable := &mockProvider{name: "unavailable", available: false, priority: 90}
		registry.Register(unavailable)

		_, err := registry.GetBest()
		if err == nil {
			t.Error("Expected error when no providers available")
		}
		if err != nil && !containsSubstring(err.Error(), "no available") {
			t.Errorf("Expected error to contain 'no available', got %s", err.Error())
		}
	})

	t.Run("returns error when registry is empty", func(t *testing.T) {
		registry := NewRegistry()

		_, err := registry.GetBest()
		if err == nil {
			t.Error("Expected error when registry is empty")
		}
	})
}

// Integration tests
func TestRegistryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("full provider lifecycle", func(t *testing.T) {
		registry := NewRegistry()

		// Register multiple providers
		providers := []*mockProvider{
			{name: "fallback", available: true, priority: 0},
			{name: "systemd", available: true, priority: 30},
			{name: "dbus", available: true, priority: 70},
			{name: "x11", available: false, priority: 50}, // unavailable
		}

		for _, p := range providers {
			err := registry.Register(p)
			if err != nil {
				t.Errorf("Failed to register provider %s: %s", p.name, err.Error())
			}
		}

		// Get best available provider
		best, err := registry.GetBest()
		if err != nil {
			t.Fatalf("Failed to get best provider: %s", err.Error())
		}
		if best.Name() != "dbus" {
			t.Errorf("Expected 'dbus' as best provider, got %s", best.Name())
		}

		// Test inhibition
		cookie, err := best.Inhibit("test reason")
		if err != nil {
			t.Errorf("Failed to inhibit: %s", err.Error())
		}
		if cookie == nil {
			t.Error("Cookie should not be nil")
		}

		// Check status
		active, err := best.Status()
		if err != nil {
			t.Errorf("Failed to get status: %s", err.Error())
		}
		if !active {
			t.Error("Provider should be active after inhibit")
		}

		// Uninhibit
		err = best.Uninhibit(cookie)
		if err != nil {
			t.Errorf("Failed to uninhibit: %s", err.Error())
		}

		// Check status again
		active, err = best.Status()
		if err != nil {
			t.Errorf("Failed to get status: %s", err.Error())
		}
		if active {
			t.Error("Provider should not be active after uninhibit")
		}
	})
}

// Benchmark tests
func BenchmarkRegistryRegister(b *testing.B) {
	registry := NewRegistry()

	for i := 0; i < b.N; i++ {
		provider := &mockProvider{
			name:      fmt.Sprintf("provider-%d", i),
			available: true,
			priority:  i % 100,
		}
		err := registry.Register(provider)
		if err != nil {
			b.Fatalf("Failed to register provider: %s", err.Error())
		}
	}
}

func BenchmarkRegistryGetBest(b *testing.B) {
	registry := NewRegistry()

	// Register some providers
	for i := 0; i < 10; i++ {
		provider := &mockProvider{
			name:      fmt.Sprintf("provider-%d", i),
			available: true,
			priority:  i * 10,
		}
		registry.Register(provider)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := registry.GetBest()
		if err != nil {
			b.Fatalf("Failed to get best provider: %s", err.Error())
		}
	}
}

func BenchmarkRegistryGetAvailable(b *testing.B) {
	registry := NewRegistry()

	// Register providers (mix of available and unavailable)
	for i := 0; i < 20; i++ {
		provider := &mockProvider{
			name:      fmt.Sprintf("provider-%d", i),
			available: i%2 == 0, // Every other provider is available
			priority:  i * 5,
		}
		registry.Register(provider)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		providers := registry.GetAvailable()
		_ = providers
	}
}

// Test utilities
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Mock provider for testing
type mockProvider struct {
	name      string
	available bool
	priority  int
	active    bool
	cookie    Cookie
}

func (p *mockProvider) Name() string {
	return p.name
}

func (p *mockProvider) Available() bool {
	return p.available
}

func (p *mockProvider) Priority() int {
	return p.priority
}

func (p *mockProvider) Inhibit(reason string) (Cookie, error) {
	if !p.available {
		return nil, fmt.Errorf("provider not available")
	}
	p.active = true
	p.cookie = StringCookie{Value: fmt.Sprintf("%s-cookie-%s", p.name, reason)}
	return p.cookie, nil
}

func (p *mockProvider) Uninhibit(cookie Cookie) error {
	if !p.active {
		return fmt.Errorf("no active inhibition")
	}
	if p.cookie != cookie {
		return fmt.Errorf("cookie mismatch")
	}
	p.active = false
	p.cookie = nil
	return nil
}

func (p *mockProvider) Status() (bool, error) {
	return p.active, nil
}
