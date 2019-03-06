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

var _goCentrifugeBuildConfigsDefault_configYaml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xc4\x58\x5b\x73\xdb\x36\x94\x7e\xd7\xaf\x38\xe3\xbc\xb4\x33\x4b\x19\x04\xaf\xd2\x4c\x67\x47\xbe\xe4\xd2\x38\xae\x7c\x49\xdd\xf8\x65\x03\x02\x87\x12\x62\x0a\x60\x00\x50\x97\xfc\xfa\x1d\x80\x94\x6b\xc7\xb1\xbb\xdb\x4e\x77\xfd\x62\x0a\xc0\xb9\x7f\xe7\xc3\xe5\x15\x9c\x60\xcd\xba\xc6\x81\xc0\x35\x36\xba\x5d\xa1\x72\xe0\xd0\x3a\x85\x0e\xd8\x82\x49\x65\x1d\x18\xa9\xee\xb0\xda\x8d\x38\x2a\x67\x64\xdd\x2d\xf0\x1c\xdd\x46\x9b\xbb\x29\x98\xce\x5a\xc9\xd4\x52\x36\xcd\x28\x28\x93\x0a\xc1\x2d\x11\xc4\xa0\x57\xf5\x2b\x2d\xb8\x25\x73\x70\x7c\xaf\x01\x56\x4c\x2a\xe7\xf5\x8f\xf6\x4b\xa6\x23\x80\x57\x70\xa6\x39\x6b\x82\x0b\x52\x2d\x80\x6b\xe5\x0c\xe3\x0e\x98\x10\x06\xad\x45\x0b\x0a\x51\x80\xd3\x50\x21\x58\x74\xb0\x91\x6e\x09\xa8\xd6\xb0\x66\x46\xb2\xaa\x41\x3b\x1e\xc1\x5e\xde\xab\x04\x90\x62\x0a\x49\x92\x84\x6f\x74\x4b\x34\xd8\xad\x86\x08\xde\x89\x29\x94\x49\xd9\xcf\x55\x5a\x3b\xeb\x0c\x6b\xe7\x88\xc6\xf6\xb2\x11\x1c\x1c\xca\x36\x3d\x8c\x69\x31\x26\x63\x32\x8e\x0f\x1d\x6f\x0f\x93\x92\x12\x7a\x28\xdb\xda\x1e\x5e\xac\xae\x2f\xb6\xd5\xe6\xae\xbb\xfd\xf4\xe9\xa4\xee\xbe\x5d\x57\xdb\xd3\xd9\x25\x5e\x9f\x1f\x9f\xe9\x6f\xbb\x5d\x96\x95\xeb\x0b\xb5\xf8\x7d\x3d\xff\xf0\xe5\xec\xd3\xdd\xc1\x5f\x28\x4d\xf6\x4a\x7f\xaf\xf3\xd3\xf3\x7c\x75\xf7\xf5\x06\xbf\xdc\xbc\xbf\xa1\x5f\xe7\x5d\x9c\xff\xd1\x8a\x37\xc9\xdd\xaf\x3a\xbe\x4e\x56\x4b\xb6\x9c\x1f\x65\x57\x98\xa9\xb8\x57\xba\x4f\xd5\x6c\x9f\xa9\x3e\x00\x1f\x3e\x2a\x27\xdd\xee\x35\xe3\x4e\x9b\xdd\x14\x0e\x0e\xbe\x9b\xb9\xc4\x85\xb4\xee\xd1\x14\x53\x7c\xa9\xcd\x25\xb6\xda\xca\xef\xa4\x5a\xb6\xf3\x30\xf9\xad\x6a\xe4\x82\x39\xa9\x55\x98\x0b\xc5\xfb\xc0\xa4\xfa\x21\x94\x86\x1a\xc3\x4f\x97\x3d\x96\x7e\x1e\xc1\x43\xec\xf4\xae\xbe\x82\xf3\x6e\x85\x46\x72\x78\x77\x02\xba\x0e\x38\x7a\x80\x98\x41\xc7\x7d\x49\xb3\x78\x90\x3a\xda\xd7\x0d\x1a\x69\x9d\x97\x54\x5a\xe0\x53\xc8\xb5\x46\xaf\x65\x98\xd0\x41\xf7\x03\x07\xf6\x8e\xfe\x25\x0e\x92\x6c\x4c\x69\x36\xa6\x84\x8c\x53\xfa\x3d\x16\x62\x7a\x92\xbc\xd7\xfa\xe6\x4c\x4a\x7e\xf1\xfb\xe6\x7a\x79\x7d\xf4\x29\xdf\xbe\xe7\x73\x7d\x56\xe7\x97\x17\x9f\x7e\x7d\xdd\x6e\xea\xd8\x14\xd9\xe6\x6c\x4b\x6f\x2f\x93\xf6\x58\xc4\x07\x3f\x52\x5f\xe6\x63\x1a\x93\xe7\xd4\x5f\xdc\x7e\x98\x95\x6f\xe6\x6f\xcd\xfa\xf4\xf6\x68\xb2\x11\x77\xfa\x23\x9f\xcd\x56\xc7\xb7\x6f\xdb\x09\xee\x76\xb7\xe9\xd5\x69\xb9\x78\x6d\x92\xe5\xf5\xf9\x1f\x07\x43\x8e\x4e\x07\xdc\xdf\x57\xe2\xdd\x09\x44\x30\x54\xe3\xb9\xce\x48\x07\xe1\x33\xe6\xd3\x03\x02\xdb\x46\xef\x50\xc0\xd5\x8a\x19\x07\xc7\x03\xe0\x2c\xd4\xda\x84\x84\x2e\xe4\x1a\xd5\xa3\x54\x3e\x05\x25\x3c\x8b\x4a\xb2\x8d\x2b\x92\x91\xb4\xc0\x32\x8e\x4b\x96\x0b\x56\xa7\x69\x5a\x53\x9a\x66\x79\x4e\xea\x4c\x24\x75\x9a\x53\xc2\x26\xe2\x05\xfc\x92\xed\x24\xcf\x09\x27\xc9\x44\x24\x71\x9c\x66\x09\xab\x89\xc8\x4a\x9e\xe5\x79\x5e\xd0\x44\x4c\x38\xad\x59\x21\x72\xe4\x2f\x20\x9d\x6c\x71\x52\x56\x84\xe7\x42\x4c\x08\xcd\x8b\x34\xa9\x8b\x24\xcf\x0b\xac\x59\x5d\xb2\x49\x46\x04\x11\x65\x9d\x4e\x04\x7d\xa9\x27\xc8\xb6\xe2\x31\xc3\x82\xd6\x55\x95\x33\x4a\x4a\x9a\xd5\x59\x9c\xe4\x9c\x57\x8c\x60\x99\x88\xbc\x2a\xb1\x28\x53\x1f\x4f\xe8\x9e\xf7\x7a\xcd\xfa\xf4\x3d\xc0\x7a\x85\x46\xb1\x66\x89\x72\xb1\x74\x03\x16\x5f\xbd\x7a\x35\x14\xa6\x97\x78\x3d\xbb\x18\x7e\x47\x70\xe3\xe9\x50\xaa\xba\x33\x0c\x76\xba\x83\x85\xe7\x71\x05\x68\x8c\x36\x1e\x65\xd7\x4b\x69\xc1\xe0\xd7\xce\x5b\x91\x16\x94\x76\x60\xbb\xb6\xd5\xc6\xa1\x80\x0a\x39\xeb\x2c\x7a\x49\x13\x9a\xc8\x2f\x31\x9d\x52\x9e\x8b\x03\xd3\x5a\xc7\x9c\xef\xa4\xce\x0f\x8d\xe1\xb2\x53\xfd\x78\x14\x0d\x63\xbf\x30\xc3\x97\x72\x8d\xe3\x83\xff\x18\x9c\x02\xd8\xf8\x46\x74\x1a\x84\xfe\xcf\x20\xc1\xa0\x09\x2c\xdf\x32\x23\xdd\xae\x37\x14\xb4\xdc\x85\x78\x70\x31\xed\x7f\x7e\x1e\x16\x44\x11\x5f\x32\xa9\x7e\xe9\xa7\xa3\xc8\x7b\xfb\x4b\x42\x12\x92\x42\x14\x6d\x98\x69\x87\x7f\x51\xc5\x8c\x91\x68\x20\xcb\x4b\x42\x08\x81\x28\x52\x3a\x62\x8a\x4b\x54\x2e\xaa\x1a\xcd\xef\x6c\x3f\x66\xd1\xac\x31\x6a\x7c\x52\x21\x8a\x56\x6c\x1b\xb5\xbe\xd7\x81\x66\x5e\xc8\x2a\xd6\xda\xa5\x76\xc3\x60\x18\x5b\x49\xf5\xe8\xa7\xf7\x99\x71\x27\xd7\x08\x51\xe4\x31\xee\x53\xa4\xeb\xfa\x69\x26\x20\x8a\x44\x15\x71\xbd\x6a\xfd\x7a\xad\xc0\x5a\xe1\x43\x62\x7c\x89\x91\x95\xdf\x10\x52\x32\xc9\x21\x8a\xbe\x58\xad\x4c\xcb\xa3\xa5\xb6\xce\x02\x6b\x9a\x07\x63\x52\x39\x34\x35\xe3\xe8\xc7\x3f\x3f\x2e\xf7\xd3\x64\xfe\xa8\xf2\x47\x3e\x7c\x14\xbe\x25\x15\xf6\x8e\x38\x0d\x37\x58\x5d\xf9\x71\x67\x21\xe4\xc4\x40\x6d\xf4\x0a\x3a\xe5\x4c\x67\x3d\x24\xb4\x91\x0b\xa9\xa6\x30\x1e\x1f\x3c\x5b\x4f\xdf\xfb\x4f\x6a\xf9\x39\x8a\x3a\x65\x59\x8d\x11\x6e\x5b\x6d\xf1\x33\xd4\x0d\x5b\x7c\x07\xe0\xff\x1d\xe1\xd3\x7f\x48\xf8\x8f\x7a\xe9\x7f\x4c\xf9\x31\x49\xc7\x71\x96\x8e\xe3\x72\x9c\x3d\xd9\xfe\xf7\x9c\x3c\xb7\xb9\x64\xf8\xb1\x7b\x7d\x7b\xde\xc5\x6f\xb6\x6b\xbb\x3b\xba\xbe\x32\xd7\x76\xb2\x76\x47\x79\xe5\x3e\xcc\xd4\xdb\xd7\xfa\xec\x4b\x75\xf7\xed\x98\x1d\xfc\x40\x7d\x36\x8e\xcb\x6c\x4c\x93\xe2\x59\x03\xc7\x6f\xf8\x46\x5e\x7f\xd1\xef\x6f\xde\xd6\x47\x2c\x2d\xe9\xc7\xb9\x63\xf8\x71\x7b\x7e\xb6\x11\xe5\xb7\x4a\x1d\xc5\x57\xc5\x06\x67\xb7\x1f\xb7\xb7\x2f\x93\x7e\x20\x8d\x67\x29\x9f\xfe\x0b\x9c\xff\x02\xe5\x4f\xf2\x3c\xce\x4b\x12\x97\x98\x12\x9e\xc6\x65\x91\x8a\x32\xe7\x28\xea\x8a\x97\x04\x31\xaf\xaa\x92\x65\x45\x9e\xbe\x48\xf9\x59\xca\x30\x29\x92\x9a\x4c\xf2\x9a\xd5\x54\x54\x79\x55\xb2\x34\x2f\xe2\x82\x93\x6a\x52\x22\xaf\x19\x29\x32\x21\x5e\xa4\xfc\xba\x28\xe2\x98\x08\x91\x23\x2d\x2b\xca\x59\x91\xc5\x49\x41\x58\x5e\x97\x94\x60\x45\x8b\x22\x15\x55\x52\xc6\xe4\x65\xca\xa7\x45\x11\xd3\x09\x47\x5a\x57\x05\x25\x05\x27\x35\xc1\x94\xe5\x44\x50\x41\xaa\xac\x2e\xbd\x2b\x55\x55\x55\x03\xe5\x5f\xea\xd6\x3a\x7c\x42\xfa\x42\x2f\x5a\xe6\xf8\xf2\xef\x9d\x8b\x92\x7f\xd8\x26\x7b\xeb\xf0\xd3\xf5\x6f\x27\xbf\x01\x37\xe8\x39\xdf\x0c\xae\xfa\x56\x09\x7a\x7e\x7e\xb6\x73\xfe\xf5\xe3\xd2\xff\xdf\x81\xa9\x4f\xc2\x73\xdd\x93\xfc\xdf\x36\x4f\xc9\xb2\x94\xa7\x85\xc8\xf3\x8c\xc6\x98\xd7\x82\x8b\x3a\x89\x79\x4a\x63\x24\x79\x4c\xe3\x2c\x15\x75\x99\xe6\xd5\xcb\xcd\x93\xc5\x39\x39\x8a\x29\x99\xc4\x3c\x8d\x4f\x8b\xe3\x34\xae\xc9\xd1\x2c\x9b\xe1\x84\x66\x8c\xa4\x34\xe7\x29\x9b\x9d\xce\x5e\x6e\x9e\x9c\xd2\x22\x2d\x0b\x12\xf3\x98\x66\x28\x28\x4d\x62\xce\xeb\xa4\xe6\x75\x55\x30\x5a\x91\xb8\x66\x65\x92\x65\x2f\x37\x8f\x10\x69\xc1\x90\x15\x69\xc1\x29\x25\xa9\xa0\x69\x4a\xfc\x89\x29\x41\x4c\x49\xc5\x69\x95\xd1\x8a\x4f\x44\x7d\x30\xf2\x97\x4d\xe6\x18\x5c\x39\x6d\xd8\x02\x47\xb6\xff\xdf\x5f\x21\xe7\xcc\x2d\x43\x8a\x1b\x7f\x13\x39\x39\x82\x5a\x36\x38\xf2\x46\xdd\x72\x0a\x87\x6e\xd5\x1e\xfe\x79\x95\xfd\x2f\xc1\x1c\x1b\x87\x95\xa2\xf2\x7a\x8f\xb5\xaa\xe5\xa2\x33\xc1\xad\x7b\x03\x3c\x8c\x5e\xfd\x7d\x33\xbd\x82\x27\xd6\x66\x9c\xeb\x4e\x39\x0b\x77\xb8\x83\x21\x8a\x11\x1b\x06\xbd\x9d\x3b\xdc\xf9\x61\x1c\x34\xee\xa7\xbc\xec\xbb\xfb\x33\xc1\xc6\x23\x31\x20\x6a\x36\x7f\x07\x4c\x09\x98\xd3\x39\x5c\xf5\x1b\xba\x6f\x7e\x54\xbe\xbb\x47\xbe\x6f\xdf\x6a\xeb\x14\x5b\xe1\x14\x48\xb8\x7c\x92\xd1\x2b\x98\x6b\xe3\x06\x25\x5e\xc1\x8f\x05\xfd\xa2\x29\x94\xa4\xa4\xde\xb8\x6f\xf7\xc8\xe9\x70\x26\x02\xfe\x30\x67\x76\xd4\xd2\xb6\x4f\xd1\x55\x8b\x5c\xd6\x3b\x38\xdd\xba\xb0\xf5\xc2\xbb\xf9\x03\x5f\xc3\x59\x81\x33\xe5\xaf\xf2\x06\xfd\x71\x48\x00\x73\x20\x6b\xa8\x70\x29\x95\x80\xf3\xd9\xb5\x57\x83\x83\xf4\xbb\xf9\x14\x36\xe3\xed\x78\x37\xfe\xd6\x17\xc0\x7b\xdd\x59\x14\xf7\xfd\xe4\xa3\x6e\xd8\x0e\x8d\x2f\x43\x70\x37\xb0\x41\x58\x7d\x2d\x57\xa8\xbb\x10\xa6\x02\xdd\xa2\x1a\xde\x17\x86\xc3\x50\x60\xbf\x70\xc0\x1b\xc1\x7e\x78\x10\x99\xc2\x41\x42\x6c\x00\xdd\x45\x87\x1d\x7e\x17\x6e\xb0\xce\xec\x4e\xf1\xa5\xd1\x4a\x77\xd6\x13\x2a\x47\x6b\xa5\x5a\x8c\xbe\x7a\x81\x3e\x19\xfd\xeb\x88\xed\x43\xef\x56\x15\x1a\x4f\xc9\x9e\x3a\xd0\xd8\x43\xae\x95\xf5\x2c\x3f\xd0\xf3\xc6\x5f\x4a\xab\x70\xda\xd3\x9c\xb9\x3e\x33\xd6\x31\xe3\xba\x76\x04\x5e\xfe\xa6\x17\x9c\x42\x1f\xde\x6b\x83\x68\xa1\x6b\xe1\x78\xfe\x11\xf8\x8e\x37\x68\xfb\x50\x7b\x03\xfe\x20\xbf\x61\x32\x3c\xaa\x78\x7f\x71\x8d\x1e\x45\x30\x4c\xdf\x30\x19\xa2\xfd\x70\x35\x85\x78\x34\x6c\x39\x83\x87\x06\x9d\x91\x18\x0e\xa4\x7a\x33\x24\x9b\x81\x63\xd6\x6f\x39\xfe\xdf\x65\xbf\x60\x0a\x31\xf1\x39\xba\x67\x4e\x1b\xaa\x2f\xf9\xe3\x7c\x8d\xf6\xbc\x39\x40\x04\x1b\xf4\x94\xb8\x59\x4a\xbe\xbc\xe7\x54\x18\x70\xee\x8b\xe2\x2f\x24\xc3\xae\xa7\x7d\xfe\x86\xed\x4a\x80\xec\x4f\x9e\xbc\xb3\x4e\xaf\x06\x23\xfb\x26\x1c\xde\x9f\x86\xf6\x3a\x0f\x78\x3f\x58\x31\xa9\x0e\xee\x5f\x99\x42\x7f\x0f\x8a\xef\xed\xf2\xc6\xdf\x15\x7a\x68\xfe\xb4\xc1\x70\x55\x92\x06\x61\x63\x41\x1b\x90\x2d\x1f\x9e\x9e\x58\xd5\xa0\xff\xe4\x61\xa3\xec\xb3\xe9\x37\x44\x2f\xf8\xf1\xf2\x6c\x0a\x4b\xe7\xda\xe9\xe1\x61\x38\x9b\xfb\x03\xfd\x74\x92\xa5\xd9\x1e\x07\xe1\x69\x6c\xc1\x7c\x2c\x92\x7b\x77\x17\xcc\xce\xfd\xa7\xcf\xe1\xfe\xef\xc9\xe2\x46\xae\xa4\xeb\x17\x9f\xf9\xcf\x29\xa4\x45\x4c\x93\xb2\x7c\x84\x6f\xa7\x43\xa1\xfb\x32\xa9\x3f\x23\x73\x86\x29\xcb\xee\x0f\xfe\x3e\x06\x21\xfa\xa7\x34\x06\xe1\x6e\x14\x88\xa3\x0f\x05\x9c\x91\x8b\x05\x1a\x14\x7d\x37\x38\xdc\xba\x3d\x46\xfa\x8e\xc8\x89\x6f\x89\xe7\x0c\x1b\x64\x02\xb4\x6a\x76\xbe\xd3\xf6\x7d\xb2\x7f\x4f\xdc\xbb\xf4\xa7\xea\x4b\x64\xe2\xb1\xfa\x38\x1b\xb4\x9f\xfb\x4a\x3c\xf4\xbd\xd5\xba\x81\x15\xdb\xde\xe3\xd2\x69\xb0\xa8\x84\xc7\xe4\x83\x65\x7a\x1d\x58\x60\xc5\xb6\xf7\xf0\xa4\x43\x4e\x7f\xac\x32\xdc\xb0\xd6\xac\x09\x7a\x77\x7d\xef\x30\xef\x20\xef\x8c\x09\x6f\x59\x0f\x24\x96\xcc\x42\x85\xa8\x40\xa0\x43\xee\x42\x9a\xf6\x0a\xbc\x3d\xbf\x2b\xd2\x21\x82\x13\x69\x03\x5a\x82\x46\xab\x57\x4f\xd0\x66\x41\xe8\x87\x17\x71\x70\xdb\xe0\x11\x6b\xa5\xef\xb0\xed\x5c\xeb\x66\xc6\x3d\xa3\x9c\x2a\xaf\x49\x4c\xc1\x99\x0e\x7d\xaf\x31\xb5\x03\x81\x55\xb7\x58\x0c\x6c\xe6\x5b\x20\x70\xc7\x42\x83\x37\x32\x0a\xb3\x7d\xab\xb5\xad\xd1\x75\x28\xcf\xbd\x88\xe7\x49\x3f\x3a\x85\x9a\x35\x16\x47\xa3\x7e\x77\x1f\x9e\x4e\x5b\x83\x5c\xaf\x02\xd2\x82\xc1\xff\x0e\x00\x00\xff\xff\x78\x36\xb5\xe4\x2f\x16\x00\x00")

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

	info := bindataFileInfo{name: "go-centrifuge/build/configs/default_config.yaml", size: 5679, mode: os.FileMode(420), modTime: time.Unix(1551869118, 0)}
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

	info := bindataFileInfo{name: "go-centrifuge/build/configs/testing_config.yaml", size: 1076, mode: os.FileMode(420), modTime: time.Unix(1551869111, 0)}
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

