// Code generated by go-bindata.
// sources:
// ../../resources/default_config.yaml
// ../../resources/testing_config.yaml
// DO NOT EDIT!

package resources

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

var _resourcesDefault_configYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x95\x4b\x77\xe2\x38\x13\x86\xf7\xfc\x0a\x1d\xbe\xf5\x07\xba\xd8\xba\x78\x47\x77\x93\xbe\x84\x64\x80\xd0\xa1\xc3\x4e\x97\x32\xa8\xc1\x16\xb1\x65\x2e\xfd\xeb\xe7\x18\xc2\xf4\x4c\xce\x84\x9e\x33\xeb\xf1\xca\x56\xa9\xde\x92\xeb\x7d\x4e\xc9\x42\x19\x2b\x9f\x37\x4b\xb8\x87\xb8\x0f\xd5\x3a\x43\x3f\x97\xa6\x4d\x5d\x7b\x5d\xae\xfc\x66\x33\x8c\xab\xa9\x2f\xd7\x60\x8e\x9d\xf2\xbc\xb1\xce\x3a\x08\xfd\x0f\x8d\x82\xd5\x1b\x14\xa1\x8e\xbe\x5c\x22\x1b\xca\x58\x69\x1b\x91\x76\xae\x82\xba\x86\x1a\x95\x00\x0e\xc5\x80\x0c\xa0\x1a\x22\xda\xfb\xb8\x42\x50\xee\xd0\x4e\x57\x5e\x9b\x0d\xd4\xbd\x0e\xba\xe4\xb7\x92\x08\x79\x97\x21\xc6\xd8\xe9\x1d\xe2\x0a\x2a\x68\x8a\x97\xd3\x7d\x76\x19\x92\x4c\x9e\x63\x26\x84\x58\xc7\x4a\x6f\xc7\x00\x55\x7d\xce\x45\xe8\xff\xa8\xdb\xf7\xdb\xa4\x4f\xa8\xe8\xe1\x1e\xee\x91\x7e\xb4\xdb\x3e\x93\x14\xd3\xbe\xdf\xe6\x75\x7f\x52\xcc\x26\x07\xb3\x5f\x37\x8b\xa7\xa7\x0f\x79\xf3\x63\x66\x0e\xc3\xc1\x14\x66\xf7\xef\x47\xe1\xc7\xf1\x98\xa6\x72\x37\x29\x97\x8f\xbb\xf1\xdd\xf7\xd1\xd3\xba\xfb\x4b\x59\x76\x91\x7d\xcc\xf9\xf0\x9e\x17\xeb\xe7\x39\x7c\x9f\xdf\xce\xe9\xf3\xb8\x21\xfc\xdb\xd6\x7d\x64\xeb\x2f\x81\xcc\x58\xb1\xd2\xab\xf1\xbb\xf4\x01\xd2\x92\x9c\x65\x2f\xed\x1a\x5c\xba\x75\xf9\x09\xef\xa0\x8c\x3e\x1e\x6f\xb4\x8d\xa1\x3a\x66\xa8\xdb\x7d\x15\x99\xc2\xd2\xd7\xf1\x2f\x21\x5d\xda\x55\xa8\xae\x04\xb6\xa1\xf6\xaf\xe4\xb6\xfa\x58\x40\x19\x7f\x33\x1b\xbf\xd4\xd1\x87\xf2\x14\xeb\xa0\x37\x29\xf8\x5c\x46\x58\x56\xe7\xad\x7f\xb8\x45\x30\xf9\xcf\xad\x3f\xbb\x85\x0f\x82\x51\x46\x4c\xa2\x18\xe6\xda\x1a\xae\x0d\x16\x89\xc6\x82\x09\x99\x33\xc1\x73\xab\x8c\x33\x54\x10\x7e\xc5\x57\x7c\x70\x5c\x5a\x8e\x85\x31\x96\x62\x89\x31\xe6\x52\x31\x92\x60\x45\xb8\xe6\x92\x92\xdc\x09\xc2\x0c\x55\xf4\x4d\x02\xf0\xc1\xa4\x82\xe4\x90\x60\x07\xd4\x71\x02\x2a\x4d\x5c\xae\xac\x4e\x40\x71\x92\x73\x95\x48\x99\x13\x96\x6a\xfc\xaf\x81\x78\x19\x0b\x3f\x61\x48\xc9\x3f\xf0\x9b\xa5\x3d\x4a\xd3\x1e\xc5\xb8\x97\xd0\xd7\x9e\x13\xfa\x81\xdd\x86\x30\x1f\x79\x6f\x27\x8f\xfb\xd9\x6a\xf6\xee\x89\x1f\x6e\xed\x38\x8c\x72\x3e\x9d\x3c\x7d\xb9\xd9\xee\x73\x52\x89\x74\x3f\x3a\xd0\xc5\x94\x6d\xdf\x3b\xf2\xda\xf9\x97\x02\x92\xf7\x28\xc1\x6f\x15\x98\x2c\xee\x06\xf2\xe3\xf8\x53\xb5\x1b\x2e\xde\xa9\xbd\x5b\x87\xaf\x76\x30\x28\xde\x2f\x3e\x6d\x15\x1c\x8f\x8b\xe4\x61\x28\x97\x37\x15\x5b\xcd\xee\xbf\x9d\x9a\xf0\xb7\x88\x27\x6f\xd0\x81\xae\xe0\xa1\xb0\xa3\x2a\x49\x05\x01\xc1\x64\x42\xb9\x12\x9a\x73\x23\xb4\x52\x1a\x2b\xe7\xb8\x15\xcc\xb1\x94\xbb\xab\x78\x28\xce\xb1\xc5\x4c\x39\x46\x48\x92\x32\x9d\x63\x97\x4a\x9b\x72\xce\x05\x65\x4e\x59\x9a\x6b\xe1\x38\xd8\x2b\x78\xb0\x84\x6a\x91\x18\x26\xa9\x23\xca\x69\x9e\x28\x29\x0d\x13\xdc\x61\x48\x34\x4f\xb9\x11\x26\xd7\xa9\x73\x57\x46\x09\x3e\x88\x5c\xb6\x58\x69\x25\x31\xa1\x4e\xe4\x3a\x4d\xad\xc4\xcc\x18\x4d\x29\xc7\xc6\x3a\x80\xc4\xa4\xe0\x7e\xc1\xd8\x73\x03\x0d\xb4\xa0\x94\x4d\x31\x0f\xd5\xba\xc5\x06\xd1\x0e\x42\xfb\xd3\xc7\x5c\xfb\x38\xf3\x05\xdc\x3d\x64\x88\x74\x3a\x17\x23\xda\x04\x07\xb9\x6e\x36\x71\x60\x6d\x68\xca\x78\xaf\x0b\xc8\x50\xb7\xd0\xbe\x6c\x2b\x96\xc1\xc1\xd7\xe9\x28\x43\xfb\x3a\xeb\xf7\x37\xed\x7d\xb5\x0a\x75\xcc\x54\x9a\xf0\x0e\x42\x4b\x5d\x8f\x2b\x6f\xa1\x9d\x62\x97\xe7\xbc\x3c\xf2\x85\x8f\x19\x4a\x04\xa1\x4c\xca\xce\xd9\x62\x38\xc4\xcb\x41\x42\x13\x33\xd4\xe5\x18\xd7\x6d\x99\x42\x1f\xa6\x10\x2b\xdf\x1a\x4f\x4f\x12\xbe\x8c\x50\xed\xf4\xa6\x5d\x6e\x1b\x45\x4f\xfb\xe2\x61\x1c\xc2\x66\x60\x2d\xd4\xf5\xb0\x6c\x6f\x40\x97\xa1\x58\x35\xd0\xe9\xfc\x1e\x00\x00\xff\xff\x9a\xbf\x17\x4e\x8d\x07\x00\x00")

