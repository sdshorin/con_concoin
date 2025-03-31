package signer

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/asn1"
	"encoding/json"
	"log"
	"math/big"
	"net/http"
)

const (
	ServerAddress     = "http://localhost"
	ServerDefaultPort = "8080"
	SendHandle        = "/add_message" // implemented in con-run
)

type MaliciousBehaviourType int

const (
	Double           MaliciousBehaviourType = iota
	InvalidSignKey                          // 1
	InvalidTransfer                         // 1
	InvalidSignature                        // 1
	DifferentData                           // cannot be fully tested internally since con-valid must check signatures for each tx
	None                                    // good ending
)

type Transaction struct {
	From   string
	To     string
	Amount int
	Key    *ecdsa.PrivateKey
}

type TxData struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

type SignedTransaction struct {
	Data      TxData `json:"data"`
	Signature []byte `json:"signature"`
}

func Sign(tx Transaction, t MaliciousBehaviourType) []byte {

	if t == InvalidTransfer {
		randomBytes := make([]byte, 64)
		_, err := rand.Read(randomBytes)
		if err != nil {
			log.Fatalf("Error generating random bytes: %v", err)
		}
		return randomBytes
	}

	data := TxData{
		From:   tx.From,
		To:     tx.To,
		Amount: tx.Amount,
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Error marshaling transaction data: %v", err)
	}

	if t == InvalidSignKey {
		tx.Key = func() *ecdsa.PrivateKey {
			privKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			return privKey
		}()
	}

	r, s, err := ecdsa.Sign(rand.Reader, tx.Key, dataBytes[:])
	if err != nil {
		log.Fatalf("Error signing transaction: %v", err)
	}

	sig, err := asn1.Marshal(struct {
		R *big.Int
		S *big.Int
	}{r, s})
	if err != nil {
		log.Fatalf("Error encoding signature: %v", err)
	}

	if t == InvalidSignature {
		sig[0] ^= 1 // corrupt the signature
	}

	if t == DifferentData {
		data.To = "MaliciousBob"
	}

	signedTx := SignedTransaction{
		Data:      data,
		Signature: sig,
	}

	output, err := json.Marshal(signedTx)
	if err != nil {
		log.Fatalf("Error marshaling signed transaction: %v", err)
	}

	return output
}

func transactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var tx Transaction
	err := json.NewDecoder(r.Body).Decode(&tx)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	maliciousHeader := r.Header.Get("Malicious")
	var maliciousType MaliciousBehaviourType

	switch maliciousHeader {
	case "Double":
		maliciousType = Double
	case "Key":
		maliciousType = InvalidSignKey
	case "Transfer":
		maliciousType = InvalidTransfer
	case "Signature":
		maliciousType = InvalidSignature
	case "DifferentData":
		maliciousType = DifferentData
	default:
		maliciousType = None // No malicious behavior
	}

	if tx.Key == nil {
		http.Error(w, "Missing private key in transaction", http.StatusBadRequest)
		return
	}

	signed_transaction := Sign(tx, maliciousType)
	req, err := http.NewRequest(http.MethodPost, "http://localhost", bytes.NewBuffer(signed_transaction))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request", http.StatusInternalServerError)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to send transaction", resp.StatusCode)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Transaction signed and sent successfully"))
}

// func main() {
// 	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	tx := Transaction{
// 		From:   "Alice",
// 		To:     "Bob",
// 		Key:    privKey,
// 		Amount: 100,
// 	}

// 	SignAndSend(tx)
// }
