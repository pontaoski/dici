package linker

import (
	"dici/builder/xattr"
	"dici/pkg"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

var roots = []string{"etc", "usr", "var", "opt"}

// LinkError describes why links could not be built
type LinkError string

func (l LinkError) Error() string {
	return string(l)
}

const (
	// NotADirectory - A passed argument isn't a directory
	NotADirectory LinkError = "The passed fileInfo isn't a directory"
)

func mkdirp(path string) error {
	err := os.MkdirAll(path, 0655)
	return err
}

func mount(file string, to string) error {
	if err := mkdirp(to); err != nil {
		return err
	}
	cmd := exec.Command("mount", file, to)
	if data, err := cmd.CombinedOutput(); err != nil {
		println(string(data))
		return err
	}
	return nil
}

func unmount(place string) error {
	cmd := exec.Command("umount", "-l", place)
	if data, err := cmd.CombinedOutput(); err != nil {
		if strings.Contains(string(data), "not mounted") {
			return nil
		}
		return err
	}
	return nil
}

func exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func overlay(root string, lowers []string, work, upper string) error {
	if len(lowers) == 0 {
		return nil
	}
	lowers = append([]string{root}, lowers...)
	if err := mkdirp(work); err != nil {
		return err
	}
	if err := mkdirp(upper); err != nil {
		return err
	}
	cmd := exec.Command("mount", "-o", fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", strings.Join(lowers, ":"), upper, work), root)
	return cmd.Run()
}

func checkpackage(mountDir string, p pkg.Package) (etc bool, usr bool, Var bool, opt bool, etcs string, usrs string, Vars string, opts string) {
	return exists(path.Join(mountDir, p.NV(), "etc")),
		exists(path.Join(mountDir, p.NV(), "usr")),
		exists(path.Join(mountDir, p.NV(), "var")),
		exists(path.Join(mountDir, p.NV(), "opt")),
		path.Join(mountDir, p.NV(), "etc"),
		path.Join(mountDir, p.NV(), "usr"),
		path.Join(mountDir, p.NV(), "var"),
		path.Join(mountDir, p.NV(), "opt")
}

// Unlink unlinks packages from the mountDirectory and the outDirectory
func Unlink(mountDirectory, outDirectory string) error {
	if fi, err := os.Stat(mountDirectory); err != nil {
		return err
	} else {
		if !fi.IsDir() {
			return NotADirectory
		}
	}
	if fi, err := os.Stat(outDirectory); err != nil {
		return err
	} else {
		if !fi.IsDir() {
			return NotADirectory
		}
	}
	fis, err := ioutil.ReadDir(mountDirectory)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		if !fi.IsDir() {
			continue
		}
		err := unmount(path.Join(mountDirectory, fi.Name()))
		if err != nil {
			return err
		}
		err = os.RemoveAll(path.Join(mountDirectory, fi.Name()))
		if err != nil {
			return err
		}
	}
	outFis, err := ioutil.ReadDir(outDirectory)
	if err != nil {
		return err
	}
	for _, fi := range outFis {
		if _, err := os.Lstat(path.Join(outDirectory, fi.Name())); err == nil {
			err := os.Remove(path.Join(outDirectory, fi.Name()))
			if err != nil {
				return err
			}
		} else if os.IsNotExist(err) {
			return err
		}
		switch {
		case fi.IsDir():
			err = os.RemoveAll(path.Join(outDirectory, fi.Name()))
		case !fi.IsDir():
			err = os.Remove(path.Join(outDirectory, fi.Name()))
		default:
			panic("unhandled case")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// Link links packages from the packageDirectory to the outDirectory
func Link(packageDirectory, mountDirectory, outDirectory string) error {
	if fi, err := os.Stat(packageDirectory); err != nil {
		return err
	} else {
		if !fi.IsDir() {
			return NotADirectory
		}
	}
	if fi, err := os.Stat(mountDirectory); err != nil {
		return err
	} else {
		if !fi.IsDir() {
			return NotADirectory
		}
	}
	if fi, err := os.Stat(outDirectory); err != nil {
		return err
	} else {
		if !fi.IsDir() {
			return NotADirectory
		}
	}
	err := Unlink(mountDirectory, outDirectory)
	if err != nil {
		return err
	}
	fis, err := ioutil.ReadDir(packageDirectory)
	if err != nil {
		return err
	}
	var pkgs []pkg.Package
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		pkgPath := path.Join(packageDirectory, fi.Name())
		var data pkg.PackageMetadata
		err := xattr.ReadXattrs(pkgPath, &data)
		pkgs = append(pkgs, pkg.Package{data, pkgPath})
		if err != nil {
			return err
		}
		if exists(path.Join(mountDirectory, data.NV())) {
			continue
		}
		err = mount(pkgPath, path.Join(mountDirectory, data.NV()))
		if err != nil {
			return err
		}
		mnt, err := filepath.Abs(path.Join(mountDirectory, data.NV()))
		if err != nil {
			return err
		}
		out, err := filepath.Abs(path.Join(outDirectory, data.NV()))
		if err != nil {
			return err
		}
		err = os.Symlink(mnt, out)
		if err != nil {
			return err
		}
	}
	for _, root := range roots {
		if err := mkdirp(path.Join(outDirectory, root)); err != nil {
			return err
		}
	}
	var etcDirs []string
	var usrDirs []string
	var varDirs []string
	var optDirs []string
	for _, pkg := range pkgs {
		etc, usr, Var, opt, etcs, usrs, Vars, opts := checkpackage(mountDirectory, pkg)
		if etc {
			etcDirs = append(etcDirs, etcs)
		}
		if usr {
			usrDirs = append(usrDirs, usrs)
		}
		if Var {
			varDirs = append(varDirs, Vars)
		}
		if opt {
			optDirs = append(optDirs, opts)
		}
	}
	return nil
}
