package walkrepo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalkRepo(t *testing.T) {
	tests := []struct {
		name         string
		files        map[string]string // map of file path to content
		gitignores   map[string]string // map of .gitignore path to content
		expectedWalk []string          // paths that should be walked
		notExpected  []string          // paths that should not be walked
	}{
		{
			name: "basic ignore file",
			files: map[string]string{
				"file1.txt":       "content",
				"file2.txt":       "content",
				"ignored.txt":     "content",
				"sub/subfile.txt": "content",
				"sub/ignored.txt": "content",
			},
			gitignores: map[string]string{
				".gitignore": "*.txt\n!file1.txt",
			},
			expectedWalk: []string{
				"file1.txt",
				"sub",
			},
			notExpected: []string{
				"file2.txt",
				"ignored.txt",
				"sub/subfile.txt",
				"sub/ignored.txt",
			},
		},
		{
			name: "nested gitignore",
			files: map[string]string{
				"file1.txt":          "content",
				"sub/file2.txt":      "content",
				"sub/ignored.txt":    "content",
				"sub/deep/file3.txt": "content",
			},
			gitignores: map[string]string{
				".gitignore":     "*.log",
				"sub/.gitignore": "ignored.txt",
			},
			expectedWalk: []string{
				"file1.txt",
				"sub",
				"sub/file2.txt",
				"sub/deep",
				"sub/deep/file3.txt",
			},
			notExpected: []string{
				"sub/ignored.txt",
			},
		},
		{
			name: "negation patterns",
			files: map[string]string{
				"build/output.txt":   "content",
				"build/keep.txt":     "content",
				"build/logs/log.txt": "content",
			},
			gitignores: map[string]string{
				".gitignore": "build/*\n!build/keep.txt",
			},
			expectedWalk: []string{
				"build",
				"build/keep.txt",
			},
			notExpected: []string{
				"build/output.txt",
				"build/logs",
				"build/logs/log.txt",
			},
		},
		{
			name: "directory specific ignores",
			files: map[string]string{
				"test/a.txt": "content",
				"test/b.log": "content",
				"prod/a.txt": "content",
				"prod/b.log": "content",
			},
			gitignores: map[string]string{
				"test/.gitignore": "*.txt",
				"prod/.gitignore": "*.log",
			},
			expectedWalk: []string{
				"test",
				"test/b.log",
				"prod",
				"prod/a.txt",
			},
			notExpected: []string{
				"test/a.txt",
				"prod/b.log",
			},
		},
		{
			name: "complex patterns",
			files: map[string]string{
				"doc/foo.md":        "content",
				"doc/bar/baz.md":    "content",
				"src/test.go":       "content",
				"src/temp/test.tmp": "content",
			},
			gitignores: map[string]string{
				".gitignore": "doc/**/*.md\n*.tmp\n",
			},
			expectedWalk: []string{
				"doc",
				"doc/bar",
				"src",
				"src/test.go",
				"src/temp",
			},
			notExpected: []string{
				"doc/foo.md",
				"doc/bar/baz.md",
				"src/temp/test.tmp",
			},
		},
		{
			name: "overlapping patterns",
			files: map[string]string{
				"logs/debug.log":     "content",
				"logs/error.log":     "content",
				"logs/important.txt": "content",
			},
			gitignores: map[string]string{
				".gitignore":      "*.log",
				"logs/.gitignore": "!error.log",
			},
			expectedWalk: []string{
				"logs",
				"logs/error.log",
				"logs/important.txt",
			},
			notExpected: []string{
				"logs/debug.log",
			},
		},
		{
			name: "empty directories",
			files: map[string]string{
				"empty/.gitkeep":     "",
				"not-empty/file.txt": "content",
			},
			gitignores: map[string]string{
				".gitignore": ".gitkeep",
			},
			expectedWalk: []string{
				"empty",
				"not-empty",
				"not-empty/file.txt",
			},
			notExpected: []string{
				"empty/.gitkeep",
			},
		},
		{
			name: "wildcard patterns",
			files: map[string]string{
				"foo.test.js":        "content",
				"bar.test.ts":        "content",
				"test.production.js": "content",
			},
			gitignores: map[string]string{
				".gitignore": "*.test.*",
			},
			expectedWalk: []string{
				"test.production.js",
			},
			notExpected: []string{
				"foo.test.js",
				"bar.test.ts",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tmpDir, err := os.MkdirTemp("", "walkrepo-test-*")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			// Create test files and directories
			for path, content := range tt.files {
				fullPath := filepath.Join(tmpDir, path)
				err := os.MkdirAll(filepath.Dir(fullPath), 0755)
				if err != nil {
					t.Fatal(err)
				}
				err = os.WriteFile(fullPath, []byte(content), 0644)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Create .gitignore files
			for path, content := range tt.gitignores {
				fullPath := filepath.Join(tmpDir, path)
				err := os.MkdirAll(filepath.Dir(fullPath), 0755)
				if err != nil {
					t.Fatal(err)
				}
				err = os.WriteFile(fullPath, []byte(content), 0644)
				if err != nil {
					t.Fatal(err)
				}
			}

			// Track walked paths
			walked := make(map[string]bool)
			err = WalkRepo(tmpDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				relPath, err := filepath.Rel(tmpDir, path)
				if err != nil {
					t.Fatal(err)
				}
				walked[relPath] = true
				return nil
			})

			if err != nil {
				t.Errorf("WalkRepo() error = %v", err)
			}

			// Verify expected paths were walked
			for _, expected := range tt.expectedWalk {
				if !walked[expected] {
					t.Errorf("expected path %q was not walked", expected)
				}
			}

			// Verify not expected paths were not walked
			for _, notExpected := range tt.notExpected {
				if walked[notExpected] {
					t.Errorf("path %q was walked but should have been ignored", notExpected)
				}
			}
		})
	}
}
