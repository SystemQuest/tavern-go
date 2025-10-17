package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ValidateJWT validates a JWT token with specified options
// This function is designed to be used as a Tavern extension validator
//
// Expected args:
//   - jwt_key: the JSON key containing the token (e.g., "token")
//   - key: the secret key used to sign the token
//   - options: validation options (verify_signature, verify_aud, verify_exp)
//   - audience: expected audience value
//
// Example usage in YAML:
//
//	response:
//	  body:
//	    $ext:
//	      function: validate_jwt
//	      extra_kwargs:
//	        jwt_key: "token"
//	        key: "CGQgaG7GYvTcpaQZqosLy4"
//	        options:
//	          verify_signature: true
//	          verify_aud: true
//	          verify_exp: true
//	        audience: testserver
func ValidateJWT(responseBody map[string]interface{}, args map[string]interface{}) error {
	// Extract the token from response body
	jwtKey, ok := args["jwt_key"].(string)
	if !ok {
		return fmt.Errorf("jwt_key must be a string")
	}

	tokenValue, ok := responseBody[jwtKey]
	if !ok {
		return fmt.Errorf("token key '%s' not found in response", jwtKey)
	}

	tokenString, ok := tokenValue.(string)
	if !ok {
		return fmt.Errorf("token value must be a string")
	}

	// Extract signing key
	signingKey, ok := args["key"].(string)
	if !ok {
		return fmt.Errorf("signing key must be a string")
	}

	// Extract options
	options, ok := args["options"].(map[string]interface{})
	if !ok {
		// Default options if not provided
		options = map[string]interface{}{
			"verify_signature": true,
			"verify_aud":       true,
			"verify_exp":       true,
		}
	}

	// Extract audience
	expectedAudience, _ := args["audience"].(string)

	// Parse and validate token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(signingKey), nil
	})

	if err != nil {
		verifySignature, _ := options["verify_signature"].(bool)
		if verifySignature {
			return fmt.Errorf("failed to parse token: %v", err)
		}
	}

	// Validate token is valid
	if token != nil && !token.Valid {
		return fmt.Errorf("token is not valid")
	}

	// Extract and validate claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		// Verify audience
		verifyAud, _ := options["verify_aud"].(bool)
		if verifyAud && expectedAudience != "" {
			aud, err := claims.GetAudience()
			if err != nil {
				return fmt.Errorf("failed to get audience: %v", err)
			}

			found := false
			for _, a := range aud {
				if a == expectedAudience {
					found = true
					break
				}
			}

			if !found {
				return fmt.Errorf("audience mismatch: expected '%s', got %v", expectedAudience, aud)
			}
		}

		// Verify expiration
		verifyExp, _ := options["verify_exp"].(bool)
		if verifyExp {
			exp, err := claims.GetExpirationTime()
			if err != nil {
				return fmt.Errorf("failed to get expiration time: %v", err)
			}
			if exp != nil && exp.Before(time.Now()) {
				return fmt.Errorf("token has expired")
			}
		}

		return nil
	}

	return fmt.Errorf("failed to extract claims from token")
}

// Note: In a real implementation, this would be registered with the extension system
// For example:
//   extension.Register("validate_jwt", ValidateJWT)
//
// However, since this is an example, we'll document how to use it in tests
// without actually registering it in the main tavern-go extension registry.
