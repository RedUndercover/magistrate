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

func HashArgs(args ...interface{}) string {
	hasher := md5.New()
	for _, arg := range args {
		jsonArg, _ := json.Marshal(arg)
		hasher.Write(jsonArg)
	}
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// hashDirectory computes the MD5 hash of all files in a directory
func hashDirectory(dirPath string) (string, error) {
	hash := md5.New()

	// walk the directory and hash all files
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(hash, file); err != nil {
				return err
			}
		}
		return nil
	})

	// check for errors
	if err != nil {
		return "", err
	}

	// return the hash as a string
	hashInBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashInBytes)
	return hashString, nil
}
