package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gluster/glusterd2/glusterd2/gdctx"

	"github.com/dgrijalva/jwt-go"
)

const (
	internalUser = "glustercli"
)

var (
	requiredClaims = []string{"iss", "exp"}
)

func getAuthSecret(issuer string) string {
	if issuer == internalUser {
		return gdctx.LocalAuthToken
	}

	// TODO: Look for issuer secret in etcd if not internal user, this depends on User management feature

	return ""
}

// Auth is a middleware which authenticates HTTP requests
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If Auth disabled Return as is
		if !gdctx.RESTAPIAuthEnabled {
			next.ServeHTTP(w, r)
			return
		}

		// Verify if Authorization header exists or not
		authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		// Verify the Authorization header format "Bearer <TOKEN>"
		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
			http.Error(w, "Authorization header format must be Bearer <TOKEN>", http.StatusUnauthorized)
			return
		}

		// Verify JWT token with additional validations for Claims
		token, err := jwt.Parse(authHeaderParts[1], func(token *jwt.Token) (interface{}, error) {
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return nil, fmt.Errorf("Unable to parse Token claims")
			}

			// Error if required claims are not sent by Client
			for _, claimName := range requiredClaims {
				if _, claimOk := claims[claimName]; !claimOk {
					return nil, fmt.Errorf("Token missing %s Claim", claimName)
				}
			}

			// Validate the JWT Signing Algo
			if _, tokenOk := token.Method.(*jwt.SigningMethodHMAC); !tokenOk {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}

			secret := getAuthSecret(claims["iss"].(string))
			if secret == "" {
				return nil, fmt.Errorf("Invalid App ID: %s", claims["iss"])
			}
			// All checks GOOD, return the Secret to validate
			return []byte(secret), nil
		})

		// Check if token is Valid
		if err != nil || !token.Valid {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// TODO: Filter URLs here if any role based control of APIs, this depends on User management feature

		// Authentication is successful, continue serving the request
		next.ServeHTTP(w, r)
	})
}
