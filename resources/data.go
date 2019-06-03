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

var _goCentrifugeBuildConfigsDefault_configYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x99\x5b\x73\xdb\x38\xb2\xc7\xdf\xf5\x29\xfa\x28\x2f\xc9\xa9\x89\x4c\x82\x77\x55\x9d\x07\xea\xe6\x24\x8e\x1d\xd9\xf2\x65\xe2\x97\x53\x20\xd9\x94\x10\x91\x04\x03\x80\xba\xf8\xd3\x6f\x01\xa4\x14\xe5\xe2\xcc\x56\x66\x67\xaa\xb6\x76\xf3\x62\x15\x88\xfe\xa3\xd1\xfd\xeb\x06\x89\xbc\x80\x09\xe6\xb4\x29\x14\x64\xb8\xc1\x82\xd7\x25\x56\x0a\x14\x4a\x55\xa1\x02\xba\xa4\xac\x92\x0a\x04\xab\xd6\x98\xec\x7b\x29\x56\x4a\xb0\xbc\x59\xe2\x15\xaa\x2d\x17\xeb\x21\x88\x46\x4a\x46\xab\x15\x2b\x8a\x9e\x11\x63\x15\x82\x5a\x21\x64\x9d\x6e\xd5\xce\x94\xa0\x56\x54\xc1\xf8\xa8\x00\x25\x65\x95\xd2\xfa\xbd\xc3\x94\x61\x0f\xe0\x05\xbc\xe7\x29\x2d\x8c\x0b\xac\x5a\x42\xca\x2b\x25\x68\xaa\x80\x66\x99\x40\x29\x51\x42\x85\x98\x81\xe2\x90\x20\x48\x54\xb0\x65\x6a\x05\x58\x6d\x60\x43\x05\xa3\x49\x81\x72\xd0\x83\x83\xbd\x96\x04\x60\xd9\x10\x1c\xc7\x31\xbf\x51\xad\x50\x60\x53\x76\x3b\x78\x9b\x0d\x21\x74\xc2\xf6\x59\xc2\xb9\x92\x4a\xd0\x7a\x8e\x28\x64\x6b\xfb\x1a\xfa\x67\xac\x76\xcf\x6c\x12\x0c\xac\x81\x35\xb0\xcf\x54\x5a\x9f\x39\x21\xb1\xc8\x19\xab\x73\x79\x76\x5d\xde\x5e\xef\x92\xed\xba\x79\xfc\xf8\x71\x92\x37\x4f\xb7\xc9\x6e\x1a\xdf\xe0\xed\xd5\xf8\x3d\x7f\xda\xef\x3d\x2f\xdc\x5c\x57\xcb\xfb\xcd\xfc\xf2\xd3\xfb\x8f\xeb\xfe\x1f\x88\x3a\x07\xd1\xfb\xdc\x9f\x5e\xf9\xe5\xfa\xf3\x03\x7e\x7a\xb8\x78\x20\x9f\xe7\x8d\xed\xff\x5e\x67\xe7\xce\xfa\x1d\xb7\x6f\x9d\x72\x45\x57\xf3\x91\xb7\x40\xaf\xb2\x5b\xd1\x43\xa8\xe2\x43\xa4\xda\x0d\xe8\xed\x63\xa5\x98\xda\xcf\x68\xaa\xb8\xd8\x0f\xa1\xdf\xef\x9e\xd0\x2a\x5d\x71\x71\x83\x35\x97\xec\x9b\x47\xac\xda\x70\x96\xe2\x5d\x55\x53\x1d\xbe\x7e\xbf\x67\xb2\x73\x49\x59\xf5\x43\x56\xba\x24\xc2\xcb\x9b\x16\x96\x57\x3d\x38\x85\xa3\xf5\xe5\x05\x5c\x35\x25\x0a\x96\xc2\xdb\x09\xf0\xdc\x80\x72\x82\x44\xa7\x71\xcc\x99\x67\x77\x56\xa3\x43\x62\xa0\x60\x52\x69\xcb\x8a\x67\xf8\x3d\x53\xb5\xe0\x1b\x66\x1e\x70\xa3\x7d\xe2\xc0\xc1\xd1\x3f\x4c\xb4\xe3\x0d\x08\xf1\x06\xc4\xb2\x06\x2e\xf9\x36\xd9\x36\x99\x38\x17\x9c\x3f\x5c\xc9\x47\xf9\x10\xdc\x26\xe9\xa3\x17\x5e\x05\xf6\xdd\xf5\xe2\xc2\x9b\x7c\x7a\xfc\x5c\xce\xd6\x6f\xe6\x6f\xb6\xbb\xd9\xc5\x6d\xbc\xe7\x77\x77\x93\x30\xcb\xfb\x3f\x92\x0f\xfd\x01\xb1\xad\xe7\xe4\x27\x48\xe4\xf6\x61\xea\xe4\x84\xbd\x4b\xee\xf0\x3a\x3a\xbf\xbb\xbb\x1e\xbd\x19\x8b\x87\xf7\xc9\x28\xa5\xd1\xe5\xf9\xe5\xe7\xbc\x4c\xc6\x4b\xd1\x24\xfd\x2e\x46\xd3\x0e\xec\x63\x26\xde\x4e\xe0\x35\x74\xd9\x78\x0e\x7d\xb7\x33\x7e\x4f\x75\x78\x20\xc3\xba\xe0\x7b\xcc\x60\x51\x52\xa1\x60\xdc\x11\x25\x21\xe7\xc2\x04\x74\xc9\x36\x58\x7d\x15\xca\xef\xa9\x83\x67\xb1\xb3\x76\x79\x18\x5a\x49\xe8\x5b\xb6\xe5\x24\x99\xeb\x51\x8f\x38\x5e\xe0\xc6\x88\x63\x2b\x18\xbb\x11\xb1\x1c\x3b\x77\x83\xd0\xfe\x09\xa0\xd6\x2e\x22\xf1\xc4\x75\x47\xa3\x70\x46\x9c\x89\x97\xd9\x24\xc2\x51\x48\xa8\x67\x65\x4e\xe8\x87\xc9\xc8\x4d\xec\x14\x67\xf6\xec\x39\x94\xad\x5d\xea\xc6\x21\x8e\x48\x90\x8f\x9c\x29\x25\x63\x2b\xf2\xbc\x59\x48\xbd\x91\xed\xdb\xde\x88\xf8\x59\xe8\xcd\xc6\x23\x0c\xb1\x83\xfe\x82\x6f\x68\xbb\xeb\x13\x44\x13\x14\x15\x2d\x56\xc8\x96\x2b\x25\x7f\x0d\x6f\xf2\x27\xf1\xfe\xca\x85\x7f\x1a\x70\xdb\x72\x07\xb6\xe7\x0e\xec\x70\xe0\x7d\xd7\xcd\x0e\x04\x2e\x92\x5d\x72\x31\x4e\x1e\x57\xd1\xbb\x7b\x25\xaf\xf7\xf7\xe7\xd9\xed\x5c\x50\xf7\xa6\x5e\xc4\xae\x4a\x36\xd2\xa7\x95\x6d\x7f\xda\x9e\xc7\xe4\xa9\xff\x03\x79\x6f\x60\x87\xde\x80\x38\xc1\x73\x0b\x5c\x97\x24\x5d\x94\x62\xca\xe8\xe2\xf2\xde\x5d\xde\x6d\x82\x87\xf3\x55\xbd\xbc\xd9\xf2\x70\xcb\x67\x0b\xf9\x66\xf5\x78\x9e\x9c\x33\x87\xc6\xe1\xee\xe7\x88\x9b\xec\x3c\x0b\x38\xf9\x0b\x08\xff\x09\xe0\xb6\xe3\x93\x69\x3a\xca\x43\x3f\x88\x88\xeb\x4c\x89\x9b\xc7\xd6\x74\xec\x12\x2f\x23\x68\x5b\xb1\x15\x12\xe2\xa4\xc1\xe4\xa7\x80\x07\x76\x68\x4d\x82\xc0\xb1\xad\x0c\xd3\x30\x1e\x91\x30\xa6\xa1\x45\xa6\xa9\x15\xcd\xf2\x98\x4c\x66\xbe\x8b\x91\x15\xa4\xcf\x03\x6e\x87\x8e\x1d\x58\x6e\x68\xfb\x6e\x98\x63\x9e\xa3\x1b\xb9\xd6\xcc\x99\xc4\x71\xe6\xd0\x20\x49\x93\xc4\x4a\xbd\x38\x9e\x75\x80\xdf\xf0\x5a\x2a\xfc\x0e\xf1\x8c\x2f\x6b\xaa\xd2\xd5\xaf\xd1\xed\xfc\x49\xba\x0f\xab\xc3\xcb\xdb\x0f\x93\x0f\x90\x0a\xa4\x0a\x41\x74\xae\x6a\xc2\x8d\xce\xab\xff\xb4\x8e\xde\x06\xe0\x39\xe0\x9d\xbf\x97\x77\x2b\x73\x22\x7b\x1a\x10\x87\x78\x63\xcc\xc6\xae\x3d\x75\x43\xcb\x73\xa6\x41\x40\xc2\x90\x86\xd1\x8c\x4c\x1d\xdb\xb6\xbd\x9f\xf2\x4e\xc6\xa1\x35\xb3\x27\x34\x9f\xd0\x80\xc6\x13\x4c\xc8\xd8\x0e\xbc\xcc\x1d\xb9\x4e\x1c\x7a\xa1\x1b\x38\x53\xdb\x0e\x6c\xe7\x79\xde\xdd\x28\xc1\xc8\xb1\xac\xb1\xe3\x8f\x73\x8f\x38\x61\x32\xf3\xa3\xa9\x3b\x76\x23\xcf\xb7\x66\xb3\x30\x0f\x66\x7e\x40\xa6\xee\xc9\x5b\x8c\x7e\x69\x39\xe5\x1d\x26\x1f\xe0\xea\xc3\x2d\xdc\x2d\xa6\xff\xd3\x03\xc0\x32\xa1\x22\xa5\x19\x0a\xae\x67\xfd\x52\x09\xd8\xd6\xb3\x6c\x7e\x85\x8f\x1d\x85\x03\x9b\x90\x81\x6d\x3f\xdb\x2e\xe3\xa5\x33\x4d\x63\x25\x3e\xde\x8f\x77\xdb\x27\x7f\xed\xcb\xdb\x88\x3d\x2e\x6e\x9e\xd4\x53\x34\x09\xf6\x77\x4f\xf5\x68\x7e\x33\x9d\x3d\x89\x3b\x7e\xdf\xff\x7e\x05\xe2\x92\x01\x21\xf6\xc0\xb6\x9f\xed\xf8\x17\xe7\x5b\xb6\xfb\x1d\xab\xe6\xf7\xf8\xfe\xf3\xfa\xdd\x45\x59\xbd\x59\xc4\xef\x26\x9f\x9e\xf2\x00\xcf\x2f\xb9\xaf\x04\x67\xcb\xc7\x5d\x19\xc4\xde\xcd\xcf\x09\x2d\xdb\xe8\x3e\x47\xa8\xfd\xf7\x12\x1a\xcf\x5c\xcf\x4f\x6d\xdf\x09\x7d\xea\xbb\x79\xe6\xce\xdc\xc4\x8f\x68\x6e\x3b\x34\xf4\x27\xb9\x35\xf2\x7c\x12\x53\xcb\xfa\x29\xa1\xbe\x13\x8c\xc2\xb1\x33\x21\x71\xec\x8c\x53\x62\xf9\x93\xc8\xf5\xec\x28\xf1\xdc\x30\x22\x56\x18\xa5\xd1\xd4\x0f\xa2\xc8\x7a\x9e\xd0\x91\x87\x2e\x71\xb2\x71\x1a\xb8\x56\x32\x1a\x87\x56\x1e\x59\xbe\xed\x38\x68\x7b\xbe\x65\xe7\x51\x68\x45\x51\xe8\x78\xfe\x37\x84\x7e\x41\xea\x04\xc8\x7f\x35\x8c\x7f\x35\x8a\xff\x05\xf1\xdf\x13\xc4\x17\x30\xa1\x8a\xc2\x42\x71\x41\x97\xd8\x93\xed\xdf\xf6\x33\x7d\x4e\xd5\xca\x44\xa6\xd0\x1f\x83\x93\x11\xe4\xac\xc0\x1e\x40\x4d\xd5\x6a\x08\x67\xaa\xac\xcf\xbe\x5c\x17\xfc\x7f\x46\x15\x1d\x98\x99\x59\xa2\x75\xc7\xbc\xca\xd9\xb2\x11\x54\x31\x5e\x1d\x17\x48\xcd\xe8\xe2\xd7\x97\x69\x05\xbe\x5b\x2d\x4e\x53\xde\x54\x4a\xc2\x1a\xf7\xd0\xed\xa2\x47\xbb\x41\xbd\xce\x1a\xf7\x7a\x18\x3b\xc5\xc3\x23\x6d\xfb\xb6\x52\x28\x72\x9a\x22\x6c\x35\x40\x06\x84\x78\xfe\x16\x68\x95\xc1\x9c\xcc\x61\x81\x62\x83\xc2\xbc\xda\x60\xa5\xdf\x5d\x7a\xfa\xad\xe4\x0d\x97\xaa\xa2\x25\x0e\xe1\xf8\x89\xdf\x7b\x01\x73\x2e\x54\x27\xa3\x25\x7e\x6c\xaa\x27\x0d\x21\xb4\x42\xa2\x97\xd7\x55\xfa\x5a\xf1\xd7\x35\xa2\x80\xf4\x34\x6a\xb2\x57\x93\xba\x0d\xd2\xa2\xc6\x94\xe5\x7b\x98\xee\x94\xf9\x22\x80\xb7\xf3\x13\x6f\xb5\x28\xa4\xb4\x82\x04\x41\x20\x4d\x57\x98\x01\x55\xc0\x72\x48\x70\xc5\xaa\x0c\xae\xe2\x5b\x2d\x83\x9d\xf5\xdb\xf9\x10\xb6\x83\xdd\x60\x3f\x78\x6a\x53\xa0\xbd\x6e\x24\x66\xc7\x42\xd0\xfb\x2e\xe8\x1e\x85\x4e\x84\x71\xd7\x94\xb1\x99\x7d\xcb\x4a\xe4\x8d\xd9\x66\x05\xbc\xc6\xaa\xbb\xc5\xa9\x30\x35\x5e\xeb\xb7\x3b\xbd\x19\xd9\x83\xc3\x70\x67\x32\x84\xbe\x63\xc9\xbe\x51\x29\x59\xc5\xca\xa6\x84\x0c\x0b\xba\x37\xeb\xe2\x06\xc5\x1e\x6a\x52\x83\x40\x59\xf3\x4a\xa2\x56\xa2\x1b\xce\x32\x50\xac\xd4\xab\x50\xa5\x68\xba\x96\x46\x80\x66\x9f\x1a\xa9\x20\xa1\xda\x6f\x5e\xc1\x8a\x4b\xa5\x2d\x79\x23\x52\x94\xf0\x72\xb1\x98\xfc\x06\xe3\xf9\xdd\x6f\x90\x72\x81\x12\x06\x83\xc1\xab\xee\xfa\x89\xaf\x81\x55\x50\xf0\xa5\xa9\xfc\x21\xf4\xb5\x7f\xda\x57\xd9\x94\x98\x41\xb2\xd7\xdb\x6a\x73\xd0\xd7\x51\xdc\xfd\xdf\xcb\x0d\x2d\x1a\xbc\x41\x9a\xc1\xff\x02\x79\x05\x4c\x42\x81\xd2\xbc\xe1\x56\x60\x9e\x41\x82\x05\xdf\xfe\xa6\xa3\x57\x41\xba\xa2\xd5\x12\x8f\xfb\x98\x98\x3d\x2a\x0e\xbb\x1e\x7c\x3d\x38\x84\xbe\x67\x59\xa5\x34\xa5\x78\xdd\x60\x83\xdf\x20\x60\x22\x43\xe5\xbe\x4a\x57\x82\x57\xbc\x91\xfa\x25\x3a\x45\x29\x59\xb5\xec\x7d\xd6\x06\x2d\x20\xed\xbd\x9c\x6c\x71\x68\xca\x04\x85\x3e\x30\x74\x1f\x44\x21\xcf\xba\xad\x89\xee\x95\x7c\xcb\x8a\x42\xb3\x42\x8b\x82\xa7\x54\xb5\xb4\x48\x45\x85\x6a\xea\x1e\x68\xfb\x87\xd6\x50\x9f\x29\x96\xd1\x9f\x09\x44\x09\x4d\xad\x23\x0a\xe9\x3e\x2d\x50\xb6\x00\xb4\x4b\xe8\x80\x6c\x29\x33\x17\x7a\x5d\x2e\x75\x75\x41\xf7\xf8\x81\x32\xc3\xc0\xe5\xa2\xed\xc9\xe6\x60\xeb\x7c\x14\xa8\x04\x43\x69\x9c\xd9\x76\x08\x52\x50\x54\xea\x83\x4d\xff\xb9\x69\x27\x98\xf3\xad\x77\x72\x10\x48\x53\x13\x2c\xfd\x3a\x62\xbd\xc3\x31\xd0\x15\x0e\x16\xa8\x3b\xfc\x76\xc5\xd2\xd5\xf1\x88\x80\xae\xfe\x75\x4e\x1a\x89\x87\xb3\x95\xeb\x08\x76\x1f\x29\x99\x46\x44\x0f\xa6\x8d\x54\xbc\xec\x16\x39\x34\xa7\xee\xee\xb3\x6b\x3b\x57\xa6\x0f\xf4\xf5\x61\xd4\x3f\xde\x70\x9a\xbe\xd7\x09\x1f\xd7\x4d\x0b\x86\x95\x6a\x0b\xf6\xe5\x56\x13\xf2\xb9\x61\x02\x61\x2b\x81\x0b\x60\x75\xda\x5d\x7b\xd2\xa4\x30\xf4\xa7\xe6\xf3\xa8\x8d\xa6\xa6\x57\x1b\xde\xdd\xbc\x1f\xc2\x4a\xa9\x7a\x78\x76\xa6\xf3\x57\x68\xf2\x87\x91\xe7\x7a\x6d\x61\xd1\x9d\x29\xac\x43\x3c\x97\x54\xef\x89\xa5\x46\xaf\xee\x6a\x8d\x82\x12\xb4\x92\xd4\x54\xac\xde\xe9\x16\x99\xb1\x26\x16\x9c\x6f\x91\x41\xc5\xb7\x3d\xd0\x5a\xe7\x54\xce\xb5\xf5\x10\x88\x75\xfc\x67\xa6\x9e\x53\x09\x05\x2b\x59\x77\x7e\x66\x2c\xcf\x51\xe8\xdd\x1d\x33\x74\xac\x22\x4d\xc2\x92\xca\xf7\x66\xf6\xe1\xc6\x76\x6c\x3e\xf7\x0c\x62\x9d\xa6\x1e\x8d\xb3\xec\x02\xf7\x43\x70\x4e\x07\x6f\x70\xc3\xd7\x68\xc6\x3d\xef\x30\xdc\x9e\x9e\x63\x5e\x96\x4c\xb7\xd3\x6f\xc6\xe7\x02\x0f\x8f\xec\x2f\x52\x55\xae\x2e\x59\xa5\x86\x10\x7d\x35\x76\xab\x83\x91\xa3\x98\x09\x5e\x9e\xcc\xff\xd2\xeb\x14\x37\x78\xb7\xb1\xab\xbe\xe4\xf3\x34\x8a\x5d\xe6\xb2\xac\xbd\xbc\xa6\x90\x14\x3c\x5d\x9b\x63\xa4\x4d\x20\x28\xc1\x96\x4b\x14\x98\xb5\x9d\x51\xe1\x4e\x1d\x2a\xa3\xed\x8e\xbe\x75\x68\x8f\x3f\x5a\x58\xe8\xf6\xc3\xab\xe2\xa4\x3d\xc9\xe3\x0d\xfe\xc1\xa5\x2f\xd2\xba\x5b\x7d\x2d\x6f\x7b\x9d\xfa\x95\xe6\xef\xd4\xf7\x9a\xf3\x42\x67\xfb\x58\x8d\x8a\x83\xc4\x2a\xfb\x06\x14\xbe\x31\x27\x42\x49\x77\xc7\xa2\x24\x5d\xa4\x7e\x2c\xc9\xf4\xd9\xba\xa1\x85\xd1\xdd\xb7\x1d\x83\x6a\x07\xd3\x46\x18\x56\x4e\x2d\x56\x54\x42\x82\x58\x41\x86\x0a\x53\x65\xc2\x74\x10\xd0\xeb\xe9\x6e\x49\xda\x56\x49\xab\x3d\x64\x98\x34\xcb\x65\x77\xf8\xe8\xda\x34\x6d\x6d\xc9\x41\x07\xa2\x67\x9e\xb6\x3d\x00\x2b\x53\x4e\x66\x44\x77\x7d\x6d\xd3\x03\xfd\x6b\x08\x39\x2d\x24\x9a\x59\x75\x2d\x78\xde\x92\x7c\x10\xd6\x87\x9f\x1e\x3d\x4c\xeb\xb5\x68\x75\xff\xeb\x50\x0b\x4c\x3b\xc2\x94\x68\xb0\xf7\x8f\x00\x00\x00\xff\xff\x00\x8f\xc7\xbd\x6a\x19\x00\x00")

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

	info := bindataFileInfo{name: "go-centrifuge/build/configs/default_config.yaml", size: 6506, mode: os.FileMode(420), modTime: time.Unix(1559587696, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _goCentrifugeBuildConfigsTesting_configYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x94\x53\xb9\x8e\xe4\x36\x10\xcd\xf5\x15\x04\x1d\x4c\xd2\x07\x4f\xf1\xc8\x1c\x1a\x0b\x3b\xb1\x81\x8d\x8b\x64\xb1\x87\xe8\xd1\x61\x92\x9a\xd9\xc6\x62\xff\xdd\x50\x6f\x8f\xbd\x99\x67\xb3\xaa\xd2\x3b\xaa\xa4\xa7\x88\x73\xaf\x25\x6f\x17\xfc\x03\xfb\xdb\x52\xaf\x9e\x74\x6c\xbd\xcc\x97\x01\xfb\x33\x56\xdc\x26\x3f\x10\x02\x31\x2e\xdb\xdc\xdb\x5e\x13\x32\x41\x99\x3d\xb9\x97\x84\x5c\xf1\xe6\xc9\xd3\x57\x0a\x29\x55\x6c\x8d\x7a\x6a\x5d\x60\x60\x47\x6d\x65\x54\x4a\x29\x88\x39\x19\x1e\xd4\x28\x91\x25\x19\xb5\x06\xe4\x8a\x0b\xd0\xf4\x40\x63\xbd\xad\x7d\xa1\xfe\x2b\x8d\x65\x7d\xc6\x4a\x3d\x05\x6c\x47\x2e\xec\x31\xf6\xba\x03\xee\xe3\x8e\x5f\x3a\xf5\x34\x1a\xe3\xb2\x95\xc6\x25\x63\x58\x72\x22\xe6\xc8\x53\x4a\x0a\x6c\x96\x3c\x69\x60\x90\xa2\xcd\x02\x58\x10\xc0\x15\xe3\xd2\xb0\x24\x47\xc9\xb2\xb4\x91\x45\x0b\xff\xea\xad\x50\x61\x6a\xbb\x6d\x79\xa5\x9e\xca\x31\xf2\xd1\xa2\x91\x21\x3b\xcb\x32\x1a\x1d\x98\x11\x26\x5b\xc7\xc0\x70\x48\xf4\xdb\x81\x5e\x53\xa6\x9e\xb6\xfb\xc2\xf4\xde\xfe\x27\x92\xae\x2f\x38\x53\x2f\xc5\x81\xce\xd4\x8b\x51\x70\xa5\x0e\x74\xa5\x9e\x1f\x68\xa5\xde\x1e\x68\x83\x97\xfd\x80\x84\x3c\x20\x1f\x51\x46\x67\xb9\x53\x2a\x71\x8c\x20\x82\x0d\xc2\xa0\xc2\x11\x59\xd0\x21\x07\x25\x03\x32\x69\x46\xd0\xc9\x5a\xeb\x32\x8c\xc6\x81\xb0\x5c\x88\x7d\x91\x09\xe2\xfe\x2a\x22\x17\x36\x58\xae\xb5\xd6\x01\x38\x42\x32\x11\xd0\xb1\x91\xa1\xb5\x4a\x40\x8e\x60\xa5\x1e\x13\x1b\x95\xd6\x21\x39\xd0\x46\x8b\x00\x63\x8e\x91\x39\x81\x79\x57\x2a\x89\x7a\xaa\x34\xb2\x91\xc1\x78\x4c\x02\xf0\xa8\x64\xb0\x47\x27\x44\x3e\x2a\x65\x85\x53\xce\x25\x69\x12\x3d\xd0\x57\xac\xad\x2c\xfb\x91\xdf\x9e\x1e\x1f\x7e\x85\xd6\xde\x96\x9a\x3c\x79\x7a\x1f\x3d\x32\xe0\xc9\x47\x23\x30\x0c\x25\xe1\xdc\x4b\xbf\xfd\x96\x3c\xa1\xec\xcb\x87\xb3\x33\x0c\xbf\x90\x5f\x1f\xa9\xdc\x33\x48\x5a\x5f\x2a\x5c\x70\xf8\x31\xaa\x57\xbc\xed\x63\xf4\xe4\xdc\xa7\xf5\xfc\xfe\x68\x18\xfe\xde\x70\xc3\x1d\x31\x6f\xd3\xe7\xa5\x5e\xb1\x36\x4f\xc4\x40\xc8\xdb\xbd\xf9\x0c\xa5\xff\x55\x26\xfc\xfd\x4f\x4f\xf8\x30\xec\x32\x3b\x78\x15\xeb\xf7\x1f\x60\xdd\xc2\x4b\x89\x9f\xf6\xe4\x9f\x4e\xe7\xd3\xe9\x1c\xb6\xf2\x92\xce\x15\xdb\xb2\xd5\x88\xed\xbc\x8a\xf5\x13\xde\x4e\xeb\x16\x4e\x2b\x4e\xdf\x39\xb5\xbc\x42\xc7\xff\x27\x5d\x77\xe2\x9d\xd4\xca\x65\x2e\xf3\xe5\x83\x9e\x0f\xf4\xcf\xfb\xfe\x40\x7c\xf7\x1e\x60\x8e\xcf\x4b\x7d\x98\xaf\x15\xe3\x32\x4d\xa5\x7b\xd2\xeb\x86\xff\x04\x00\x00\xff\xff\x1f\xaf\xbe\x5d\x34\x04\x00\x00")

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

	info := bindataFileInfo{name: "go-centrifuge/build/configs/testing_config.yaml", size: 1076, mode: os.FileMode(420), modTime: time.Unix(1552484064, 0)}
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

