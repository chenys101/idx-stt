package algo

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

type aesEncryptor struct {
	key []byte
}

// AESEncryptECB 使用 AES - ECB 模式加密
func (a *aesEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return []byte(""), err
	}
	paddedPlaintext := PKCS7Padding(plaintext, block.BlockSize())
	ciphertext := make([]byte, len(paddedPlaintext))
	mode := newECBEncrypter(block)
	mode.CryptBlocks(ciphertext, paddedPlaintext)
	return ciphertext, nil
}

func (a *aesEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("malformed ciphertext")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// PKCS7Padding 实现 PKCS7 填充
func PKCS7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

// PKCS7UnPadding 实现 PKCS7 去填充
func PKCS7UnPadding(data []byte) []byte {
	length := len(data)
	unpadding := int(data[length-1])
	return data[:(length - unpadding)]
}

// AESEncryptECB 使用 AES - ECB 模式加密
func AESEncryptECB(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	paddedPlaintext := PKCS7Padding(plaintext, block.BlockSize())
	ciphertext := make([]byte, len(paddedPlaintext))
	mode := newECBEncrypter(block)
	mode.CryptBlocks(ciphertext, paddedPlaintext)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AESDecryptECB 使用 AES - ECB 模式解密
func AESDecryptECB(ciphertextBase64 string, key []byte) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decrypted := make([]byte, len(ciphertext))
	mode := newECBDecrypter(block)
	mode.CryptBlocks(decrypted, ciphertext)
	return PKCS7UnPadding(decrypted), nil
}

// newECBEncrypter 创建 ECB 加密模式
func newECBEncrypter(b cipher.Block) cipher.BlockMode {
	return ecbEncrypter{
		b: b,
	}
}

type ecbEncrypter struct {
	b cipher.Block
}

func (x ecbEncrypter) BlockSize() int { return x.b.BlockSize() }

func (x ecbEncrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.BlockSize() != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	for len(src) > 0 {
		x.b.Encrypt(dst, src[:x.BlockSize()])
		src = src[x.BlockSize():]
		dst = dst[x.BlockSize():]
	}
}

// newECBDecrypter 创建 ECB 解密模式
func newECBDecrypter(b cipher.Block) cipher.BlockMode {
	return ecbDecrypter{
		b: b,
	}
}

type ecbDecrypter struct {
	b cipher.Block
}

func (x ecbDecrypter) BlockSize() int { return x.b.BlockSize() }

func (x ecbDecrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.BlockSize() != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	for len(src) > 0 {
		x.b.Decrypt(dst, src[:x.BlockSize()])
		src = src[x.BlockSize():]
		dst = dst[x.BlockSize():]
	}
}
