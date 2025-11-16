package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	// ErrInvalidCiphertext is returned when the ciphertext is invalid
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	// ErrInvalidKeySize is returned when the encryption key size is invalid
	ErrInvalidKeySize = errors.New("encryption key must be 32 bytes for AES-256")
)

// getEncryptionKey retrieves the encryption key from environment variables
// The key must be 32 bytes (256 bits) for AES-256
func getEncryptionKey() ([]byte, error) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	keyString := os.Getenv("ENCRYPTION_KEY")
	if keyString == "" {
		return nil, errors.New("ENCRYPTION_KEY environment variable is not set")
	}

	// Base64デコード
	key, err := base64.StdEncoding.DecodeString(keyString)
	if err != nil {
		return nil, errors.New("ENCRYPTION_KEY must be a valid base64 string")
	}

	// AES-256には32バイト（256ビット）の鍵が必要
	if len(key) != 32 {
		return nil, ErrInvalidKeySize
	}

	return key, nil
}

// EncryptToken encrypts a token using AES-256-GCM
// Returns base64-encoded ciphertext
func EncryptToken(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	key, err := getEncryptionKey()
	if err != nil {
		return "", err
	}

	// AES暗号ブロックを作成
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// GCM (Galois/Counter Mode) を使用
	// GCMは認証付き暗号化を提供し、改ざん検知が可能
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// ノンス（nonce）を生成
	// GCMでは各暗号化に一意のnonceが必要
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// 暗号化を実行
	// nonceを暗号文の先頭に付加（復号時に必要）
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Base64エンコードして返す
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptToken decrypts a token using AES-256-GCM
// Accepts base64-encoded ciphertext
func DecryptToken(ciphertextBase64 string) (string, error) {
	if ciphertextBase64 == "" {
		return "", nil
	}

	key, err := getEncryptionKey()
	if err != nil {
		return "", err
	}

	// Base64デコード
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", err
	}

	// AES暗号ブロックを作成
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// GCMモードを使用
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// nonceサイズを確認
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	// nonceと暗号文を分離
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 復号を実行
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	return string(plaintext), nil
}
