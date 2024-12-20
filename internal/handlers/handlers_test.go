package handlers

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
		code     int
		location string
	}
	tests := []struct { // the array of structures
		name string
		want want
	}{
		{
			name: "simple POST_test #1",
			want: want{
				code: 201,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8080/", bytes.NewReader([]byte("https://practicum.yandex.ru/")))
			request.Header.Set("Content-Type", "text/plain")
			// create a new Recorder
			w := httptest.NewRecorder()
			CreateShortURLHandler(w, request)

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
		code     int
		location string
	}
	tests := []struct { // the array of structures
		name string
		want want
	}{
		{
			name: "simple GET_test #1",
			want: want{
				code: 400,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8080/", bytes.NewReader([]byte("https://practicum.yandex.ru/")))
			request.Header.Set("Content-Type", "text/plain")
			// create a new Recorder
			w := httptest.NewRecorder()
			GetLongURLHandler(w, request)

			res := w.Result()
			// check response code
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
