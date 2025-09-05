package internal

import (
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

// ParseToken parses an HMAC JWT and returns claims as a map.
// Returns an error if secret is empty, token is invalid, or claims are not map claims.
func ParseToken(tokenString, secret string) (map[string]interface{}, error) {
    if secret == "" {
        return nil, errors.New("jwt secret not set")
    }
    tok, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
        // Ensure HMAC signing method
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return []byte(secret), nil
    })
    if err != nil {
        return nil, err
    }
    if !tok.Valid {
        return nil, errors.New("invalid token")
    }
    claims, ok := tok.Claims.(jwt.MapClaims)
    if !ok {
        return nil, errors.New("invalid token claims")
    }
    out := make(map[string]interface{}, len(claims))
    for k, v := range claims {
        out[k] = v
    }
    return out, nil
}

// GenerateHMACToken creates a signed HMAC JWT with "sub" and optional extra claims.
// ttl specifies token lifetime; pass 0 for no exp claim.
func GenerateHMACToken(sub, secret string, ttl time.Duration, extras map[string]interface{}) (string, error) {
    if secret == "" {
        return "", errors.New("jwt secret not set")
    }
    claims := jwt.MapClaims{}
    if sub != "" {
        claims["sub"] = sub
    }
    now := time.Now().UTC()
    claims["iat"] = now.Unix()
    if ttl > 0 {
        claims["exp"] = now.Add(ttl).Unix()
    }
    for k, v := range extras {
        claims[k] = v
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}