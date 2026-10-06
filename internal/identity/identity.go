// Package identity verifies OAuth bearer tokens against a configured resource.
package identity

import (
	"context"
	"net/http"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/modelcontextprotocol/go-sdk/auth"
)

func NewVerifier(ctx context.Context, issuer, resource, jwksURL string) (auth.TokenVerifier, error) {
	keys, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, err
	}
	return Verifier(keys.Keyfunc, issuer, resource), nil
}
func Verifier(key jwt.Keyfunc, issuer, resource string) auth.TokenVerifier {
	return func(ctx context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(token, claims, key, jwt.WithValidMethods([]string{"RS256"}), jwt.WithIssuer(issuer), jwt.WithAudience(resource), jwt.WithExpirationRequired())
		if err != nil {
			return nil, auth.ErrInvalidToken
		}
		subject, err := claims.GetSubject()
		if err != nil || strings.TrimSpace(subject) == "" || len(subject) > 512 {
			return nil, auth.ErrInvalidToken
		}
		expiration, err := claims.GetExpirationTime()
		if err != nil || expiration == nil {
			return nil, auth.ErrInvalidToken
		}
		scopes := []string{}
		switch value := claims["scope"].(type) {
		case string:
			scopes = strings.Fields(value)
		case []any:
			for _, v := range value {
				if s, ok := v.(string); ok {
					scopes = append(scopes, s)
				}
			}
		}
		return &auth.TokenInfo{UserID: subject, Expiration: expiration.Time, Scopes: scopes}, nil
	}
}
