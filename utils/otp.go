package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// GenerateNumericOTP generates a cryptographically secure n-digit numeric string (e.g. 6 digits: 100000 - 999999).
func GenerateNumericOTP(digits int) (string, error) {
	if digits <= 0 {
		digits = 6
	}
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	format := fmt.Sprintf("%%0%dd", digits)
	return fmt.Sprintf(format, n.Int64()), nil
}
