package delete_test

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	deleteHandler "url-shortener/internal/http-server/handlers/delete"
	"url-shortener/internal/http-server/handlers/delete/mocks"
	"url-shortener/internal/lib/api"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
)

func TestDeleteHandler(t *testing.T) {
	cases := []struct {
		name       string
		alias      string
		respError  string
		mockError  error
		statusCode int
	}{
		{
			name:       "Success",
			alias:      "test_alias",
			statusCode: http.StatusNoContent,
		},
		{
			name:       "DeleteURL Error",
			alias:      "test_alias",
			respError:  "internal error",
			mockError:  errors.New("unexpected error"),
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlDeleterMock := mocks.NewURLDeleter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlDeleterMock.On("DeleteURL", tc.alias).
					Return(tc.mockError).
					Once()
			}

			r := chi.NewRouter()
			r.Delete("/{alias}", deleteHandler.New(slogdiscard.NewDiscardLogger(), urlDeleterMock))

			ts := httptest.NewServer(r)
			defer ts.Close()

			status, err := api.DeleteURL(ts.URL + "/" + tc.alias)
			require.NoError(t, err)

			require.Equal(t, tc.statusCode, status)
		})
	}
}
