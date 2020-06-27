package builder

import (
	"dici/builder/xattr"
	"dici/pkg"
	"os"
	"os/exec"
)

// CompressPackageError describes why a package failed to compress
type CompressPackageError string

func (c CompressPackageError) Error() string {
	return string(c)
}

const (
	// NotADirectory - The passed fileInfo isn't a directory
	NotADirectory CompressPackageError = "The passed fileInfo isn't a directory"
)

func squashPackage(input string, output string) error {
	cmd := exec.Command("mksquashfs", input, output, "-noappend", "-comp", "gzip", "-no-fragments", "-no-progress")
	err := cmd.Run()
	return err
}

// SquashPackage compresses a directory and produces a package file
func SquashPackage(input string, output string, p pkg.PackageMetadata) error {
	fi, err := os.Stat(input)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return NotADirectory
	}
	err = squashPackage(input, output)
	if err != nil {
		return err
	}
	return xattr.ApplyXattrs(output, p)
}
