package documents

import (
	"bytes"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-substrate-rpc-client/scale"
	logging "github.com/ipfs/go-log"
	"github.com/perlin-network/life/exec"
)

var computeLog = logging.Logger("compute_fields")

const (
	// ErrComputeFieldInvalidWASM is a sentinel error for invalid WASM blob
	ErrComputeFieldInvalidWASM = errors.Error("Invalid WASM blob")

	// ErrComputeFieldAllocateNotFound is a sentinel error when WASM doesn't expose 'allocate' function
	ErrComputeFieldsAllocateNotFound = errors.Error("'allocate' function not exported")

	// ErrComputeFieldComputeNotFound is a sentinel error when WASM doesn't expose 'compute' function
	ErrComputeFieldsComputeNotFound = errors.Error("'compute' function not exported")
)

// validateWASM checks WASM if the required exported fields are present
// `allocate`: allocate function to allocate the required bytes on WASM
// `compute`: compute function to compute the 32byte value from the passed attributes
func validateWASM(wasm []byte) error {
	i, err := exec.NewVirtualMachine(wasm, exec.VMConfig{}, &exec.NopResolver{}, nil)
	if err != nil {
		return ErrComputeFieldInvalidWASM
	}

	_, ok := i.GetFunctionExport("allocate")
	if !ok {
		err = errors.AppendError(err, ErrComputeFieldsAllocateNotFound)
	}

	_, ok = i.GetFunctionExport("compute")
	if !ok {
		err = errors.AppendError(err, ErrComputeFieldsComputeNotFound)
	}

	return err
}

// executeWASM encodes the passed attributes and executes WASM.
// returns a 32byte value. If the WASM exits with an error, returns a zero 32byte value
func executeWASM(wasm []byte, attributes []Attribute) (result [32]byte) {
	if err := validateWASM(wasm); err != nil {
		computeLog.Error(err)
		return result
	}

	i, err := exec.NewVirtualMachine(wasm, exec.VMConfig{}, &exec.NopResolver{}, nil)
	if err != nil {
		computeLog.Error(err)
		return result
	}

	cattrs, err := toComputeFieldsAttributes(attributes)
	if err != nil {
		computeLog.Error(err)
		return result
	}

	var buf bytes.Buffer
	enc := scale.NewEncoder(&buf)
	err = enc.Encode(cattrs)
	if err != nil {
		computeLog.Error(err)
		return result
	}

	// allocate memory
	f, _ := i.GetFunctionExport("allocate")
	res, err := i.Run(f, int64(buf.Len()))
	if err != nil {
		computeLog.Error(err)
		return result
	}

	// copy encoded attributes to memory
	mem := i.Memory[res:]
	copy(mem, buf.Bytes())

	// execute compute
	f, _ = i.GetFunctionExport("compute")
	res, err = i.Run(f, res, int64(buf.Len()))
	if err != nil {
		computeLog.Error(err)
		return result
	}

	// copy result from the wasm
	d := i.Memory[res : res+32]
	copy(result[:], d)
	return result
}

type computeSigned struct {
	Identity, DocumentVersion, Value []byte
	Type                             string
	Signature, PublicKey             []byte
}

type computeAttribute struct {
	Key    string
	Type   string
	Value  []byte
	Signed computeSigned
}

func toComputeFieldsAttributes(attrs []Attribute) (cattrs []computeAttribute, err error) {
	for _, attr := range attrs {
		cattr, err := toComputeFieldsAttribute(attr)
		if err != nil {
			return nil, err
		}

		cattrs = append(cattrs, cattr)
	}

	return cattrs, nil
}

// toComputeFieldsAttribute convert attribute of type `string`, `bytes`, `integer`, `signed` to compute field attribute
func toComputeFieldsAttribute(attr Attribute) (cattr computeAttribute, err error) {
	cattr = computeAttribute{
		Key:  attr.KeyLabel,
		Type: attr.Value.Type.String()}

	switch attr.Value.Type {
	case AttrSigned:
		s := attr.Value.Signed
		cattr.Signed = computeSigned{
			Identity:        s.Identity.ToAddress().Bytes(),
			DocumentVersion: s.DocumentVersion,
			Value:           s.Value,
			Type:            s.Type.String(),
			Signature:       s.Signature,
			PublicKey:       s.PublicKey,
		}
	case AttrBytes, AttrInt256, AttrString:
		cattr.Value, err = attr.Value.ToBytes()
	default:
		err = errors.New("'%s' attribute type not supported by compute fields", attr.Value.Type)
	}

	return cattr, err
}
