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

var _goCentrifugeBuildConfigsDefault_configYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x99\x59\x73\xdb\xba\x92\xc7\xdf\xf5\x29\x7a\x94\x97\x64\xea\x44\xe6\xbe\xa8\x6a\x1e\xa8\xcd\x49\x1c\x3b\xb2\xe5\xe5\xc4\x2f\x53\x20\xd9\x94\x10\x93\x00\x03\x80\x5a\xfc\xe9\xa7\x00\x52\x8a\x12\x47\xb9\x35\xb9\x73\x4f\xd5\xa9\xb9\x7e\xb1\x0b\xcb\x1f\x8d\xee\x5f\x37\x40\xf8\x15\x4c\xb0\x20\x4d\xa9\x20\xc7\x35\x96\xbc\xae\x90\x29\x50\x28\x15\x43\x05\x64\x49\x28\x93\x0a\x04\x65\x4f\x98\xee\x7a\x19\x32\x25\x68\xd1\x2c\xf1\x0a\xd5\x86\x8b\xa7\x21\x88\x46\x4a\x4a\xd8\x8a\x96\x65\xcf\x88\x51\x86\xa0\x56\x08\x79\xa7\xcb\xda\x91\x12\xd4\x8a\x28\x18\x1f\x14\xa0\x22\x94\x29\xad\xdf\xdb\x0f\x19\xf6\x00\x5e\xc1\x47\x9e\x91\xd2\x98\x40\xd9\x12\x32\xce\x94\x20\x99\x02\x92\xe7\x02\xa5\x44\x09\x0c\x31\x07\xc5\x21\x45\x90\xa8\x60\x43\xd5\x0a\x90\xad\x61\x4d\x04\x25\x69\x89\x72\xd0\x83\xfd\x7c\x2d\x09\x40\xf3\x21\xb8\xae\x6b\xfe\x46\xb5\x42\x81\x4d\xd5\xed\xe0\x7d\x3e\x84\xc8\x8d\xda\xbe\x94\x73\x25\x95\x20\xf5\x1c\x51\xc8\x76\xee\x5b\xe8\x9f\xd1\xda\x3b\xb3\x9d\x70\x60\x0d\xac\x81\x7d\xa6\xb2\xfa\xcc\x8d\x1c\xcb\x39\xa3\x75\x21\xcf\xae\xab\xdb\xeb\x6d\xba\x79\x6a\x1e\x3f\x7f\x9e\x14\xcd\xf3\x6d\xba\x9d\x26\x37\x78\x7b\x35\xfe\xc8\x9f\x77\x3b\xdf\x8f\xd6\xd7\x6c\x79\xbf\x9e\x5f\x7e\xf9\xf8\xf9\xa9\xff\x0f\x44\xdd\xbd\xe8\x7d\x11\x4c\xaf\x82\xea\xe9\xeb\x03\x7e\x79\xb8\x78\x70\xbe\xce\x1b\x3b\xf8\xb3\xce\xcf\xdd\xa7\x0f\xdc\xbe\x75\xab\x15\x59\xcd\x47\xfe\x02\x7d\x66\xb7\xa2\x7b\x57\x25\x7b\x4f\xb5\x1b\xd0\xdb\x47\xa6\xa8\xda\xcd\x48\xa6\xb8\xd8\x0d\xa1\xdf\xef\x7a\x08\xcb\x56\x5c\xdc\x60\xcd\x25\xfd\xa1\x8b\xb2\x35\xa7\x19\xde\xb1\x9a\x68\xf7\xf5\xfb\x3d\x13\x9d\x4b\x42\xd9\x4f\x59\xe9\x82\x08\xaf\x6f\x5a\x58\xde\xf4\xe0\x18\x8e\xd6\x96\x57\x70\xd5\x54\x28\x68\x06\xef\x27\xc0\x0b\x03\xca\x11\x12\x9d\xc6\x21\x66\xbe\xdd\xcd\x1a\xed\x03\x03\x25\x95\x4a\xcf\x64\x3c\xc7\x97\x4c\xd5\x82\xaf\xa9\xe9\xe0\x46\xfb\xc8\x80\xbd\xa1\xff\x30\xd0\xae\x3f\x70\xbc\x60\x60\xbb\xde\x20\xf6\x7f\x0c\xb6\xed\x4c\xdc\x0b\xce\x1f\xae\xe4\xa3\x7c\x08\x6f\xd3\xec\xd1\x8f\xae\x42\xfb\xee\x7a\x71\xe1\x4f\xbe\x3c\x7e\xad\x66\x4f\xef\xe6\xef\x36\xdb\xd9\xc5\x6d\xb2\xe3\x77\x77\x93\x28\x2f\xfa\x3f\xca\xdb\x71\x34\xb0\x03\x7b\x60\x47\xd6\x29\xfd\x09\x3a\x72\xf3\x30\x75\x0b\x87\x7e\x48\xef\xf0\x3a\x3e\xbf\xbb\xbb\x1e\xbd\x1b\x8b\x87\x8f\xe9\x28\x23\xf1\xe5\xf9\xe5\xd7\xa2\x4a\xc7\x4b\xd1\xa4\xfd\xce\x49\xd3\x8e\xec\x43\x28\xde\x4f\xe0\x2d\x74\xe1\x38\xc5\xbe\xd7\x4d\xfe\x48\xb4\x7f\x20\xc7\xba\xe4\x3b\xcc\x61\x51\x11\xa1\x60\xdc\x21\x25\xa1\xe0\xc2\x78\x74\x49\xd7\xc8\xbe\xf3\xe5\x4b\xec\xe0\x24\x77\xd6\xb6\x88\x22\x2b\x8d\x02\xcb\xb6\xdc\x34\xf7\x7c\xe2\x3b\xae\x1f\x7a\x09\xe2\xd8\x0a\xc7\x5e\xec\x58\xae\x5d\x78\x61\x64\xff\x82\x50\x6b\x1b\x3b\xc9\xc4\xf3\x46\xa3\x68\xe6\xb8\x13\x3f\xb7\x9d\x18\x47\x91\x43\x7c\x2b\x77\xa3\x20\x4a\x47\x5e\x6a\x67\x38\xb3\x67\xa7\x58\xb6\xb6\x99\x97\x44\x38\x72\xc2\x62\xe4\x4e\x89\x33\xb6\x62\xdf\x9f\x45\xc4\x1f\xd9\x81\xed\x8f\x9c\x20\x8f\xfc\xd9\x78\x84\x11\x76\xd4\x5f\xf0\x35\x69\x77\x7d\xc4\x68\x8a\x82\x91\x72\x85\x74\xb9\x52\xf2\xf7\xf8\x76\xfe\x49\xbe\xbf\x33\xe1\x7f\x47\xb8\x33\x70\x5c\x6b\x60\xdb\xc1\x29\x04\x17\xe9\x36\xbd\x18\xa7\x8f\xab\xf8\xc3\xbd\x92\xd7\xbb\xfb\xf3\xfc\x76\x2e\x88\x77\x53\x2f\x12\x4f\xa5\x6b\x19\x10\x66\xdb\x5f\x36\xe7\x89\xf3\xfc\x02\x71\xc7\xf5\x06\xa1\x33\xb0\x9d\xf0\x94\xfc\x75\xe5\x64\x8b\x4a\x4c\x29\x59\x5c\xde\x7b\xcb\xbb\x75\xf8\x70\xbe\xaa\x97\x37\x1b\x1e\x6d\xf8\x6c\x21\xdf\xad\x1e\xcf\xd3\x73\xea\x92\x24\xda\xfe\x9a\x70\x13\x9c\x93\x7c\x3b\xff\x02\xc0\x7f\xc1\xb7\xed\x06\xce\x34\x1b\x15\x51\x10\xc6\x8e\xe7\x4e\x1d\xaf\x48\xac\xe9\xd8\x73\xfc\xdc\x41\xdb\x4a\xac\xc8\x71\xdc\x2c\x9c\xfc\x92\xef\xd0\x8e\xac\x49\x18\xba\xb6\x95\x63\x16\x25\x23\x27\x4a\x48\x64\x39\xd3\xcc\x8a\x67\x45\xe2\x4c\x66\x81\x87\xb1\x15\x66\xa7\xf9\xb6\x23\xd7\x0e\x2d\x2f\xb2\x03\x2f\x2a\xb0\x28\xd0\x8b\x3d\x6b\xe6\x4e\x92\x24\x77\x49\x98\x66\x69\x6a\x65\x7e\x92\xcc\x3a\xbe\x6f\x78\x2d\x15\xbe\x20\x3c\xe7\xcb\x9a\xa8\x6c\xf5\x7b\x70\xbb\xff\x24\xdc\xfb\xd5\xe1\xf5\xed\xa7\xc9\x27\xc8\x04\x12\x85\x20\x3a\x53\x35\xe0\x46\xe7\xcd\xff\xbb\x8a\xde\x7a\xe0\x14\xf1\xee\x5f\x0b\xbc\x95\xbb\xb1\x3d\x0d\x1d\xd7\xf1\xc7\x98\x8f\x3d\x7b\xea\x45\x96\xef\x4e\xc3\xd0\x89\x22\x12\xc5\x33\x67\xea\xda\xb6\xed\xff\x12\x78\x67\x1c\x59\x33\x7b\x42\x8a\x09\x09\x49\x32\xc1\xd4\x19\xdb\xa1\x9f\x7b\x23\xcf\x4d\x22\x3f\xf2\x42\x77\x6a\xdb\xa1\xed\x9e\x06\xde\x8b\x53\x8c\x5d\xcb\x1a\xbb\xc1\xb8\xf0\x1d\x37\x4a\x67\x41\x3c\xf5\xc6\x5e\xec\x07\xd6\x6c\x16\x15\xe1\x2c\x08\x9d\xa9\x77\x74\x8d\xd1\xb7\x96\x63\xe0\x61\xf2\x09\xae\x3e\xdd\xc2\xdd\x62\xfa\x1f\x3d\x00\xac\x52\x22\x32\x92\xa3\xe0\x7a\xd4\x6f\xe5\x80\x6d\x9d\x84\xf3\x25\x3f\x8e\x33\xb0\xed\x93\xf5\x32\x59\xba\xd3\x2c\x51\xe2\xf3\xfd\x78\xbb\x79\x0e\x9e\x02\x79\x1b\xd3\xc7\xc5\xcd\xb3\x7a\x8e\x27\xe1\xee\xee\xb9\x1e\xcd\x6f\xa6\xb3\x67\x71\xc7\xef\xfb\x2f\x57\x30\x05\xdf\xb1\x07\xb6\xfd\xe2\x02\xbb\x5f\xe1\xe2\x7c\x43\xb7\x7f\x22\x6b\xfe\x4c\xee\xbf\x3e\x7d\xb8\xa8\xd8\xbb\x45\xf2\x61\xf2\xe5\xb9\x08\xf1\xfc\x92\x07\x4a\x70\xba\x7c\xdc\x56\x61\xe2\xdf\xfc\x9a\xd0\xaa\xf5\xee\x29\x42\xed\xbf\x96\xd0\x64\xe6\xf9\x41\x66\x07\x6e\x14\x90\xc0\x2b\x72\x6f\xe6\xa5\x41\x4c\x0a\xdb\x25\x51\x30\x29\xac\x91\x1f\x38\x09\xb1\xac\x5f\x12\x1a\xb8\xe1\x28\x1a\xbb\x13\x27\x49\xdc\x71\xe6\x58\xc1\x24\xf6\x7c\x3b\x4e\x7d\x2f\x8a\x1d\x2b\x8a\xb3\x78\x1a\x84\x71\x6c\x9d\x26\x74\xe4\xa3\xe7\xb8\xf9\x38\x0b\x3d\x2b\x1d\x8d\x23\xab\x88\xad\xc0\x76\x5d\xb4\xfd\xc0\xb2\x8b\x38\xb2\xe2\x38\x72\xfd\xe0\x07\x42\xbf\x21\x75\x04\xe4\xff\x35\x8c\xff\x6a\x14\xff\x0d\xe2\xdf\x13\xc4\x57\x30\x21\x8a\xc0\x42\x71\x41\x96\xd8\x93\xed\xef\xf6\x3b\x7d\x4e\xd4\xca\x78\xa6\xd4\x5f\x83\x93\x11\x14\xb4\xc4\x1e\x40\x4d\xd4\x6a\x08\x67\xaa\xaa\xcf\xbe\xbd\x17\xfc\x77\x4e\x14\x19\x98\x91\x79\xaa\x75\xc7\x9c\x15\x74\xd9\x08\xa2\x28\x67\x87\x05\x32\xd3\xba\xf8\xfd\x65\x5a\x81\x17\xab\x25\x59\xc6\x1b\xa6\x24\x3c\xe1\x0e\xba\x5d\xf4\x48\xd7\xa8\xd7\x79\xc2\x9d\x6e\xc6\x4e\x71\xdf\xa5\xe7\xbe\x67\x0a\x45\x41\x32\x84\x8d\x06\xc8\x80\x90\xcc\xdf\x03\x61\x39\xcc\x9d\x39\x2c\x50\xac\x51\x98\xbb\x0d\x32\x7d\x79\xe9\xe9\x6b\xc9\x3b\x2e\x15\x23\x15\x0e\xe1\xf0\x8d\xdf\x7b\x05\x73\x2e\x54\x27\xa3\x25\x7e\x3e\x55\x0f\x1a\x42\x64\x45\x8e\x5e\x5e\x67\xe9\x5b\xc5\xdf\xd6\x88\x02\xb2\x63\xaf\xc9\x5e\xed\xd4\xad\x93\x16\x35\x66\xb4\xd8\xc1\x74\xab\xcc\x17\x01\xbc\x9f\x1f\x59\xab\x45\x21\x23\x0c\x52\x04\x81\x24\x5b\x61\x0e\x44\x01\x2d\x20\xc5\x15\x65\x39\x5c\x25\xb7\x5a\x06\xbb\xd9\xef\xe7\x43\xd8\x0c\xb6\x83\xdd\xe0\xb9\x0d\x81\xb6\xba\x91\x98\x1f\x12\x41\xef\xbb\x24\x3b\x14\x3a\x10\xc6\x5c\x93\xc6\x66\xf4\x2d\xad\x90\x37\x66\x9b\x0c\x78\x8d\xac\x7b\xc6\x61\x98\x19\xab\xf5\xf5\x4e\x6f\x46\xf6\x60\xdf\xdc\x4d\x19\x42\xdf\xb5\x64\xdf\xa8\x54\x94\xd1\xaa\xa9\x20\xc7\x92\xec\xcc\xba\xb8\x46\xb1\x83\xda\xa9\x41\xa0\xac\x39\x93\xa8\x95\xc8\x9a\xd3\x1c\x14\xad\xf4\x2a\x44\x29\x92\x3d\x49\x23\x40\xf2\x2f\x8d\x54\x90\x12\x6d\x37\x67\xb0\xe2\x52\xe9\x99\xbc\x11\x19\x4a\x78\xbd\x58\x4c\xfe\x80\xf1\xfc\xee\x0f\xc8\xb8\x40\x09\x83\xc1\xe0\x4d\xf7\xfe\xc4\x9f\x80\x32\x28\xf9\xd2\x64\xfe\x10\xfa\xda\x3e\x6d\xab\x6c\x2a\xcc\x21\xdd\xe9\x6d\xb5\x31\xe8\x6b\x2f\x6e\xff\xeb\xf5\x9a\x94\x0d\xde\x20\xc9\xe1\x3f\xc1\x79\x03\x54\x42\x89\xd2\x5c\x71\x19\x98\x3e\x48\xb1\xe4\x9b\x3f\xb4\xf7\x18\x64\x2b\xc2\x96\x78\xd8\xc7\xc4\xec\x51\x71\xd8\xf6\xe0\xfb\xc6\x21\xf4\x7d\xcb\xaa\xa4\x49\xc5\xeb\x06\x1b\xfc\x01\x01\xe3\x19\x22\x77\x2c\x5b\x09\xce\x78\x23\xf5\x2d\x3a\x43\x29\x29\x5b\xf6\xbe\xea\x09\x2d\x20\xed\xc3\x9c\x6c\x71\x68\xaa\x14\x85\x3e\x30\x74\x1d\x44\x21\xcf\xba\xad\x89\xee\x4e\xbe\xa1\x65\xa9\x59\x21\x65\xc9\x33\xa2\x5a\x5a\xa4\x22\x42\x35\x75\x0f\xf4\xfc\x87\x76\xa2\x3e\x53\x2c\xa3\x3f\x13\x88\x12\x9a\x5a\x7b\x14\xb2\x5d\x56\xa2\x6c\x01\x68\x97\xd0\x0e\xd9\x10\x6a\x5e\xf4\xba\x58\xea\xec\x82\xae\xfb\x81\x50\xc3\xc0\xe5\xa2\xad\xc9\xe6\x60\xeb\x6c\x14\xa8\x04\x45\x69\x8c\xd9\x74\x08\x12\x50\x44\xea\x83\x4d\xff\xba\x69\x07\x98\xf3\x4d\x17\x16\x64\x6a\xbc\x22\x94\x81\x34\x49\x41\xb3\xef\x5d\x66\x1e\x31\xcd\x00\xed\x19\x9d\x1a\x77\x37\x1f\x87\xb0\x91\xc3\xb3\x6f\xcf\x71\xc3\x38\xf6\xbc\xd6\x10\x9d\x3b\x4a\x10\x26\x89\xc1\x17\x6a\xce\x4b\xa8\xc8\xf6\x60\x98\xe2\x20\x91\xe5\xda\xa8\xa3\x61\x7c\x6d\x92\xa3\x22\xdb\x83\x7d\x4e\xe7\xab\x9f\x4b\x52\x5d\x66\xd6\xa4\x34\xba\xbb\xd6\x79\x44\x9b\x9e\x35\x42\x98\x87\xb6\xa3\x19\x2b\x22\x21\x45\x64\x90\xa3\xc2\x4c\x61\xde\x83\x83\x80\x5e\x4f\x83\xe3\xb4\xd4\x1c\x0e\xc6\x13\xee\xd8\x1f\x8b\x5d\x21\xc1\x12\xf5\x89\xb7\x59\xd1\x6c\x75\x38\x32\xa1\xab\x87\x7a\xaf\x8d\xc4\xfd\x5d\x83\x6b\xa2\xba\xaf\xb6\x5c\xa7\x8c\x6e\xcc\x1a\xa9\x78\xd5\x2d\xb2\x2f\xd6\xdd\x63\x70\x57\x86\xaf\x4c\x5d\xec\xeb\xc3\xb9\x7f\x78\xf2\x6d\xbd\xd6\x0a\x1f\xd6\xcd\x4a\xaa\xb7\x6e\x0a\xd8\xeb\x8d\xce\x98\xaf\x0d\x15\x08\x1b\x09\x5c\x00\xad\xb3\xee\x1d\x98\xa4\xa5\xa9\x06\x99\xf9\x5e\x6c\xe9\x7a\x73\x1c\xde\x95\x52\xf5\xf0\xec\x4c\xf3\x5c\xea\x4a\x30\x8c\x7d\xcf\x6f\x0b\x0d\xd9\x9a\x42\xb3\xe7\x6b\x49\xf4\x9e\x68\x66\xf4\xea\xae\xf6\x7c\x1f\x5b\xca\x60\x83\xd4\xcc\x76\x2c\x38\xdf\x20\x05\xc6\x37\x6d\xb4\xcf\x89\x9c\xeb\xd9\x26\xdc\xfb\x1f\x33\xf4\x9c\x48\x28\x69\x45\xbb\xfb\x44\x4e\x8b\x02\x4d\x60\x0f\x11\x3a\x54\x15\x9d\x19\x4b\x22\x3f\x9a\xd1\xfb\x27\xec\xb1\xf9\xfe\x35\x29\xd7\x69\xea\xd6\x24\xcf\x2f\x70\x37\x04\xf7\xb8\xf1\x06\xd7\xfc\x09\x4d\xbb\xef\xef\x9b\xdb\xdb\xc4\x98\x57\x15\xd5\xc7\xcb\x0f\xed\x73\x81\xfb\x2e\xfb\x9b\x14\x2b\xd4\x25\x65\x6a\x08\xf1\x77\x6d\xb7\xda\x19\x05\x8a\x99\xe0\xd5\x10\x6c\xff\xb0\xc7\x7d\xed\x57\xdc\xa4\x7b\xeb\x3b\xf6\x2d\x9e\xc7\x5e\xec\x22\x97\xe7\xed\x6b\x3e\x81\xb4\xe4\xd9\x93\x39\x56\xdb\x00\x82\x12\x74\xb9\x44\x61\xe8\xd6\x17\x2d\xdc\xaa\x7d\xa5\x68\x4f\x8b\xc0\xda\x1f\x17\x3f\x5b\x58\xe8\x72\xcc\x59\x79\x54\xae\xe5\xe1\x5f\x1a\x7b\x93\xbe\x49\xeb\xea\xfd\xbd\xbc\xed\x77\xea\x7f\xfb\x22\x40\xd8\x0e\x72\x4c\x9b\xe5\xb2\x3b\x8c\x75\x6e\x9a\x32\xbf\xe4\xa0\x1d\xd1\x33\xbd\x6d\x0d\x40\x66\xd2\xc9\xb4\xe8\x53\x50\xcf\xe9\x81\xfe\x6b\x08\x05\x29\x25\x9a\x51\x75\x2d\x78\xd1\x92\xbc\x17\xd6\x97\x01\xdd\xba\x1f\xd6\x6b\xd1\xea\xfe\x0d\x53\x0b\xcc\x3a\xc2\x94\x68\xb0\xf7\x3f\x01\x00\x00\xff\xff\x99\x8d\x03\x22\x7b\x1a\x00\x00")

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

	info := bindataFileInfo{name: "go-centrifuge/build/configs/default_config.yaml", size: 6779, mode: os.FileMode(420), modTime: time.Unix(1574877192, 0)}
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

	info := bindataFileInfo{name: "go-centrifuge/build/configs/testing_config.yaml", size: 1263, mode: os.FileMode(420), modTime: time.Unix(1574948666, 0)}
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
