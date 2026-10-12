package benchmarks

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

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

func BenchmarkPhotoHandler404(b *testing.B) {
	b.ReportAllocs()
	s, err := website.New("test", b.TempDir())
	h := s.Handler()
	if err != nil {
		panic(err)
	}
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		// assuming http.NotFound() responses:
		req := httptest.NewRequest(http.MethodGet, "/testimonial-photos/10", nil)
		rr := httptest.NewRecorder()
		for p.Next() {
			h.ServeHTTP(rr, req)
		}
	})
}

func BenchmarkPhotoHandler(b *testing.B) {
	b.ReportAllocs()
	s, err := website.New("test", b.TempDir())
	h := s.Handler()
	if err != nil {
		panic(err)
	}
	id := "00000000000000000000000000000010" // 32 bit integer "10"
	dir := filepath.Join(s.DataDir, "approved", id)
	photoData := unsafe.Slice(unsafe.StringData("example photo"), len("example photo"))
	if err3 := os.MkdirAll(dir, os.ModePerm); err3 != nil {
		panic(err3)
	}
	if err2 := os.WriteFile(filepath.Join(dir, "photo"), photoData, os.ModePerm); err2 != nil {
		panic(err2)
	}
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			req := httptest.NewRequest(http.MethodGet, "/testimonial-photos/"+id, nil)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				b.Fatalf("failed with status %d", rr.Code)
				return
			}
		}
	})
}

func BenchmarkInstillationHandler(b *testing.B) {
	b.ReportAllocs()
	s, err := website.New("test", b.TempDir())
	h := s.Handler()
	if err != nil {
		panic(err)
	}
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		req := httptest.NewRequest(http.MethodGet, "/install.sh", nil)
		rr := httptest.NewRecorder()
		for p.Next() {
			h.ServeHTTP(rr, req)
		}
	})
}
