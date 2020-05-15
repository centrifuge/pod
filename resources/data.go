// Code generated by go-bindata.
// sources:
// ../build/configs/default_config.yaml
// ../build/configs/testing_config.yaml
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

var _goCentrifugeBuildConfigsDefault_configYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x58\xdd\x6f\xdb\xb0\x11\x7f\xd7\x5f\x71\x70\x5e\xda\xa1\xb5\x2d\xf9\x23\x8e\x80\x3d\x38\x71\xe2\xa6\xf9\x80\x13\xa7\x49\xdb\x97\x81\xa6\x4e\x12\x6b\x89\x54\x49\xca\xb6\xf2\xd7\x0f\x47\x49\x4e\xd2\xb4\xeb\xd6\x61\x03\x06\x2c\x2f\x11\x48\xde\xef\x8e\x77\xbf\xfb\xa0\x0f\x60\x86\x31\x2b\x33\x0b\x11\x6e\x30\x53\x45\x8e\xd2\x82\x45\x63\x25\x5a\x60\x09\x13\xd2\x58\x58\xab\x0d\x93\x1e\x47\x69\xb5\x88\xcb\x04\xaf\xd1\x6e\x95\x5e\x87\x10\x67\x42\x5a\xcf\x81\x08\x89\x60\x53\x84\xa8\xc1\x93\xf5\x19\x03\x36\x65\x16\x4e\xf6\xb2\x90\x33\x21\x2d\xe1\x7a\xed\x91\xd0\x03\x38\x80\x4b\xc5\x59\xe6\x54\x0b\x99\x00\x57\xd2\x6a\xc6\x2d\xb0\x28\xd2\x68\x0c\x1a\x90\x88\x11\x58\x05\x2b\x04\x83\x16\xb6\xc2\xa6\x80\x72\x03\x1b\xa6\x05\x5b\x65\x68\xba\x1e\xb4\xf2\x04\x09\x20\xa2\x10\x06\x83\x81\xfb\x46\x9b\xa2\xc6\x32\x6f\x6c\x3f\x8f\x42\x98\x0c\x26\xf5\xde\x4a\x29\x6b\xac\x66\xc5\x02\x51\x9b\x5a\xf6\x3d\x74\x7a\xa2\x18\xf6\xfc\xe0\xb0\xdb\xef\xf6\xbb\x7e\xcf\xf2\xa2\x37\x98\x04\xfd\xa0\x27\x8a\xd8\xf4\x6e\xf2\xbb\x9b\xdd\x6a\xbb\x2e\xbf\x7e\xf9\x32\x8b\xcb\xc7\xbb\xd5\xee\x74\x7a\x8b\x77\xd7\x27\x97\xea\xb1\xaa\x46\xa3\xc9\xe6\x46\x26\xf7\x9b\xc5\xd5\xb7\xcb\x2f\xeb\xce\x6f\x40\x07\x2d\xe8\x7d\x3c\x3e\xbd\x1e\xe7\xeb\xef\x0f\xf8\xed\xe1\xe2\x21\xf8\xbe\x28\xfd\xf1\xe7\x22\x9a\x0f\xd6\x1f\x95\x7f\x37\xc8\x53\x96\x2e\x8e\x47\x4b\x1c\x49\xbf\x06\x6d\x5d\x35\x6d\x3d\x55\x5f\x80\xae\x8f\xd2\x0a\x5b\x9d\x31\x6e\x95\xae\x42\xe8\x74\x3c\xe7\xea\x2b\x26\xe4\xab\x80\xb7\x11\x83\x37\x17\x14\xee\xb7\x1e\xd4\xe1\xad\xd1\x0e\xe0\xba\xcc\x51\x0b\x0e\xe7\x33\x50\xb1\x0b\xf5\xb3\xa0\x36\xb2\x7b\xaf\xfb\x41\x23\x75\xdc\xba\x16\x32\x61\x2c\x49\x4a\x15\xe1\x6b\x56\x14\x5a\x6d\x84\xdb\x50\x0e\xdb\xa9\x6e\x89\xf8\xdb\x20\x0d\x46\xdd\x60\x18\x74\x83\x41\xbf\xeb\xfb\xe3\x1f\x23\xe5\x07\xb3\xc1\x85\x52\x0f\xcb\xd5\x6e\x75\x71\xb2\xfa\x9a\x1e\x7d\xbc\xb7\xe6\xa6\xba\x9f\x47\x77\x0b\xcd\x86\xb7\xc5\x72\x3a\xb4\xab\x8d\x19\x33\xe9\xfb\xdf\xb6\xf3\x69\xf0\xd8\x79\x85\x3f\x18\x76\x0f\x83\xae\x1f\x1c\xfe\x0a\xfe\x26\x0f\xf8\x32\xd7\xa7\x82\x2d\xaf\xee\x87\xc9\xa7\xcd\xe1\xc3\x3c\x2d\x92\xdb\xad\x9a\x6c\xd5\xd9\xd2\x7c\x48\xbf\xce\x57\x73\x31\x60\xd3\xc9\xae\xd3\xb8\xe7\xb4\x61\xe5\xde\xf9\xe7\x33\x78\x0f\x2e\x00\xbf\x62\xed\xb0\x75\xed\x25\x73\x61\x8b\xb0\xc8\x54\x85\x11\x2c\x73\xa6\x2d\x9c\x34\x6c\x30\x10\x2b\xed\x5c\x99\x88\x0d\xca\x17\xae\xfc\x17\x18\xd3\xdf\xf9\x83\x71\x70\xca\x8f\xe3\xc9\xf8\xf0\x28\x18\x0e\x4e\x83\x61\x3c\xed\x9f\x9e\x0c\x83\x51\x14\xa0\xdf\x9f\xf6\x27\x41\x30\xe0\x87\xb3\xe7\xdc\x32\x96\x25\x94\xc5\xaf\x29\xc5\xf2\x15\xea\x3f\xa3\x94\xff\x6f\x52\xca\xa9\xfe\x2d\xa5\xfe\xf3\xa4\xfa\x3f\xad\xfe\x90\x56\xd4\x92\x9e\x58\x91\xd7\x2b\x7f\xc6\xa5\xfe\x3f\x53\x52\xfc\xa3\x49\xd7\x0f\x82\xae\xef\xff\x32\x38\xd3\x64\x70\xca\xa7\x56\x7f\xb9\x3f\xd9\x6d\x1f\xc7\xeb\xb1\xb9\x3b\x12\x5f\x97\xb7\x8f\xf6\xf1\x68\x76\x58\x7d\x7a\x2c\x8e\x17\xb7\xa7\x67\x8f\xfa\x93\xba\x7f\x5d\x52\x88\x5d\x81\xdf\xf5\xfd\x57\xcd\xa5\xc5\xbf\x98\x6f\xc5\xee\x33\xca\xf2\xf3\xf4\xfe\xfb\xfa\xe3\x45\x2e\x3f\x2c\xa7\x1f\x67\xdf\x1e\xe3\x43\x9c\x5f\xa9\xb1\xd5\x4a\x24\x5f\x77\xf9\xe1\x74\x74\xfb\x8f\x83\xdf\xb8\xeb\x57\xe1\xf7\xff\xbb\xd1\x9f\x9e\x0d\x47\x63\xee\x8f\x07\x93\x31\x1b\x0f\xe3\x68\x78\x36\x5c\x8d\x8f\x58\xec\x0f\xd8\x64\x3c\x8b\xfb\xc7\xa3\x71\x30\x65\xfd\x7e\xc7\xa3\xe9\x82\x59\x06\x4b\xab\x34\x4b\xd0\x33\xf5\xff\x7a\x66\x58\x30\x9b\x3a\x93\x32\x6a\x66\xb3\x63\x88\x45\x86\x1e\x40\xc1\x6c\x1a\x42\xcf\xe6\x45\xef\x69\x6a\xf9\x5b\xc4\x2c\xeb\xba\x93\xd1\x8a\x70\x4f\x94\x8c\x45\x52\x6a\x66\x85\x92\x7b\x05\xdc\xad\x2e\xff\x5c\x4d\x0d\xf0\x4a\xdb\x94\x73\x55\x4a\x6b\x60\x8d\x15\x34\xb7\xf0\x58\xb3\x48\x7a\xd6\x58\xd1\x32\x36\x88\xed\x16\xc9\x9e\x4b\x8b\x3a\x66\x1c\x61\x4b\x91\x73\x11\x98\x2e\xce\x81\xc9\x08\x16\xc1\x02\x96\xa8\x37\xa8\x5d\x3d\x44\x49\x05\xcf\xa3\x92\xf8\x41\x19\x2b\x59\x8e\xd4\x8e\x9b\x79\xc3\x3b\x80\x85\xd2\xb6\x81\x21\x88\x9f\x8b\xd2\xa1\x10\x26\xfd\x49\x40\xea\x29\x3d\xde\x5b\xf5\xbe\x40\xd4\xc0\x9f\x7b\xcd\x78\x45\x50\xd4\x4e\x5a\x16\xc8\x45\x5c\xc1\xe9\xce\xa2\x96\x2c\x83\xf3\xc5\x33\x6b\x09\x14\x38\x93\x34\xbd\x69\x64\x3c\xc5\x08\x98\x05\x11\xc3\x0a\x53\x21\x23\xb8\x9e\xde\x11\x0c\x36\xd2\xe7\x8b\x10\xb6\xdd\x5d\xb7\xea\x3e\xd6\x21\x20\xab\x4b\x83\xd1\x9e\x81\x74\xef\x8c\x55\xa8\x29\x10\xce\x5c\x97\x3f\xee\xf4\x9d\xc8\x51\x95\xee\x9a\x12\x54\x81\xb2\x19\x29\x25\x72\x67\x35\xb5\x04\xba\x8c\xf1\xa0\x5d\x6e\x44\x42\xe8\x0c\xfa\xa6\xe3\x50\x72\x21\x45\x5e\xe6\x10\x61\xc6\x2a\xa7\x17\x37\xa8\x2b\x28\x82\x02\x34\x9a\x42\x49\x83\x84\xc4\x36\x4a\x44\x60\x45\x4e\x5a\x98\xb5\x8c\xaf\x8d\x03\x60\xd1\xb7\xd2\x58\x58\x31\xb2\x5b\x49\x48\x95\xb1\x24\xa9\x4a\xcd\xd1\xc0\x9b\xe5\x72\xf6\x0e\x4e\x16\x9f\xde\x01\x57\x1a\x0d\x74\xbb\xdd\xb7\xcd\x2c\xac\xd6\x20\x24\x64\x2a\x71\x29\x17\x42\x87\xec\x23\x5b\x4d\x99\x63\x04\xab\x8a\xae\x55\xc7\xa0\x43\x5e\xdc\xfd\xf5\xcd\x86\x65\x25\xde\x22\x8b\xe0\x2f\x10\xbc\x05\x61\x20\x43\xe3\xda\xa2\x04\xb7\x07\x2b\xcc\xd4\xf6\x1d\x79\x4f\x02\x4f\x99\x4c\x70\x7f\x8f\x99\xbb\xa3\x55\xb0\xf3\xe0\xe5\x62\x08\x9d\x51\xbf\x9f\x1b\x97\x8a\x37\x25\x96\xf8\x03\x05\x9c\x67\x98\xa9\x24\x4f\xb5\x92\xaa\x34\xd4\x79\x39\x1a\x23\x64\xe2\x7d\x27\x81\x9a\x20\xf5\x23\xc1\xd4\x74\x28\x5d\x33\x56\x31\x50\x01\x42\x6d\x7a\xcd\xd5\x74\xd3\xc7\xb7\x22\xcb\x88\x2b\x2c\xcb\x14\x67\xb6\x66\x8b\xb1\x4c\xdb\xb2\xf0\x80\xe4\x1f\x6a\x41\x2a\xe6\x7d\x87\x7f\xa6\x11\x0d\x94\x05\x79\x14\x78\xc5\x33\x34\x35\x01\x6a\x15\xe4\x90\x2d\x13\xee\x75\xd1\xc4\x92\xb2\x0b\x9a\xed\x07\x26\x1c\x07\xae\x96\x75\x31\x3c\x80\x69\x4e\xf9\xe7\xba\x09\xf9\x9e\x81\x65\x66\x4d\x28\x1b\x96\x89\x08\x62\xad\x72\x77\x17\xae\xd1\x39\xc2\x83\x7a\xe7\xcc\xc5\xcb\x0f\xd2\x8e\xe7\xaa\x0c\x4a\x7b\x92\xba\xa9\xc8\x65\x88\xe0\x2f\xfd\xe7\xde\x55\xee\x00\xb9\x89\xf2\xe4\xd3\xed\x65\x08\x5b\x13\xf6\x9e\xde\x09\xe1\xd1\xd1\x70\xe8\xac\xba\xa6\x44\xb2\x9a\x49\xc3\x1c\x97\xa1\x50\x2a\x83\x9c\xed\x40\xa3\xd5\xa2\x1e\x77\x0c\xca\x88\x0c\x7e\x76\x4c\x6d\x5c\xa6\xe4\x6c\x77\x5b\x9f\x0b\x21\x68\x1c\xf7\x73\x48\x41\x35\x67\xc3\x32\x87\x5b\xd5\x9e\x64\x64\x3a\x2f\xb5\x76\x8f\x86\x67\x12\x29\x33\xb0\x42\xa4\x57\x85\x45\x6e\x31\xf2\x60\x0f\x40\xfa\x88\x45\x41\x93\x56\xed\x8b\x33\x13\x31\x36\xc4\xb4\x8a\x72\xbb\xd6\xc1\x55\x9e\x0b\xeb\xc2\xc4\x24\x30\xc9\x53\xa2\x57\xf3\x12\x75\xfe\x46\x69\xb9\x73\xe8\x7b\xf0\xa1\x42\x46\xf7\xaa\xcf\x5d\x8a\x18\x4d\xc1\x64\x08\x9d\xc9\xe1\xb8\x9f\x3a\xce\xee\xfb\xe1\x2f\xfc\xdf\x76\xc3\xa6\x8c\x61\x86\xd4\xe8\xb6\xa9\xe0\xe9\xbe\x53\x42\x53\x8d\x5b\x4b\x9b\x11\x43\x11\x9f\x9b\x39\x33\xa2\x84\x75\xf6\x95\xc6\xaa\xbc\x51\xd2\xb6\x8a\xe6\x59\xdc\x34\x81\x6b\x57\x95\x3b\xd4\x93\x3b\xfb\xc7\x6f\x1d\xa6\x1a\x78\xaf\x97\x67\x82\x7c\xed\xca\xe7\x9b\x2d\xe5\xeb\xf7\x52\x68\x84\xad\x01\xa5\x41\x14\xbc\x79\x11\xd3\x03\x98\x3e\x39\xb3\x64\xb6\xe3\xf6\xdb\xe7\x7c\x4a\xad\x2d\xc2\x5e\x8f\xb2\x29\xa3\x3a\x14\x1e\x8d\x86\xa3\xba\xcc\xb1\x9d\x2b\x73\x94\x6a\x5b\x8c\x20\x61\x74\x27\xc1\x1d\x5e\xd1\x54\xbe\x97\x64\x12\x12\xb6\x28\x9c\x74\xd0\x87\xf9\x16\x05\x48\xb5\xad\xe9\x35\x67\x66\x41\xd2\x8e\x5f\xed\x9f\x3b\x3a\x67\x06\x32\x91\x8b\x66\x8c\x88\x44\x1c\xa3\x63\xd2\x3e\x42\xfb\x9a\x46\x79\x99\x30\x73\xe9\x4e\xb7\x8f\xf9\x13\x4a\x34\x74\x09\xdf\x60\xd2\xea\x34\x8a\x2e\xb0\x0a\x61\xf0\x7c\xf1\x16\x37\x6a\x8d\x6e\x7d\x34\x6a\x97\x6b\x8e\x9c\x38\x7e\x85\x30\xf9\x61\x7d\xa1\xb1\xdd\xf2\x9f\xa0\x64\x6c\xaf\xe8\x11\x0c\x47\x2f\xd6\xee\xc8\x19\x31\xea\x33\xad\xf2\x10\xfc\xd1\x7e\x8f\x19\x83\x76\x59\xb7\xf1\x31\xad\xc2\xc1\xbe\x96\x69\xcc\xd5\x86\x2a\x99\x01\xa3\x94\xa4\xff\x2b\x2d\xa2\x04\xa9\xa8\x50\xb6\x24\x9a\xd5\xa9\xf3\xd4\xc1\xac\x72\x45\xab\x8e\x81\x7c\xe2\xc5\xf3\x68\x34\x0c\x88\xa2\xfa\xf7\x11\x06\xab\x4c\xf1\xb5\x1b\x0e\x6a\x22\x80\xd5\x22\x49\x50\x3b\x6c\x9a\xd3\x70\x67\xdb\x7a\x57\xf7\xbc\x71\xbf\x6d\x7a\x3f\x53\xac\xa9\xa9\x28\x99\x3d\x6b\x3a\x66\x9f\x92\xad\x49\x4f\xd0\xd4\x83\x5e\xc2\xfb\xa3\x06\xfd\x7f\xbb\x7a\x79\x07\xc0\x64\x05\x11\xae\xca\x24\x69\x46\x0a\xca\x71\x17\xe0\x44\x01\x39\xc2\x73\xbb\x75\x2d\x41\xe9\xd2\xd2\xad\x50\x2f\x27\x19\x0f\xe8\x2b\x84\x98\x65\x06\xdd\xa9\xa2\xd0\x2a\xae\x33\xa2\x05\xa6\x91\x86\x56\xdb\x63\x5e\x4d\xd1\xe6\x87\xad\x42\x23\x6f\x98\x6a\x75\x89\xde\xdf\x03\x00\x00\xff\xff\x9a\x8d\x48\xb5\xc5\x13\x00\x00")

