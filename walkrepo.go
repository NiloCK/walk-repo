package walkrepo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

// walkRepo walks through the repository directory, applying .gitignore rules.
func WalkRepo(root string, walkFn filepath.WalkFunc) error {
	var ps []gitignore.Pattern
	domain := []string{}

	walk := func(path string, domain []string, patterns []gitignore.Pattern) error {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		files, err := f.Readdir(-1)
		if err != nil {
			return err
		}

		// First, check for .gitignore in this directory and process it
		localPatterns := make([]gitignore.Pattern, len(patterns))
		copy(localPatterns, patterns)

		for _, file := range files {
			if file.Name() == ".gitignore" {
				filePath := filepath.Join(path, file.Name())
				filePatterns, err := parseFilePatterns(filePath, domain)
				if err != nil {
					return err
				}
				localPatterns = append(localPatterns, filePatterns...)
			}
		}
		matcher := gitignore.NewMatcher(localPatterns)

		// Then process all other files
		for _, file := range files {
			if file.Name() == ".gitignore" {
				continue
			}

			filePath := filepath.Join(path, file.Name())
			// Get relative path components for matching
			relPath, err := filepath.Rel(root, filePath)
			if err != nil {
				return err
			}
			pathComponents := strings.Split(relPath, string(filepath.Separator))
			isIgnored := matcher.Match(pathComponents, file.IsDir())

			if !isIgnored {
				err := walkFn(filePath, file, nil)
				if err != nil {
					if err == filepath.SkipDir && file.IsDir() {
						continue
					}
					return err
				}

				if file.IsDir() {
					newDomain := append(domain, file.Name())
					err := walk(filePath, newDomain, localPatterns)
					if err != nil {
						return err
					}
				}
			}
		}

		return nil
	}

	return walk(root, domain, ps)
}

// parseFilePatterns parses the .gitignore file and returns a list of gitignore.Patterns.
func parseFilePatterns(path string, domain []string) ([]gitignore.Pattern, error) {
	if !strings.HasSuffix(path, ".gitignore") {
		return nil, fmt.Errorf("file %s is not a .gitignore file", path)
	}

	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	filePatterns := []gitignore.Pattern{}

	// Split the contents of the .gitignore file into rawPatterns
	rawPatterns := strings.Split(string(fileBytes), "\n")
	for _, rawPattern := range rawPatterns {
		// Ignore empty lines and comments
		if rawPattern == "" || strings.HasPrefix(rawPattern, "#") {
			continue
		}
		pattern := gitignore.ParsePattern(rawPattern, domain)

		filePatterns = append(filePatterns, pattern)
	}
	return filePatterns, nil
}
