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
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v5"
	"golang.org/x/oauth2"
)

const testSecret = "test-access-secret-with-at-least-32-bytes"
const testUserID = "4c3b9d7e-91ef-4b2b-9e28-2408153715d9"

type memoryUsers struct {
	ids      map[string]string
	sessions map[[32]byte]string
	upserts  int
	lastUser GoogleUser
}

func (m *memoryUsers) CreateRefreshSession(_ context.Context, userID string, hash []byte) error {
	if m.sessions == nil {
		m.sessions = make(map[[32]byte]string)
	}
	var key [32]byte
	copy(key[:], hash)
	m.sessions[key] = userID
	return nil
}

func (m *memoryUsers) RotateRefreshSession(_ context.Context, oldHash, newHash []byte) (string, error) {
	var oldKey, newKey [32]byte
	copy(oldKey[:], oldHash)
	copy(newKey[:], newHash)
	userID, ok := m.sessions[oldKey]
	if !ok {
		return "", pgx.ErrNoRows
	}
	delete(m.sessions, oldKey)
	m.sessions[newKey] = userID
	return userID, nil
}

func (m *memoryUsers) DeleteRefreshSession(_ context.Context, hash []byte) error {
	var key [32]byte
	copy(key[:], hash)
	delete(m.sessions, key)
	return nil
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

func (m *memoryUsers) GetUserByID(_ context.Context, id string) (User, error) {
	for _, storedID := range m.ids {
		if storedID == id {
			return User{ID: id, Email: m.lastUser.Email, DisplayName: m.lastUser.DisplayName, AvatarURL: m.lastUser.AvatarURL}, nil
		}
	}
	return User{}, pgx.ErrNoRows
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
	NewHandler(oauth, verifier, NewService(f.users, []byte(testSecret)), "http://localhost:3000/dashboard").Register(e)
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
	f.signClaims(t, jwt.MapClaims{
		"iss": "https://accounts.google.com", "aud": "test-client",
		"sub": "stable-google-subject", "exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(), "nonce": nonce,
		"email": "user@example.com", "email_verified": true,
		"name": "Example User", "picture": "https://example.com/avatar.png",
	})
}

func (f *oauthFixture) signClaims(t *testing.T, claims jwt.MapClaims) {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
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
		if response.Code != http.StatusSeeOther || response.Header().Get("Location") != "http://localhost:3000/dashboard" {
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
		var refresh *http.Cookie
		for _, candidate := range response.Result().Cookies() {
			if candidate.Name == refreshCookieName {
				refresh = candidate
			}
		}
		if refresh == nil || !refresh.HttpOnly || refresh.Path != "/api/v1/auth" || refresh.Value == "" {
			t.Fatalf("missing refresh cookie: %+v", response.Result().Cookies())
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
		name         string
		query        func(string) string
		cookie       bool
		cookieValue  string
		secondCookie bool
		wrongNonce   bool
		badCode      bool
		claims       func(jwt.MapClaims)
		badIDToken   bool
	}{
		{name: "missing cookie", query: func(s string) string { return "code=good-code&state=" + s }},
		{name: "wrong state", query: func(string) string { return "code=good-code&state=wrong" }, cookie: true},
		{name: "missing state", query: func(string) string { return "code=good-code" }, cookie: true},
		{name: "duplicate state", query: func(s string) string { return "code=good-code&state=" + s + "&state=" + s }, cookie: true},
		{name: "duplicate code", query: func(s string) string { return "code=good-code&code=good-code&state=" + s }, cookie: true},
		{name: "malformed query", query: func(s string) string { return "code=good-code&state=" + s + "&%zz" }, cookie: true},
		{name: "missing code", query: func(s string) string { return "state=" + s }, cookie: true},
		{name: "empty code", query: func(s string) string { return "code=&state=" + s }, cookie: true},
		{name: "malformed flow cookie", query: func(s string) string { return "code=good-code&state=" + s }, cookie: true, cookieValue: "invalid"},
		{name: "duplicate flow cookie", query: func(s string) string { return "code=good-code&state=" + s }, cookie: true, secondCookie: true},
		{name: "provider error", query: func(s string) string { return "state=" + s + "&code=good-code&error=access_denied" }, cookie: true},
		{name: "wrong nonce", query: func(s string) string { return "code=good-code&state=" + s }, cookie: true, wrongNonce: true},
		{name: "failed exchange", query: func(s string) string { return "code=good-code&state=" + s }, cookie: true, badCode: true},
		{name: "invalid ID token", query: func(s string) string { return "code=good-code&state=" + s }, cookie: true, badIDToken: true},
		{name: "wrong issuer", query: func(s string) string { return "code=good-code&state=" + s }, cookie: true, claims: func(c jwt.MapClaims) { c["iss"] = "https://other.example.com" }},
		{name: "wrong audience", query: func(s string) string { return "code=good-code&state=" + s }, cookie: true, claims: func(c jwt.MapClaims) { c["aud"] = "other-client" }},
		{name: "expired ID token", query: func(s string) string { return "code=good-code&state=" + s }, cookie: true, claims: func(c jwt.MapClaims) { c["exp"] = time.Now().Add(-time.Hour).Unix() }},
		{name: "unverified email", query: func(s string) string { return "code=good-code&state=" + s }, cookie: true, claims: func(c jwt.MapClaims) { c["email_verified"] = false }},
		{name: "missing subject", query: func(s string) string { return "code=good-code&state=" + s }, cookie: true, claims: func(c jwt.MapClaims) { delete(c, "sub") }},
	} {
		t.Run(test.name, func(t *testing.T) {
			f := newOAuthFixture(t)
			cookie, nonce := f.start(t)
			if test.wrongNonce {
				nonce = "incorrect-nonce"
			}
			claims := jwt.MapClaims{
				"iss": "https://accounts.google.com", "aud": "test-client",
				"sub": "stable-google-subject", "exp": time.Now().Add(time.Hour).Unix(),
				"iat": time.Now().Unix(), "nonce": nonce,
				"email": "user@example.com", "email_verified": true,
				"name": "Example User", "picture": "https://example.com/avatar.png",
			}
			if test.claims != nil {
				test.claims(claims)
			}
			f.signClaims(t, claims)
			if test.badIDToken {
				f.issuedJWT = "invalid"
			}
			f.badCode = test.badCode
			state := strings.Split(cookie.Value, ".")[0]
			if test.cookieValue != "" {
				cookie.Value = test.cookieValue
			}
			if !test.cookie {
				cookie = nil
			}
			var response *httptest.ResponseRecorder
			if test.secondCookie {
				request := httptest.NewRequest(http.MethodGet, "/auth/google/callback?"+test.query(state), nil)
				request.AddCookie(cookie)
				request.AddCookie(&http.Cookie{Name: flowCookieName, Value: cookie.Value})
				response = httptest.NewRecorder()
				f.handler.ServeHTTP(response, request)
			} else {
				response = f.callback(cookie, test.query(state))
			}
			if response.Code != http.StatusBadRequest || f.users.upserts != 0 {
				t.Fatalf("invalid callback accepted: status=%d, users=%d", response.Code, f.users.upserts)
			}
			if f.requests != 0 && !test.badCode && !test.wrongNonce && test.claims == nil && !test.badIDToken {
				t.Fatal("malformed callback reached token exchange")
			}
		})
	}
}

func TestMeReturnsAuthenticatedUser(t *testing.T) {
	f := newOAuthFixture(t)
	flowCookie, nonce := f.start(t)
	f.signIDToken(t, nonce)
	state := strings.Split(flowCookie.Value, ".")[0]
	signIn := f.callback(flowCookie, "code=good-code&state="+url.QueryEscape(state))
	if signIn.Code != http.StatusSeeOther {
		t.Fatalf("sign-in status = %d", signIn.Code)
	}
	var access *http.Cookie
	for _, cookie := range signIn.Result().Cookies() {
		if cookie.Name == accessCookieName {
			access = cookie
		}
	}
	if access == nil {
		t.Fatal("missing access cookie")
	}
	for _, bearer := range []bool{false, true} {
		request := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
		if bearer {
			request.Header.Set("Authorization", "Bearer "+access.Value)
		} else {
			request.AddCookie(access)
		}
		response := httptest.NewRecorder()
		f.handler.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("me status = %d, body = %s", response.Code, response.Body.String())
		}
		var user User
		if err := json.Unmarshal(response.Body.Bytes(), &user); err != nil {
			t.Fatal(err)
		}
		if user.ID != testUserID || user.Email != "user@example.com" || user.DisplayName != "Example User" || user.AvatarURL != "https://example.com/avatar.png" {
			t.Fatalf("unexpected user: %+v", user)
		}
	}
}

func TestMeRejectsMissingInvalidAndDeletedUser(t *testing.T) {
	f := newOAuthFixture(t)
	validToken, err := NewService(f.users, []byte(testSecret)).issueAccessToken(testUserID)
	if err != nil {
		t.Fatal(err)
	}
	expiredToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject: testUserID, ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
	}).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, authorization string
		cookie              *http.Cookie
	}{
		{name: "missing"},
		{name: "invalid bearer", authorization: "Bearer invalid"},
		{name: "expired bearer", authorization: "Bearer " + expiredToken},
		{name: "missing user", authorization: "Bearer " + validToken},
		{name: "invalid cookie", cookie: &http.Cookie{Name: accessCookieName, Value: "invalid"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
			if tc.authorization != "" {
				request.Header.Set("Authorization", tc.authorization)
			}
			if tc.cookie != nil {
				request.AddCookie(tc.cookie)
			}
			response := httptest.NewRecorder()
			f.handler.ServeHTTP(response, request)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("me status = %d, body = %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestLogoutClearsSessionAndPendingOAuthCookies(t *testing.T) {
	for _, test := range []struct {
		name, redirectURL string
		secure            bool
	}{
		{name: "local HTTP", redirectURL: "http://localhost:8080/api/v1/auth/google/callback"},
		{name: "production HTTPS", redirectURL: "https://api.example.com/api/v1/auth/google/callback", secure: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			e := echo.New()
			NewHandler(oauth2.Config{RedirectURL: test.redirectURL}, nil, nil, "").Register(e)
			response := httptest.NewRecorder()
			e.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/auth/logout", nil))
			if response.Code != http.StatusNoContent {
				t.Fatalf("logout status = %d", response.Code)
			}
			cookies := response.Result().Cookies()
			if len(cookies) != 3 {
				t.Fatalf("logout cookies = %+v", cookies)
			}
			paths := map[string]string{accessCookieName: "/api/v1", refreshCookieName: "/api/v1/auth", flowCookieName: "/api/v1/auth/google"}
			for _, cookie := range cookies {
				if cookie.Path != paths[cookie.Name] || cookie.MaxAge >= 0 || cookie.Value != "" ||
					!cookie.HttpOnly || cookie.Secure != test.secure || cookie.SameSite != http.SameSiteLaxMode {
					t.Fatalf("cookie was not cleared: %+v", cookie)
				}
			}
		})
	}
}

func TestRefreshRotatesTokenAndLogoutRevokesIt(t *testing.T) {
	f := newOAuthFixture(t)
	flowCookie, nonce := f.start(t)
	f.signIDToken(t, nonce)
	state := strings.Split(flowCookie.Value, ".")[0]
	signIn := f.callback(flowCookie, "code=good-code&state="+url.QueryEscape(state))
	if signIn.Code != http.StatusSeeOther {
		t.Fatalf("sign-in status = %d", signIn.Code)
	}
	var original *http.Cookie
	for _, cookie := range signIn.Result().Cookies() {
		if cookie.Name == refreshCookieName {
			original = cookie
		}
	}
	if original == nil || len(f.users.sessions) != 1 {
		t.Fatalf("refresh session was not created: cookies=%v, sessions=%d", signIn.Result().Cookies(), len(f.users.sessions))
	}
	call := func(path string, cookie *http.Cookie) *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodPost, path, nil)
		if cookie != nil {
			request.AddCookie(cookie)
		}
		response := httptest.NewRecorder()
		f.handler.ServeHTTP(response, request)
		return response
	}
	renewed := call("/auth/refresh", original)
	if renewed.Code != http.StatusNoContent || len(f.users.sessions) != 1 {
		t.Fatalf("refresh status = %d, sessions = %d", renewed.Code, len(f.users.sessions))
	}
	var rotated *http.Cookie
	for _, cookie := range renewed.Result().Cookies() {
		if cookie.Name == refreshCookieName {
			rotated = cookie
		}
	}
	if rotated == nil || rotated.Value == original.Value || rotated.Path != "/api/v1/auth" {
		t.Fatalf("refresh token was not rotated: %+v", rotated)
	}
	if replay := call("/auth/refresh", original); replay.Code != http.StatusUnauthorized {
		t.Fatalf("old refresh token replay status = %d", replay.Code)
	}
	if missing := call("/auth/refresh", nil); missing.Code != http.StatusUnauthorized {
		t.Fatalf("missing refresh token status = %d", missing.Code)
	}
	if logout := call("/auth/logout", rotated); logout.Code != http.StatusNoContent || len(f.users.sessions) != 0 {
		t.Fatalf("logout status = %d, sessions = %d", logout.Code, len(f.users.sessions))
	}
	if revoked := call("/auth/refresh", rotated); revoked.Code != http.StatusUnauthorized {
		t.Fatalf("revoked refresh token status = %d", revoked.Code)
	}
}
