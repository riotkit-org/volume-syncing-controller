package helpers

import (
	"io"
	"os"
)

func GetEnvOrDefault(name string, defaultValue interface{}) interface{} {
	value, exists := os.LookupEnv(name)
	if !exists {
		return defaultValue
	}
	if value == "1" || value == "true" || value == "TRUE" || value == "yes" || value == "YES" || value == "Y" || value == "y" {
		return true
	}
	if value == "0" || value == "false" || value == "FALSE" || value == "no" || value == "NO" || value == "N" || value == "n" {
		return false
	}
	return value
}

// IsLocalDirEmpty: https://stackoverflow.com/a/30708914/6782994
func IsLocalDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}
