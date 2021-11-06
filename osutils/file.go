package osutils

import "os"

func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

func FileNotExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return true
	}
	return false
}

func IsDir(path string) bool {
	if s, err := os.Stat(path); err == nil {
		return s.IsDir()
	}
	return false
}
