package auth

import (
	"encoding/base32"
	"fmt"
	"image/png"
	"os"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// MFA represents Multi-Factor Authentication
type MFA struct {
	issuer string
}

// NewMFA creates a new MFA instance
func NewMFA(issuer string) *MFA {
	return &MFA{
		issuer: issuer,
	}
}

// GenerateTOTPSecret generates a new TOTP secret for a user
func (m *MFA) GenerateTOTPSecret(userEmail string) (*otp.Key, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      m.issuer,
		AccountName: userEmail,
		Period:      30, // 30 seconds
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %w", err)
	}

	return key, nil
}

// GenerateQRCode generates a QR code image for the TOTP secret
func (m *MFA) GenerateQRCode(key *otp.Key, filename string) error {
	img, err := key.Image(200, 200)
	if err != nil {
		return fmt.Errorf("failed to generate QR code image: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create QR code file: %w", err)
	}
	defer file.Close()

	return png.Encode(file, img)
}

// ValidateTOTP validates a TOTP code
func (m *MFA) ValidateTOTP(secret, code string) (bool, error) {
	// Decode base32 secret
	decodedSecret, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return false, fmt.Errorf("invalid secret format: %w", err)
	}

	// Validate code
	valid := totp.Validate(code, string(decodedSecret))
	return valid, nil
}

// GenerateTOTPCode generates a TOTP code from a secret (for testing)
func (m *MFA) GenerateTOTPCode(secret string) (string, error) {
	decodedSecret, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("invalid secret format: %w", err)
	}

	code, err := totp.GenerateCode(string(decodedSecret), time.Now())
	if err != nil {
		return "", fmt.Errorf("failed to generate TOTP code: %w", err)
	}

	return code, nil
}
