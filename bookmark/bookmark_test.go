package bookmark

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveDuplicates(t *testing.T) {
	bookmarks := Bookmarks{
		{Title: "Google", URI: "https://google.com", Tags: []string{"search"}},
		{Title: "Google", URI: "https://google.com", Tags: []string{"web"}},
		{Title: "GitHub", URI: "https://github.com", Tags: []string{"code"}},
	}

	result := bookmarks.RemoveDuplicates()
	assert.Len(t, result, 2)

	// Check that the Google bookmark has merged tags
	for _, b := range result {
		if b.Title == "Google" {
			assert.Len(t, b.Tags, 2)
			assert.Contains(t, b.Tags, "search")
			assert.Contains(t, b.Tags, "web")
		}
	}
}

func TestRemoveDuplicatesNoDuplicates(t *testing.T) {
	bookmarks := Bookmarks{
		{Title: "Google", URI: "https://google.com"},
		{Title: "GitHub", URI: "https://github.com"},
		{Title: "Reddit", URI: "https://reddit.com"},
	}

	result := bookmarks.RemoveDuplicates()
	assert.Len(t, result, 3)
}

func TestRemoveDuplicatesEmpty(t *testing.T) {
	bookmarks := Bookmarks{}
	result := bookmarks.RemoveDuplicates()
	assert.Len(t, result, 0)
}

func TestRemoveDuplicatesSameTitleDifferentURI(t *testing.T) {
	bookmarks := Bookmarks{
		{Title: "Test", URI: "https://example.com/1"},
		{Title: "Test", URI: "https://example.com/2"},
	}

	result := bookmarks.RemoveDuplicates()
	assert.Len(t, result, 2)
}

func TestRemoveDuplicatesMergesDuplicateTags(t *testing.T) {
	bookmarks := Bookmarks{
		{Title: "Test", URI: "https://example.com", Tags: []string{"a", "b"}},
		{Title: "Test", URI: "https://example.com", Tags: []string{"b", "c"}},
	}

	result := bookmarks.RemoveDuplicates()
	assert.Len(t, result, 1)
	assert.Len(t, result[0].Tags, 3)
	assert.Contains(t, result[0].Tags, "a")
	assert.Contains(t, result[0].Tags, "b")
	assert.Contains(t, result[0].Tags, "c")
}

func TestRemoveDuplicatesPreservesSource(t *testing.T) {
	bookmarks := Bookmarks{
		{Title: "Test", URI: "https://example.com", Source: "Firefox"},
		{Title: "Test", URI: "https://example.com", Source: "Chrome"},
	}

	result := bookmarks.RemoveDuplicates()
	assert.Len(t, result, 1)
	// First occurrence wins
	assert.Equal(t, "Firefox", result[0].Source)
}

func TestBookmarksLen(t *testing.T) {
	bookmarks := Bookmarks{
		{Title: "A"},
		{Title: "B"},
	}
	assert.Equal(t, 2, bookmarks.Len())
}

func TestRemoveDuplicatesNilTags(t *testing.T) {
	bookmarks := Bookmarks{
		{Title: "Test", URI: "https://example.com"},
		{Title: "Test", URI: "https://example.com", Tags: []string{"tag1"}},
	}

	result := bookmarks.RemoveDuplicates()
	assert.Len(t, result, 1)
	assert.Contains(t, result[0].Tags, "tag1")
}
