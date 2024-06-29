package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type HashError struct {
	Err error
}

func (e *HashError) Error() string {
	return e.Err.Error()
}

func HashArgs(args ...interface{}) string {
	hasher := md5.New()
	for _, arg := range args {
		jsonArg, _ := json.Marshal(arg)
		hasher.Write(jsonArg)
	}
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// hashDirectory computes the MD5 hash of a directory structure and all of its files, nested directories, and their files.
func HashDirectory(dir string) (string, error) {
	hasher := md5.New()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return &HashError{Err: err}
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return &HashError{Err: err}
		}
		defer file.Close()

		_, err = io.Copy(hasher, file)
		if err != nil {
			return &HashError{Err: err}
		}

		return nil
	})
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
