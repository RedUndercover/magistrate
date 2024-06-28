package cache

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// TestHashArgs tests the Hash function.
func TestHashArgs(t *testing.T) {
	// define hashes
	hash11 := HashArgs("test", 123.4)
	hash12 := HashArgs("test", 123.4)

	hash21 := HashArgs("test", 123, "ABCDEF")
	hash22 := HashArgs("test", 123, "ABCDEF")

	hash31 := HashArgs("test", 123, "ABCDEF", 456)
	hash32 := HashArgs("test", 123, "ABCDEF", 456)

	// test equal hashes
	if hash11 != hash12 {
		t.Errorf("Expected hashes to be equal: %s and %s", hash11, hash12)
	}

	if hash21 != hash22 {
		t.Errorf("Expected hashes to be equal: %s and %s", hash21, hash22)
	}

	if hash31 != hash32 {
		t.Errorf("Expected hashes to be equal: %s and %s", hash31, hash32)
	}

	// test different hashes
	if hash21 == hash31 {
		t.Errorf("Expected hashes to be different: %s and %s", hash11, hash31)
	}

	if hash11 == hash31 {
		t.Errorf("Expected hashes to be different: %s and %s", hash11, hash31)
	}
}

// TestHashDirectory tests the hashDirectory function
func TestHashDirectory(t *testing.T) {
	// Create a temporary directory
	dir, err := ioutil.TempDir("", "testdir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir) // Clean up

	// Create temporary files in the directory
	fileContents := map[string]string{
		"file1.txt": "Hello, World!",
		"file2.txt": "Go is awesome!",
		"file3.txt": "Test content",
	}

	for fileName, content := range fileContents {
		filePath := filepath.Join(dir, fileName)
		err := ioutil.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write file %s: %v", filePath, err)
		}
	}

	// Compute the hash of the directory
	hash, err := hashDirectory(dir)
	if err != nil {
		t.Fatalf("Failed to hash directory: %v", err)
	}

	// Computed hash for the test files
	expectedHash := "b375d4bb463d9bfa1bc7a872d6e38f87"

	// Check if the computed hash matches the expected hash
	if hash != expectedHash {
		t.Errorf("Hash mismatch: got %s, want %s", hash, expectedHash)
	}
}
