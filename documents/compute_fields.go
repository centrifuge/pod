package documents

import (
	"bytes"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-substrate-rpc-client/scale"
	logging "github.com/ipfs/go-log"
	"github.com/wasmerio/go-ext-wasm/wasmer"
)

var computeLog = logging.Logger("compute_fields")

const (
	// ErrComputeFieldsInvalidWASM is a sentinel error for invalid WASM blob
	ErrComputeFieldsInvalidWASM = errors.Error("Invalid WASM blob")

	// ErrComputeFieldsAllocateNotFound is a sentinel error when WASM doesn't expose 'allocate' function
	ErrComputeFieldsAllocateNotFound = errors.Error("'allocate' function not exported")

	// ErrComputeFieldsComputeNotFound is a sentinel error when WASM doesn't expose 'compute' function
	ErrComputeFieldsComputeNotFound = errors.Error("'compute' function not exported")
)

// fetchComputeFunctions checks WASM if the required exported fields are present
// `allocate`: allocate function to allocate the required bytes on WASM
// `compute`: compute function to compute the 32byte value from the passed attributes
// and returns both functions along with the VM instance
func fetchComputeFunctions(wasm []byte) (i wasmer.Instance, allocate, compute func(...interface{}) (wasmer.Value, error), err error) {
	i, err = wasmer.NewInstance(wasm)
	if err != nil {
		return i, allocate, compute, errors.AppendError(nil, ErrComputeFieldsInvalidWASM)
	}

	allocate, ok := i.Exports["allocate"]
	if !ok {
		err = errors.AppendError(err, ErrComputeFieldsAllocateNotFound)
	}

	compute, ok = i.Exports["compute"]
	if !ok {
		err = errors.AppendError(err, ErrComputeFieldsComputeNotFound)
	}

	return i, allocate, compute, err
}

// executeWASM encodes the passed attributes and executes WASM.
// returns a 32byte value. If the WASM exits with an error, returns a zero 32byte value
// execution is allowed to run for upto timeout. Once the timeout is reached, VM is stopped and returned a zero value.
func executeWASM(wasm []byte, attributes []Attribute, timeout time.Duration) (result [32]byte) {
	i, allocate, compute, err := fetchComputeFunctions(wasm)
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

	// start the timer
	go func() {
		t := time.NewTimer(timeout)
		<-t.C
		computeLog.Error("timout exceeded: WASM took too long to compute")
		i.Close()
	}()

	// allocate memory
	res, err := allocate(buf.Len())
	if err != nil {
		computeLog.Error(err)
		return result
	}

	// copy encoded attributes to memory
	mem := i.Memory.Data()[res.ToI32():]
	copy(mem, buf.Bytes())

	// execute compute
	res, err = compute(res.ToI32(), buf.Len())
	if err != nil {
		computeLog.Error(err)
		return result
	}

	// copy result from the wasm
	d := i.Memory.Data()[res.ToI32() : res.ToI32()+32]
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

// executeComputeFields executes all the compute fields and updates the document with target attributes
// each WASM is executed at a max of timeout duration.
// TODO: Add tests
func (cd *CoreDocument) executeComputeFields(timeout time.Duration) (*CoreDocument, error) {
	var computeFields []*coredocumentpb.TransitionRule
	for _, rule := range cd.Document.TransitionRules {
		if rule.Action != coredocumentpb.TransitionAction_TRANSITION_ACTION_COMPUTE {
			continue
		}

		computeFields = append(computeFields, rule)
	}

	ncd := cd
	for _, computeField := range computeFields {
		var attrs []Attribute

		// filter attributes
		for _, attr := range computeField.ComputeFields {
			key, err := AttrKeyFromBytes(attr)
			if err != nil {
				return nil, err
			}

			attrs = append(attrs, ncd.Attributes[key])
		}

		// execute WASM
		result := executeWASM(computeField.ComputeCode, attrs, timeout)

		// set result into the target attribute
		targetKey, err := AttrKeyFromLabel(string(computeField.ComputeTargetField))
		if err != nil {
			return nil, err
		}

		targetAttr := Attribute{
			KeyLabel: string(computeField.ComputeTargetField),
			Key:      targetKey,
			Value: AttrVal{
				Type:  AttrBytes,
				Bytes: result[:],
			},
		}

		ncd, err = ncd.AddAttributes(CollaboratorsAccess{}, false, nil, targetAttr)
		if err != nil {
			return nil, err
		}
	}

	return ncd, nil
}
