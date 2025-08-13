package schema

import (
	"fmt"
	"sync"
)

// Registry manages configuration schemas
type Registry struct {
	schemas map[string]*Schema
	mu      sync.RWMutex
}

// NewRegistry creates a new schema registry
func NewRegistry() *Registry {
	return &Registry{
		schemas: make(map[string]*Schema),
	}
}

// Register adds a schema to the registry
func (r *Registry) Register(domain string, schema *Schema) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if schema == nil {
		return fmt.Errorf("cannot register nil schema")
	}

	r.schemas[domain] = schema
	return nil
}

// GetSchema retrieves a schema by domain
func (r *Registry) GetSchema(domain string) *Schema {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.schemas[domain]
}

// HasSchema checks if a schema exists for a domain
func (r *Registry) HasSchema(domain string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.schemas[domain]
	return exists
}

// Remove removes a schema from the registry
func (r *Registry) Remove(domain string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.schemas, domain)
}

// ListDomains returns all registered domains
func (r *Registry) ListDomains() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	domains := make([]string, 0, len(r.schemas))
	for domain := range r.schemas {
		domains = append(domains, domain)
	}
	return domains
}

// Clear removes all schemas from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.schemas = make(map[string]*Schema)
}
