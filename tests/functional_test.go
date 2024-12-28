package tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
	"url-shortener/internal/http-server/handlers/login"
	"url-shortener/internal/http-server/handlers/register"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/lib/api"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/random"
)

const (
	host = "localhost:8082"
)

func TestURLShortener_HappyPathToSave(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	e.POST("/register").
		WithJSON(register.Request{
			Email:    "testadmin@gmail.com",
			Password: "admin",
		}).
		Expect().
		Status(http.StatusCreated)

	r := e.POST("/login").
		WithJSON(login.Request{
			Email:    "testadmin@gmail.com",
			Password: "admin",
		}).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object()

	r.Keys().ContainsOnly("status", "token")

	r.Value("status").String().IsEqual(resp.StatusOK)

	token := r.Value("token").String().Raw()

	r = e.POST("/url").
		WithJSON(save.Request{
			URL:   gofakeit.URL(),
			Alias: random.NewRandomString(10),
		}).
		WithHeader("Authorization", "Bearer "+token).
		Expect().
		Status(http.StatusOK).
		JSON().
		Object()

	r.Keys().ContainsOnly("status", "alias")

	r.Value("status").String().IsEqual(resp.StatusOK)

}

func TestURLShortener_SaveRedirect(t *testing.T) {
	testCases := []struct {
		name  string
		url   string
		alias string
		error string
	}{
		{
			name:  "Valid URL",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
		},
		{
			name:  "Invalid URL",
			url:   "invalid_url",
			alias: gofakeit.Word(),
			error: "field URL is not a valid URL",
		},
		{
			name:  "Empty Alias",
			url:   gofakeit.URL(),
			alias: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}

			e := httpexpect.Default(t, u.String())

			// Login

			r := e.POST("/login").
				WithJSON(login.Request{
					Email:    "testadmin@gmail.com",
					Password: "admin",
				}).
				Expect().
				Status(http.StatusOK).
				JSON().
				Object()

			token := r.Value("token").String().Raw()

			// Save

			r = e.POST("/url").
				WithJSON(save.Request{
					URL:   tc.url,
					Alias: tc.alias,
				}).
				WithHeader("Authorization", "Bearer "+token).
				Expect().Status(http.StatusOK).
				JSON().Object()

			if tc.error != "" {
				r.NotContainsKey("alias")

				r.Value("error").String().IsEqual(tc.error)

				return
			}

			alias := tc.alias

			if tc.alias != "" {
				r.Value("alias").String().IsEqual(alias)
			} else {
				r.Value("alias").String().NotEmpty()

				alias = r.Value("alias").String().Raw()
			}

			// Redirect

			testRedirect(t, alias, tc.url)

			// Delete
			path := "/" + alias
			e.DELETE(path).
				WithHeader("Authorization", "Bearer "+token).
				Expect().
				Status(http.StatusNoContent)

			// Redirect again

			testRedirectNotFound(t, alias)

		})
	}

}

func testRedirect(t *testing.T, alias string, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	redirectedToURL, err := api.GetRedirect(u.String())
	require.NoError(t, err)

	require.Equal(t, urlToRedirect, redirectedToURL)

}

func testRedirectNotFound(t *testing.T, alias string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	_, err := api.GetRedirect(u.String())
	require.ErrorIs(t, err, api.ErrInvalidStatusCode)
}
