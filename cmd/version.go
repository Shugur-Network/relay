package main

import "fmt"

// GetVersion returns the current version information
func GetVersion() string {
	return version
}

// GetFullVersionInfo returns detailed version information
func GetFullVersionInfo() string {
	return fmt.Sprintf("Version: %s\nCommit: %s\nBuilt: %s", version, commit, date)
}

// GetVersionWithPrefix returns version with "shugur relay version: " prefix
func GetVersionWithPrefix() string {
	return fmt.Sprintf("shugur relay version: %s", version)
}
