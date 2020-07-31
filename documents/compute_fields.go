package documents

import (
	"bytes"
	"context"
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

	// computeFieldsTimeout is the max time we let the WASM computation to be run.
	computeFieldsTimeout = time.Second * 30
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
	//defer i.Close()

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

	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()

	resultCh := make(chan [32]byte)

	go func(resultCh chan<- [32]byte) {
		var result [32]byte

		// allocate memory
		res, err := allocate(buf.Len())
		if err != nil {
			computeLog.Error(err)
			resultCh <- result
			return
		}

		// copy encoded attributes to memory
		mem := i.Memory.Data()[res.ToI32():]
		copy(mem, buf.Bytes())

		// execute compute
		res, err = compute(res.ToI32(), buf.Len())
		if err != nil {
			computeLog.Error(err)
			resultCh <- result
			return
		}

		// copy result from the wasm
		d := i.Memory.Data()[res.ToI32() : res.ToI32()+32]
		copy(result[:], d)
		resultCh <- result
	}(resultCh)

	select {
	case <-ctx.Done():
		computeLog.Error("timout exceeded: WASM took too long to compute")
	case result = <-resultCh:
		computeLog.Info("WASM execution is successful")
	}

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

// ExecuteComputeFields executes all the compute fields and updates the document with target attributes
// each WASM is executed at a max of timeout duration.
func (cd *CoreDocument) ExecuteComputeFields(timeout time.Duration) error {
	computeFields := cd.GetComputeFieldsRules()

	ncd := cd
	for _, computeField := range computeFields {
		targetAttr, err := executeComputeField(computeField, ncd.Attributes, timeout)
		if err != nil {
			return err
		}

		ncd, err = ncd.AddAttributes(CollaboratorsAccess{}, false, nil, targetAttr)
		if err != nil {
			return err
		}
	}

	*cd = *ncd
	return nil
}

func executeComputeField(rule *coredocumentpb.TransitionRule, attributes map[AttrKey]Attribute, timeout time.Duration) (result Attribute, err error) {
	var attrs []Attribute

	// filter attributes
	for _, attr := range rule.ComputeFields {
		key, err := AttrKeyFromBytes(attr)
		if err != nil {
			return result, err
		}

		fa, ok := attributes[key]
		if !ok {
			continue
		}

		attrs = append(attrs, fa)
	}

	// execute WASM
	r := executeWASM(rule.ComputeCode, attrs, timeout)

	// set result into the target attribute
	targetKey, err := AttrKeyFromLabel(string(rule.ComputeTargetField))
	if err != nil {
		return result, err
	}

	result = Attribute{
		KeyLabel: string(rule.ComputeTargetField),
		Key:      targetKey,
		Value: AttrVal{
			Type:  AttrBytes,
			Bytes: r[:],
		},
	}
	return result, nil
}
