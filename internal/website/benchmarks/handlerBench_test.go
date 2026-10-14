package benchmarks

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
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
	s := BenchmarkSetupHelper(b)
	h := s.Handler()
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(p *testing.PB) {
		req := httptest.NewRequest(http.MethodGet, "/testimonials.json", nil)
		for p.Next() {
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
		}
	})
}

func BenchmarkPhotoHandler404(b *testing.B) {
	s := BenchmarkSetupHelper(b)
	h := s.Handler()
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(p *testing.PB) {
		// assuming http.NotFound() responses:
		req := httptest.NewRequest(http.MethodGet, "/testimonial-photos/10", nil)
		for p.Next() {
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
		}
	})
}

func BenchmarkPhotoHandler(b *testing.B) {
	s := BenchmarkSetupHelper(b)
	h := s.Handler()
	id := "00000000000000000000000000000010" // 32 bit integer "10"
	dir := filepath.Join(s.DataDir, "approved", id)
	photoData := unsafe.Slice(unsafe.StringData("example photo"), len("example photo"))
	if err3 := os.MkdirAll(dir, os.ModePerm); err3 != nil {
		panic(err3)
	}
	if err2 := os.WriteFile(filepath.Join(dir, "photo"), photoData, os.ModePerm); err2 != nil {
		panic(err2)
	}
	b.ReportAllocs()
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
	s := BenchmarkSetupHelper(b)
	h := s.Handler()
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(p *testing.PB) {
		req := httptest.NewRequest(http.MethodGet, "/install.sh", nil)
		for p.Next() {
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
		}
	})
}

// write-included handlers.

func BenchmarkSubmitHandler(b *testing.B) {
	s := BenchmarkSetupHelper(b)
	h := s.Handler()

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	field := map[string]string{
		"name":    "eg.1",
		"quote":   "some testimonialssss",
		"consent": "on",
	}

	for key, value := range field {
		if err := mw.WriteField(key, value); err != nil {
			panic(err)
		}
	}

	if err := mw.Close(); err != nil {
		panic(err)
	}

	payload := append([]byte(nil), body.Bytes()...)
	contentType := mw.FormDataContentType()

	b.ReportAllocs()
	b.SetBytes(int64(len(payload)))
	b.ResetTimer()

	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			req := httptest.NewRequest(http.MethodPost, "/api/testimonials", bytes.NewReader(payload))
			req.Header.Set("Content-Type", contentType)
			req.Header.Set("x-cully-submission", "1")
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != http.StatusCreated {
				b.Errorf("failed with status %d", rr.Code)
				return
			}
		}
	})
}

func BenchmarkProfileImportHandler(b *testing.B) {
	s := BenchmarkSetupHelper(b)
	s.Client.Transport = &examReq{}
	h := s.Handler()
	type Test struct {
		URL string `json:"url"`
	}
	n := &Test{
		URL: "https://www.linkedin.com/in/linkedin",
	}
	js, err := json.Marshal(n)
	if err != nil {
		panic(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(js)))
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/api/linkedin-profile", bytes.NewReader(js))
			req.Header.Set("x-cully-submission", "1")
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)

			if rr.Code != http.StatusCreated {
				b.Errorf("failed with status %d", rr.Code)
				return
			}
		}
	})
}

type examReq struct{}

func (s *examReq) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader([]byte("example data"))),
		Request:    req,
	}, nil
}

func BenchmarkSetupHelper(b *testing.B) *website.Server {
	b.Helper()
	r := b.TempDir()
	if err := os.MkdirAll(r, os.ModePerm); err != nil {
		b.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join("test", "index.html"))
	if err != nil {
		b.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(r, "index.html"), data, os.ModePerm); err != nil {
		b.Fatal(err)
	}
	s, err := website.New(r, b.TempDir())
	if err != nil {
		b.Fatal(err)
	}
	return s
}
