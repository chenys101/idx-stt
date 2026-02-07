package encrypt

import (
	"crypto/rsa"
	"errors"
	"stt/internal/pkg/encrypt/algo"
)

// Encryptor interface
type Encryptor interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

// 算法类型枚举
const (
	AES256  = "aes-256-gcm"
	RSA4096 = "rsa-4096-oaep"
)

// 错误类型
var (
	ErrUnsupportedAlgo = errors.New("unsupported algorithm")
)

// 工厂方法
func NewEncryptor(algoMethod string, key any) (Encryptor, error) {
	switch algoMethod {
	case AES256:
		keyBytes, ok := key.([]byte)
		if !ok {
			return nil, errors.New("invalid key type for AES256")
		}
		return algo.NewAesEncryptor(keyBytes), nil
	case RSA4096:
		pubKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("invalid key type for RSA4096")
		}
		return algo.NewRsaEncryptor(pubKey, nil), nil
	default:
		return nil, ErrUnsupportedAlgo
	}
}
