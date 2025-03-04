package key

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestKeys() (privPath, pubPath string, cleanup func(), err error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", nil, err
	}

	privBytes, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return "", "", nil, err
	}

	privPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
	privPath = filepath.Join(os.TempDir(), "test_private.pem")
	err = os.WriteFile(privPath, privPEM, 0600)
	if err != nil {
		return "", "", nil, err
	}

	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		return "", "", nil, err
	}

	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	pubPath = filepath.Join(os.TempDir(), "test_public.pem")
	err = os.WriteFile(pubPath, pubPEM, 0600)
	if err != nil {
		return "", "", nil, err
	}

	cleanup = func() {
		os.Remove(privPath)
		os.Remove(pubPath)
	}

	return privPath, pubPath, cleanup, nil
}

func TestLoad(t *testing.T) {
	privPath, pubPath, cleanup, err := generateTestKeys()
	require.NoError(t, err)

	defer cleanup()

	privKey, err := Load(privPath, PrivateKey)
	assert.NoError(t, err)
	assert.IsType(t, &ecdsa.PrivateKey{}, privKey)

	pubKey, err := Load(pubPath, PublicKey)
	assert.NoError(t, err)
	assert.IsType(t, &ecdsa.PublicKey{}, pubKey)

	_, err = Load("nonexistent.pem", PrivateKey)
	assert.Error(t, err)

	_, err = Load(pubPath, PrivateKey)
	assert.Error(t, err)
}
