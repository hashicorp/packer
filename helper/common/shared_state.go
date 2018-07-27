package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Used to set variables which we need to access later in the build, where
// state bag and config information won't work
func sharedStateFilename(suffix string, buildName string) string {
	uuid := os.Getenv("PACKER_RUN_UUID")
	return filepath.Join(os.TempDir(), fmt.Sprintf("packer-%s-%s-%s", uuid, suffix, buildName))
}

func SetSharedState(key string, value string, buildName string) error {
	uuid := os.Getenv("PACKER_RUN_UUID")

	// Encrypt the value using the run uuid. This is probably good enough as an
	// encryption key because we only keep it in memory until after the point
	// at which this storage file is wiped.
	encryptionKey := []byte(uuid)[0:32]
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	encryptedVal, err := gcm.Seal(nonce, nonce, []byte(value), nil), nil

	if err != nil {
		return fmt.Errorf("Error encrypting sensitive variable: %s", err)
	}

	// Write encrypted value to the storage file.
	err = ioutil.WriteFile(sharedStateFilename(key, buildName), encryptedVal, 0600)
	return err
}

func RetrieveSharedState(key string, buildName string) (string, error) {
	uuid := os.Getenv("PACKER_RUN_UUID")

	// Decrypt the stored item.
	encryptionKey := []byte(uuid)[0:32]
	encryptedValue, err := ioutil.ReadFile(sharedStateFilename(key, buildName))

	block, err := aes.NewCipher([]byte(encryptionKey))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedValue) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	// nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	nonce, encryptedValue := encryptedValue[:nonceSize], encryptedValue[nonceSize:]

	value, err := gcm.Open(nil, nonce, encryptedValue, nil)
	if err != nil {
		return "", err
	}

	return string(value), nil
}

func RemoveSharedStateFile(key string, buildName string) {
	os.Remove(sharedStateFilename(key, buildName))
}
