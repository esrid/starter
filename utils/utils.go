package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/mail"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword génère un hachage sécurisé pour le mot de passe fourni en utilisant bcrypt.
func HashPassword(password string) (string, error) {
	// Générer le hachage avec bcrypt en utilisant le coût par défaut
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("erreur lors du hachage du mot de passe : %w", err)
	}
	return string(hashedPassword), nil
}

// CompareHash vérifie si le mot de passe correspond au hachage fourni.
func CompareHash(hashedPassword, password string) (bool, error) {
	// Comparer le hachage avec le mot de passe fourni
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		// Si les hachages ne correspondent pas, retourner false sans erreur
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		// Pour toute autre erreur, la retourner
		return false, fmt.Errorf("erreur lors de la comparaison du hachage : %w", err)
	}
	return true, nil
}

// GenPin génère un code PIN aléatoire à 6 chiffres.
func GenPin() (string, error) {
	const pinLength = 6
	const charset = "0123456789"

	pin := make([]byte, pinLength)
	for i := range pin {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("erreur lors de la génération du PIN : %w", err)
		}
		pin[i] = charset[num.Int64()]
	}

	return string(pin), nil
}

// CSRFToken génère un token CSRF sécurisé.
func CSRFToken() (string, error) {
	const tokenLength = 32

	token := make([]byte, tokenLength)
	if _, err := rand.Read(token); err != nil {
		return "", fmt.Errorf("erreur lors de la génération du token CSRF : %w", err)
	}

	return base64.RawStdEncoding.EncodeToString(token), nil
}

// CleanString trims leading and trailing spaces and converts the string to lowercase.
func CleanString(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}

func IsValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// IsValidPasswordLength returns true if the password length is between 8 and 72 bytes (inclusive).
func IsValidPasswordLength(password string) bool {
	length := len([]byte(password)) // counts bytes, not runes
	return length >= 8 && length <= 72
}

// Returns the IP address as a slice of bytes ([]byte)
func GetIPAddressBytes(r *http.Request) []byte {
	// Check X-Forwarded-For header (may contain multiple IPs)
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		ips := strings.Split(xForwardedFor, ",")
		ip := strings.TrimSpace(ips[0])
		parsedIP := net.ParseIP(ip)
		if parsedIP != nil {
			return parsedIP
		}
	}
	// Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		parsedIP := net.ParseIP(ip)
		if parsedIP != nil {
			return parsedIP
		}
	}
	return nil
}
