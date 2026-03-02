package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type (
	compressData struct {
		statusCode int
		gzipWriter *gzip.Writer
	}

	compressResponseWriter struct {
		http.ResponseWriter
		compressData *compressData
	}
)

func (r *compressResponseWriter) WriteHeader(statusCode int) {
	if r.compressData.statusCode != 0 {
		return
	}
	r.compressData.statusCode = statusCode
	contentType := r.Header().Get("Content-Type")
	allowCompress := strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")
	if allowCompress {
		r.compressData.gzipWriter = gzip.NewWriter(r.ResponseWriter)
		r.Header().Set("Content-Encoding", "gzip")
	}
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *compressResponseWriter) Write(b []byte) (int, error) {
	if r.compressData.gzipWriter != nil {
		return r.compressData.gzipWriter.Write(b)
	}
	return r.ResponseWriter.Write(b)
}

func (r *compressResponseWriter) Close() error {
	if r.compressData.gzipWriter != nil {
		return r.compressData.gzipWriter.Close()
	}
	return nil
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &compressReader{r: r, zr: zr}, nil
}

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	return c.zr.Close()
}

func WithCompress(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := &compressResponseWriter{
				ResponseWriter: w,
				compressData:   &compressData{},
			}
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				http.Error(w, "invalid gzip body", http.StatusBadRequest)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		handler.ServeHTTP(ow, r)
	})
}
