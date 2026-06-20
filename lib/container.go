package lib

import (
	"fmt"
	"iter"
	"strings"
	"sync"
)

// Container is a container for domain list entries
type Container interface {
	Add(entry *Entry) error
	Get(name string) (*Entry, bool)
	GetEntry(name string) (*Entry, bool)
	Has(name string) bool
	Loop() iter.Seq[*Entry]
	Len() int
	GetNames() []string
}

// SimpleContainer is a simple implementation of Container
type SimpleContainer struct {
	entries map[string]*Entry
	mu      sync.RWMutex
}

// NewSimpleContainer creates a new SimpleContainer
func NewSimpleContainer() *SimpleContainer {
	return &SimpleContainer{
		entries: make(map[string]*Entry),
	}
}

// Add adds an entry to the container
func (c *SimpleContainer) Add(entry *Entry) error {
	if entry == nil {
		return fmt.Errorf("entry is nil")
	}

	name := strings.ToUpper(strings.TrimSpace(entry.GetName()))
	if name == "" {
		return fmt.Errorf("entry name is empty")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, found := c.entries[name]; found {
		// Merge domains
		existing.AddDomains(entry.GetDomains())
	} else {
		c.entries[name] = entry
	}

	return nil
}

// Get retrieves an entry by name
func (c *SimpleContainer) Get(name string) (*Entry, bool) {
	return c.GetEntry(name)
}

// GetEntry retrieves an entry by name
func (c *SimpleContainer) GetEntry(name string) (*Entry, bool) {
	name = strings.ToUpper(strings.TrimSpace(name))
	
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, found := c.entries[name]
	return entry, found
}

// Has checks if an entry exists
func (c *SimpleContainer) Has(name string) bool {
	name = strings.ToUpper(strings.TrimSpace(name))
	
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, found := c.entries[name]
	return found
}

// Loop iterates over all entries
func (c *SimpleContainer) Loop() iter.Seq[*Entry] {
	return func(yield func(*Entry) bool) {
		c.mu.RLock()
		defer c.mu.RUnlock()

		for _, entry := range c.entries {
			if !yield(entry) {
				return
			}
		}
	}
}

// Len returns the number of entries
func (c *SimpleContainer) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

// GetNames returns all entry names
func (c *SimpleContainer) GetNames() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.entries))
	for name := range c.entries {
		names = append(names, name)
	}
	return names
}
