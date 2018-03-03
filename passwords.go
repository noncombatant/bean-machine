// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/scrypt"
	"log"
	"os"
	"path"
	"strings"
)

const (
	saltSize     = 16
	scryptLength = 16
	scryptN      = 16384
	scryptP      = 1
	scryptR      = 8
)

type Credentials map[string]string

// TODO: Add a `getStoredCredential` method that stores the passwords in memory,
// and only re-reads the whole database if it has changed since last read.
func readPasswordDatabase(pathname string) Credentials {
	passwords := make(Credentials)

	file, e := os.OpenFile(pathname, os.O_RDONLY, 0600)
	if e != nil {
		if os.IsNotExist(e) {
			return passwords
		}
		log.Fatalf("Could not open %q: %v", pathname, e)
	}
	defer file.Close()

	var username, password string
	for {
		_, e = fmt.Fscanf(file, "%s %s\n", &username, &password)
		if e != nil {
			break
		}
		passwords[username] = password
	}
	return passwords
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
		log.Fatalf("Could not obfuscate the password: %v", e)
	}
	return obfuscated
}

func setPassword() {
	salt := makeRandomBytes(saltSize)

	pathname := path.Join(configurationPathname, passwordsBasename)
	stored := readPasswordDatabase(pathname)

	file, e := os.OpenFile(pathname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if e != nil {
		log.Fatalf("Could not open %q: %v", pathname, e)
	}
	defer file.Close()

	username, password := promptForCredentials()
	obfuscated := obfuscatePassword([]byte(password), salt)
	stored[strings.ToLower(username)] = hex.EncodeToString(salt) + hex.EncodeToString(obfuscated)

	writePasswordDatabase(file, stored)
}

func mustWriteString(file *os.File, s string) {
	_, e := file.WriteString(s)
	if e != nil {
		log.Fatalf("Could not write string to file: %v", e)
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
		log.Fatalf("Could not decode stored credential: %v", e)
	}
	return decodedCredential[:saltSize], decodedCredential[saltSize:]
}

func checkPassword(stored Credentials, username, password string) bool {
	storedCredential, ok := stored[strings.ToLower(username)]
	// BUG: Timing oracle for username existence.
	if !ok {
		return false
	}

	salt, scrypted := getSaltAndScrypted(storedCredential)
	obfuscated := obfuscatePassword([]byte(password), salt)

	return 1 == subtle.ConstantTimeEq(1, int32(subtle.ConstantTimeCompare(obfuscated, scrypted)))
}
