// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
)

const (
	configurationBasename     = ".bean-machine"
	serverKeyBasename         = "server-key.pem"
	serverCertificateBasename = "server-certificate.pem"
	passwordsBasename         = "passwords"
	httpPort                  = ":1234"
)

var (
	frontEndFiles = []string{
		"login.html",
		"help.html",

		"manifest.json",

		"bean-machine.css",
		"bean-machine.js",
		"index.html",
		"search.js",

		"clef-192.png",
		"clef-144.png",
		"clef-96.png",
		"clef-64.png",
		"favicon.ico",
		"pause.png",
		"play.png",
		"unknown-album.png",
	}

	homePathname          = getHomePathname()
	configurationPathname = path.Join(homePathname, configurationBasename)
	bindToIPv4            = true
	bindToIPv6            = false

	musicRoot = ""
)

func installFrontEndFiles(root string) {
	assertValidRootPathname(root)
	for _, f := range frontEndFiles {
		copyFile(f, path.Join(root, string(os.PathSeparator), f))
	}
}

func generateServerCredentials(hosts []string) (string, string) {
	certificatePathname := path.Join(configurationPathname, serverCertificateBasename)
	keyPathname := path.Join(configurationPathname, serverKeyBasename)

	_, e1 := os.Stat(certificatePathname)
	_, e2 := os.Stat(keyPathname)
	if e1 == nil && e2 == nil {
		return certificatePathname, keyPathname
	}

	certificateFile, err := os.Create(certificatePathname)
	if err != nil {
		log.Fatalf("generateServerCredentials: Failed to open %q for writing: %s", certificatePathname, err)
	}

	keyFile, err := os.OpenFile(keyPathname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("generateServerCredentials: Failed to open %q for writing: %s", keyPathname, err)
	}

	generateCertificate(hosts, false, keyFile, certificateFile)
	certificateFile.Close()
	keyFile.Close()
	return certificatePathname, keyPathname
}

func getHomePathname() string {
	homes := []string{
		"HOME",
		"USERPROFILE",
	}

	for _, home := range homes {
		pathname := os.Getenv(home)
		if "" != pathname {
			return pathname
		}
	}

	log.Fatal("getHomePathname: No HOME environment variable is set.")
	return ""
}

func establishConfiguration() {
	if homePathname == "" {
		log.Fatal("establishConfiguration: No HOME environment variable is set.")
	}

	if e := os.MkdirAll(configurationPathname, 0755); e != nil {
		log.Fatalf("establishConfiguration: Could not create %q: %v", configurationPathname, e)
	}
}

func serveApp(root string) {
	assertValidRootPathname(root)
	addresses, e := net.InterfaceAddrs()
	if e != nil || 0 == len(addresses) {
		log.Fatal("serveApp: Can't find any network interfaces to run the web server on. Giving up.")
	}

	message := "Starting the web server. Point your browser to any of these addresses:"
	if 1 == len(addresses) {
		message = "Starting the web server. Point your browser to this address:"
	}
	log.Printf("serveApp: %s", message)

	var hosts []string
	for _, address := range addresses {
		switch a := address.(type) {
		case *net.IPNet:
			if !bindToIPv4 && a.IP.To4() != nil || !bindToIPv6 && a.IP.To4() == nil {
				continue
			}
			names, e := net.LookupAddr(a.IP.String())
			if e != nil || 0 == len(names) {
				log.Printf("    https://%s%s/", a.IP, httpPort)
				hosts = append(hosts, fmt.Sprintf("%s", a.IP))
			} else {
				for _, name := range names {
					log.Printf("    https://%s%s/", name, httpPort)
					hosts = append(hosts, fmt.Sprintf("%s", name))
				}
			}
		}
	}

	certificatePathname, keyPathname := generateServerCredentials(hosts)
	handler := AuthenticatingFileHandler{Root: root}
	log.Fatal(http.ListenAndServeTLS(httpPort, certificatePathname, keyPathname, handler))
}

func printHelp() {
	fmt.Println(`Usage:

  bean-machine -m music-directory
  bean-machine -m music-directory command [command...]
  bean-machine set-password

Invoking bean-machine with no command is equivalent to invoking it with the
"catalog", "install", and "serve" commands (see below). The commands are:

  catalog
    Scans music-directory for music files, and creates a database of their
    metadata in music-directory/catalog.tsv.

  install
    Installs the web front-end files in music-directory.

  serve
    Starts a web server rooted at music-directory, and prints out the URL(s)
    of the Bean Machine web app.

There are 2 additional commands for managing the password authentication for the
web app:

  set-password
    Prompts for a username and password, and sets the password for the given
    username.`)
	os.Exit(1)
}

func main() {
	establishConfiguration()

	if os.Getenv("IPV6") != "" {
		bindToIPv6 = true
	}

	needsHelp1 := flag.Bool("help", false, "Print the help message.")
	needsHelp2 := flag.Bool("h", false, "Print the help message.")
	root := flag.String("m", "", "Set the music directory.")
	flag.Parse()
	musicRoot = *root

	if *needsHelp1 || *needsHelp2 || flag.NArg() == 0 {
		printHelp()
	}

	status := 0
	for i := 0; i < flag.NArg(); i++ {
		command := flag.Arg(i)
		switch command {
		case "catalog":
			buildCatalog(musicRoot)
		case "help":
			printHelp()
		case "install":
			installFrontEndFiles(musicRoot)
		case "serve":
			serveApp(musicRoot)
		case "set-password":
			setPassword()
		default:
			printHelp()
		}
	}
	os.Exit(status)
}
