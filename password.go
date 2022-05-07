// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See web/index.html for more information.

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

type Credentials map[string]string

func normalizeUsername(username string) string {
	return strings.TrimSpace(strings.ToLower(username))
}

func ReadCredentials(pathname string) (Credentials, error) {
	file, _, e := OpenFileAndInfo(pathname)
	if e != nil {
		if os.IsNotExist(e) {
			return make(Credentials), nil
		}
		return nil, e
	}

	credentials := make(Credentials)
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

// TODO: This should go in main.go. TODO: Consider getting rid of username.
func promptForCredentials() (string, string) {
	var username, password string
	fmt.Print("Username: ")
	fmt.Scanln(&username)
	fmt.Print("Password: ")
	fmt.Scanln(&password)
	return username, password
}

func ObfuscatePassword(password, salt []byte) ([]byte, error) {
	return scrypt.Key(password, salt, scryptN, scryptR, scryptP, scryptLength)
}

func WriteCredentialsByPathname(pathname string, cs Credentials) error {
	w, e := os.OpenFile(pathname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if e != nil {
		return e
	}
	if e := WriteCredentials(w, cs); e != nil {
		w.Close()
		return e
	}
	return w.Close()
}

func WriteCredentials(w io.Writer, cs Credentials) error {
	for k, v := range cs {
		if _, e := fmt.Fprintf(w, "%s %s\n", k, v); e != nil {
			return e
		}
	}
	return nil
}

// TODO: This should be changed to (username, password string) error
func SetPassword(configurationPathname string) error {
	username, password := promptForCredentials()

	// TODO: Don't use Must; return the error
	salt := MustMakeRandomBytes(saltSize)
	obfuscated, e := ObfuscatePassword([]byte(password), salt)
	if e != nil {
		return nil
	}

	pathname := path.Join(configurationPathname, passwordsBasename)
	credentials, e := ReadCredentials(pathname)
	if e != nil {
		return e
	}
	credentials[normalizeUsername(username)] = hex.EncodeToString(salt) + hex.EncodeToString(obfuscated)
	return WriteCredentialsByPathname(pathname, credentials)
}

// Returns the salt and the obfuscated password.
func decodeObfuscated(obfuscated string) ([]byte, []byte, error) {
	decoded, e := hex.DecodeString(obfuscated)
	if e != nil {
		return nil, nil, e
	}
	return decoded[:saltSize], decoded[saltSize:], nil
}

func CheckPassword(stored Credentials, username, password string) (bool, error) {
	username = normalizeUsername(username)
	storedCredential, ok := stored[username]
	if !ok {
		return false, nil
	}
	salt, scrypted, e := decodeObfuscated(storedCredential)
	if e != nil {
		return false, e
	}
	obfuscated, e := ObfuscatePassword([]byte(password), salt)
	if e != nil {
		return false, e
	}
	return subtle.ConstantTimeEq(1, int32(subtle.ConstantTimeCompare(obfuscated, scrypted))) == 1, nil
}
