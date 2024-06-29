package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

// TestInstallDependencies tests the InstallDependencies function
func TestInstallDependencies(t *testing.T) {
	// Create a temporary directory
	dir := t.TempDir()

	// Create a dummy go.mod file in the temporary directory
	goModPath := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run InstallDependencies
	err := InstallDependencies(dir, "./test")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
}

// Mock plugin structure to satisfy plugin interface requirements
type MockPlugin interface{}

// TestLoadPlugin tests the LoadPlugin function
func TestLoadPlugin(t *testing.T) {
	// Setup a temporary plugin file
	pluginDir := t.TempDir()
	pluginFile := filepath.Join(pluginDir, "main.go")
	err := os.WriteFile(pluginFile, []byte(`package main; func main() {}`), 0644)
	os.WriteFile(filepath.Join(pluginDir, "go.mod"), []byte("module test1"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Load the plugin
	_, err = LoadPlugin[MockPlugin](pluginFile, "./test")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
}

// TestLoadPlugins tests the LoadPlugins function
func TestLoadPlugins(t *testing.T) {
	// Setup a temporary plugin file
	pluginDir := t.TempDir()
	pluginFile := filepath.Join(pluginDir, "plugin.go")
	module_file := filepath.Join(pluginDir, "go.mod")
	_ = os.WriteFile(pluginFile, []byte(`package main; func main() {}`), 0644)
	err := os.WriteFile(module_file, []byte("module test1"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Load plugins
	pluginFolders := []string{pluginDir}
	plugins, err := LoadPlugins[MockPlugin](pluginFolders, "./test")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Validate the loaded plugins count and type
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, but got %d", len(plugins))
	}
}

// TestFindGoModDirs tests the findGoModDirs function
func TestFindGoModDirs(t *testing.T) {
	// Create a temporary directory structure with go.mod files
	root := t.TempDir()
	dir1 := filepath.Join(root, "dir1")
	dir2 := filepath.Join(root, "dir2")
	os.Mkdir(dir1, 0755)
	os.Mkdir(dir2, 0755)
	os.WriteFile(filepath.Join(dir1, "go.mod"), []byte("module test1"), 0644)
	os.WriteFile(filepath.Join(dir2, "go.mod"), []byte("module test2"), 0644)

	dirs, err := findGoModDirs(root)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	expected := 2
	if len(dirs) != expected {
		t.Errorf("Expected %d directories, but got %d", expected, len(dirs))
	}
}

// TestLoadPluginsRecursive tests the LoadPluginsRecursive function
func TestLoadPluginsRecursive(t *testing.T) {
	// Setup a temporary plugin directory structure
	root := t.TempDir()
	pluginDir := filepath.Join(root, "plugin")
	os.Mkdir(pluginDir, 0755)
	pluginFile := filepath.Join(pluginDir, "plugin.go")
	os.WriteFile(pluginFile, []byte(`package main; func main() {}`), 0644)
	os.WriteFile(filepath.Join(pluginDir, "go.mod"), []byte("module test"), 0644)

	// Load plugins recursively
	pluginFolders := []string{root}
	plugins, err := LoadPluginsRecursive[MockPlugin](pluginFolders, "./test")
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Validate the loaded plugins count and type
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, but got %d", len(plugins))
	}
}
