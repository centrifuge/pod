package merkleroots

import (
	"testing"
	"fmt"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"crypto/rand"
)
func createRandomByte32() (out []byte) {
	r := make([]byte, 32)
	_, err := rand.Read(r)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		panic(err)
	}
	return r[:]
}

func TestMerkleDoc(t *testing.T) {
	saltedInvoiceData := invoice.SaltedInvoiceData{
		Recipient: &invoice.SaltedBytes{createRandomByte32(), createRandomByte32()},
		Amount: &invoice.SaltedInt64{1000, createRandomByte32()},
		Currency: &invoice.SaltedString{"CHF", createRandomByte32()},
		Country:  &invoice.SaltedString{"CH", createRandomByte32()},
		Sender:  &invoice.SaltedBytes{createRandomByte32(), createRandomByte32()},
		Discounts: []*invoice.SaltedString{
			{"LineItem1", createRandomByte32()},
			{"LineItem1", createRandomByte32()},
		},


	}

	//docFlat := [][]byte{[]byte(doc.A), []byte(doc.B), []byte(doc.C)}
	//
	//var root []byte
	//root = make([]byte, 32)
	//tree := merkle.NewTree()
	//tree.Generate(docFlat, sha3.New256())
	//copy(root[:32], tree.Root().Hash[:32])
	//fmt.Println(root)
	fmt.Println(GenerateMerkleRoot(&saltedInvoiceData))
	//if !bytes.Equal(GenerateMerkleRoot(doc), root) {
	//	t.Fatal("Roots are not equal")
	//}
	fmt.Println("DONE")

	t.Fatal("all good")
}