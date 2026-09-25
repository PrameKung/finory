package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
	"golang.org/x/oauth2"
)

const testSecret = "test-access-secret-with-at-least-32-bytes"
const testUserID = "4c3b9d7e-91ef-4b2b-9e28-2408153715d9"

type memoryUsers struct {
	ids      map[string]string
	upserts  int
	lastUser GoogleUser
}

func (m *memoryUsers) UpsertGoogleUser(_ context.Context, user GoogleUser) (string, error) {
	m.upserts++
	m.lastUser = user
	if m.ids == nil {
		m.ids = make(map[string]string)
	}
	if id := m.ids[user.Subject]; id != "" {
		return id, nil
	}
	m.ids[user.Subject] = testUserID
	return testUserID, nil
}

type oauthFixture struct {
	handler   http.Handler
	provider  *httptest.Server
	key       *rsa.PrivateKey
	users     *memoryUsers
	issuedJWT string
	requests  int
	badCode   bool
}

func newOAuthFixture(t *testing.T) *oauthFixture {
	t.Helper()
	f := &oauthFixture{users: &memoryUsers{}}
	var err error
	f.key, err = rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	f.provider = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/keys":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
				"kty": "RSA", "kid": "test-key", "alg": "RS256", "use": "sig",
				"n": base64.RawURLEncoding.EncodeToString(f.key.PublicKey.N.Bytes()),
				"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(f.key.PublicKey.E)).Bytes()),
			}}})
		case "/token":
			f.requests++
			if f.badCode || r.FormValue("code") != "good-code" || r.FormValue("code_verifier") == "" {
				http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": "google-access-token", "token_type": "Bearer",
				"id_token": f.issuedJWT,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(f.provider.Close)
	keySet := oidc.NewRemoteKeySet(context.Background(), f.provider.URL+"/keys")
	verifier := oidc.NewVerifier("https://accounts.google.com", keySet, &oidc.Config{ClientID: "test-client"})
	oauth := oauth2.Config{
		ClientID: "test-client", ClientSecret: "test-secret",
		RedirectURL: "http://localhost:8080/api/v1/auth/google/callback",
		Scopes:      []string{"openid", "email", "profile"},
		Endpoint:    oauth2.Endpoint{AuthURL: f.provider.URL + "/authorize", TokenURL: f.provider.URL + "/token"},
	}
	e := echo.New()
	NewHandler(oauth, verifier, NewService(f.users, []byte(testSecret)), "http://localhost:3000/").Register(e)
	f.handler = e
	return f
}

func (f *oauthFixture) start(t *testing.T) (*http.Cookie, string) {
	t.Helper()
	response := httptest.NewRecorder()
	f.handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/auth/google", nil))
	if response.Code != http.StatusFound {
		t.Fatalf("authorization status = %d, body = %s", response.Code, response.Body.String())
	}
	location, err := url.Parse(response.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	query := location.Query()
	if query.Get("response_type") != "code" || query.Get("scope") != "openid email profile" ||
		query.Get("code_challenge_method") != "S256" || query.Get("code_challenge") == "" ||
		query.Get("nonce") == "" || query.Get("state") == "" ||
		query.Get("redirect_uri") != "http://localhost:8080/api/v1/auth/google/callback" {
		t.Fatalf("unexpected authorization URL: %s", location)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != flowCookieName || !cookies[0].HttpOnly || cookies[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("unexpected flow cookie: %+v", cookies)
	}
	return cookies[0], query.Get("nonce")
}

func (f *oauthFixture) signIDToken(t *testing.T, nonce string) {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": "https://accounts.google.com", "aud": "test-client",
		"sub": "stable-google-subject", "exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(), "nonce": nonce,
		"email": "user@example.com", "email_verified": true,
		"name": "Example User", "picture": "https://example.com/avatar.png",
	})
	token.Header["kid"] = "test-key"
	var err error
	f.issuedJWT, err = token.SignedString(f.key)
	if err != nil {
		t.Fatal(err)
	}
}

func (f *oauthFixture) callback(cookie *http.Cookie, query string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, "/auth/google/callback?"+query, nil)
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	f.handler.ServeHTTP(response, request)
	return response
}

func TestGoogleOAuthCreatesAndReusesUser(t *testing.T) {
	f := newOAuthFixture(t)
	for attempt := 1; attempt <= 2; attempt++ {
		cookie, nonce := f.start(t)
		f.signIDToken(t, nonce)
		state := strings.Split(cookie.Value, ".")[0]
		response := f.callback(cookie, "code=good-code&state="+url.QueryEscape(state))
		if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "http://localhost:3000/" {
			t.Fatalf("callback status = %d, location = %q, body = %s", response.Code, response.Header().Get("Location"), response.Body.String())
		}
		var access *http.Cookie
		for _, candidate := range response.Result().Cookies() {
			if candidate.Name == accessCookieName {
				access = candidate
			}
		}
		if access == nil || !access.HttpOnly || access.Path != "/api/v1" {
			t.Fatalf("missing application session cookie: %+v", response.Result().Cookies())
		}
		claims := &jwt.RegisteredClaims{}
		parsed, err := jwt.ParseWithClaims(access.Value, claims, func(*jwt.Token) (any, error) {
			return []byte(testSecret), nil
		}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
		if err != nil || !parsed.Valid || claims.Subject != testUserID {
			t.Fatalf("invalid application session: valid=%v, subject=%q, error=%v", parsed != nil && parsed.Valid, claims.Subject, err)
		}
	}
	if len(f.users.ids) != 1 || f.users.upserts != 2 || f.users.lastUser.Subject != "stable-google-subject" {
		t.Fatalf("unexpected users: %+v", f.users)
	}
}

func TestGoogleOAuthRejectsInvalidCallbacks(t *testing.T) {
	for _, test := range []struct {
		name       string
		query      func(string) string
		cookie     bool
		wrongNonce bool
		badCode    bool
	}{
		{name: "missing cookie", query: func(s string) string { return "code=good-code&state=" + s }},
		{name: "wrong state", query: func(string) string { return "code=good-code&state=wrong" }, cookie: true},
		{name: "missing code", query: func(s string) string { return "state=" + s }, cookie: true},
		{name: "provider error", query: func(s string) string { return "state=" + s + "&code=good-code&error=access_denied" }, cookie: true},
		{name: "wrong nonce", query: func(s string) string { return "code=good-code&state=" + s }, cookie: true, wrongNonce: true},
		{name: "failed exchange", query: func(s string) string { return "code=good-code&state=" + s }, cookie: true, badCode: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := newOAuthFixture(t)
			cookie, nonce := f.start(t)
			if test.wrongNonce {
				nonce = "incorrect-nonce"
			}
			f.signIDToken(t, nonce)
			f.badCode = test.badCode
			state := strings.Split(cookie.Value, ".")[0]
			if !test.cookie {
				cookie = nil
			}
			response := f.callback(cookie, test.query(state))
			if response.Code != http.StatusBadRequest || f.users.upserts != 0 {
				t.Fatalf("invalid callback accepted: status=%d, users=%d", response.Code, f.users.upserts)
			}
		})
	}
}
