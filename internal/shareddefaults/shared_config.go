package shareddefaults

import (
	"os"
	"os/user"
	"path/filepath"
)

// SharedCredentialsFilename returns the SDK's default file path
// for the shared credentials file.
//
// Builds the shared config file path based on the OS's platform.
//
//   - Linux/Unix: $HOME/.cybr/credentials
//   - Windows: %USERPROFILE%\.cybr\credentials
func SharedCredentialsFilename() string {
	return filepath.Join(UserHomeDir(), ".cybr", "credentials")
}

// SharedConfigFilename returns the SDK's default file path for
// the shared config file.
//
// Builds the shared config file path based on the OS's platform.
//
//   - Linux/Unix: $HOME/.cybr/config
//   - Windows: %USERPROFILE%\.cybr\config
func SharedConfigFilename() string {
	return filepath.Join(UserHomeDir(), ".cybr", "config")
}

// UserHomeDir returns the home directory for the user the process is
// running under.
func UserHomeDir() string {
	// Ignore errors since we only care about Windows and *nix.
	home, _ := os.UserHomeDir()

	if len(home) > 0 {
		return home
	}

	currUser, _ := user.Current()
	if currUser != nil {
		home = currUser.HomeDir
	}

	return home
}
