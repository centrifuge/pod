package merkleroots

import (
	"github.com/xsleonard/go-merkle"
	"golang.org/x/crypto/sha3"
	"reflect"
	"strconv"
	"fmt"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"strings"
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"sort"
	"errors"
)

func getFieldOrderFromProtobuf(tag string) (order int64, err error){
	fmt.Println("TAG", tag)
	tagList := strings.Split(tag, ",")
	order, err = strconv.ParseInt(tagList[1], 10, 64)
	if err != nil {
		return order, err
	}
	order = order-1 // Protobuf starts fields with 1
	return order, nil
}

func AppendByte(slice []byte, data ...byte) []byte {
	m := len(slice)
	n := m + len(data)
	if n > cap(slice) { // if necessary, reallocate
		// allocate double what's needed, for future growth.
		newSlice := make([]byte, (n+1)*2)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0:n]
	copy(slice[m:n], data)
	return slice
}

func getBytesFromValue(value reflect.Value) (bytes []byte, err error) {
	typeInt64 := reflect.TypeOf((*invoice.SaltedInt64)(nil)).Elem()
	typeString := reflect.TypeOf((*invoice.SaltedString)(nil)).Elem()
	typeTimestamp := reflect.TypeOf((*invoice.SaltedTimestamp)(nil)).Elem()
	typeBytes := reflect.TypeOf((*invoice.SaltedBytes)(nil)).Elem()
	salt := value.Elem().FieldByName("Salt").Bytes()

	switch t := value.Type().Elem(); t {
	case typeInt64:
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(value.Elem().FieldByName("Value").Int()))
		bytes = AppendByte(salt, b...)
	case typeString:
		bytes = AppendByte(salt, []byte(value.Elem().FieldByName("Value").String())...)
	case typeTimestamp:
		fmt.Println("Got timestamp")
	case typeBytes:
		bytes = AppendByte(salt, value.Elem().FieldByName("Value").Bytes()...)
	default:
		fmt.Println(t)
		panic("Unknown type")
	}

	return
}

type OrderedRoot struct {
	Order int64
	Bytes [][]byte
}

type MerkleLeaves []OrderedRoot

func (l MerkleLeaves) Len() int {
	return len(l)
}
func (l MerkleLeaves) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func (s MerkleLeaves) Less(i, j int) bool {
	return s[i].Order > s[j].Order
}

func (l MerkleLeaves) ToArray() [][]byte {
	var leaveArray [][]byte

	for _, leaf := range l {
		leaveArray = append(leaveArray, leaf.Bytes[:]...)
	}
	return leaveArray
}

func GenerateLeavesForField(value reflect.Value) (leaves [][]byte, err error) {
	typeInt64 := reflect.TypeOf((*invoice.SaltedInt64)(nil)).Elem()
	typeString := reflect.TypeOf((*invoice.SaltedString)(nil)).Elem()
	typeTimestamp := reflect.TypeOf((*invoice.SaltedTimestamp)(nil)).Elem()
	typeBytes := reflect.TypeOf((*invoice.SaltedBytes)(nil)).Elem()

	// if not say it needs to be salted protobuf message
	bytes := [][]byte{}
	switch t := value.Type().Elem(); t {
	case typeInt64:
		salt := value.Elem().FieldByName("Salt").Bytes()
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(value.Elem().FieldByName("Value").Int()))
		bytes = append(bytes, AppendByte(salt, b...))
	case typeString:
		salt := value.Elem().FieldByName("Salt").Bytes()
		bytes = append(bytes, AppendByte(salt, []byte(value.Elem().FieldByName("Value").String())...))
	case typeTimestamp:
		//salt := value.Elem().FieldByName("Salt").Bytes()
		fmt.Println("Got timestamp")
	case typeBytes:
		salt := value.Elem().FieldByName("Salt").Bytes()
		bytes = append(bytes, AppendByte(salt, value.Elem().FieldByName("Value").Bytes()...))
	default:
		bytes, err = GenerateMerkleLeaves(value)
		if err != nil {
			return nil, err
		}
	}
	return bytes, nil
}

func GenerateMerkleLeavesForArray(items reflect.Value) (leaves [][]byte, err error) {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(items.Len()))
	leaves = append(leaves, b)
	for i := 0; i < items.Len(); i++ {
		fmt.Println("ARRAY LINE")
		item := items.Index(i)
		fieldLeaves, err := GenerateLeavesForField(item)
		if err != nil {
			return leaves, err
		}
		leaves = append(leaves, fieldLeaves...)
	}

	return
}

func GenerateMerkleLeaves(data interface{}) (leaves [][]byte, err error) {
	merkleLeaves := MerkleLeaves{}
	var v reflect.Value
	if reflect.ValueOf(data).Kind() == reflect.Ptr {
		v = reflect.ValueOf(data).Elem()
	} else {
		v = reflect.ValueOf(data)

	}

	// Iterate over all available fields and read the tag value
	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i)
		if value.IsNil() {
			continue
		}
		tag := v.Type().Field(i).Tag.Get("protobuf")
		fieldNo, err := getFieldOrderFromProtobuf(tag)
		if err != nil {
			return leaves, err
		}

		var bytes [][]byte
		if value.Kind() == reflect.Slice {
			bytes, err = GenerateMerkleLeavesForArray(value)
			if err != nil {
				return bytes, err
			}
		// The struct should not have any members that are not valid protobuf messages
		} else if !value.Type().Implements(reflect.TypeOf((*proto.Message)(nil)).Elem()) {
			return nil, errors.New("Field is not valid protobuf message")
		} else {
			bytes, err = GenerateLeavesForField(value)
			if err != nil {
				return nil, err
			}
		}
		merkleLeaves = append(merkleLeaves, OrderedRoot{
			fieldNo,
			bytes,
		})
	}
	sort.Sort(merkleLeaves)
	leaves = merkleLeaves.ToArray()
	return
}

// GenerateMerkleRoot creates a merkle root for a salted document
func GenerateMerkleRoot(doc proto.Message) (root []byte) {

	flat, err := GenerateMerkleLeaves(doc)
	if err != nil {
		panic(err)
	}
	fmt.Println("Length", len(flat))
	root = make([]byte, 32)
	tree := merkle.NewTree()
	tree.Generate(flat, sha3.New256())
	copy(root[:32], tree.Root().Hash[:32])
	return
}