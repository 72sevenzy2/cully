package identity

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestVerification(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	verify := Verifier(func(*jwt.Token) (any, error) { return &key.PublicKey, nil }, "https://issuer.test", "https://resource.test/cully/mcp")
	tests := []struct {
		name   string
		change func(jwt.MapClaims)
		valid  bool
	}{
		{"valid", func(jwt.MapClaims) {}, true},
		{"wrong audience", func(c jwt.MapClaims) { c["aud"] = "https://other.test/mcp" }, false},
		{"wrong issuer", func(c jwt.MapClaims) { c["iss"] = "https://other.test" }, false},
		{"expired", func(c jwt.MapClaims) { c["exp"] = time.Now().Add(-time.Minute).Unix() }, false},
		{"missing expiration", func(c jwt.MapClaims) { delete(c, "exp") }, false},
		{"missing subject", func(c jwt.MapClaims) { delete(c, "sub") }, false},
		{"empty subject", func(c jwt.MapClaims) { c["sub"] = " " }, false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c := jwt.MapClaims{"iss": "https://issuer.test", "aud": "https://resource.test/cully/mcp", "sub": "owner-a", "exp": time.Now().Add(time.Hour).Unix(), "scope": "tools:read tools:write"}
			tc.change(c)
			raw, err := jwt.NewWithClaims(jwt.SigningMethodRS256, c).SignedString(key)
			if err != nil {
				t.Fatal(err)
			}
			ti, err := verify(context.Background(), raw, &http.Request{})
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v error=%v", tc.valid, err)
			}
			if tc.valid && (ti.UserID != "owner-a" || len(ti.Scopes) != 2) {
				t.Fatal("identity not preserved")
			}
		})
	}
}
