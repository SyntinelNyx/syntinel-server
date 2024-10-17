package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

type KeyType int

const (
	PrivateKey KeyType = iota
	PublicKey
)

func loadECDSAKey(path string, keyType KeyType) (interface{}, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read key file: %v", err)
	}

	block, _ := pem.Decode(keyData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	switch keyType {
	case PrivateKey:
		if block.Type != "EC PRIVATE KEY" {
			return nil, fmt.Errorf("unexpected PEM block type for private key: %s", block.Type)
		}
		key, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("could not parse EC private key: %v", err)
		}
		return key, nil

	case PublicKey:
		if block.Type != "PUBLIC KEY" {
			return nil, fmt.Errorf("unexpected PEM block type for public key: %s", block.Type)
		}
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("could not parse public key: %v", err)
		}
		ecKey, ok := key.(*ecdsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("not an EC public key")
		}
		return ecKey, nil

	default:
		return nil, fmt.Errorf("unknown key type")
	}
}
