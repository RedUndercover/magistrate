package plugin

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"

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

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	return err
}

// Recursively copy a directory from src to dst
func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relativePath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destinationPath := filepath.Join(dst, relativePath)

		if info.IsDir() {
			return os.MkdirAll(destinationPath, info.Mode())
		}

		return copyFile(path, destinationPath)
	})
}

func InstallDependencies(dir string, gopath string) error {
	// create the gopath directory if it doesn't exist
	if _, err := os.Stat(gopath); os.IsNotExist(err) {
		if err := os.MkdirAll(gopath, 0755); err != nil {
			return err
		}
	}

	// run 'go mod tidy' and 'go mod download' in the plugin's directory
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

	// run 'go mod vendor' in the plugin's directory
	cmd = exec.Command("go", "mod", "vendor")
	cmd.Dir = dir
	output, err = cmd.CombinedOutput()
	if err != nil {
		return &dependenciesError{Plugin: dir, msg: fmt.Sprintf("error running 'go mod vendor': %s", output)}
	}

	// copy the vendor directory to the gopath if vendor directory exists
	_, err = os.Stat(filepath.Join(dir, "vendor"))
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return &dependenciesError{Plugin: dir, msg: fmt.Sprintf("error checking for vendor directory: %s", err)}
	} else {
		// copy the vendor directory to the gopath
		err = copyDirectory(filepath.Join(dir, "vendor"), filepath.Join(gopath, "src"))
		if err != nil {
			return &dependenciesError{Plugin: dir, msg: fmt.Sprintf("error copying vendor directory: %s", err)}
		}
		return nil
	}
}

// LoadPlugin loads a plugin from the specified location.
// It checks if the plugin is already loaded in the cache and returns it if found.
// Otherwise, it loads the plugin and caches it.
func LoadPlugin[T any](pluginLocation string, gopath string) (T, error) {
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

	// Install the plugin's dependencies
	err = InstallDependencies(filepath.Dir(pluginLocation), gopath)
	if err != nil {
		return zero, err
	}

	// Create a new interpreter
	i := interp.New(interp.Options{GoPath: gopath})
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
func LoadPlugins[T any](pluginFolders []string, gopath string) ([]T, error) {
	var plugins []T

	for _, folder := range pluginFolders {
		plugin, err := LoadPlugin[T](folder, gopath)
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
func LoadPluginsRecursive[T any](pluginFolders []string, gopath string) ([]T, error) {
	var plugins []T
	var valid_directories []string
	// under the plugin directories, recursively get all folders that have a go.mod file
	// make a new one, lazy bones
	for _, folder := range pluginFolders {
		// get all directories with a go.mod file
		dirs, err := findGoModDirs(folder)
		if err != nil {
			return nil, err
		}
		valid_directories = append(valid_directories, dirs...)
	}

	// load all plugins in each directory
	for _, dir := range valid_directories {
		plugin, err := LoadPlugin[T](dir, gopath)
		// ignore contract mismatch errors but catch all other errors
		if _, ok := err.(*ContractVerifyError); !ok && err != nil {
			return nil, err
		}

		plugins = append(plugins, plugin)
	}
	return plugins, nil
}
