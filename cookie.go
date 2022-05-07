// Copyright 2022 by Chris Palmer (https://noncombatant.org), and released
// under the terms of the GNU GPL3. See web/index.html for more information.

package main

import (
	"net/http"
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
