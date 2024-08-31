// Code generated for package bindata by go-bindata DO NOT EDIT. (@generated)
// sources:
// internal/assets/cli-templates/component_config.yaml
package bindata

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _internalAssetsCliTemplatesComponent_configYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x74\xce\x31\xeb\xc2\x30\x10\x05\xf0\xbd\x9f\xe2\xc8\x7f\xce\xd1\x7f\x26\xc9\xe6\xd0\x41\x10\x05\x5b\x5c\xcb\xb5\x3d\x6c\xa1\xcd\x85\x4b\xc8\xe7\x17\xc5\x41\x44\xd7\xf7\x7e\x3c\xde\x1f\x1c\xf7\x5d\xd3\x76\xfd\xb5\xb9\xb4\x87\xf3\xa9\x1a\x65\x8b\x12\x38\xe4\xe4\x2b\x80\x91\x35\xdb\x8d\x02\xdd\x58\x3d\x18\x87\x0e\x5d\xaf\xbc\x32\x25\x36\x15\xc0\x2a\x23\xad\x36\x52\x9e\x6d\x54\x29\x4b\x5a\x24\x3c\x65\xa9\xb1\x46\xb7\xfb\x6d\xec\xc4\xc5\x83\xd9\x28\x65\xd6\x07\x93\xc1\x4e\x94\xe6\x41\x48\x27\x0f\xa6\x46\x87\xff\xaf\x5c\x22\x2b\x65\xf9\xfa\xe0\xad\xfe\x98\xbc\x07\x00\x00\xff\xff\x18\xff\xc3\xfa\xdb\x00\x00\x00")

func internalAssetsCliTemplatesComponent_configYamlBytes() ([]byte, error) {
	return bindataRead(
		_internalAssetsCliTemplatesComponent_configYaml,
		"internal/assets/cli-templates/component_config.yaml",
	)
}

func internalAssetsCliTemplatesComponent_configYaml() (*asset, error) {
	bytes, err := internalAssetsCliTemplatesComponent_configYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "internal/assets/cli-templates/component_config.yaml", size: 219, mode: os.FileMode(436), modTime: time.Unix(1725072637, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"internal/assets/cli-templates/component_config.yaml": internalAssetsCliTemplatesComponent_configYaml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//
//	data/
//	  foo.txt
//	  img/
//	    a.png
//	    b.png
//
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"internal": &bintree{nil, map[string]*bintree{
		"assets": &bintree{nil, map[string]*bintree{
			"cli-templates": &bintree{nil, map[string]*bintree{
				"component_config.yaml": &bintree{internalAssetsCliTemplatesComponent_configYaml, map[string]*bintree{}},
			}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
