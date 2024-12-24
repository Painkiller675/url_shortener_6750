package gzip

import (
	"compress/gzip"
	"github.com/Painkiller675/url_shortener_6750/internal/middleware/logger"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func GzMW(h http.Handler) http.Handler {
	gzipFunc := func(res http.ResponseWriter, req *http.Request) {
		// copy original request
		or := req
		// check, that the client sent to the server compressed data in gzip format
		contentEncoding := req.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := newCompressReader(req.Body)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				// TODO mb I should call handler with original res and req here?
				return
			}
			// меняем тело запроса на новое
			req.Body = cr
			defer cr.Close()
		}
		//h.ServeHTTP(res, req)

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := req.Header.Get("Accept-Encoding") // это выставляет клиент
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if !supportsGzip {
			// continue without gzip
			h.ServeHTTP(res, or)
			return
		}
		// TODO проверяю тут req, хотя Вы говорили про res, всё ли ОК??
		if !(strings.Contains(req.Header.Get("Content-Type"), "application/json") || strings.Contains(req.Header.Get("Content-Type"), "text/html")) {
			logger.Log.Info("[INFO]", zap.String("[INFO]", "gzip IS NOT supported by the client!"), zap.String("method", req.Method), zap.String("url", req.URL.Path))
			//  continue without gzip
			h.ServeHTTP(res, or)
			return
		}
		// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
		cres := newCompressWriter(res)
		// не забываем отправить клиенту все сжатые данные после завершения middleware
		defer cres.Close()
		// call handler with modified res and req
		h.ServeHTTP(cres, req)

	}
	return http.HandlerFunc(gzipFunc)
}
