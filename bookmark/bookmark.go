package bookmark

import (
	"net/http"
)

type Bookmark struct {
	Title       string
	Path        string
	Description string
	URI         string
	Domain      string // Domain for favicon lookup
	Tags        []string
	Source      string
	Icon        string // Path to cached favicon
}

type Bookmarks []Bookmark

func (b Bookmarks) Len() int {
	return len(b)
}

// RemoveDuplicates removes bookmarks with duplicate title and URI while merging their tags
func (b Bookmarks) RemoveDuplicates() Bookmarks {
	seen := make(map[string]*Bookmark)
	var result Bookmarks

	// Helper function to create unique key from title and URI
	makeKey := func(bm Bookmark) string {
		return bm.Title + "|" + bm.URI
	}

	// Helper function to merge tags without duplicates
	mergeTags := func(existing []string, new []string) []string {
		tagMap := make(map[string]bool)
		for _, tag := range existing {
			tagMap[tag] = true
		}
		for _, tag := range new {
			tagMap[tag] = true
		}

		merged := make([]string, 0, len(tagMap))
		for tag := range tagMap {
			merged = append(merged, tag)
		}
		return merged
	}

	// Process all bookmarks
	for _, bookmark := range b {
		key := makeKey(bookmark)
		if existing, exists := seen[key]; exists {
			// Merge tags if this is a duplicate
			existing.Tags = mergeTags(existing.Tags, bookmark.Tags)
		} else {
			// Create a copy of the bookmark to avoid modifying the original
			bookmarkCopy := bookmark
			seen[key] = &bookmarkCopy
			result = append(result, bookmarkCopy)
		}
	}

	return result
}

// URLIsValid checks whether the bookmark's URI is reachable.
func (b Bookmark) URLIsValid() bool {
	httpClient := new(http.Client)
	resp, err := httpClient.Head(b.URI)
	if err != nil {
		return false
	}
	// Ensure the response body is closed to avoid resource leaks.
	resp.Body.Close()
	return true
}
