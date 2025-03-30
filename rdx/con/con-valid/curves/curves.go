package curves

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"errors"
	"math/big"
)

func UnmarshalPublicKey(ecdsaCurve elliptic.Curve, bytes []byte) (*ecdsa.PublicKey, error) {
	var curve ecdh.Curve
	switch ecdsaCurve {
	case elliptic.P256():
		curve = ecdh.P256()
	case elliptic.P384():
		curve = ecdh.P384()
	case elliptic.P521():
		curve = ecdh.P521()
	default:
		return nil, errors.New("non-NIST curve")
	}

	// For error checking.
	key, err := curve.NewPublicKey(bytes)
	if err != nil {
		return nil, err
	}

	// https://github.com/golang/go/issues/63963#issuecomment-1794706080
	rawKey := key.Bytes()
	switch key.Curve() {
	case ecdh.P256():
		return &ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     big.NewInt(0).SetBytes(rawKey[1:33]),
			Y:     big.NewInt(0).SetBytes(rawKey[33:]),
		}, nil
	case ecdh.P384():
		return &ecdsa.PublicKey{
			Curve: elliptic.P384(),
			X:     big.NewInt(0).SetBytes(rawKey[1:49]),
			Y:     big.NewInt(0).SetBytes(rawKey[49:]),
		}, nil
	case ecdh.P521():
		return &ecdsa.PublicKey{
			Curve: elliptic.P521(),
			X:     big.NewInt(0).SetBytes(rawKey[1:67]),
			Y:     big.NewInt(0).SetBytes(rawKey[67:]),
		}, nil
	default:
		return nil, errors.New("cannot convert non-NIST *ecdh.PublicKey to *ecdsa.PublicKey")
	}
}
