package models

import (
	"encoding/json"
	"errors"
	"time"
)

// Memory represents a stored memory with its embedding and metadata
type Memory struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Embedding []float32              `json:"-"` // Not stored in JSON metadata
	Namespace string                 `json:"namespace"`
	Tags      []string               `json:"tags,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// SearchResult represents a memory search result with scoring information
type SearchResult struct {
	Memory   Memory  `json:"memory"`
	Score    float32 `json:"score"`    // Similarity score
	Reranked bool    `json:"reranked"` // Whether result was re-ranked
	Rank     int     `json:"rank"`     // Final rank after re-ranking
}

// MemoryImport represents a memory being imported from a file
// It includes the embedding field which can be serialized during export/import
type MemoryImport struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Embedding []float32              `json:"embedding,omitempty"` // Included for import/export
	Namespace string                 `json:"namespace"`
	Tags      []string               `json:"tags,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// Validate checks if the memory has valid required fields
func (m *Memory) Validate() error {
	if m.ID == "" {
		return errors.New("memory ID cannot be empty")
	}
	if m.Content == "" {
		return errors.New("memory content cannot be empty")
	}
	if m.Namespace == "" {
		return errors.New("memory namespace cannot be empty")
	}
	if len(m.Embedding) == 0 {
		return errors.New("memory embedding cannot be empty")
	}
	if m.CreatedAt.IsZero() {
		return errors.New("memory created_at cannot be zero")
	}
	if m.UpdatedAt.IsZero() {
		return errors.New("memory updated_at cannot be zero")
	}
	return nil
}

// ToJSON serializes the memory to JSON (excluding embedding)
func (m *Memory) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON deserializes JSON to a memory (embedding will be empty)
func (m *Memory) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}

// Clone creates a deep copy of the memory
func (m *Memory) Clone() (*Memory, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	var clone Memory
	if err := json.Unmarshal(data, &clone); err != nil {
		return nil, err
	}

	// Copy embedding separately (not included in JSON)
	clone.Embedding = make([]float32, len(m.Embedding))
	copy(clone.Embedding, m.Embedding)

	return &clone, nil
}

// AddTag adds a tag to the memory if it doesn't already exist
func (m *Memory) AddTag(tag string) {
	for _, t := range m.Tags {
		if t == tag {
			return
		}
	}
	m.Tags = append(m.Tags, tag)
}

// RemoveTag removes a tag from the memory
func (m *Memory) RemoveTag(tag string) {
	for i, t := range m.Tags {
		if t == tag {
			m.Tags = append(m.Tags[:i], m.Tags[i+1:]...)
			return
		}
	}
}

// HasTag checks if the memory has a specific tag
func (m *Memory) HasTag(tag string) bool {
	for _, t := range m.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// SetMetadata sets a metadata key-value pair
func (m *Memory) SetMetadata(key string, value interface{}) {
	if m.Metadata == nil {
		m.Metadata = make(map[string]interface{})
	}
	m.Metadata[key] = value
}

// GetMetadata gets a metadata value by key
func (m *Memory) GetMetadata(key string) (interface{}, bool) {
	if m.Metadata == nil {
		return nil, false
	}
	val, ok := m.Metadata[key]
	return val, ok
}

// Touch updates the UpdatedAt timestamp to the current time
func (m *Memory) Touch() {
	m.UpdatedAt = time.Now()
}

// NewMemory creates a new memory with the given content, namespace, and embedding
func NewMemory(id, content, namespace string, embedding []float32, tags []string) (*Memory, error) {
	now := time.Now()
	memory := &Memory{
		ID:        id,
		Content:   content,
		Embedding: embedding,
		Namespace: namespace,
		Tags:      tags,
		Metadata:  make(map[string]interface{}),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := memory.Validate(); err != nil {
		return nil, err
	}

	return memory, nil
}

// SearchResults is a slice of SearchResult with sorting methods
type SearchResults []SearchResult

// Len returns the number of results
func (r SearchResults) Len() int {
	return len(r)
}

// Less compares two results by score (higher is better)
func (r SearchResults) Less(i, j int) bool {
	return r[i].Score > r[j].Score
}

// Swap swaps two results
func (r SearchResults) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

// SortByScore sorts results by score in descending order
func (r SearchResults) SortByScore() {
	// Simple bubble sort for now (can be optimized later)
	n := len(r)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if r[j].Score < r[j+1].Score {
				r[j], r[j+1] = r[j+1], r[j]
			}
		}
	}
}

// UpdateRanks updates the Rank field based on current order
func (r SearchResults) UpdateRanks() {
	for i := range r {
		r[i].Rank = i + 1
	}
}

// TopN returns the top N results
func (r SearchResults) TopN(n int) SearchResults {
	if n > len(r) {
		n = len(r)
	}
	return r[:n]
}

// FilterByScore filters results to only those with score >= threshold
func (r SearchResults) FilterByScore(threshold float32) SearchResults {
	var filtered SearchResults
	for _, result := range r {
		if result.Score >= threshold {
			filtered = append(filtered, result)
		}
	}
	return filtered
}