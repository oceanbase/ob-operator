/*
Copyright (c) 2025 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// AESEncrypt encrypts the given content using AES-256 with the provided key.
func AESEncrypt(key string, content string) (string, error) {
	// Ensure the key is 32 bytes for AES-256
	if len(key) != 32 {
		return "", errors.New("key must be 32 bytes for AES-256")
	}

	// Create the AES cipher
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", errors.New("failed to create AES cipher")
	}

	// Create a new IV
	iv := make([]byte, block.BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", errors.New("failed to generate IV")
	}

	// Pad the content to be a multiple of the block size
	paddedContent := PKCS7Pad([]byte(content), block.BlockSize())

	// Encrypt the content
	encryptedData := make([]byte, len(paddedContent))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encryptedData, paddedContent)

	// Combine the IV and encrypted data
	encryptedData = append(iv, encryptedData...)

	// Encode the result in base64
	return base64.StdEncoding.EncodeToString(encryptedData), nil
}

// AESDecrypt decrypts the given content using AES-256 with the provided key.
func AESDecrypt(key string, content string) (string, error) {
	// Decode the base64 encoded content
	encryptedData, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", errors.New("failed to decode base64 content")
	}

	// Ensure the key is 32 bytes for AES-256
	if len(key) != 32 {
		return "", errors.New("key must be 32 bytes for AES-256")
	}

	// Create the AES cipher
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", errors.New("failed to create AES cipher")
	}

	// The IV (Initialization Vector) should be the first block size bytes
	blockSize := block.BlockSize()
	if len(encryptedData) < blockSize {
		return "", errors.New("encrypted data is too short")
	}
	iv := encryptedData[:blockSize]
	encryptedData = encryptedData[blockSize:]

	// Decrypt the content
	if len(encryptedData)%blockSize != 0 {
		return "", errors.New("encrypted data is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(encryptedData, encryptedData)

	// Unpad the decrypted data
	decryptedData := PKCS7Unpad(encryptedData)

	return string(decryptedData), nil
}

// PKCS7Pad pads the data to be a multiple of the block size.
func PKCS7Pad(data []byte, blockSize int) []byte {
	paddingLen := blockSize - (len(data) % blockSize)
	padding := bytes.Repeat([]byte{byte(paddingLen)}, paddingLen)
	return append(data, padding...)
}

// PKCS7Unpad removes the PKCS#7 padding from the data.
func PKCS7Unpad(data []byte) []byte {
	paddingLen := int(data[len(data)-1])
	return data[:len(data)-paddingLen]
}
