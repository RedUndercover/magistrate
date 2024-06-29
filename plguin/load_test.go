package plugin

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RedUndercover/magistrate/cache"
	"github.com/RedUndercover/magistrate/plugin"
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
	err := plugin.InstallDependencies(dir)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
}

// Mock plugin structure to satisfy plugin interface requirements
type MockPlugin struct{}

func (MockPlugin) SomeMethod() {}

// TestLoadPlugin tests the LoadPlugin function
func TestLoadPlugin(t *testing.T) {
	// Setup a temporary plugin file
	pluginDir := t.TempDir()
	pluginFile := filepath.Join(pluginDir, "plugin.go")
	err := os.WriteFile(pluginFile, []byte(`package main; func main() {}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Compute the directory hash
	dirHash, err := cache.HashDirectory(pluginDir)
	if err != nil {
		t.Fatal(err)
	}

	// Cache a mock plugin
	var mp MockPlugin
	cache.PluginRegistry.Set(dirHash, mp)

	// Load the plugin
	plugin, err := plugin.LoadPlugin[MockPlugin](pluginFile)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Validate the loaded plugin type
	if _, ok := plugin.(MockPlugin); !ok {
		t.Errorf("Expected plugin to be of type MockPlugin, but got %T", plugin)
	}
}

// TestLoadPlugins tests the LoadPlugins function
func TestLoadPlugins(t *testing.T) {
	// Setup a temporary plugin file
	pluginDir := t.TempDir()
	pluginFile := filepath.Join(pluginDir, "plugin.go")
	err := os.WriteFile(pluginFile, []byte(`package main; func main() {}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Compute the directory hash
	dirHash, err := cache.HashDirectory(pluginDir)
	if err != nil {
		t.Fatal(err)
	}

	// Cache a mock plugin
	var mp MockPlugin
	cache.PluginRegistry.Set(dirHash, mp)

	// Load plugins
	pluginFolders := []string{pluginFile}
	plugins, err := plugin.LoadPlugins[MockPlugin](pluginFolders)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Validate the loaded plugins count and type
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, but got %d", len(plugins))
	}

	if _, ok := plugins[0].(MockPlugin); !ok {
		t.Errorf("Expected plugin to be of type MockPlugin, but got %T", plugins[0])
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

	dirs, err := plugin.FindGoModDirs(root)
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

	// Compute the directory hash
	dirHash, err := cache.HashDirectory(pluginDir)
	if err != nil {
		t.Fatal(err)
	}

	// Cache a mock plugin
	var mp MockPlugin
	cache.PluginRegistry.Set(dirHash, mp)

	// Load plugins recursively
	pluginFolders := []string{root}
	plugins, err := plugin.LoadPluginsRecursive[MockPlugin](pluginFolders)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}

	// Validate the loaded plugins count and type
	if len(plugins) != 1 {
		t.Errorf("Expected 1 plugin, but got %d", len(plugins))
	}

	if _, ok := plugins[0].(MockPlugin); !ok {
		t.Errorf("Expected plugin to be of type MockPlugin, but got %T", plugins[0])
	}
}