func resourcesDefault_configYamlBytes() ([]byte, error) {
	return bindataRead(
		_resourcesDefault_configYaml,
		"resources/default_config.yaml",
	)
}

func resourcesDefault_configYaml() (*asset, error) {
	bytes, err := resourcesDefault_configYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/default_config.yaml", size: 1933, mode: os.FileMode(420), modTime: time.Unix(1539941400, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _resourcesTesting_configYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x9c\x52\xcb\x6e\xe3\x3a\x0c\xdd\xfb\x2b\x0c\x6e\xb2\x71\x52\xbd\x5f\x7f\x70\x71\x71\x57\x77\x80\xae\x29\x89\x6a\x8c\xc4\x8e\x47\xb6\xdb\x06\x45\xff\x7d\xe0\x34\x9d\x6e\xa7\x03\x69\x41\x12\x3a\x87\x87\xe2\xa1\xe5\x48\x95\xd6\x21\x34\x6d\x8b\x29\x5d\xd6\x71\x99\xb7\xb8\x6d\x07\xec\xc7\xd0\xde\xc2\xb6\x3d\xd1\x35\xb4\xbb\x37\xc0\x9c\x2b\xcd\x33\x04\x70\x3e\x32\x74\x46\x3b\x99\x94\x52\x0a\x53\xc9\x96\x47\x65\x24\xb1\x2c\x93\xd6\x48\x5c\x71\x81\x1a\x3a\x48\xf5\x3a\x2d\x17\x08\x6f\x90\xfa\xe9\x48\x15\x02\x20\xcd\x7b\x2e\xdc\x3e\x2d\x75\x7b\x70\x2b\x2f\xf4\xba\x40\x80\x64\xad\x2f\x4e\x5a\x9f\xad\x65\xd9\x8b\x54\x12\xcf\x39\x2b\x74\x45\xf2\xac\x91\x61\x4e\xae\x08\x64\x51\x20\x57\x8c\x4b\xcb\xb2\x34\x92\x15\xe9\x12\x4b\x0e\x7f\xf3\x4d\x58\x71\x98\xb7\xb6\xfd\x33\x04\x90\x26\x71\xe3\xc8\xca\x58\xbc\x63\x85\xac\x8e\xcc\x0a\x5b\x9c\x67\x68\x39\x66\x78\xef\xe0\x94\x0b\x04\x98\x6f\x82\xe1\x96\x7e\x91\xe4\xd3\x99\x46\x08\x52\x74\x30\x42\x10\x46\x70\xa5\x3a\x98\x20\xf0\x0e\x2a\x04\xd7\xc1\x8c\xe7\x6d\x80\x4c\x3c\x12\x37\x24\x93\x77\xdc\x2b\x95\x39\x25\x14\xd1\x45\x61\x49\x91\x21\x16\x75\x2c\x51\xc9\x48\x4c\x5a\x83\x3a\x3b\xe7\x7c\x41\x63\x3d\x0a\xc7\x85\xd8\x84\x0c\x98\xb6\xaf\x48\x5c\xb8\xe8\xb8\xd6\x5a\x47\xe4\x84\xd9\x26\x24\xcf\x0c\x23\xe7\x94\xc0\x92\xd0\x49\x6d\x32\x33\x4a\xeb\x98\x3d\x6a\xab\x45\x44\x53\x52\x62\x5e\x50\xd9\x98\xfa\x0c\x01\x94\x26\x66\x18\x9a\x7d\x16\x48\x7b\x25\xa3\xdb\x7b\x21\xca\x5e\x29\x27\xbc\xf2\x3e\x4b\x9b\xa1\x83\x67\xaa\x73\x7f\xd9\x86\x7c\xdf\xdd\x17\x3f\xe1\x3c\xbf\x5c\x6a\x0e\xed\xee\xb3\x74\xf7\x40\x68\xff\xd4\x02\x4d\xd3\x67\x1a\x97\x7e\xb9\xfe\xb3\xf1\xb0\x57\xc6\xbf\xce\xae\x69\x7e\xae\xb4\xd2\x66\xba\x71\x1d\x1e\x2f\xf5\x44\x75\x0e\xad\x68\xda\xf6\xe5\x96\x3c\x62\xbf\xfc\xe8\x07\xfa\xef\xff\xd0\xf2\xa6\x39\xd1\xf5\xe6\xd0\xb9\x7f\x1a\xfb\xf1\xe9\xc3\xac\xd3\x1a\xcf\x7d\xfa\x77\x73\xe9\xe1\xf0\xf0\x71\xe9\x15\x87\xe9\x4c\x0f\x95\xe6\xcb\x5a\x13\xcd\x0f\x1b\x04\x97\xb5\x12\x3f\x4c\x6b\x3c\x4c\x34\x7c\x80\x6b\xff\x8c\x0b\x7d\x03\x7d\xa2\xeb\x1d\x4d\xcb\x11\xd7\xe5\xf8\x1d\x15\x77\xc8\xdf\x48\xf8\x84\x7e\xf6\xff\x15\x00\x00\xff\xff\x2a\x32\x0b\x61\xbe\x03\x00\x00")

func resourcesTesting_configYamlBytes() ([]byte, error) {
	return bindataRead(
		_resourcesTesting_configYaml,
		"resources/testing_config.yaml",
	)
}

func resourcesTesting_configYaml() (*asset, error) {
	bytes, err := resourcesTesting_configYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "resources/testing_config.yaml", size: 958, mode: os.FileMode(420), modTime: time.Unix(1539936343, 0)}
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
	"resources/default_config.yaml": resourcesDefault_configYaml,
	"resources/testing_config.yaml": resourcesTesting_configYaml,
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
	"resources": &bintree{nil, map[string]*bintree{
		"default_config.yaml": &bintree{resourcesDefault_configYaml, map[string]*bintree{}},
		"testing_config.yaml": &bintree{resourcesTesting_configYaml, map[string]*bintree{}},
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
