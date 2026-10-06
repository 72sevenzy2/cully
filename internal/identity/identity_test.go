package identity

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestMCPAuthSDKVerifier(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	jwks := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
			"kty": "RSA", "kid": "cully-test", "alg": "RS256", "use": "sig",
			"n": base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes()),
			"e": "AQAB",
		}}})
	}))
	defer jwks.Close()
	verify, err := NewVerifier(context.Background(), "https://auth.example/mcp-auth", "https://mcp.example/cully/mcp", jwks.URL)
	if err != nil {
		t.Fatal(err)
	}
	sign := func(audience string, expiration time.Time) string {
		t.Helper()
		claims := jwt.MapClaims{"iss": "https://auth.example/mcp-auth", "aud": audience, "sub": "owner-a", "exp": expiration.Unix(), "scope": "tools:read tools:write"}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = "cully-test"
		raw, err := token.SignedString(key)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	info, err := verify(context.Background(), sign("https://mcp.example/cully/mcp", time.Now().Add(time.Hour)), &http.Request{})
	if err != nil || info.UserID != "owner-a" || len(info.Scopes) != 2 {
		t.Fatalf("valid MCP Auth token: info=%+v err=%v", info, err)
	}
	if _, err := verify(context.Background(), sign("https://other.example/mcp", time.Now().Add(time.Hour)), &http.Request{}); err == nil {
		t.Fatal("accepted token for another resource")
	}
	if _, err := verify(context.Background(), sign("https://mcp.example/cully/mcp", time.Now().Add(-2*time.Minute)), &http.Request{}); err == nil {
		t.Fatal("accepted expired token")
	}
}
