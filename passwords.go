// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See web/index.html for more information.

package main

import (
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"log"
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

func readCredentials(pathname string) Credentials {
	file, _, e := OpenFileAndInfo(pathname)
	if e != nil {
		if os.IsNotExist(e) {
			return make(Credentials)
		}
		log.Fatal(e)
	}
	defer file.Close()

	credentials := make(Credentials)
	var username, password string
	for {
		_, e := fmt.Fscanf(file, "%s %s\n", &username, &password)
		if e != nil {
			break
		}
		credentials[normalizeUsername(username)] = password
	}
	return credentials
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
	credentials := readCredentials(pathname)
	credentials[normalizeUsername(username)] = hex.EncodeToString(salt) + hex.EncodeToString(obfuscated)
	return WriteCredentialsByPathname(pathname, credentials)
}

func getSaltAndScrypted(storedCredential string) ([]byte, []byte) {
	decodedCredential, e := hex.DecodeString(storedCredential)
	if e != nil {
		log.Fatal(e)
	}
	return decodedCredential[:saltSize], decodedCredential[saltSize:]
}

func CheckPassword(stored Credentials, username, password string) (bool, error) {
	username = normalizeUsername(username)
	storedCredential, ok := stored[username]
	if !ok {
		return false, nil
	}
	salt, scrypted := getSaltAndScrypted(storedCredential)
	obfuscated, e := ObfuscatePassword([]byte(password), salt)
	if e != nil {
		return false, e
	}
	return subtle.ConstantTimeEq(1, int32(subtle.ConstantTimeCompare(obfuscated, scrypted))) == 1, nil
}
