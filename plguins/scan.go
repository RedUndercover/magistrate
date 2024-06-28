package plugin

import (
	"fmt"
	"path/filepath"
	"reflect"
)

func LoadPlugin[T any](goFilePath, infoFilePath string, pluginInterface reflect.Type) (T, error) {
	var zero T
	pluginInfo, err := LoadPluginInfo(infoFilePath)
	if err != nil {
		return zero, fmt.Errorf("failed to load plugin info: %w", err)
	}

	err = InstallDependencies(filepath.Dir(goFilePath))
	if err != nil {
		return zero, fmt.Errorf("failed to install dependencies: %w", err)
	}

	i := interp.New(interp.Options{})
	i.Use(stdlib.Symbols)

	_, err = i.EvalPath(goFilePath)
	if err != nil {
		return zero, fmt.Errorf("failed to evaluate path: %w", err)
	}

	v, err := i.Eval("main." + pluginInfo.MainSymbol)
	if err != nil {
		return zero, fmt.Errorf("failed to evaluate main symbol: %w", err)
	}

	plugin := v.Interface()
	if !reflect.TypeOf(plugin).Implements(pluginInterface) {
		return zero, fmt.Errorf("loaded type does not implement the provided plugin interface")
	}

	return plugin.(T), nil
}
