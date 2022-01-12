package JWT

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateToken(t *testing.T) {
	generatedToken, err := GenerateToken("AccountId")
	assert.NoError(t, err)

	os.Setenv("Token", generatedToken)
}

func TestValidateToken(t *testing.T) {
	token := os.Getenv("Token")

	claim, err := ValidateToken(token)
	assert.NoError(t, err)

	assert.Equal(t, "AccountId", claim.Account_Id)
}
