package benchmarks

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcp-runtime/cully/internal/website"
)

func BenchmarkServer(t *testing.B) {
	//t.ReportAllocs()
	s, err := website.New("test", "test2")
	if err != nil {
		panic(err)
	}
	t.ResetTimer()
	h := s.Handler()
	rec := httptest.NewRequest(http.MethodGet, "/", nil)

	for t.Loop() {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, rec)
	}
}
