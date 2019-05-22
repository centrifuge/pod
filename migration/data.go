// Code generated by go-bindata.
// sources:
// files/00Initial.go
// DO NOT EDIT!

package migration

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

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _migrationFiles00initialGo = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x4c\x8f\x4d\x6a\xe3\x40\x10\x85\xd7\xee\x53\x3c\x7a\x25\x0d\x58\xf2\x05\x66\x31\x83\x37\x86\x78\x93\x1b\xb4\xa4\x52\xa9\x70\xbb\xca\xe9\x1f\x07\x13\x72\xf7\x60\x45\x90\xac\x0a\xea\x7b\xbc\x9f\x5b\x18\x2f\x81\x09\xb3\x44\xca\xce\xf5\x3d\xfe\xe9\x03\xe3\x12\x94\x29\xa3\x18\xca\x22\x79\xa5\x48\xf4\x56\x25\x7d\x7f\x99\x94\x52\x28\x84\xc0\x41\x14\x6c\x98\x42\x09\x18\x44\x27\x51\xce\x08\x19\x79\xb1\x77\x85\x28\xce\xe1\x42\x4f\x03\xe7\xe4\x7a\xb3\x54\xd0\xb8\x5d\x34\x66\x51\x86\x67\x29\x4b\x1d\xba\xd1\xae\xbd\xdc\xe6\xdc\xb3\xed\xa3\xb1\x77\xbb\xdf\x24\x3f\x74\x2a\xa9\x67\x8b\x74\xa7\x38\x0d\xfd\x76\xbd\x6b\x9d\xbb\x87\x84\x68\x8c\xbf\xd8\x4c\xbb\x17\x63\xa6\xd4\xf8\xab\xf0\xb3\xe3\x7e\xdd\xe6\xdb\x75\xdd\x49\xa5\x48\x88\x87\x03\x8e\x46\x19\x6a\x65\x11\x65\x37\x57\x1d\x7f\x58\x33\x0d\xf8\xb3\x65\x74\xc7\xff\x2d\x28\x25\x4b\xf8\x58\x7b\x77\x27\x9d\x6d\x6e\xfc\x61\x93\xe3\xbc\xc6\x88\x29\x5e\xab\x22\xd7\x71\xa4\x9c\xe7\x1a\xe3\xc3\xb7\x6e\x97\xa8\xd4\xa4\x50\x89\xee\xd3\x7d\x05\x00\x00\xff\xff\x1b\x0e\xa0\x8c\x6e\x01\x00\x00")

func migrationFiles00initialGoBytes() ([]byte, error) {
	return bindataRead(
		_migrationFiles00initialGo,
		"migration/files/00Initial.go",
	)
}

func migrationFiles00initialGo() (*asset, error) {
	bytes, err := migrationFiles00initialGoBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "migration/files/00Initial.go", size: 366, mode: os.FileMode(420), modTime: time.Unix(1558548238, 0)}
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
	"migration/files/00Initial.go": migrationFiles00initialGo,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
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
	"migration": &bintree{nil, map[string]*bintree{
		"files": &bintree{nil, map[string]*bintree{
			"00Initial.go": &bintree{migrationFiles00initialGo, map[string]*bintree{}},
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
