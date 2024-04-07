package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
)

func AesGCMEncrypt(data []byte, key []byte) []byte {
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		panic(err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}
	ciphertext := aead.Seal(nil, nonce, data, nil)

	return append(nonce, ciphertext...)
}

func AesGCMDecrypt(data []byte, key []byte) []byte {
	nonce := data[:12]
	ciphertext := data[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		panic(err)
	}
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err)
	}

	return plaintext
}
