package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type HTTPAdminSession struct {
	UUID      string
	Username  string
	LoginTime time.Time
}

func (h *HTTPAdminSession) IsValid() bool {
	return h.LoginTime.Add(time.Hour).Before(time.Now())
}

type HTTPAdminJWTToken struct {
	Token     string
	SessionID string
	Expires   time.Time
}

func (h *HTTPAdminJWTToken) IsValid() bool {
	return h.Expires.Before(time.Now())
}

type HTTPAdminTokenData struct {
	Access  *HTTPAdminJWTToken
	Refresh *HTTPAdminJWTToken
}

func (s *HTTPDService) adminLoginTokenExtract(r *http.Request) string {
	cookie, err := r.Cookie("token")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func (s *HTTPDService) adminLoginTokenVerify(r *http.Request) (*jwt.Token, error) {
	tokenString := s.adminLoginTokenExtract(r)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.Config.Values.HTTPD.Secrets.Authentication), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (s *HTTPDService) adminIsTokenValid(r *http.Request) error {
	token, err := s.adminLoginTokenVerify(r)
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); !ok || !token.Valid {
		return err
	}
	return nil
}

func (s *HTTPDService) adminExtractTokenData(r *http.Request) (string, error) {
	token, err := s.adminLoginTokenVerify(r)
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		uid, ok := claims["session"].(string)
		if !ok {
			return "", err
		}
		return uid, nil
	}
	return "", err
}

func (s *HTTPDService) adminCreateToken() (*HTTPAdminTokenData, error) {
	td := &HTTPAdminTokenData{
		Access: &HTTPAdminJWTToken{
			SessionID: uuid.New().String(),
			Expires:   time.Now().Add(time.Hour),
		},
		Refresh: &HTTPAdminJWTToken{
			SessionID: uuid.New().String(),
			Expires:   time.Now().Add(time.Hour * 24),
		},
	}

	var err error
	accessClaims := jwt.MapClaims{}
	accessClaims["session"] = td.Access.SessionID
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	td.Access.Token, err = accessToken.SignedString([]byte(s.Config.Values.HTTPD.Secrets.Authentication))
	if err != nil {
		return nil, err
	}

	refreshClaims := jwt.MapClaims{}
	refreshClaims["session"] = td.Refresh.SessionID
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	td.Refresh.Token, err = refreshToken.SignedString([]byte(s.Config.Values.HTTPD.Secrets.Refresh))
	if err != nil {
		return nil, err
	}

	return td, nil
}

func (s *HTTPDService) adminCreateSession(username string, td *HTTPAdminTokenData) error {
	cache := s.cache[HTTPCacheAdminSessions].(map[string]*HTTPAdminSession)
	accessSesh := &HTTPAdminSession{
		UUID:      td.Access.SessionID,
		Username:  username,
		LoginTime: td.Access.Expires,
	}
	cache[accessSesh.UUID] = accessSesh

	refreshSesh := &HTTPAdminSession{
		UUID:      td.Access.SessionID,
		Username:  username,
		LoginTime: td.Refresh.Expires,
	}
	cache[accessSesh.UUID] = refreshSesh

	return nil
}
