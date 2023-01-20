package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"testingCourserWeb/pkg/data"
	"time"
)

func Test_app_authenticate(t *testing.T) {
	var theTests = []struct {
		name               string
		requestBody        string
		expectedStatusCode int
	}{
		{"valid user", `{"email": "admin@example.com", "password":"secret"}`, http.StatusOK},
		{"not json", `This is not json`, http.StatusUnauthorized},
		{"empty json", `{}`, http.StatusUnauthorized},
		{"empty email", `{"email": "", "password":"secret"}`, http.StatusUnauthorized},
		{"no password", `{"email": "admin@example.com", "password":""}`, http.StatusUnauthorized},
		{"another domain", `{"email": "admin@someexample.com", "password":"secret"}`, http.StatusUnauthorized},
	}

	for _, e := range theTests {
		var reader io.Reader
		reader = strings.NewReader(e.requestBody)
		req, _ := http.NewRequest("POST", "/auth", reader)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(app.authenticate)

		handler.ServeHTTP(rr, req)
		if e.expectedStatusCode != rr.Code {
			t.Errorf("%s; returned wrong status code; expected %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
	}
}

func Test_app_refresh(t *testing.T) {
	var tests = []struct {
		name               string
		token              string
		expectedStatusCode int
		resetRefreshTime   bool
	}{
		{"valid", "", http.StatusOK, true},
		{"valid but not yet ready to expire", "", http.StatusTooEarly, false},
		{"expired token", expiredToken, http.StatusBadRequest, false},
	}
	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	oldRefreshTime := refreshTokenExpiry

	for _, e := range tests {
		var tkn string
		if e.token == "" {
			if e.resetRefreshTime {
				refreshTokenExpiry = time.Second * 1
			}
			tokens, _ := app.generateTokenPair(&testUser)
			tkn = tokens.RefreshToken
		} else {
			tkn = e.token
		}

		postedData := url.Values{
			"refresh_token": {tkn},
		}

		req, _ := http.NewRequest("POST", "/refresh-token", strings.NewReader(postedData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(app.refresh)
		handler.ServeHTTP(rr, req)
		if rr.Code != e.expectedStatusCode {
			t.Errorf("%s: expected status of %d, but got %d", e.name, e.expectedStatusCode, rr.Code)
		}
		refreshTokenExpiry = oldRefreshTime
	}
}

func Test_app_userHandlers(t *testing.T) {
	var tests = []struct {
		name           string
		method         string
		json           string
		paramID        string
		handler        http.HandlerFunc
		expectedStatus int
	}{
		{"allUsers", "GET", "", "", app.allUsers, http.StatusOK},
		{"deleteUser", "DELETE", "", "1", app.deleteUser, http.StatusNoContent},
		{"deleteUser var url param", "DELETE", "", "x", app.deleteUser, http.StatusBadRequest},
		{"get user valid", "GET", "", "1", app.getUser, http.StatusOK},
		{"get user invalid", "GET", "", "100", app.getUser, http.StatusBadRequest},
		{"get user bad url param", "GET", "", "1y", app.getUser, http.StatusBadRequest},
		{"update user valid",
			"PATCH",
			`{
					"id": 1,
					"first_name": "Administrator",
					"last_name": "User",
					"email": "admin@example.com"
			}`, "1", app.updateUser,
			http.StatusNoContent,
		},
		{"update user invalid",
			"PATCH",
			`{
					"id": 20,
					"first_name": "Administrator",
					"last_name": "User",
					"email": "admin@example.com"
			}`, "1", app.updateUser,
			http.StatusBadRequest,
		},
		{"update user invalid JSON",
			"PATCH",
			`{
					"id": 1,
					first_name: "Administrator",
					"last_name": "User",
					"email": "admin@example.com"
			}`, "1", app.updateUser,
			http.StatusBadRequest,
		},
		{"insert user valid",
			"PUT",
			`{
					"id": 1,
					"first_name": "Jack",
					"last_name": "User",
					"email": "admin@example.com"
			}`, "1", app.insertUser,
			http.StatusNoContent,
		},
		{"insert user invalid",
			"PUT",
			`{
					"id": 1,
					"foo": "bar",
					"first_name": "Jack",
					"last_name": "User",
					"email": "admin@example.com"
			}`, "1", app.insertUser,
			http.StatusBadRequest,
		},
		{"insert user invalid JSON",
			"PUT",
			`{
					"id": 1,
					first_name: "Jack",
					"last_name": "User",
					"email": "admin@example.com"
			}`, "1", app.insertUser,
			http.StatusBadRequest,
		},
	}

	for _, e := range tests {
		var req *http.Request
		if e.json == "" {
			req, _ = http.NewRequest(e.method, "/", nil)
		} else {
			req, _ = http.NewRequest(e.method, "/", strings.NewReader(e.json))
		}

		if e.paramID != "" {
			chiCtx := chi.NewRouteContext()
			chiCtx.URLParams.Add("userID", e.paramID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
		}
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(e.handler)

		handler.ServeHTTP(rr, req)

		if rr.Code != e.expectedStatus {
			t.Errorf("%s; wrong status returned; expected %d, but got %d", e.name, e.expectedStatus, rr.Code)
		}

	}

}

func Test_app_refreshUsingCookie(t *testing.T) {
	testUser := data.User{
		ID:        1,
		FirstName: "Admin",
		LastName:  "User",
		Email:     "admin@example.com",
	}

	tokens, _ := app.generateTokenPair(&testUser)

	testCookie := &http.Cookie{
		Name:     "__Host-refresh_token",
		Path:     "/",
		Value:    tokens.RefreshToken,
		Expires:  time.Now().Add(refreshTokenExpiry),
		MaxAge:   int(refreshTokenExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   "localhost",
		HttpOnly: true,
		Secure:   true,
	}

	badCookie := &http.Cookie{
		Name:     "__Host-refresh_token",
		Path:     "/",
		Value:    "somebadstring",
		Expires:  time.Now().Add(refreshTokenExpiry),
		MaxAge:   int(refreshTokenExpiry.Seconds()),
		SameSite: http.SameSiteStrictMode,
		Domain:   "localhost",
		HttpOnly: true,
		Secure:   true,
	}

	var tests = []struct {
		name           string
		addCookie      bool
		cookie         *http.Cookie
		expectedStatus int
	}{
		{"valid cookie", true, testCookie, http.StatusOK},
		{"invalid cookie", true, badCookie, http.StatusBadRequest},
		{"no cookie", false, nil, http.StatusUnauthorized},
	}

	for _, e := range tests {
		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		if e.addCookie {
			req.AddCookie(e.cookie)
		}
		handler := http.HandlerFunc(app.refreshUsingCookie)
		handler.ServeHTTP(rr, req)
		if rr.Code != e.expectedStatus {
			t.Errorf("%s: wrong status code returned; expected %d but got %d", e.name, e.expectedStatus, rr.Code)
		}
	}
}

func Test_app_deleteRefreshCookie(t *testing.T) {
	req, _ := http.NewRequest("GET", "/logout", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(app.deleteRefreshCookie)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Errorf("wrong status: expected %d but got %d", http.StatusAccepted, rr.Code)
	}

	foundCookie := false
	for _, c := range rr.Result().Cookies() {
		if c.Name == "__Host-refresh_token" {
			foundCookie = true
			if c.Expires.After(time.Now()) {
				t.Errorf("cookie expiration in future, and should not be: %v", c.Expires.UTC())
			}
		}
	}

	if !foundCookie {
		t.Error("__Host-refresh_token cookie not found")
	}

}
