package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"os/exec"

	"github.com/RedUndercover/magistrate/cache"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type ContractVerifyError struct {
	Contract string
	Plugin   string
	msg      string
}

func (e *ContractVerifyError) Error() string {
	return e.msg
}

type dependenciesError struct {
	Plugin string
	msg    string
}

func (e dependenciesError) Error() string {
	return e.msg
}

// installDependencies installs the dependencies of a plugin.
// It runs 'go mod tidy' and 'go mod download' in the plugin's directory.
func InstallDependencies(dir string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &dependenciesError{Plugin: dir, msg: fmt.Sprintf("error running 'go mod tidy': %s", output)}
	}

	cmd = exec.Command("go", "mod", "download")
	cmd.Dir = dir
	output, err = cmd.CombinedOutput()
	if err != nil {
		return &dependenciesError{Plugin: dir, msg: fmt.Sprintf("error running 'go mod download': %s", output)}
	}

	return nil
}

// LoadPlugin loads a plugin from the specified location.
// It checks if the plugin is already loaded in the cache and returns it if found.
func LoadPlugin[T any](pluginLocation string) (T, error) {
	var zero T

	// Check if the plugin is already loaded
	dir_hash, err := cache.HashDirectory(filepath.Dir(pluginLocation))
	if err != nil {
		return zero, err
	}

	value, found := cache.PluginRegistry.Get(dir_hash)

	// If the plugin is already loaded, return it
	if found {
		return value.(T), nil
	}

	// Grab plugin type
	var contract = reflect.TypeOf((*T)(nil)).Elem()

	// Create a new interpreter
	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)

	// Load the plugin
	plugin_var, err := i.EvalPath(pluginLocation) // runs the plugin's main function

	if err != nil {
		return zero, fmt.Errorf("error loading plugin %s: %v", pluginLocation, err)
	}

	plugin := plugin_var.Interface()

	if !reflect.TypeOf(plugin).Implements(contract) {
		return zero, &ContractVerifyError{
			Contract: contract.String(),
			Plugin:   pluginLocation,
			msg:      fmt.Sprintf("plugin %s does not implement the %s interface", plugin, contract.String()),
		}
	}

	// Cache the plugin
	cache.PluginRegistry.Set(dir_hash, plugin)

	return plugin.(T), nil
}

// LoadAllPlugins loads all plugins of the specified type from the specified folder.
func LoadPlugins[T any](pluginFolders []string) ([]T, error) {
	var plugins []T

	for _, folder := range pluginFolders {
		plugin, err := LoadPlugin[T](folder)
		// ignore contract mismatch errors but catch all other errors
		if _, ok := err.(*ContractVerifyError); !ok && err != nil {
			return nil, err
		}

		plugins = append(plugins, plugin)
	}

	return plugins, nil
}

func findGoModDirs(root string) ([]string, error) {
	var dirs []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Check if the current file is a directory and contains a go.mod file.
		if info.IsDir() {
			goModPath := filepath.Join(path, "go.mod")
			if _, err := os.Stat(goModPath); err == nil {
				// go.mod exists in this directory, add it to the list.
				dirs = append(dirs, path)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return dirs, nil
}

// LoadPluginsRecursive loads all plugins of the specified type from the specified folders and their subfolders.
func LoadPluginsRecursive[T any](pluginFolders []string) ([]T, error) {
	var plugins []T
	// under the plugin directories, recursively get all folders that have a go.mod file
	// make a new one, lazy bones
	for _, folder := range pluginFolders {
		// get all directories with a go.mod file
		dirs, err := findGoModDirs(folder)
		if err != nil {
			return nil, err
		}

		// load all plugins in each directory
		for _, dir := range dirs {
			plugin, err := LoadPlugin[T](dir)
			// ignore contract mismatch errors but catch all other errors
			if _, ok := err.(*ContractVerifyError); !ok && err != nil {
				return nil, err
			}

			plugins = append(plugins, plugin)
		}
	}
	return plugins, nil
}
