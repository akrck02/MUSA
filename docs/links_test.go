// Package docs holds no code: only the tests that keep the documentation true.
package docs

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// relativeLink is a Markdown link or image: its target is the second group.
var relativeLink = regexp.MustCompile(`!?\[[^\]]*\]\(([^)\s]+)\)`)

// TestEveryDocsLinkLeadsToAFile follows each relative link in the README and the docs, so a page
// renamed or moved never leaves a reader at a dead end.
func TestEveryDocsLinkLeadsToAFile(test *testing.T) {
	pages, err := filepath.Glob("../*.md")
	if err != nil {
		test.Fatal(err)
	}
	err = filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err == nil && strings.HasSuffix(path, ".md") {
			pages = append(pages, path)
		}
		return err
	})
	if err != nil {
		test.Fatal(err)
	}
	for _, page := range pages {
		for _, target := range brokenLinks(test, page) {
			test.Errorf("%s links to %s, which does not exist", page, target)
		}
	}
}

// brokenLinks are the relative links of a page whose file is missing.
func brokenLinks(test *testing.T, page string) []string {
	text, err := os.ReadFile(page)
	if err != nil {
		test.Fatal(err)
	}
	var broken []string
	for _, match := range relativeLink.FindAllStringSubmatch(withoutCode(string(text)), -1) {
		target := strings.SplitN(match[1], "#", 2)[0]
		if target == "" || strings.Contains(target, "://") || strings.HasPrefix(target, "mailto:") {
			continue
		}
		if _, err := os.Stat(filepath.Join(filepath.Dir(page), target)); err != nil {
			broken = append(broken, match[1])
		}
	}
	return broken
}

// withoutCode is a page with its fenced code blocks taken out: a link shown as an example of
// Markdown is not a link.
func withoutCode(page string) string {
	var kept []string
	inCode := false
	for _, line := range strings.Split(page, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCode = !inCode
			continue
		}
		if !inCode {
			kept = append(kept, line)
		}
	}
	return strings.Join(kept, "\n")
}
