package gzip

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

/*type GzipLogger struct {
	logger *zap.Logger
}

func NewGzipLogger(logger *zap.Logger) *GzipLogger {
	return &GzipLogger{logger: logger}
}
*/
// compressWriter implements interface  http.ResponseWriter и позволяет прозрачно для сервера
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

// compressReader implements interface io.ReadCloser и позволяет прозрачно для сервера
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
		// check, that the client has sent to the server compressed data in gzip format
		contentEncoding := req.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			fmt.Println("[INFO] Content-Encoding is gzip")
			// wrap request body into io.Reader with decompression available
			fmt.Println("[INFO] req.Body = ", req.Body)
			cr, err := newCompressReader(req.Body)
			if err != nil {
				fmt.Println(err)
				fmt.Println("[ERROR] 500")
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
			//GzipLogger.logger.Info("[INFO]", zap.String("[INFO]", "gzip IS NOT supported by the client!"), zap.String("method", req.Method), zap.String("url", req.URL.Path))
			fmt.Println("Unacceptable Content-Type => continue without gzip")
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
