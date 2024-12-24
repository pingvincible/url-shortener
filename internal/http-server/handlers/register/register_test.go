package register_test

import (
	"bytes"
	"context"
	"encoding/json"
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
		name      string
		email     string
		password  string
		respError string
		mockError error
	}{
		{
			name:     "Success",
			email:    "test@gmail.com",
			password: "123456",
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

			req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, http.StatusCreated, rr.Code)

			body := rr.Body.String()
			var resp register.Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)
		})
	}
}
