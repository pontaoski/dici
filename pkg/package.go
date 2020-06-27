package pkg

import "fmt"

// PackageMetadata describes the metadata associated with a package
type PackageMetadata struct {
	Name        string `xattr:"name"`
	Description string `xattr:"description"`
	Version     string `xattr:"version"`
}

// NV gets a name-version string naming the package
func (m PackageMetadata) NV() string {
	return fmt.Sprintf("%s-%s", m.Name, m.Version)
}

// Package stores data about a package
type Package struct {
	PackageMetadata
	Location string
}
