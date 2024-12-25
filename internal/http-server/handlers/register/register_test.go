package register_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/http-server/handlers/register"
	"url-shortener/internal/http-server/handlers/register/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
)

func TestRegisterHandler(t *testing.T) {
	cases := []struct {
		name       string
		email      string
		body       string
		password   string
		statusCode int
		respError  string
		mockError  error
	}{
		{
			name:       "Success",
			email:      "test@gmail.com",
			password:   "123456",
			statusCode: http.StatusCreated,
		},
		{
			name:       "Empty password",
			email:      "test@gmail.com",
			password:   "",
			respError:  "field Password is a required field",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Empty email",
			email:      "",
			password:   "123456",
			respError:  "field Email is a required field",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Empty email and password",
			email:      "",
			password:   "",
			respError:  "field Email is a required field; field Password is a required field",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "User already exists",
			email:      "test@gmail.com",
			password:   "123456",
			mockError:  errors.New("user already exists"),
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Malformed body",
			body:       "malformed body message #%$^@#{}",
			respError:  "failed to decode request",
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userRegistererMock := mocks.NewUserRegisterer(t)

			if tc.respError == "" || tc.mockError != nil {
				userRegistererMock.
					On("Register", context.Background(), tc.email, tc.password).
					Return(int64(0), tc.mockError).
					Once()
			}

			handler := register.New(slogdiscard.NewDiscardLogger(), userRegistererMock)

			input := fmt.Sprintf(`{"email":"%s","password":"%s"}`, tc.email, tc.password)
			if tc.body != "" {
				input = tc.body
			}

			req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.statusCode, rr.Code)

			body := rr.Body.String()
			var resp register.Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)
		})
	}
}
