package xlator

/*
#cgo LDFLAGS: -ldl

#include <stdlib.h>    // free()
#include <dlfcn.h>     // dlopen(), dlclose()
#include "options.h"   // volume_option_t
*/
import "C"

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unsafe"
)

const (
	maxXlatorOptions = 100
)

func structifyOption(option *C.volume_option_t) Option {
	var x Option

	for _, k := range option.key {
		if k != nil {
			x.Key = append(x.Key, C.GoString(k))
		}
	}

	for _, k := range option.value {
		if k != nil {
			x.Value = append(x.Value, C.GoString(k))
		}
	}

	x.Type = OptionType(option.otype)
	x.Min = float64(option.min)
	x.Max = float64(option.max)
	x.DefaultValue = C.GoString(option.default_value)
	x.Description = C.GoString(option.description)
	x.Validate = OptionValidateType(option.validate)

	return x
}

func loadXlatorOptions(xlator string) ([]Option, error) {

	csXlator := C.CString(xlator)
	defer C.free(unsafe.Pointer(csXlator))

	handle := C.dlopen(csXlator, C.RTLD_LAZY|C.RTLD_LOCAL)
	if handle == nil {
		return nil, fmt.Errorf("dlopen(%s) failed; dlerror = %s",
			xlator, C.GoString((*C.char)(C.dlerror())))
	}
	defer C.dlclose(handle)

	csSym := C.CString("options")
	defer C.free(unsafe.Pointer(csSym))

	p := C.dlsym(handle, csSym)
	if p == nil {
		// .so is not an xlator
		return nil, nil
	}

	xlatorOptions := (*[maxXlatorOptions]C.volume_option_t)(p)
	var vopts []Option
	for _, option := range xlatorOptions {

		// identify sentinel NULL key which marks the end of options
		if option.key[0] == nil {
			break
		}

		// &option i.e *C.volume_option_t still points to an address
		// in memory where that symbol resides as mmap()ed by the call
		// to dlsym(). We need to copy the contents of that C structure
		// to its equivalent Go struct before dlclose() happens.
		vopts = append(vopts, structifyOption(&option))
	}

	return vopts, nil
}

func getXlatorsDir() string {

	// glusterfs gets the path to xlator dir from a compile time flag named
	// 'XLATORDIR' which gets passed through a -D flag to GCC. This isn't
	// available to external programs via gluster CLI yet. When one or more
	// versions of gluster are installed from source or otherwise, the
	// following is the most fool-proof but hacky way to get the xlator dir
	// location without making assumptions.

	cmd := "strings -d `which glusterfsd` | awk '/glusterfs\\/.*\\/xlator$/'"
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}

func getAllXlatorOptions() (map[string][]Option, error) {

	xlatorsDir := getXlatorsDir()
	s, err := os.Stat(xlatorsDir)
	if err != nil {
		return nil, err
	}
	if !s.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", xlatorsDir)
	}

	xlatorsOpMap := make(map[string][]Option)

	// NOTE: The following shared objects are symlinks and hence duplicated
	// disperse.so -> ec.so
	// distribute.so -> dht.so
	// replicate.so -> afr.so
	// stat-prefetch.so -> md-cache.so
	// posix-locks.so -> locks.so
	// access-control.so -> ../system/posix-acl.so

	actor := func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".so") {
			xopts, err := loadXlatorOptions(path)
			if err != nil {
				// Can also skip over instead. Should we ?
				return err
			}
			if xopts != nil {
				// only include xlators that does have options
				xlatorsOpMap[path[len(xlatorsDir)+1:len(path)-len(".so")]] = xopts
				// example of key usage: xlatorsOpMap["features/upcall"]
			}
		}
		return nil
	}

	if err := filepath.Walk(xlatorsDir, actor); err != nil {
		return nil, err
	}

	return xlatorsOpMap, nil
}
