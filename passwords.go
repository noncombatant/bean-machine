// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

// TODO: Make this an in-memory database class. Wrap a class around gob that
// reloads the gobs file when the file is newer than the last read. Use it for
// passwords, the catalog, and whatever else.

import (
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/scrypt"
	"os"
	"path"
	"strings"
	"time"
)

const (
	saltSize     = 16
	scryptLength = 16
	scryptN      = 16384
	scryptP      = 1
	scryptR      = 8
)

type Credentials map[string]string

var (
	lastCredentialRead time.Time
	credentials        = make(Credentials)
)

func normalizeUsername(username string) string {
	return strings.TrimSpace(strings.ToLower(username))
}

func readPasswordDatabase(pathname string) Credentials {
	file, info, e := OpenFileAndInfo(pathname)
	if e != nil {
		if os.IsNotExist(e) {
			return credentials
		}
		Logger.Fatalf("Could not open %q: %v", pathname, e)
	}
	defer file.Close()

	if info.ModTime().After(lastCredentialRead) {
		credentials = make(Credentials)
		var username, password string
		for {
			_, e := fmt.Fscanf(file, "%s %s\n", &username, &password)
			if e != nil {
				break
			}
			credentials[normalizeUsername(username)] = password
		}
		lastCredentialRead = info.ModTime()
	}

	return credentials
}

func promptForCredentials() (string, string) {
	var username, password string
	fmt.Print("Username: ")
	fmt.Scanln(&username)
	fmt.Print("Password: ")
	fmt.Scanln(&password)
	return username, password
}

func obfuscatePassword(password, salt []byte) []byte {
	obfuscated, e := scrypt.Key(password, salt, scryptN, scryptR, scryptP, scryptLength)
	if e != nil {
		Logger.Fatalf("Could not obfuscate the password: %v", e)
	}
	return obfuscated
}

func setPassword(configurationPathname string) {
	salt := MustMakeRandomBytes(saltSize)

	pathname := path.Join(configurationPathname, passwordsBasename)
	stored := readPasswordDatabase(pathname)

	file, e := os.OpenFile(pathname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if e != nil {
		Logger.Fatalf("Could not open %q: %v", pathname, e)
	}
	defer file.Close()

	username, password := promptForCredentials()
	obfuscated := obfuscatePassword([]byte(password), salt)
	stored[normalizeUsername(username)] = hex.EncodeToString(salt) + hex.EncodeToString(obfuscated)

	writePasswordDatabase(file, stored)
}

func mustWriteString(file *os.File, s string) {
	_, e := file.WriteString(s)
	if e != nil {
		Logger.Fatalf("Could not write to file: %v", e)
	}
}

func writePasswordDatabase(file *os.File, toBeStored Credentials) {
	for key, value := range toBeStored {
		mustWriteString(file, key)
		mustWriteString(file, " ")
		mustWriteString(file, value)
		mustWriteString(file, "\n")
	}
}

func getSaltAndScrypted(storedCredential string) ([]byte, []byte) {
	decodedCredential, e := hex.DecodeString(storedCredential)
	if e != nil {
		Logger.Fatalf("Could not decode stored credential: %v", e)
	}
	return decodedCredential[:saltSize], decodedCredential[saltSize:]
}

func checkPassword(stored Credentials, username, password string) bool {
	storedCredential, ok := stored[normalizeUsername(username)]
	// BUG: Timing oracle for username existence.
	if !ok {
		return false
	}

	salt, scrypted := getSaltAndScrypted(storedCredential)
	obfuscated := obfuscatePassword([]byte(password), salt)

	return 1 == subtle.ConstantTimeEq(1, int32(subtle.ConstantTimeCompare(obfuscated, scrypted)))
}
