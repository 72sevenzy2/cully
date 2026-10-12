package benchmarks

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcp-runtime/cully/internal/website"
)

// read-only routes.

func BenchmarkPublishedHandler(b *testing.B) {
	b.ReportAllocs()
	s, err := website.New("test", b.TempDir())
	h := s.Handler()
	if err != nil {
		panic(err)
	}
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		req := httptest.NewRequest(http.MethodGet, "/testimonials.json", nil)
		rr := httptest.NewRecorder()
		for p.Next() {
			h.ServeHTTP(rr, req)
		}
	})
}
