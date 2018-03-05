package xlator

/*
#cgo LDFLAGS: -ldl

#include <stdlib.h>    // free()
#include <dlfcn.h>     // dlopen(), dlsym(), dlclose(), dlerror()
#include "xlator.h"    // xlator_api_t, volume_option_t
*/
import "C"

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/gluster/glusterd2/glusterd2/xlator/options"
)

const (
	maxOptions = 100
)

// structifyOption creates an options.Option from the given C.volume_option_t
func structifyOption(cOpt *C.volume_option_t) *options.Option {
	var opt options.Option

	for _, k := range cOpt.key {
		if k != nil {
			opt.Key = append(opt.Key, C.GoString(k))
		}
	}

	for _, k := range cOpt.value {
		if k != nil {
			opt.Value = append(opt.Value, C.GoString(k))
		}
	}

	for _, k := range cOpt.op_version {
		opt.OpVersion = append(opt.OpVersion, uint32(k))
	}

	for _, k := range cOpt.deprecated {
		opt.Deprecated = append(opt.Deprecated, uint32(k))
	}

	for _, k := range cOpt.tags {
		if k != nil {
			opt.Tags = append(opt.Tags, C.GoString(k))
		}
	}

	opt.Type = options.OptionType(cOpt.otype)
	opt.Min = float64(cOpt.min)
	opt.Max = float64(cOpt.max)
	opt.DefaultValue = C.GoString(cOpt.default_value)
	opt.Description = C.GoString(cOpt.description)
	opt.ValidateType = options.OptionValidateType(cOpt.validate)
	opt.Flags = uint32(cOpt.flags)
	opt.SetKey = C.GoString(cOpt.setkey)
	opt.Level = options.OptionLevel(cOpt.level)

	return &opt
}

// loadXlator loads the xlator at the given path and returns a Xlator
func loadXlator(xlPath string) (*Xlator, error) {

	cXlPath := C.CString(xlPath)
	defer C.free(unsafe.Pointer(cXlPath))

	handle := C.dlopen(cXlPath, C.RTLD_LAZY|C.RTLD_LOCAL)
	if handle == nil {
		return nil, fmt.Errorf("dlopen(%s) failed; dlerror = %s",
			xlPath, C.GoString((*C.char)(C.dlerror())))
	}
	defer C.dlclose(handle)

	xl := new(Xlator)
	xl.ID = strings.TrimSuffix(filepath.Base(xlPath), filepath.Ext(xlPath))

	xlSym := C.CString("xlator_api")
	defer C.free(unsafe.Pointer(xlSym))

	p := C.dlsym(handle, xlSym)
	if p != nil {
		xp := (*C.xlator_api_t)(p)
		// FIXME: It's named "server-protocol" instead of "server" in server.so
		//        https://review.gluster.org/18879
		// xl.ID = C.GoString(xp.identifier)
		xl.rawID = uint32(xp.xlator_id)
		xl.Flags = uint32(xp.flags)
		for _, k := range xp.op_version {
			xl.OpVersion = append(xl.OpVersion, uint32(k))
		}
		p = unsafe.Pointer(xp.options)
	} else {
		optsSym := C.CString("options")
		defer C.free(unsafe.Pointer(optsSym))
		p = C.dlsym(handle, optsSym)
		if p == nil {
			return xl, nil
		}
	}

	soOptions := (*[maxOptions]C.volume_option_t)(p)
	for _, option := range soOptions {

		// identify sentinel NULL key which marks the end of options
		if option.key[0] == nil {
			break
		}

		// &option i.e *C.volume_option_t still points to an address
		// in memory where that symbol resides as mmap()ed by the call
		// to dlsym(). We need to copy the contents of that C structure
		// to its equivalent Go struct before dlclose() happens.
		xl.Options = append(xl.Options, structifyOption(&option))
	}

	return xl, nil
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

// loadAllXlators loads available xlators and returns a map of Xlators indexed
// by Xlator.ID
func loadAllXlators() (map[string]*Xlator, error) {

	xlatorsDir := getXlatorsDir()

	if xlatorsDir == "" {
		return nil, fmt.Errorf("No xlators dir found")
	}

	s, err := os.Stat(xlatorsDir)
	if err != nil {
		return nil, err
	}
	if !s.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", xlatorsDir)
	}
	xlatorsParentDir := filepath.Dir(xlatorsDir)

	xlMap := make(map[string]*Xlator)

	// NOTE: The following shared objects are symlinks and hence duplicated
	// disperse.so -> ec.so
	// distribute.so -> dht.so
	// replicate.so -> afr.so
	// stat-prefetch.so -> md-cache.so
	// posix-locks.so -> locks.so
	// access-control.so -> ../system/posix-acl.so

	actor := func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".so") {
			xl, err := loadXlator(path)
			if err != nil {
				return err
			}
			xlMap[xl.ID] = xl
		}
		return nil
	}

	if err := filepath.Walk(xlatorsParentDir, actor); err != nil {
		return nil, err
	}

	return xlMap, nil
}
