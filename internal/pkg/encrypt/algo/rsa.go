package algo

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

type rsaEncryptor struct {
	pubKey  *rsa.PublicKey
	privKey *rsa.PrivateKey
}

func (r *rsaEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, r.pubKey, plaintext, nil)
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

func (r *rsaEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	hash := sha256.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, r.privKey, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
