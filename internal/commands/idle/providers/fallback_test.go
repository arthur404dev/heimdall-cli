package providers

import (
	"testing"
)

func TestNewFallbackProvider(t *testing.T) {
	t.Run("creates fallback provider", func(t *testing.T) {
		provider := NewFallbackProvider()

		if provider == nil {
			t.Fatal("Provider should not be nil")
		}
		if provider.Name() != "fallback" {
			t.Errorf("Expected name 'fallback', got %s", provider.Name())
		}
		if !provider.Available() {
			t.Error("Fallback provider should always be available")
		}
		if provider.Priority() != 0 {
			t.Errorf("Expected priority 0, got %d", provider.Priority())
		}
	})
}

func TestFallbackProviderInhibit(t *testing.T) {
	t.Run("creates inhibition successfully", func(t *testing.T) {
		provider := NewFallbackProvider()

		cookie, err := provider.Inhibit("test reason")
		if err != nil {
			t.Errorf("Failed to inhibit: %s", err.Error())
		}
		if cookie == nil {
			t.Error("Cookie should not be nil")
		}

		// Check cookie string representation
		cookieStr := cookie.String()
		if !containsSubstring(cookieStr, "fallback") {
			t.Errorf("Expected cookie to contain 'fallback', got %s", cookieStr)
		}
	})

	t.Run("returns same cookie for multiple calls", func(t *testing.T) {
		provider := NewFallbackProvider()

		cookie1, err := provider.Inhibit("test reason 1")
		if err != nil {
			t.Errorf("Failed to inhibit first time: %s", err.Error())
		}

		cookie2, err := provider.Inhibit("test reason 2")
		if err != nil {
			t.Errorf("Failed to inhibit second time: %s", err.Error())
		}

		// Should return the same cookie (already active)
		if cookie1.String() != cookie2.String() {
			t.Errorf("Expected same cookie, got %s and %s", cookie1.String(), cookie2.String())
		}
	})
}

func TestFallbackProviderUninhibit(t *testing.T) {
	t.Run("uninhibits successfully", func(t *testing.T) {
		provider := NewFallbackProvider()

		// First inhibit
		cookie, err := provider.Inhibit("test reason")
		if err != nil {
			t.Errorf("Failed to inhibit: %s", err.Error())
		}

		// Then uninhibit
		err = provider.Uninhibit(cookie)
		if err != nil {
			t.Errorf("Failed to uninhibit: %s", err.Error())
		}
	})

	t.Run("fails with wrong cookie type", func(t *testing.T) {
		provider := NewFallbackProvider()

		// Try to uninhibit with wrong cookie type
		wrongCookie := StringCookie{Value: "wrong"}
		err := provider.Uninhibit(wrongCookie)
		if err == nil {
			t.Error("Expected error with wrong cookie type")
		}
		if err != nil && !containsSubstring(err.Error(), "invalid cookie type") {
			t.Errorf("Expected error to contain 'invalid cookie type', got %s", err.Error())
		}
	})

	t.Run("fails with mismatched cookie", func(t *testing.T) {
		provider := NewFallbackProvider()

		// Inhibit first
		_, err := provider.Inhibit("test reason")
		if err != nil {
			t.Errorf("Failed to inhibit: %s", err.Error())
		}

		// Try to uninhibit with different cookie
		wrongCookie := &FallbackCookie{id: "wrong-id"}
		err = provider.Uninhibit(wrongCookie)
		if err == nil {
			t.Error("Expected error with mismatched cookie")
		}
		if err != nil && !containsSubstring(err.Error(), "cookie mismatch") {
			t.Errorf("Expected error to contain 'cookie mismatch', got %s", err.Error())
		}
	})
}

func TestFallbackProviderStatus(t *testing.T) {
	t.Run("returns correct status", func(t *testing.T) {
		provider := NewFallbackProvider()

		// Initially not active
		active, err := provider.Status()
		if err != nil {
			t.Errorf("Failed to get status: %s", err.Error())
		}
		if active {
			t.Error("Provider should not be active initially")
		}

		// After inhibit, should be active
		cookie, err := provider.Inhibit("test reason")
		if err != nil {
			t.Errorf("Failed to inhibit: %s", err.Error())
		}

		active, err = provider.Status()
		if err != nil {
			t.Errorf("Failed to get status after inhibit: %s", err.Error())
		}
		if !active {
			t.Error("Provider should be active after inhibit")
		}

		// After uninhibit, should not be active
		err = provider.Uninhibit(cookie)
		if err != nil {
			t.Errorf("Failed to uninhibit: %s", err.Error())
		}

		active, err = provider.Status()
		if err != nil {
			t.Errorf("Failed to get status after uninhibit: %s", err.Error())
		}
		if active {
			t.Error("Provider should not be active after uninhibit")
		}
	})
}

func TestFallbackCookie(t *testing.T) {
	t.Run("creates cookie with correct string representation", func(t *testing.T) {
		cookie := FallbackCookie{id: "test-id"}

		result := cookie.String()
		expected := "fallback:test-id"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})

	t.Run("handles empty ID", func(t *testing.T) {
		cookie := FallbackCookie{id: ""}

		result := cookie.String()
		expected := "fallback:"
		if result != expected {
			t.Errorf("Expected %s, got %s", expected, result)
		}
	})
}

// Integration test
func TestFallbackProviderIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	t.Run("full lifecycle", func(t *testing.T) {
		provider := NewFallbackProvider()

		// Verify initial state
		if !provider.Available() {
			t.Error("Fallback provider should always be available")
		}

		active, err := provider.Status()
		if err != nil {
			t.Errorf("Failed to get initial status: %s", err.Error())
		}
		if active {
			t.Error("Provider should not be active initially")
		}

		// Inhibit
		cookie, err := provider.Inhibit("integration test")
		if err != nil {
			t.Errorf("Failed to inhibit: %s", err.Error())
		}
		if cookie == nil {
			t.Error("Cookie should not be nil")
		}

		// Verify active
		active, err = provider.Status()
		if err != nil {
			t.Errorf("Failed to get status after inhibit: %s", err.Error())
		}
		if !active {
			t.Error("Provider should be active after inhibit")
		}

		// Uninhibit
		err = provider.Uninhibit(cookie)
		if err != nil {
			t.Errorf("Failed to uninhibit: %s", err.Error())
		}

		// Verify inactive
		active, err = provider.Status()
		if err != nil {
			t.Errorf("Failed to get status after uninhibit: %s", err.Error())
		}
		if active {
			t.Error("Provider should not be active after uninhibit")
		}
	})
}

// Benchmark tests
func BenchmarkFallbackProviderInhibit(b *testing.B) {
	provider := NewFallbackProvider()

	for i := 0; i < b.N; i++ {
		cookie, err := provider.Inhibit("benchmark test")
		if err != nil {
			b.Fatalf("Failed to inhibit: %s", err.Error())
		}
		provider.Uninhibit(cookie)
	}
}

func BenchmarkFallbackProviderStatus(b *testing.B) {
	provider := NewFallbackProvider()
	cookie, _ := provider.Inhibit("benchmark test")
	defer provider.Uninhibit(cookie)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := provider.Status()
		if err != nil {
			b.Fatalf("Failed to get status: %s", err.Error())
		}
	}
}

// Test utilities - using containsSubstring from provider_test.go
