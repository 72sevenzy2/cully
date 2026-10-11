package benchmarks

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"unsafe"

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
	tear_down(t, "test2")
}

func tear_down(r *testing.B, f1 string) {
	if err := os.RemoveAll(f1); err != nil {
		// zero copy err logging to affect benchmark allocations as litle as possible.
		s := unsafe.Slice(unsafe.StringData(err.Error()), len(err.Error()))
		r.Fatal(unsafe.String(unsafe.SliceData(s), len(s)))
	}
}
