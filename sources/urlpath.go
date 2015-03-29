package sources

import (
	"net/url"
	"path"
	"strings"
)

// URLJoin joins two URL path segments.
func URLJoin(base, path string) string {
	if strings.Index(path, "://") != -1 {
		return path
	}
	if base == "" {
		return path
	}
	if base[len(base)-1] == '/' {
		return base + path
	}
	return base + "/" + path
}

// URLTargetPath returns a relative local file path for the remote file; the
// server hostname in baseURL being replaced by the server alias. Panics
// if the base URL is malformed.
func URLTargetPath(alias, baseURL, remotePath string) string {
	fullURL, err := url.ParseRequestURI(URLJoin(baseURL, remotePath))
	if err != nil {
		panic(err)
	}
	return path.Join(alias, fullURL.Path)
}
