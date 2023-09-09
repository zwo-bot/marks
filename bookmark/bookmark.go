package bookmark

import (
	"net/http"
)

type Bookmark struct {
	Title       string
	Path        string
	Description string
	URI         string
	Tags        []string
	Source      string
}

type Bookmarks []Bookmark

func (b Bookmarks) Len() int {
	return len(b)
}

func (b Bookmark) UrlIsValid(i int) bool {
	httpClient := new(http.Client)
	_, err := httpClient.Head(b.URI)
	if err != nil {
		return false
	}
	return true
}
