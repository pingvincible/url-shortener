package delete_test

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	deleteHandler "url-shortener/internal/http-server/handlers/delete"
	"url-shortener/internal/http-server/handlers/delete/mocks"
	mocks2 "url-shortener/internal/http-server/middleware/authenticator/mocks"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
)

func TestDeleteHandler(t *testing.T) {
	cases := []struct {
		name                    string
		alias                   string
		urlDeleterMockError     error
		userId                  int64
		shouldCallIsAdmin       bool
		isAdmin                 bool
		isAdminCheckerMockError error
		statusCode              int
	}{
		{
			name:              "Success",
			alias:             "test_alias",
			statusCode:        http.StatusNoContent,
			shouldCallIsAdmin: true,
			userId:            int64(1),
			isAdmin:           true,
		},
		{
			name:                    "Empty alias",
			alias:                   "",
			statusCode:              http.StatusNotFound,
			shouldCallIsAdmin:       false,
			userId:                  int64(1),
			isAdmin:                 false,
			isAdminCheckerMockError: errors.New("mock must not be called"),
		},
		{
			name:              "User is not admin",
			alias:             "test_alias",
			statusCode:        http.StatusForbidden,
			shouldCallIsAdmin: true,
			userId:            int64(1),
			isAdmin:           false,
		},
		{
			name:                    "Error in IsAdmin method",
			alias:                   "test_alias",
			statusCode:              http.StatusInternalServerError,
			shouldCallIsAdmin:       true,
			userId:                  int64(1),
			isAdmin:                 false,
			isAdminCheckerMockError: errors.New("unexpected error"),
		},
		{
			name:                "DeleteURL Error",
			alias:               "test_alias",
			urlDeleterMockError: errors.New("unexpected error"),
			shouldCallIsAdmin:   true,
			userId:              int64(1),
			isAdmin:             true,
			statusCode:          http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			t.Parallel()

			isAdminCheckerMock := mocks.NewIsAdminChecker(t)
			if tc.shouldCallIsAdmin {
				isAdminCheckerMock.On(
					"IsAdmin",
					context.Background(),
					tc.userId,
				).
					Return(tc.isAdmin, tc.isAdminCheckerMockError).
					Once()
			}
			urlDeleterMock := mocks.NewURLDeleter(t)

			if tc.isAdmin && tc.isAdminCheckerMockError == nil {
				urlDeleterMock.On("DeleteURL", tc.alias).
					Return(tc.urlDeleterMockError).
					Once()
			}

			// Creating router and route with handler
			r := chi.NewRouter()
			r.Use(mocks2.UserIdAdder(tc.userId))
			r.Delete(
				"/{alias}",
				deleteHandler.New(
					slogdiscard.NewDiscardLogger(),
					urlDeleterMock,
					isAdminCheckerMock,
				),
			)

			// Act
			req, err := http.NewRequest("DELETE", "/"+tc.alias, nil)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			// Assert
			require.NoError(t, err)

			require.Equal(t, tc.statusCode, rr.Code)
		})
	}
}
