package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestGenerateResetToken tests if a token is generated correctly.
func TestGenerateResetToken(t *testing.T) {
	token, err := generateResetToken()
	assert.NoError(t, err, "generateResetToken should not return an error")
	assert.NotEmpty(t, token, "Generated token should not be empty")
	// Basic check for expected format (32 hex characters)
	assert.Regexp(t, `^[a-f0-9]{32}$`, token, "Token should be 32 hexadecimal characters")
}

// TestStoreAndValidateResetToken tests the storing and validation logic.
func TestStoreAndValidateResetToken(t *testing.T) {
	userID := 123
	testToken, err := storeResetToken(userID)
	assert.NoError(t, err, "storeResetToken should not return an error")
	assert.NotEmpty(t, testToken, "Stored token should not be empty")

	// --- Test Valid Token ---
	validatedUserID, valid := validateResetToken(testToken)
	assert.True(t, valid, "Token should be valid immediately after storing")
	assert.Equal(t, userID, validatedUserID, "Validated user ID should match the stored user ID")

	// --- Test Invalid Token ---
	invalidToken := "invalidtokenstring1234567890abcdef"
	_, valid = validateResetToken(invalidToken)
	assert.False(t, valid, "An invalid token string should not validate")

	// --- Test Expired Token ---
	// Manually add an expired token (adjust expiry)
	expiredToken, _ := generateResetToken() // Generate a new one to avoid race conditions
	tokenMutex.Lock()
	resetTokens[expiredToken] = ResetTokenInfo{
		UserID: 999,
		Expiry: time.Now().Add(-2 * time.Minute), // Expired 2 minutes ago
	}
	tokenMutex.Unlock()

	_, valid = validateResetToken(expiredToken)
	assert.False(t, valid, "An expired token should not validate")

	// Check if the expired token was removed
	tokenMutex.RLock()
	_, exists := resetTokens[expiredToken]
	tokenMutex.RUnlock()
	assert.False(t, exists, "Expired token should have been removed during validation check")

	// --- Test Invalidate Token ---
	// Use the still valid token from the beginning
	invalidateResetToken(testToken)
	tokenMutex.RLock()
	_, exists = resetTokens[testToken]
	tokenMutex.RUnlock()
	assert.False(t, exists, "Invalidated token should not exist in the store")

	// Clean up any remaining test tokens if necessary (though validation/invalidation should handle it)
	tokenMutex.Lock()
	delete(resetTokens, testToken) // Ensure the first token is gone
	delete(resetTokens, expiredToken) // Ensure expired one is gone if removal failed somehow
	tokenMutex.Unlock()

}

// TestValidateResetToken_NotFound tests validation for a non-existent token.
func TestValidateResetToken_NotFound(t *testing.T) {
	nonExistentToken := "nonexistenttoken1234567890abcdef"
	_, valid := validateResetToken(nonExistentToken)
	assert.False(t, valid, "A non-existent token should not validate")
}

// TestInvalidateResetToken tests removing a token.
func TestInvalidateResetToken(t *testing.T) {
	userID := 456
	tokenToInvalidate, err := storeResetToken(userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenToInvalidate)

	// Check it exists first
	tokenMutex.RLock()
	_, exists := resetTokens[tokenToInvalidate]
	tokenMutex.RUnlock()
	assert.True(t, exists, "Token should exist before invalidation")

	// Invalidate it
	invalidateResetToken(tokenToInvalidate)

	// Check it's gone
	tokenMutex.RLock()
	_, exists = resetTokens[tokenToInvalidate]
	tokenMutex.RUnlock()
	assert.False(t, exists, "Token should not exist after invalidation")
} 