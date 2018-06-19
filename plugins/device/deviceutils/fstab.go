package deviceutils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"syscall"
)

// FstabMount represents entry in Fstab file
type FstabMount struct {
	Device           string
	MountPoint       string
	FilesystemFormat string
	MountOptions     string
	DumpValue        string
	FsckOption       string
}

// Fstab represents Fstab entries
type Fstab struct {
	Filename string
	Mounts   []FstabMount
}

func lock(filename string) (*os.File, error) {
	lockfile, err := os.Create(filename + ".lock")
	if err != nil {
		return nil, err
	}

	err = syscall.Flock(int(lockfile.Fd()), syscall.LOCK_EX)
	if err != nil {
		return nil, err
	}
	return lockfile, nil
}

func (f *Fstab) save() error {
	data := "\n"
	for _, mnt := range f.Mounts {
		data += fmt.Sprintf("%s %s %s %s %s %s\n",
			mnt.Device,
			mnt.MountPoint,
			mnt.FilesystemFormat,
			mnt.MountOptions,
			mnt.DumpValue,
			mnt.FsckOption,
		)
	}
	err := ioutil.WriteFile(f.Filename+".tmp", []byte(data), 0600)
	if err != nil {
		return err
	}
	return os.Rename(f.Filename+".tmp", f.Filename)
}

func (f *Fstab) load() error {
	out, err := ioutil.ReadFile(f.Filename)
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		var parts []string
		for _, p := range strings.Split(line, " ") {
			if p != "" {
				parts = append(parts, p)
			}
		}

		if len(parts) != 6 {
			return errors.New("invalid mount entry: " + line)
		}
		f.Mounts = append(f.Mounts, FstabMount{
			Device:           parts[0],
			MountPoint:       parts[1],
			FilesystemFormat: parts[2],
			MountOptions:     parts[3],
			DumpValue:        parts[4],
			FsckOption:       parts[5],
		})
	}
	return nil
}

func (f *Fstab) mountExists(mountpoint string) bool {
	for _, mnt := range f.Mounts {
		if mnt.MountPoint == mountpoint {
			return true
		}
	}
	return false
}

// FstabAddMount adds mount point entry to fstab
func FstabAddMount(filename string, mnt FstabMount) error {
	fstab := Fstab{Filename: filename}

	file, err := lock(fstab.Filename)
	if err != nil {
		return err
	}

	defer file.Close()

	err = fstab.load()
	if err != nil {
		return err
	}

	if mnt.FilesystemFormat == "" {
		mnt.FilesystemFormat = "xfs"
	}

	if mnt.DumpValue == "" {
		mnt.DumpValue = "0"
	}

	if mnt.FsckOption == "" {
		mnt.FsckOption = "0"
	}

	if mnt.MountOptions == "" {
		mnt.MountOptions = "defaults"
	}

	if !fstab.mountExists(mnt.MountPoint) {
		fstab.Mounts = append(fstab.Mounts, mnt)
	}
	return fstab.save()
}

// FstabRemoveMount removes mountpoint from fstab
func FstabRemoveMount(filename, mountpoint string) error {
	fstab := Fstab{Filename: filename}
	file, err := lock(fstab.Filename)
	if err != nil {
		return err
	}

	defer file.Close()

	err = fstab.load()
	if err != nil {
		return err
	}

	var mounts []FstabMount
	for _, mnt := range fstab.Mounts {
		if mnt.MountPoint != mountpoint {
			mounts = append(mounts, mnt)
		}
	}
	fstab.Mounts = mounts

	return fstab.save()
}
