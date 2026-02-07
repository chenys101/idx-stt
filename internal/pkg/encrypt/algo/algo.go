package algo

import (
	"crypto/rsa"
)

// NewAesEncryptor 创建 AES 加密器
func NewAesEncryptor(key []byte) *aesEncryptor {
	return &aesEncryptor{key: key}
}

// NewRsaEncryptor 创建 RSA 加密器
func NewRsaEncryptor(pubKey *rsa.PublicKey, privKey *rsa.PrivateKey) *rsaEncryptor {
	return &rsaEncryptor{pubKey: pubKey, privKey: privKey}
}
