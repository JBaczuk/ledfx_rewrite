package constants

import (
	"os"
	"path/filepath"
	"runtime"
)

var CONFIG_DIR = ".ledfx"

func GetOsConfigDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return filepath.Join(os.Getenv("HOME"), CONFIG_DIR)
}