package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lucasvo/go-centrifuge/centrifuge/constellation"
	"github.com/lucasvo/go-centrifuge/centrifuge/signingkeys"
	"github.com/lucasvo/go-centrifuge/centrifuge/witness"
)

var publicKeyPath string = ""
var nodeSocketPath string = ""

type SendStruct struct {
	PublicTo string `json:"publicTo,omitempty"`
	Payload  string `json:"payload,omitempty"`
}

type UpdateAndSendStruct struct {
	PublicTo        string `json:"publicTo,omitempty"`
	Payload         string `json:"payload,omitempty"`
	ConfirmationKey string `json:"confirmationKey"`
}

type SendResponseStruct struct {
	ConfirmationKey string `json:"confirmationKey"`
}

type ReceiveStruct struct {
	ConfirmationKey string `json:"confirmationKey"`
}
type ReceiveResponseStruct struct {
	Payload string `json:"payload"`
}

func SignPayload(payload string) (signedPayload string) {
	keyFiles := signingkeys.KeyFiles{"signing.pub", "signing.key"}
	privateKey := signingkeys.GetPrivateKey(keyFiles)
	publicKey := signingkeys.GetPublicKey(keyFiles)
	document := witness.PrepareDocument(payload)
	document.Sign(privateKey, publicKey)
	// Check signature is right
	verified := document.Verify(publicKey)
	if !verified {
		log.Fatal("Can't verify signature")
	}
	signedPayload, err := document.SerializeDocument()
	if err != nil {
		log.Fatal(err)
	}
	return
}

func SendMessage(w http.ResponseWriter, r *http.Request) {
	//params := mux.Vars(r)
	var sendStruct SendStruct
	_ = json.NewDecoder(r.Body).Decode(&sendStruct)
	target := []string{sendStruct.PublicTo}
	Client, err := constellation.NewClient(publicKeyPath, nodeSocketPath)
	if err != nil {
		log.Fatal(err)
	}

	// Logic that signs the payload before sending it over
	payload := SignPayload(sendStruct.Payload)

	log.Println("Sending Data ...")
	v, err := Client.SendPayload([]byte(payload), "", target)
	if err != nil {
		log.Fatal(err)
	}
	response := new(SendResponseStruct)
	response.ConfirmationKey = base64.StdEncoding.EncodeToString(v)
	json.NewEncoder(w).Encode(response)
}

func UpdateAndSend(w http.ResponseWriter, r *http.Request) {
	//params := mux.Vars(r)
	var updateAndSendStruct UpdateAndSendStruct
	_ = json.NewDecoder(r.Body).Decode(&updateAndSendStruct)
	target := []string{updateAndSendStruct.PublicTo}
	Client, err := constellation.NewClient(publicKeyPath, nodeSocketPath)
	if err != nil {
		log.Fatal(err)
	}
	confByes, err := base64.StdEncoding.DecodeString(updateAndSendStruct.ConfirmationKey)
	if err != nil {
		log.Fatal(err)
	}
	vs, err := Client.ReceivePayload(confByes)
	// Create new document deserializing json
	document := new(witness.SignedDocument)
	document.DeserializeDocument(string(vs))

	newDocument := witness.UpdateDocument(document)
	newDocument.Payload = updateAndSendStruct.Payload
	payload := SignPayload(newDocument.Payload)
	log.Println("Sending Data ...")
	v, err := Client.SendPayload([]byte(payload), "", target)
	if err != nil {
		log.Fatal(err)
	}
	response := new(SendResponseStruct)
	response.ConfirmationKey = base64.StdEncoding.EncodeToString(v)
	json.NewEncoder(w).Encode(response)
}

func ReceiveMessage(w http.ResponseWriter, r *http.Request) {
	var recvStruct ReceiveStruct
	_ = json.NewDecoder(r.Body).Decode(&recvStruct)
	Client, err := constellation.NewClient(publicKeyPath, nodeSocketPath)
	if err != nil {
		log.Fatal(err)
	}
	confByes, err := base64.StdEncoding.DecodeString(recvStruct.ConfirmationKey)
	if err != nil {
		log.Fatal(err)
	}
	vs, err := Client.ReceivePayload(confByes)
	// Create new document deserializing json
	document := new(witness.SignedDocument)
	document.DeserializeDocument(string(vs))

	response := new(ReceiveResponseStruct)
	response.Payload = document.Payload
	json.NewEncoder(w).Encode(response)
}

func RunNode(publicKeyPath, nodeSocketPath, cfgPath, port string) {
	var buffer bytes.Buffer
	buffer.WriteString(":")
	buffer.WriteString(port)

	// Start Constellation Node
	log.Println("Starting Constellation Node ...")
	err := constellation.RunNode(cfgPath, nodeSocketPath)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Constellation Node Started")
	//

	// Start REST API Server
	router := mux.NewRouter()
	router.HandleFunc("/send", SendMessage).Methods("POST")
	router.HandleFunc("/updateAndSend", UpdateAndSend).Methods("POST")
	router.HandleFunc("/receive", ReceiveMessage).Methods("POST")

	log.Printf("Listening on port %v ...", buffer.String())
	log.Fatal(http.ListenAndServe(buffer.String(), router))
}
