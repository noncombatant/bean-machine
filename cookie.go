// Copyright 2022 by Chris Palmer (https://noncombatant.org), and released
// under the terms of the GNU GPL3. See web/index.html for more information.

package main

import (
	"encoding/base64"
	"net/http"
	"os"
	"path"
	"time"
)

var (
	cookieLifetime, _ = time.ParseDuration("24000h")
)

func getCookieLifetime() time.Time {
	return (time.Now()).Add(cookieLifetime)
}

func GetCookie(token string) *http.Cookie {
	return &http.Cookie{Name: "token", Value: token, Secure: true, HttpOnly: true, Expires: getCookieLifetime(), Path: "/"}
}

func CheckToken(token string, sessionsDirectoryPathname string) bool {
	if len(token) != encodedTokenLength {
		return false
	}
	if data, e := base64.URLEncoding.DecodeString(token); e != nil || len(data) != tokenLength {
		return false
	}
	_, e := os.Stat(path.Join(sessionsDirectoryPathname, token))
	return e == nil
}

func CreateToken(sessionsDirectoryPathname string) (string, error) {
	bytes := MustMakeRandomBytes(tokenLength)
	token := base64.URLEncoding.EncodeToString(bytes)
	pathname := path.Join(sessionsDirectoryPathname, token)

	if file, e := os.Create(pathname); e != nil {
		return "", e
	} else {
		if e := file.Close(); e != nil {
			return "", e
		}
	}
	return token, nil
}
