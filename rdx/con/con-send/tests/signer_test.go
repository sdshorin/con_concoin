package tests

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/asn1"
	"encoding/json"
	"math/big"
	"signer"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidSign(t *testing.T) {
	tests := []struct {
		name  string
		input signer.Transaction
	}{
		{
			name: "valid_tx_1",
			input: signer.Transaction{
				From:   "Alice",
				To:     "Bob",
				Amount: 100,
				Key: func() *ecdsa.PrivateKey {
					privKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
					return privKey
				}(),
			},
		},
		{
			name: "valid_tx_2",
			input: signer.Transaction{
				From:   "Bob",
				To:     "Alice",
				Amount: 100,
				Key: func() *ecdsa.PrivateKey {
					privKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
					return privKey
				}(),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			signature := signer.Sign(tc.input, signer.None)

			var signedTx signer.SignedTransaction
			err := json.Unmarshal(signature, &signedTx)
			require.NoError(t, err, "Failed to unmarshal signed transaction")

			r := new(big.Int)
			s := new(big.Int)
			sigStruct := struct {
				R, S *big.Int
			}{}
			_, err = asn1.Unmarshal(signedTx.Signature, &sigStruct)
			require.NoError(t, err, "Failed to unmarshal signature")

			r = sigStruct.R
			s = sigStruct.S

			dataBytes, err := json.Marshal(signedTx.Data)
			require.NoError(t, err, "Failed to marshal transaction data")

			pubKey := tc.input.Key.Public().(*ecdsa.PublicKey)
			isValid := ecdsa.Verify(pubKey, dataBytes, r, s)

			require.True(t, isValid, "Signature verification failed")

			expectedData := signer.TxData{
				From:   tc.input.From,
				To:     tc.input.To,
				Amount: tc.input.Amount,
			}
			require.Equal(t, expectedData, signedTx.Data, "Transaction data does not match expected values")
		})
	}
}

func TestMalicious(t *testing.T) {
	tests := []struct {
		name          string
		input         signer.Transaction
		maliciousType signer.MaliciousBehaviourType
	}{
		{
			name: "tx_invalid_key",
			input: signer.Transaction{
				From:   "Alice",
				To:     "Bob",
				Amount: 100,
				Key: func() *ecdsa.PrivateKey {
					privKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
					return privKey
				}(),
			},
			maliciousType: signer.InvalidSignKey,
		},
		{
			name: "tx_invalid_transfer",
			input: signer.Transaction{
				From:   "Alice",
				To:     "Bob",
				Amount: 100,
				Key: func() *ecdsa.PrivateKey {
					privKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
					return privKey
				}(),
			},
			maliciousType: signer.InvalidTransfer,
		},
		{
			name: "tx_invalid_signature",
			input: signer.Transaction{
				From:   "Alice",
				To:     "Bob",
				Amount: 100,
				Key: func() *ecdsa.PrivateKey {
					privKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
					return privKey
				}(),
			},
			maliciousType: signer.InvalidSignature,
		},
		{
			name: "tx_different_data",
			input: signer.Transaction{
				From:   "Alice",
				To:     "Bob",
				Amount: 100,
				Key: func() *ecdsa.PrivateKey {
					privKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
					return privKey
				}(),
			},
			maliciousType: signer.DifferentData,
		},
		{
			name: "tx_good",
			input: signer.Transaction{
				From:   "Alice",
				To:     "Bob",
				Amount: 100,
				Key: func() *ecdsa.PrivateKey {
					privKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
					return privKey
				}(),
			},
			maliciousType: signer.None, // make sure it does not fail
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			signature := signer.Sign(tc.input, tc.maliciousType)
			require.NotNil(t, signature, "Signature should not be nil")
			require.NotEmpty(t, signature, "Signature should not be empty")

			var signedTx signer.SignedTransaction
			err := json.Unmarshal(signature, &signedTx)

			if tc.maliciousType == signer.InvalidTransfer {
				require.Error(t, err, "Expected error while unmarshaling signed transaction")
				return
			}
			require.NoError(t, err, "Failed to unmarshal signed transaction")

			r := new(big.Int)
			s := new(big.Int)
			sigStruct := struct {
				R, S *big.Int
			}{}
			_, err = asn1.Unmarshal(signedTx.Signature, &sigStruct)

			if tc.maliciousType == signer.InvalidSignature {
				require.Error(t, err, "Expected error while unmarshaling signature")
				return
			}
			require.NoError(t, err, "Failed to unmarshal signature")

			r = sigStruct.R
			s = sigStruct.S

			dataBytes, err := json.Marshal(signedTx.Data)
			require.NoError(t, err, "Failed to marshal transaction data")

			pubKey := tc.input.Key.Public().(*ecdsa.PublicKey)
			isValid := ecdsa.Verify(pubKey, dataBytes, r, s)

			// include DifferentData as it affects signature
			if tc.maliciousType == signer.InvalidSignKey || tc.maliciousType == signer.DifferentData {
				require.False(t, isValid, "Signature verification should have failed")
				return
			}

			require.True(t, isValid, "Signature expected to be correct")

			expectedData := signer.TxData{
				From:   tc.input.From,
				To:     tc.input.To,
				Amount: tc.input.Amount,
			}

			if tc.maliciousType == signer.DifferentData {
				require.NotEqual(t, expectedData, signedTx.Data, "Transaction data does not match expected values")
				return
			}

		})
	}
}
