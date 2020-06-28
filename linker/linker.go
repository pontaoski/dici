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

type errorer struct {
	e error
}

func (e errorer) Error() error {
	return e.e
}

func (e errorer) Call(f func() error) {
	if e.e != nil {
		return
	}
	e.e = f()
}

func mkdirp(path string) (string, error) {
	err := os.MkdirAll(path, 0655)
	return path, err
}

func mount(file string, to string) error {
	if _, err := mkdirp(to); err != nil {
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
	cmd.Env = []string{"LANGUAGE=C"}
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
	if _, err := mkdirp(work); err != nil {
		return err
	}
	if _, err := mkdirp(upper); err != nil {
		return err
	}
	cmd := exec.Command("mount", "-o", fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", strings.Join(lowers, ":"), upper, work), root)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to overlay: %s\n%w", string(out), err)
	}
	return nil
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
			continue
		} else if os.IsNotExist(err) {
			continue
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
		if _, err := mkdirp(path.Join(outDirectory, root)); err != nil {
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

	etcDirWork, err := mkdirp(path.Join(mountDirectory, "etc"))
	if err != nil {
		return err
	}
	usrDirWork, err := mkdirp(path.Join(mountDirectory, "usr"))
	if err != nil {
		return err
	}
	varDirWork, err := mkdirp(path.Join(mountDirectory, "var"))
	if err != nil {
		return err
	}
	optDirWork, err := mkdirp(path.Join(mountDirectory, "opt"))
	if err != nil {
		return err
	}
	etcDirUpper, err := mkdirp(path.Join(mountDirectory, "etc-upper"))
	if err != nil {
		return err
	}
	usrDirUpper, err := mkdirp(path.Join(mountDirectory, "usr-upper"))
	if err != nil {
		return err
	}
	varDirUpper, err := mkdirp(path.Join(mountDirectory, "var-upper"))
	if err != nil {
		return err
	}
	optDirUpper, err := mkdirp(path.Join(mountDirectory, "opt-upper"))
	if err != nil {
		return err
	}
	var e errorer
	e.Call(func() error { return overlay(path.Join(outDirectory, "etc"), etcDirs, etcDirWork, etcDirUpper) })
	e.Call(func() error { return overlay(path.Join(outDirectory, "var"), varDirs, varDirWork, varDirUpper) })
	e.Call(func() error { return overlay(path.Join(outDirectory, "opt"), optDirs, optDirWork, optDirUpper) })
	e.Call(func() error { return overlay(path.Join(outDirectory, "usr"), usrDirs, usrDirWork, usrDirUpper) })
	if e.Error() != nil {
		return e.Error()
	}
	return nil
}
