// +build unit

package centchain

import (
	"fmt"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/types"
	"github.com/stretchr/testify/assert"
)

func TestExtrinsicStatusTask_ParseKwargs(t *testing.T) {
	task := ExtrinsicStatusTask{}
	kwargs := map[string]interface{}{}

	// Parse JobID error - missing JobID
	decoded, err := utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.Error(t, err)

	// Parse AccountDID error - missing account ID
	jobID := jobs.NewJobID().String()
	kwargs[jobs.JobIDParam] = jobID
	decoded, err = utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.Error(t, err)

	// Parse AccountDID error - malformed DID format
	kwargs[TransactionAccountParam] = "0x1234ZZ"
	decoded, err = utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.Error(t, err)

	// Parse Extrinsic Hash error - missing ExtHashParam
	did := testingidentity.GenerateRandomDID()
	kwargs[TransactionAccountParam] = did.String()
	decoded, err = utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.Error(t, err)

	// Parse ExtrinsicHash error - malformed ExtHashParam
	kwargs[TransactionExtHashParam] = 123456
	decoded, err = utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.Error(t, err)

	// Parse FromBlock error - missing FromBlockParam
	kwargs[TransactionExtHashParam] = "0xd18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515"
	decoded, err = utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.Error(t, err)

	// Parse FromBlock error - malformed FromBlockParam
	kwargs[TransactionFromBlockParam] = "wrong"
	decoded, err = utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.Error(t, err)

	// Parse ExtSignature error - missing ExtSignatureParam
	kwargs[TransactionFromBlockParam] = uint32(5)
	decoded, err = utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.Error(t, err)

	// Parse ExtSignature error - malformed ExtSignatureParam
	kwargs[TransactionExtSignatureParam] = 5
	decoded, err = utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.Error(t, err)

	// Parse ExtSignature error - wrong hex ExtSignatureParam
	kwargs[TransactionExtSignatureParam] = "0xZZOOPP"
	decoded, err = utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.Error(t, err)

	// Success
	kwargs[TransactionExtSignatureParam] = "0xd18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515d18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515"
	decoded, err = utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.NoError(t, err)
	assert.Equal(t, kwargs[TransactionAccountParam], task.accountID.String())
	assert.Equal(t, kwargs[TransactionExtHashParam], task.extHash)
	assert.Equal(t, kwargs[TransactionFromBlockParam], task.fromBlock)
	assert.Equal(t, kwargs[TransactionExtSignatureParam], task.extSignature.Hex())

}

func TestExtrinsicStatusTask_ProcessRunTask(t *testing.T) {
	t.Skip()
	task := NewExtrinsicStatusTask(1*time.Second, 10, nil, getBlockHash, getBlock, getMetadataLatest, getStorage)
	jobID := jobs.NewJobID().String()
	did := testingidentity.GenerateRandomDID()
	kwargs := map[string]interface{}{
		jobs.JobIDParam:              jobID,
		TransactionAccountParam:      did,
		TransactionExtHashParam:      "0xd18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515",
		TransactionFromBlockParam:    uint32(2), // Error block not ready
		TransactionExtSignatureParam: "0xd18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515d18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515",
	}
	decoded, err := utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.NoError(t, err)

	// Error getting block hash - failed but retriable
	_, err = task.processRunTask()
	assert.Error(t, err)

	// Error - not retriable
	kwargs[TransactionFromBlockParam] = uint32(3)
	decoded, err = utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.NoError(t, err)
	_, err = task.processRunTask()
	assert.Error(t, err)

	// Error getting block - some error fetching block
	kwargs[TransactionFromBlockParam] = uint32(4)
	decoded, err = utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.NoError(t, err)
	_, err = task.processRunTask()
	assert.Error(t, err)

	// Error - extrinsic not in block
	kwargs[TransactionFromBlockParam] = uint32(5)
	decoded, err = utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.NoError(t, err)
	_, err = task.processRunTask()
	assert.Error(t, err)
	assert.Equal(t, uint32(6), task.fromBlock) //Incremented block number for next iteration

	// Failure - extrinsic found in block with fail status
	kwargs[TransactionFromBlockParam] = uint32(6)
	decoded, err = utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.NoError(t, err)
	_, err = task.processRunTask()
	assert.EqualError(t, err, fmt.Sprintf("extrinsic %s failed {true 14 0}", kwargs[TransactionExtHashParam]))

	// Success - extrinsic found in block with success status
	kwargs[TransactionFromBlockParam] = uint32(7)
	decoded, err = utils.SimulateJSONDecodeForGocelery(kwargs)
	assert.Nil(t, err, "json decode should not thrown an error")
	err = task.ParseKwargs(decoded)
	assert.NoError(t, err)
	_, err = task.processRunTask()
	assert.NoError(t, err)
}

// Mocks
func getBlockHash(blockNumber uint64) (types.Hash, error) {
	hh := "0xd18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515" // only success, extrinsic in block
	switch blockNumber {
	case 2:
		return types.Hash{}, ErrBlockNotReady
	case 3:
		return types.Hash{}, errors.New("very bad error, not retriable")
	case 4:
		hh = "0xf18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515" // will cause error in next getBlock call
	case 5:
		hh = "0xf18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573516" // extrinsic not in block
	case 6:
		hh = "0xf18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573519" // extrinsic found in block with fail status
	}
	return types.NewHashFromHexString(hh)
}

func getBlock(blockHash types.Hash) (*types.SignedBlock, error) {
	bb1, _ := types.HexDecodeString("0xf18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515d18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515")
	bb2, _ := types.HexDecodeString("0xd18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515d18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515")
	switch blockHash.Hex() {
	case "0xf18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515":
		return nil, errors.New("some error fetching block")
	case "0xf18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573516": //Extrinsic not in block
		return &types.SignedBlock{}, nil
	case "0xf18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573519": // failed extrinsic
		return &types.SignedBlock{
			Block: types.Block{
				Extrinsics: []types.Extrinsic{
					{Signature: types.ExtrinsicSignatureV5{Signature: types.MultiSignature{IsSr25519: true, AsSr25519: types.NewSignature(bb1)}}},
					{Signature: types.ExtrinsicSignatureV5{Signature: types.MultiSignature{IsSr25519: true, AsSr25519: types.NewSignature(bb2)}}},
				},
			},
		}, nil
	case "0xd18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573515": //Extrinsic in block
		return &types.SignedBlock{
			Block: types.Block{
				Extrinsics: []types.Extrinsic{
					{Signature: types.ExtrinsicSignatureV5{Signature: types.MultiSignature{IsSr25519: true, AsSr25519: types.NewSignature(bb1)}}},
					{Signature: types.ExtrinsicSignatureV5{Signature: types.MultiSignature{IsSr25519: true, AsSr25519: types.NewSignature(bb2)}}},
				},
			},
		}, nil
	}
	return nil, nil
}

func getMetadataLatest() (*types.Metadata, error) {
	return MetaDataWithCall("Anchor.commit"), nil
}

func getStorage(key types.StorageKey, target interface{}, blockHash types.Hash) error {
	rawStorage := "0800000000000000000001000000000000"
	switch blockHash.Hex() {
	case "0xf18036d7c1fe109af377e8ce1d9096e69a5df0741fba7e4f3507f8e6aa573519": //failed status
		rawStorage = "08000000000000000000010000000001010e0000"
	}

	bb, err := types.HexDecodeString(rawStorage)
	if err != nil {
		return err
	}

	return types.DecodeFromBytes(bb, target)
}
