package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var jwtSecret = []byte("your-256-bit-secret-key-change-this-in-production")

type Claims struct {
	UserID   string   `json:"user_id"`
	Email    string   `json:"email"`
	Role     string   `json:"role"`
	Scopes   []string `json:"scopes"`
	IssuedAt int64    `json:"iat"`
	ExpireAt int64    `json:"exp"`
}

// GenerateJWT generates a new JWT token
func GenerateJWT(userID, email, role string, scopes []string) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	claims := Claims{
		UserID:   userID,
		Email:    email,
		Role:     role,
		Scopes:   scopes,
		IssuedAt: time.Now().Unix(),
		ExpireAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	headerJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	message := headerEncoded + "." + claimsEncoded
	
	mac := hmac.New(sha256.New, jwtSecret)
	mac.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return message + "." + signature, nil
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Verify signature
	message := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, jwtSecret)
	mac.Write([]byte(message))
	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if parts[2] != expectedSignature {
		return nil, fmt.Errorf("invalid signature")
	}

	// Decode claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %v", err)
	}

	var claims Claims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %v", err)
	}

	// Check expiration
	if time.Now().Unix() > claims.ExpireAt {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

// JWTAuthMiddleware is HTTP middleware for JWT authentication
func JWTAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		claims, err := ValidateJWT(token)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		r.Header.Set("X-User-ID", claims.UserID)
		r.Header.Set("X-User-Email", claims.Email)
		r.Header.Set("X-User-Role", claims.Role)

		next(w, r)
	}
}

// RequireScope checks if the user has a specific scope
func RequireScope(scope string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			token := strings.TrimPrefix(authHeader, "Bearer ")
			
			claims, err := ValidateJWT(token)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			hasScope := false
			for _, s := range claims.Scopes {
				if s == scope {
					hasScope = true
					break
				}
			}

			if !hasScope {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next(w, r)
		}
	}
}
