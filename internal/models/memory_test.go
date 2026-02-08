package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMemoryValidate(t *testing.T) {
	tests := []struct {
		name    string
		memory  *Memory
		wantErr bool
	}{
		{
			name: "valid memory",
			memory: &Memory{
				ID:        "test-id",
				Content:   "test content",
				Embedding: []float32{0.1, 0.2, 0.3},
				Namespace: "default",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			memory: &Memory{
				Content:   "test content",
				Embedding: []float32{0.1, 0.2, 0.3},
				Namespace: "default",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing content",
			memory: &Memory{
				ID:        "test-id",
				Embedding: []float32{0.1, 0.2, 0.3},
				Namespace: "default",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing namespace",
			memory: &Memory{
				ID:        "test-id",
				Content:   "test content",
				Embedding: []float32{0.1, 0.2, 0.3},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing embedding",
			memory: &Memory{
				ID:        "test-id",
				Content:   "test content",
				Namespace: "default",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing created_at",
			memory: &Memory{
				ID:        "test-id",
				Content:   "test content",
				Embedding: []float32{0.1, 0.2, 0.3},
				Namespace: "default",
				UpdatedAt: time.Now(),
			},
			wantErr: true,
		},
		{
			name: "missing updated_at",
			memory: &Memory{
				ID:        "test-id",
				Content:   "test content",
				Embedding: []float32{0.1, 0.2, 0.3},
				Namespace: "default",
				CreatedAt: time.Now(),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.memory.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Memory.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewMemory(t *testing.T) {
	embedding := []float32{0.1, 0.2, 0.3}
	tags := []string{"tag1", "tag2"}

	memory, err := NewMemory("test-id", "test content", "default", embedding, tags)
	if err != nil {
		t.Fatalf("NewMemory() error = %v", err)
	}

	if memory.ID != "test-id" {
		t.Errorf("NewMemory() ID = %v, want %v", memory.ID, "test-id")
	}
	if memory.Content != "test content" {
		t.Errorf("NewMemory() Content = %v, want %v", memory.Content, "test content")
	}
	if memory.Namespace != "default" {
		t.Errorf("NewMemory() Namespace = %v, want %v", memory.Namespace, "default")
	}
	if len(memory.Tags) != 2 {
		t.Errorf("NewMemory() Tags length = %v, want %v", len(memory.Tags), 2)
	}
	if memory.Metadata == nil {
		t.Errorf("NewMemory() Metadata should be initialized")
	}
	if time.Since(memory.CreatedAt) > time.Second {
		t.Errorf("NewMemory() CreatedAt should be recent")
	}
	if time.Since(memory.UpdatedAt) > time.Second {
		t.Errorf("NewMemory() UpdatedAt should be recent")
	}
}

func TestMemoryClone(t *testing.T) {
	original := &Memory{
		ID:        "test-id",
		Content:   "test content",
		Embedding: []float32{0.1, 0.2, 0.3},
		Namespace: "default",
		Tags:      []string{"tag1", "tag2"},
		Metadata:  map[string]interface{}{"key": "value"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	clone, err := original.Clone()
	if err != nil {
		t.Fatalf("Clone() error = %v", err)
	}

	if clone.ID != original.ID {
		t.Errorf("Clone() ID = %v, want %v", clone.ID, original.ID)
	}
	if clone.Content != original.Content {
		t.Errorf("Clone() Content = %v, want %v", clone.Content, original.Content)
	}
	if len(clone.Embedding) != len(original.Embedding) {
		t.Errorf("Clone() Embedding length = %v, want %v", len(clone.Embedding), len(original.Embedding))
	}
	for i := range original.Embedding {
		if clone.Embedding[i] != original.Embedding[i] {
			t.Errorf("Clone() Embedding[%d] = %v, want %v", i, clone.Embedding[i], original.Embedding[i])
		}
	}

	// Modify clone to ensure deep copy
clone.Content = "modified"
	if original.Content == "modified" {
		t.Error("Clone() did not create deep copy - original was modified")
	}
}

func TestMemoryAddTag(t *testing.T) {
	memory := &Memory{
		ID:        "test-id",
		Content:   "test content",
		Embedding: []float32{0.1, 0.2, 0.3},
		Namespace: "default",
		Tags:      []string{"tag1"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add new tag
	memory.AddTag("tag2")
	if !memory.HasTag("tag2") {
		t.Error("AddTag() did not add tag")
	}
	if len(memory.Tags) != 2 {
		t.Errorf("AddTag() Tags length = %v, want %v", len(memory.Tags), 2)
	}

	// Add duplicate tag (should not duplicate)
	memory.AddTag("tag2")
	if len(memory.Tags) != 2 {
		t.Errorf("AddTag() duplicated tag, Tags length = %v, want %v", len(memory.Tags), 2)
	}
}

func TestMemoryRemoveTag(t *testing.T) {
	memory := &Memory{
		ID:        "test-id",
		Content:   "test content",
		Embedding: []float32{0.1, 0.2, 0.3},
		Namespace: "default",
		Tags:      []string{"tag1", "tag2", "tag3"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Remove existing tag
	memory.RemoveTag("tag2")
	if memory.HasTag("tag2") {
		t.Error("RemoveTag() did not remove tag")
	}
	if len(memory.Tags) != 2 {
		t.Errorf("RemoveTag() Tags length = %v, want %v", len(memory.Tags), 2)
	}

	// Remove non-existing tag (should be no-op)
	initialLen := len(memory.Tags)
	memory.RemoveTag("nonexistent")
	if len(memory.Tags) != initialLen {
		t.Error("RemoveTag() modified tags when removing non-existent tag")
	}
}

func TestMemoryHasTag(t *testing.T) {
	memory := &Memory{
		ID:        "test-id",
		Content:   "test content",
		Embedding: []float32{0.1, 0.2, 0.3},
		Namespace: "default",
		Tags:      []string{"tag1", "tag2"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if !memory.HasTag("tag1") {
		t.Error("HasTag() should return true for existing tag")
	}
	if memory.HasTag("nonexistent") {
		t.Error("HasTag() should return false for non-existing tag")
	}
}

func TestMemoryMetadata(t *testing.T) {
	memory := &Memory{
		ID:        "test-id",
		Content:   "test content",
		Embedding: []float32{0.1, 0.2, 0.3},
		Namespace: "default",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set metadata
	memory.SetMetadata("key1", "value1")
	memory.SetMetadata("key2", 123)

	// Get metadata
	val, ok := memory.GetMetadata("key1")
	if !ok || val != "value1" {
		t.Errorf("GetMetadata() = %v, %v; want value1, true", val, ok)
	}

	val, ok = memory.GetMetadata("key2")
	if !ok || val != 123 {
		t.Errorf("GetMetadata() = %v, %v; want 123, true", val, ok)
	}

	// Get non-existing metadata
	val, ok = memory.GetMetadata("nonexistent")
	if ok {
		t.Errorf("GetMetadata() for non-existent key should return false, got %v", val)
	}
}

func TestMemoryTouch(t *testing.T) {
	memory := &Memory{
		ID:        "test-id",
		Content:   "test content",
		Embedding: []float32{0.1, 0.2, 0.3},
		Namespace: "default",
		CreatedAt: time.Now().Add(-time.Hour),
		UpdatedAt: time.Now().Add(-time.Hour),
	}

	oldUpdatedAt := memory.UpdatedAt
	time.Sleep(10 * time.Millisecond) // Ensure time difference

	memory.Touch()

	if !memory.UpdatedAt.After(oldUpdatedAt) {
		t.Error("Touch() did not update UpdatedAt")
	}
}

func TestMemoryToJSON(t *testing.T) {
	memory := &Memory{
		ID:        "test-id",
		Content:   "test content",
		Embedding: []float32{0.1, 0.2, 0.3},
		Namespace: "default",
		Tags:      []string{"tag1"},
		Metadata:  map[string]interface{}{"key": "value"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := memory.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("ToJSON() returned empty data")
	}

	// Verify embedding is not in JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if _, ok := parsed["embedding"]; ok {
		t.Error("ToJSON() should not include embedding field")
	}
}

func TestMemoryFromJSON(t *testing.T) {
	jsonData := `{
		"id": "test-id",
		"content": "test content",
		"namespace": "default",
		"tags": ["tag1"],
		"metadata": {"key": "value"},
		"created_at": "2024-01-01T00:00:00Z",
		"updated_at": "2024-01-01T00:00:00Z"
	}`

	var memory Memory
	err := memory.FromJSON([]byte(jsonData))
	if err != nil {
		t.Fatalf("FromJSON() error = %v", err)
	}

	if memory.ID != "test-id" {
		t.Errorf("FromJSON() ID = %v, want test-id", memory.ID)
	}
	if memory.Content != "test content" {
		t.Errorf("FromJSON() Content = %v, want 'test content'", memory.Content)
	}
	if len(memory.Embedding) != 0 {
		t.Error("FromJSON() Embedding should be empty")
	}
}

func TestSearchResultsSortByScore(t *testing.T) {
	results := SearchResults{
		{Score: 0.5},
		{Score: 0.9},
		{Score: 0.3},
		{Score: 0.7},
	}

	results.SortByScore()

	if results[0].Score != 0.9 {
		t.Errorf("SortByScore() first score = %v, want 0.9", results[0].Score)
	}
	if results[1].Score != 0.7 {
		t.Errorf("SortByScore() second score = %v, want 0.7", results[1].Score)
	}
	if results[2].Score != 0.5 {
		t.Errorf("SortByScore() third score = %v, want 0.5", results[2].Score)
	}
	if results[3].Score != 0.3 {
		t.Errorf("SortByScore() fourth score = %v, want 0.3", results[3].Score)
	}
}

func TestSearchResultsTopN(t *testing.T) {
	results := SearchResults{
		{Score: 0.9},
		{Score: 0.7},
		{Score: 0.5},
		{Score: 0.3},
	}

	top2 := results.TopN(2)
	if len(top2) != 2 {
		t.Errorf("TopN() length = %v, want 2", len(top2))
	}

	top10 := results.TopN(10)
	if len(top10) != 4 {
		t.Errorf("TopN() length should be capped at total results, got %v, want 4", len(top10))
	}
}

func TestSearchResultsFilterByScore(t *testing.T) {
	results := SearchResults{
		{Score: 0.9},
		{Score: 0.7},
		{Score: 0.5},
		{Score: 0.3},
	}

	filtered := results.FilterByScore(0.6)
	if len(filtered) != 2 {
		t.Errorf("FilterByScore() length = %v, want 2", len(filtered))
	}

	if filtered[0].Score != 0.9 || filtered[1].Score != 0.7 {
		t.Error("FilterByScore() did not filter correctly")
	}
}

func TestSearchResultsUpdateRanks(t *testing.T) {
	results := SearchResults{
		{Score: 0.9, Rank: 0},
		{Score: 0.7, Rank: 0},
		{Score: 0.5, Rank: 0},
	}

	results.UpdateRanks()

	if results[0].Rank != 1 {
		t.Errorf("UpdateRanks() first rank = %v, want 1", results[0].Rank)
	}
	if results[1].Rank != 2 {
		t.Errorf("UpdateRanks() second rank = %v, want 2", results[1].Rank)
	}
	if results[2].Rank != 3 {
		t.Errorf("UpdateRanks() third rank = %v, want 3", results[2].Rank)
	}
}
