// Package identity verifies OAuth bearer tokens against a configured resource.
package identity

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Agent-Hellboy/mcp-auth/auth-client/go/mcpauth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/modelcontextprotocol/go-sdk/auth"
)

func NewVerifier(_ context.Context, issuer, resource, jwksURL string) (auth.TokenVerifier, error) {
	verifier := &mcpauth.JWTVerifier{Issuer: issuer, Audience: resource, JWKSURL: jwksURL}
	return func(ctx context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
		claims, err := verifier.VerifyContext(ctx, token)
		if err != nil || strings.TrimSpace(claims.Subject) == "" || strings.TrimSpace(claims.Subject) != claims.Subject || len(claims.Subject) > 512 {
			return nil, auth.ErrInvalidToken
		}
		exp, ok := claims.Raw["exp"].(float64)
		if !ok {
			return nil, auth.ErrInvalidToken
		}
		scopes := make([]string, 0, len(claims.Scopes))
		for scope := range claims.Scopes {
			scopes = append(scopes, scope)
		}
		sort.Strings(scopes)
		return &auth.TokenInfo{UserID: claims.Subject, Expiration: time.Unix(int64(exp), 0), Scopes: scopes}, nil
	}, nil
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
