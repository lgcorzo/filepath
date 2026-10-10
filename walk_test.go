// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the https://golang.org/LICENSE file.

package filepath

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWalk(t *testing.T) {
	tempDir := t.TempDir()

	// Create structure:
	// tempDir/
	//   a/
	//     sub_file.txt
	//   b.txt
	//   a.txt
	err := os.MkdirAll(filepath.Join(tempDir, "a"), 0755)
	if err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "a", "sub_file.txt"), []byte("sub"), 0644)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "b.txt"), []byte("b"), 0644)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "a.txt"), []byte("a"), 0644)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	var visited []string
	err = Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(tempDir, path)
		visited = append(visited, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	// Flat key sorting order:
	// "a/" comes after "a.txt" because 'a/' > 'a.' ('/' ascii 47, '.' ascii 46)
	// So order should be: ".", "a.txt", "a", "a/sub_file.txt", "b.txt"
	expected := []string{".", "a.txt", "a", filepath.Join("a", "sub_file.txt"), "b.txt"}
	if !reflect.DeepEqual(visited, expected) {
		t.Errorf("visited paths = %v, expected %v", visited, expected)
	}
}

func TestWalkSkipDir(t *testing.T) {
	tempDir := t.TempDir()

	// Create structure:
	// tempDir/
	//   dir1/
	//     file1.txt
	//   dir2/
	//     file2.txt
	err := os.MkdirAll(filepath.Join(tempDir, "dir1"), 0755)
	if err != nil {
		t.Fatalf("failed to create dir1: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "dir1", "file1.txt"), []byte("1"), 0644)
	if err != nil {
		t.Fatalf("failed to create dir1/file1.txt: %v", err)
	}
	err = os.MkdirAll(filepath.Join(tempDir, "dir2"), 0755)
	if err != nil {
		t.Fatalf("failed to create dir2: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "dir2", "file2.txt"), []byte("2"), 0644)
	if err != nil {
		t.Fatalf("failed to create dir2/file2.txt: %v", err)
	}

	var visited []string
	err = Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(tempDir, path)
		visited = append(visited, rel)
		if rel == "dir1" {
			return ErrSkipDir
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed with unexpected error: %v", err)
	}

	expected := []string{".", "dir1", "dir2", filepath.Join("dir2", "file2.txt")}
	if !reflect.DeepEqual(visited, expected) {
		t.Errorf("visited paths = %v, expected %v", visited, expected)
	}
}

func TestWalkSkipFile(t *testing.T) {
	tempDir := t.TempDir()

	err := os.WriteFile(filepath.Join(tempDir, "f1.txt"), []byte("1"), 0644)
	if err != nil {
		t.Fatalf("failed to create f1: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "f2.txt"), []byte("2"), 0644)
	if err != nil {
		t.Fatalf("failed to create f2: %v", err)
	}

	var visited []string
	err = Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(tempDir, path)
		visited = append(visited, rel)
		if rel == "f1.txt" {
			return ErrSkipFile
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed with unexpected error: %v", err)
	}

	expected := []string{".", "f1.txt", "f2.txt"}
	if !reflect.DeepEqual(visited, expected) {
		t.Errorf("visited paths = %v, expected %v", visited, expected)
	}
}

func TestFlatKeySorting(t *testing.T) {
	tempDir := t.TempDir()

	// Directory "foo" and file "foo-bar"
	// "foo/" vs "foo-bar": '/' (47) vs '-' (45) => "foo-bar" comes before "foo"
	err := os.MkdirAll(filepath.Join(tempDir, "foo"), 0755)
	if err != nil {
		t.Fatalf("failed to create foo: %v", err)
	}
	err = os.WriteFile(filepath.Join(tempDir, "foo-bar"), []byte("data"), 0644)
	if err != nil {
		t.Fatalf("failed to create foo-bar: %v", err)
	}

	var visited []string
	err = Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(tempDir, path)
		if rel != "." {
			visited = append(visited, rel)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	expected := []string{"foo-bar", "foo"}
	if !reflect.DeepEqual(visited, expected) {
		t.Errorf("visited paths = %v, expected %v", visited, expected)
	}
}
