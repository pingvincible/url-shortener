package login_test

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
	"url-shortener/internal/http-server/handlers/login"
	"url-shortener/internal/http-server/handlers/login/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
)

func TestLoginHandler(t *testing.T) {
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
			statusCode: http.StatusOK,
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
			name:       "Email does not exist",
			email:      "notexistent@gmail.com",
			password:   "123456",
			mockError:  errors.New("invalid email or password"),
			respError:  "invalid email or password",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Password is incorrect",
			email:      "good@gmail.com",
			password:   "123456incorrect",
			mockError:  errors.New("invalid email or password"),
			respError:  "invalid email or password",
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

			userLoginerMock := mocks.NewUserLoginer(t)

			if tc.respError == "" || tc.mockError != nil {
				userLoginerMock.
					On("Login", context.Background(), tc.email, tc.password).
					Return("", tc.mockError).
					Once()
			}

			handler := login.New(slogdiscard.NewDiscardLogger(), userLoginerMock)

			input := fmt.Sprintf(`{"email":"%s","password":"%s"}`, tc.email, tc.password)
			if tc.body != "" {
				input = tc.body
			}

			req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.statusCode, rr.Code)

			body := rr.Body.String()
			var resp login.Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)
		})
	}
}
