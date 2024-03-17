// Copyright 2016 by Chris Palmer (https://noncombatant.org)
// SPDX-License-Identifier: GPL-3.0

package main

import (
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"golang.org/x/crypto/scrypt"
)

const (
	saltSize     = 16
	scryptLength = 16
	scryptN      = 16384
	scryptP      = 1
	scryptR      = 8
)

type credentials map[string]string

func normalizeUsername(username string) string {
	return strings.TrimSpace(strings.ToLower(username))
}

func readCredentials(pathname string) (credentials, error) {
	file, _, e := openFileAndInfo(pathname)
	if e != nil {
		if os.IsNotExist(e) {
			return make(credentials), nil
		}
		return nil, e
	}

	credentials := make(credentials)
	var username, password string
	for {
		_, e := fmt.Fscanf(file, "%s %s\n", &username, &password)
		if e != nil {
			if e != io.EOF {
				return nil, e
			}
			break
		}
		credentials[normalizeUsername(username)] = password
	}

	if e := file.Close(); e != nil {
		return nil, e
	}
	return credentials, nil
}

func obfuscatePassword(password, salt []byte) ([]byte, error) {
	return scrypt.Key(password, salt, scryptN, scryptR, scryptP, scryptLength)
}

func writeCredentialsByPathname(pathname string, cs credentials) error {
	w, e := os.OpenFile(pathname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if e != nil {
		return e
	}
	if e := writeCredentials(w, cs); e != nil {
		_ = w.Close()
		return e
	}
	return w.Close()
}

func writeCredentials(w io.Writer, cs credentials) error {
	for k, v := range cs {
		if _, e := fmt.Fprintf(w, "%s %s\n", k, v); e != nil {
			return e
		}
	}
	return nil
}

func setPassword(configurationPathname, username, password string) error {
	salt := getRandomBytes(saltSize)
	obfuscated, e := obfuscatePassword([]byte(password), salt)
	if e != nil {
		return e
	}

	pathname := path.Join(configurationPathname, passwordsBasename)
	credentials, e := readCredentials(pathname)
	if e != nil {
		return e
	}
	credentials[normalizeUsername(username)] = hex.EncodeToString(salt) + hex.EncodeToString(obfuscated)
	return writeCredentialsByPathname(pathname, credentials)
}

// Returns the salt and the obfuscated password.
func decodeObfuscated(obfuscated string) ([]byte, []byte, error) {
	decoded, e := hex.DecodeString(obfuscated)
	if e != nil {
		return nil, nil, e
	}
	return decoded[:saltSize], decoded[saltSize:], nil
}

func checkPassword(stored credentials, username, password string) (bool, error) {
	username = normalizeUsername(username)
	storedCredential, ok := stored[username]
	if !ok {
		return false, nil
	}
	salt, scrypted, e := decodeObfuscated(storedCredential)
	if e != nil {
		return false, e
	}
	obfuscated, e := obfuscatePassword([]byte(password), salt)
	if e != nil {
		return false, e
	}
	return subtle.ConstantTimeEq(1, int32(subtle.ConstantTimeCompare(obfuscated, scrypted))) == 1, nil
}
