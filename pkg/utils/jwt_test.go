package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"

	"github.com/bdarge/auth/pkg/models"
	"github.com/stretchr/testify/assert"
)

type EllipticCurve struct {
    pubKeyCurve elliptic.Curve // http://golang.org/pkg/crypto/elliptic/#P256
    privateKey  *ecdsa.PrivateKey
    publicKey   *ecdsa.PublicKey
}

// New EllipticCurve instance
func New(curve elliptic.Curve) *EllipticCurve {
    return &EllipticCurve{
        pubKeyCurve: curve,
        privateKey:  new(ecdsa.PrivateKey),
    }
}

func (ec *EllipticCurve) GenerateKeys() (privKey *ecdsa.PrivateKey, pubKey *ecdsa.PublicKey, err error) {

    privKey, err = ecdsa.GenerateKey(ec.pubKeyCurve, rand.Reader)

    if err == nil {
        ec.privateKey = privKey
        ec.publicKey = &privKey.PublicKey
    }

    return
}

// EncodePrivate private key
func (ec *EllipticCurve) EncodePrivate(privKey *ecdsa.PrivateKey) (key string, err error) {

    encoded, err := x509.MarshalECPrivateKey(privKey)

    if err != nil {
        return
    }
    pemEncoded := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: encoded})

    key = string(pemEncoded)

    return
}

// EncodePublic public key
func (ec *EllipticCurve) EncodePublic(pubKey *ecdsa.PublicKey) (key string, err error) {

    encoded, err := x509.MarshalPKIXPublicKey(pubKey)

    if err != nil {
        return
    }
    pemEncodedPub := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: encoded})

    key = string(pemEncodedPub)
    return
}

func TestLogin(t *testing.T) {
	ec := New(elliptic.P256())
	ec.GenerateKeys()
	s, _ := ec.EncodePrivate(ec.privateKey)
	p, _ := ec.EncodePublic(ec.publicKey)

	readerFunc := func(filename string) ([]byte, error) {
		if strings.HasSuffix(filename, "pub") {
			return []byte(p), nil
		}
		return []byte(s), nil
	}
	
	w := JwtWrapper {
		TokenExpOn: 2,
		RefreshTokenExpOn: 10,
		FileReader: FileReaderFunc(readerFunc),
	}

	result, _ := w.GenerateToken(models.User{}, nil)
	if assert.NotNil(t, result) {
		claims, _ := w.ValidateToken(result.Token)
		assert.NotNil(t, claims)

		c, _ := w.RefreshToken(result.RefreshToken)
		assert.NotNil(t, c)
	}
}
