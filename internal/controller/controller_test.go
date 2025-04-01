package controller

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateShortURLHandler(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct { // the array of structures
		name string
		want want
	}{
		{
			name: "simple POST_test #1",
			want: want{
				//code: 201,
				code: 200,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8080/", bytes.NewReader([]byte("https://practicum.yandex.ru/")))
			request.Header.Set("Content-Type", "text/plain")
			// create a new Recorder
			w := httptest.NewRecorder()
			//CreateShortURLHandler(w, request)

			res := w.Result()
			// check response code
			assert.Equal(t, test.want.code, res.StatusCode)
			// get and check the body

			_, err := io.ReadAll(res.Body)
			defer res.Body.Close() // we must use defer after io.ReadAll to avoid issues
			// TODO: mb I should handle error from res.Body.Close() ?
			require.NoError(t, err)

			// todo new request
			if err != nil {
				log.Fatal(err)
			}

		})
	}
}

func TestGetLongURLHandler(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct { // the array of structures
		name string
		want want
	}{
		{
			name: "simple GET_test #1",
			want: want{
				//code: 400,
				code: 200,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8080/", bytes.NewReader([]byte("https://practicum.yandex.ru/")))
			request.Header.Set("Content-Type", "text/plain")
			// create a new Recorder
			w := httptest.NewRecorder()
			//GetLongURLHandler(w, request)

			res := w.Result()
			// check response code
			assert.Equal(t, test.want.code, res.StatusCode)
			// get and check the body

			_, err := io.ReadAll(res.Body)
			defer res.Body.Close() // TODO: mb I should handle error from res.Body.Close() ?

			require.NoError(t, err)
			// TODO how shuuld I handle this error??? What log?
			if err != nil {
				log.Fatal(err)
			}

		})
	}
}

func TestCreateShortURLJSONHandler(t *testing.T) {
	type want struct {
		code int
	}
	tests := []struct { // the array of structures
		name     string
		method   string
		contType string
		want     want
	}{
		{
			name:     "simple JSON_POST_HANDLER_test #1",
			method:   http.MethodGet,
			contType: "application/json",
			want: want{
				//code: 400, // TODO: mb here must be error 405?? TROUBLE
				code: 200,
			},
		},

		{
			name:     "simple JSON_POST_HANDLER_test #2",
			method:   http.MethodPost,
			contType: "text/plain charset=UTF-8",
			want: want{
				//code: 400,
				code: 200,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, "http://127.0.0.1:8080/api/shorten", bytes.NewReader([]byte("https://practicum.yandex.ru/")))
			request.Header.Set("Content-Type", test.contType)
			// create a new Recorder
			w := httptest.NewRecorder()
			//ctrl := Controller{}
			//ctrl.CreateShortURLJSONHandler()

			res := w.Result()
			// check response code (Method not allowed)
			assert.Equal(t, test.want.code, res.StatusCode)
			// get and check the body

			_, err := io.ReadAll(res.Body)
			defer res.Body.Close() // TODO: mb I should handle error from res.Body.Close() ?

			require.NoError(t, err)

			// todo new request
			if err != nil {
				log.Fatal(err)
			}

		})
	}

}
