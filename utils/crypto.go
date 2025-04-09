package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

// AES key for encryption/decryption
var aesKey []byte

// File name and permissions for the key storage
const keyFileName = "encryption.key"
const keyFilePermission = 0600 // Read/write for owner only

// List of potential key storage locations in order of preference
var keyStorageLocations = []string{
	"/app/keys",             // Docker volume mount point
	"/data",                 // Common Docker volume mount
	"/app",                  // Docker app directory
	"/var/lib/securesignin", // System directory
}

// getKeyFilePaths returns all possible paths where the key file could be stored
func getKeyFilePaths() []string {
	paths := []string{}

	// Add standard Docker locations
	for _, location := range keyStorageLocations {
		path := filepath.Join(location, keyFileName)
		paths = append(paths, path)
		log.Printf("Adding potential key path: %s", path)
	}

	// Try user's home directory if available
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configDir := filepath.Join(homeDir, ".securesignin")
		// Add to paths but don't create directory yet - we'll create only when saving
		paths = append(paths, filepath.Join(configDir, keyFileName))
	}

	// Add current directory as last resort
	currDir, err := os.Getwd()
	if err == nil {
		paths = append(paths, filepath.Join(currDir, keyFileName))
	}

	return paths
}

// loadKeyFromFile tries to load the key from any of the potential key file locations
func loadKeyFromFile() ([]byte, error) {
	paths := getKeyFilePaths()
	log.Printf("Searching for key file in %d locations", len(paths))

	var lastErr error
	for _, keyPath := range paths {
		// Check if file exists
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			log.Printf("Key file not found at: %s", keyPath)
			lastErr = err
			continue
		}

		// Try to read the key file
		keyData, err := os.ReadFile(keyPath)
		if err != nil {
			log.Printf("Failed to read key file at %s: %v", keyPath, err)
			lastErr = err
			continue
		}

		// Key should be 32 bytes for AES-256
		if len(keyData) != 32 {
			log.Printf("Invalid key file at %s: expected 32 bytes but got %d", keyPath, len(keyData))
			lastErr = fmt.Errorf("invalid key file format, expected 32 bytes but got %d", len(keyData))
			continue
		}

		log.Printf("Successfully loaded key from: %s", keyPath)
		return keyData, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to load key from any location: %w", lastErr)
	}
	return nil, errors.New("no key file found in any location")
}

// saveKeyToFile saves the key to the first writable location
func saveKeyToFile(key []byte) error {
	paths := getKeyFilePaths()

	var lastErr error
	for _, keyPath := range paths {
		// Make sure directory exists
		dir := filepath.Dir(keyPath)
		if err := os.MkdirAll(dir, 0700); err != nil {
			log.Printf("Failed to create directory %s: %v", dir, err)
			lastErr = err
			continue
		}

		// Try to write the key file
		err := os.WriteFile(keyPath, key, keyFilePermission)
		if err != nil {
			log.Printf("Failed to write key to %s: %v", keyPath, err)
			lastErr = err
			continue
		}

		log.Printf("Successfully saved encryption key to: %s", keyPath)
		return nil
	}

	if lastErr != nil {
		return fmt.Errorf("failed to save key to any location: %w", lastErr)
	}
	return errors.New("could not save key file to any location")
}

// generateNewKey creates a new random key
func generateNewKey() ([]byte, error) {
	key := make([]byte, 32) // AES-256
	_, err := rand.Read(key)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}
	return key, nil
}

func init() {
	log.Printf("Initializing encryption system")
	var err error

	// Try to load key from file first
	aesKey, err = loadKeyFromFile()
	if err != nil {
		log.Printf("Could not load existing key: %v. Generating new key.", err)

		// Generate a new key
		aesKey, err = generateNewKey()
		if err != nil {
			log.Fatalf("Critical error: Failed to generate new AES key: %v", err)
		}

		// Save the new key to file
		err = saveKeyToFile(aesKey)
		if err != nil {
			log.Printf("Warning: Failed to save key file: %v - encryption will work for this session only", err)
		}
	} else {
		log.Println("Loaded encryption key from disk successfully")
	}

	// Always print key length for debugging but not the key itself
	log.Printf("Using %d-bit AES key (%d bytes)", len(aesKey)*8, len(aesKey))
}

// Encrypt encrypts plaintext using AES-GCM.
func Encrypt(plaintext string) (string, error) {
	if len(aesKey) == 0 {
		return "", errors.New("AES key is not initialized")
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher block: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-GCM.
func Decrypt(ciphertextHex string) (string, error) {
	if len(aesKey) == 0 {
		return "", errors.New("AES key is not initialized")
	}
	ciphertext, err := hex.DecodeString(ciphertextHex)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex ciphertext: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher block: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, encryptedMessage := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintextBytes, err := aesGCM.Open(nil, nonce, encryptedMessage, nil)
	if err != nil {
		// Log decryption failures carefully - avoid leaking info
		log.Printf("Decryption failed (potential tampering or wrong key): %v", err)
		return "", errors.New("decryption failed") // Generic error to user
	}

	return string(plaintextBytes), nil
}

// HashPassword generates a bcrypt hash of the password.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a plaintext password with a stored bcrypt hash.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