func goCentrifugeBuildConfigsDefault_configYamlBytes() ([]byte, error) {
	return bindataRead(
		_goCentrifugeBuildConfigsDefault_configYaml,
		"go-centrifuge/build/configs/default_config.yaml",
	)
}

func goCentrifugeBuildConfigsDefault_configYaml() (*asset, error) {
	bytes, err := goCentrifugeBuildConfigsDefault_configYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "go-centrifuge/build/configs/default_config.yaml", size: 5061, mode: os.FileMode(420), modTime: time.Unix(1589520430, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _goCentrifugeBuildConfigsTesting_configYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x94\x53\xc9\x6e\x1c\x37\x10\xbd\xf7\x57\x10\xcc\xc1\x97\x59\xb8\x6f\x37\x43\xf1\x12\x08\x11\xe2\x38\x80\x9c\x63\x91\xac\x96\x1a\xa3\x5e\xc2\x66\x4b\x1e\x1b\xfe\xf7\xa0\x67\x46\xb6\x6f\x51\x6e\x64\x75\xbd\xf7\x58\xd5\xef\x25\x1c\x6a\xe9\xda\xe5\x0e\x6f\xb0\x3e\x8d\xe5\x10\x48\xc5\xb9\x76\xc3\x5d\x83\xf5\x1e\x0b\x2e\x7d\x68\x08\x81\x94\xc6\x65\xa8\xf3\x7a\x26\xa4\x87\x6e\x08\xe4\x74\x24\xe4\x80\xc7\x40\x5e\x7d\xa5\x90\x73\xc1\x79\xa6\x81\x3a\x1f\x19\x38\xa3\x9d\x4c\x4a\x29\x05\xa9\xcd\x96\x47\x65\x24\xb2\x2c\x93\xd6\x80\x5c\x71\x01\x9a\x6e\x68\x2a\xc7\xa9\x8e\x34\x7c\xa5\xa9\x9b\xee\xb1\xd0\x40\x01\xe7\x2d\x17\x6e\x9b\x6a\x59\x1b\x4e\xe5\x8a\x9f\x2b\x0d\x34\x59\xeb\x5b\x27\xad\xcf\xd6\xb2\xec\x45\x6a\x13\xcf\x39\x2b\x70\xad\xe4\x59\x03\x83\x9c\x5c\x2b\x80\x45\x01\x5c\x31\x2e\x2d\xcb\xd2\x48\xd6\x4a\x97\x58\x72\xf0\x9d\x6f\x82\x02\xfd\xbc\xca\x76\x8f\x34\x50\x69\x12\x37\x0e\xad\x8c\xad\x77\xac\x45\xab\x23\xb3\xc2\xb6\xce\x33\xb0\x1c\x32\xfd\xb6\xa1\x87\xdc\xd2\x40\xe7\xd3\x83\xe9\xe9\xfa\x83\x24\x1f\x1e\x70\xa0\x41\x8a\x0d\x1d\x68\x10\x46\x70\xa5\x36\x74\xa2\x81\x6f\x68\xa1\xc1\x6d\xe8\x0c\x0f\xeb\x00\x19\x79\x44\x6e\x50\x26\xef\xb8\x57\x2a\x73\x4c\x20\xa2\x8b\xc2\xa2\x42\x83\x2c\xea\xd8\x46\x25\x23\x32\x69\x0d\xe8\xec\x9c\xf3\x2d\x18\xeb\x41\x38\x2e\xc4\xfa\x90\x1e\xd2\xba\x8a\xc4\x85\x8b\x8e\x6b\xad\x75\x04\x8e\x90\x6d\x02\xf4\xcc\x30\x74\x4e\x09\x68\x13\x38\xa9\x4d\x66\x46\x69\x1d\xb3\x07\x6d\xb5\x88\x60\xda\x94\x98\x17\xd8\xae\x4c\x5d\xa6\x81\x2a\x8d\xcc\x30\x30\xdb\x2c\x00\xb7\x4a\x46\xb7\xf5\x42\xb4\x5b\xa5\x9c\xf0\xca\xfb\x2c\x6d\xa6\x1b\xfa\x88\x65\xee\xc6\x75\xc8\x6f\xaf\x2e\x3f\x7e\x82\x79\x7e\x1a\x4b\x0e\xe4\xd5\x73\xe9\xe2\x81\x40\x5e\x6a\x81\xa6\xe9\x32\x0e\xb5\xab\xc7\xdf\x72\x20\x94\x7d\x7e\xb1\x77\x9a\x66\xb5\xee\xd5\xfd\x6a\xc5\x1f\x06\x3d\xfb\xb3\x3b\x73\x65\x25\xb5\x97\xc9\x72\xdd\xe6\x2c\x79\x32\x9c\x2b\x0e\x31\x33\x05\xde\xb7\xd9\x38\x21\x92\xd3\xda\x39\xad\x52\xca\x28\x3d\x68\xe3\x14\x5a\xd0\x26\x83\xb0\x99\x9e\xc8\x66\x4c\x05\x6b\x20\x74\xbf\x7f\xfd\xd0\x25\x3c\x57\xbf\x4f\x4a\xf5\xbb\xf2\xf4\x08\x6f\xde\xea\x2f\x9f\xa2\x30\x6f\xbf\xf8\x92\x3e\x4c\xbf\xde\x7e\xd4\xf6\xaa\xbe\xf9\xf3\xfd\x74\x83\xf7\x9f\xae\xfe\x48\x37\xe3\xfb\x77\xd7\x4b\xfd\xf0\x37\x6d\x9a\x5f\xc8\xeb\x4b\x9e\xd6\xf4\x90\xb9\x8e\x05\xee\xb0\xf9\x39\x64\x07\x3c\xae\x65\x0c\x64\x5f\xfb\x69\xff\xfc\xa9\x69\xfe\x59\x70\xc1\xb5\x63\x58\xfa\xdb\xb1\x1c\xb0\xcc\x81\x88\x86\x90\xa7\xd3\xe5\x16\xba\xfa\x57\xd7\xe3\xef\x1f\x03\xe1\x4d\xb3\xd2\xac\xcd\x93\x98\xce\xab\x99\x96\xf8\xd0\xa5\xeb\x35\xb3\xbb\xdd\x7e\xb7\xdb\xc7\xa5\x7b\xc8\xfb\x82\xf3\xb8\x94\x84\xf3\x7e\x12\xd3\x35\x1e\x77\xd3\x12\x77\x13\xf6\x67\x4c\xe9\x1e\xa1\xe2\x7f\x83\x0e\x2b\xf0\x04\x9a\xbb\xbb\xa1\x1b\xee\x5e\xa8\x79\xe9\xfe\xff\xba\x3f\x01\x9f\xb5\x1b\x18\xd2\xfd\x58\x2e\xe2\x53\xc1\x34\xf6\x7d\x57\x03\xa9\x65\xc1\xe6\xdf\x00\x00\x00\xff\xff\xdc\x3c\xc5\xc4\xef\x04\x00\x00")

func goCentrifugeBuildConfigsTesting_configYamlBytes() ([]byte, error) {
	return bindataRead(
		_goCentrifugeBuildConfigsTesting_configYaml,
		"go-centrifuge/build/configs/testing_config.yaml",
	)
}

func goCentrifugeBuildConfigsTesting_configYaml() (*asset, error) {
	bytes, err := goCentrifugeBuildConfigsTesting_configYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "go-centrifuge/build/configs/testing_config.yaml", size: 1263, mode: os.FileMode(420), modTime: time.Unix(1583245934, 0)}
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
	"go-centrifuge/build/configs/default_config.yaml": goCentrifugeBuildConfigsDefault_configYaml,
	"go-centrifuge/build/configs/testing_config.yaml": goCentrifugeBuildConfigsTesting_configYaml,
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
	"go-centrifuge": &bintree{nil, map[string]*bintree{
		"build": &bintree{nil, map[string]*bintree{
			"configs": &bintree{nil, map[string]*bintree{
				"default_config.yaml": &bintree{goCentrifugeBuildConfigsDefault_configYaml, map[string]*bintree{}},
				"testing_config.yaml": &bintree{goCentrifugeBuildConfigsTesting_configYaml, map[string]*bintree{}},
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

