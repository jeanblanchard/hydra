package google

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"net/http"
	"net/http/httptest"
	"testing"
)

var mock = &google{
	id: "123",
	conf: &oauth2.Config{
		ClientID:     "client",
		ClientSecret: "secret",
		RedirectURL:  "/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "/oauth2/authorize",
			TokenURL: "/oauth2/token",
		},
	},
}

func TestNew(t *testing.T) {
	m := New("321", "client", "secret", "/callback")
	assert.Equal(t, "321", m.id)
	assert.Equal(t, "client", m.conf.ClientID)
	assert.Equal(t, "secret", m.conf.ClientSecret)
	assert.Equal(t, "/callback", m.conf.RedirectURL)
}
func TestGetID(t *testing.T) {
	assert.Equal(t, "123", mock.GetID())
}

func TestGetAuthCodeURL(t *testing.T) {
	require.NotEmpty(t, mock.GetAuthenticationURL("state"))
}

func TestExchangeCode(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"access_token": "ABCDEFG", "token_type": "bearer", "uid": "12345", "id_token": "foobar"}`)
	})
	router.HandleFunc("/plus/v1/people/me", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
	"kind": "plus#person",
	"etag": "\"foobar\"",
	"gender": "male",
	"emails": [
		{
		"value": "foobar@gmail.com",
		"type": "account"
		}
	],
	"objectType": "person",
	"id": "foobarid",
	"displayName": "foobar",
	"name": {
		"familyName": "foobar",
		"givenName": "foobar"
	},
	"url": "https://plus.google.com/foobar",
	"image": {
		"url": "https://lh3.googleusercontent.com/foobar/photo.jpg?sz=50",
		"isDefault": true
	},
	"isPlusUser": true,
	"language": "de",
	"ageRange": {
		"min": 21
	},
	"circledByCount": 6,
	"verified": false
}`)
	})
	ts := httptest.NewServer(router)

	mock.api = ts.URL
	mock.conf.Endpoint.TokenURL = ts.URL + mock.conf.Endpoint.TokenURL

	code := "testcode"
	ses, err := mock.FetchSession(code)
	require.Nil(t, err, "%s", err)
	assert.Equal(t, "foobarid", ses.GetRemoteSubject())
}
