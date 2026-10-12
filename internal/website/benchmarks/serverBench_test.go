package benchmarks

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcp-runtime/cully/internal/website"
)

func BenchmarkServer(b *testing.B) {
	b.ReportAllocs()
	s, err := website.New("test", b.TempDir())
	if err != nil {
		panic(err)
	}
	h := s.Handler()
	rec := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	b.ResetTimer()

	for b.Loop() {
		h.ServeHTTP(rr, rec)
	}
}

func BenchmarkServerConcurrency(b *testing.B) {
	b.ReportAllocs()

	s, err := website.New("test", b.TempDir())
	if err != nil {
		panic(err)
	}
	h := s.Handler()
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			rec := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, rec)
		}
	})
}
