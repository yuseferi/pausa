// Package update checks GitHub releases for a newer Pausa version.
//
// Privacy: Pausa makes no network requests at runtime. This check only runs
// when the user explicitly clicks "Check for updates" — never automatically.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pausa/internal/version"
)

const (
	// repoAPI is the GitHub API endpoint for the latest release.
	repoAPI = "https://api.github.com/repos/yuseferi/pausa/releases/latest"
	// releasesPage is the fallback URL shown when the API is unreachable.
	releasesPage = "https://github.com/yuseferi/pausa/releases/latest"
)

// Info is the result of a version check. Error is non-empty when the check
// itself failed (network, parsing); Available is false in that case.
type Info struct {
	Current   string `json:"current"`
	Latest    string `json:"latest"`
	Available bool   `json:"available"`
	URL       string `json:"url"`
	Error     string `json:"error,omitempty"`
}

// Check queries GitHub for the latest release and compares it to the
// built-in version. It never returns a Go error — failures are reported via
// Info.Error so the Wails-bound promise always resolves.
func Check(ctx context.Context) Info {
	current := version.Version
	info := Info{Current: current, Latest: current, URL: releasesPage}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, repoAPI, nil)
	if err != nil {
		info.Error = err.Error()
		return info
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "pausa-update-check/"+current)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		info.Error = err.Error()
		return info
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		info.Error = fmt.Sprintf("github api: %s", resp.Status)
		return info
	}

	var payload struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		info.Error = err.Error()
		return info
	}
	latest := strings.TrimSpace(payload.TagName)
	if latest == "" {
		info.Error = "empty release tag"
		return info
	}
	info.Latest = strings.TrimPrefix(latest, "v")
	if payload.HTMLURL != "" {
		info.URL = payload.HTMLURL
	}
	info.Available = Compare(info.Latest, current) > 0
	return info
}

// Compare returns -1, 0, or +1 depending on whether a is older, equal, or
// newer than b. Versions are parsed as dot-separated integers with an
// optional leading "v"; non-numeric suffixes are ignored ("1.0.4-rc.1" ==
// "1.0.4" for ordering purposes, extra numeric parts count).
func Compare(a, b string) int {
	pa := parse(strings.TrimPrefix(strings.TrimSpace(a), "v"))
	pb := parse(strings.TrimPrefix(strings.TrimSpace(b), "v"))
	n := len(pa)
	if len(pb) > n {
		n = len(pb)
	}
	for i := 0; i < n; i++ {
		var x, y int
		if i < len(pa) {
			x = pa[i]
		}
		if i < len(pb) {
			y = pb[i]
		}
		if x < y {
			return -1
		}
		if x > y {
			return 1
		}
	}
	return 0
}

func parse(v string) []int {
	// Strip pre-release/build metadata for ordering (1.0.4-rc.1 -> 1.0.4).
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			out = append(out, 0)
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			// Take leading digits ("4rc1" -> 4).
			digits := ""
			for _, r := range p {
				if r < '0' || r > '9' {
					break
				}
				digits += string(r)
			}
			n, _ = strconv.Atoi(digits)
		}
		out = append(out, n)
	}
	return out
}
