// Copyright 2016 by Chris Palmer (https://noncombatant.org)
// SPDX-License-Identifier: GPL-3.0

package main

import (
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"
)

const (
	configurationBasename     = ".bean-machine"
	serverKeyBasename         = "server-key.pem"
	serverCertificateBasename = "server-certificate.pem"
	passwordsBasename         = "passwords"
	sessionsDirectoryName     = "sessions"
)

func generateServerCredentials(hosts []string, configurationPathname string) (string, string) {
	certificatePathname := path.Join(configurationPathname, serverCertificateBasename)
	keyPathname := path.Join(configurationPathname, serverKeyBasename)

	_, e1 := os.Stat(certificatePathname)
	_, e2 := os.Stat(keyPathname)
	if e1 == nil && e2 == nil {
		return certificatePathname, keyPathname
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	key, der, e := generateCertificate(hosts, "Bean Machine Server", notBefore, notAfter)
	if e != nil {
		log.Fatal(e)
	}

	keyFile, e := os.OpenFile(keyPathname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if e != nil {
		log.Fatal(e)
	}
	if keyPEM, e := pemBlockForKey(key); e != nil {
		log.Fatal(e)
	} else {
		if e := pem.Encode(keyFile, keyPEM); e != nil {
			log.Fatal(e)
		}
	}
	if e := keyFile.Close(); e != nil {
		log.Fatal(e)
	}

	certificateFile, e := os.Create(certificatePathname)
	if e != nil {
		log.Fatal(e)
	}
	if e := pem.Encode(certificateFile, pemBlockForCertificate(der)); e != nil {
		log.Fatal(e)
	}
	if e := certificateFile.Close(); e != nil {
		log.Fatal(e)
	}

	return certificatePathname, keyPathname
}

func promptForCredentials(r io.Reader, w io.Writer) (string, string) {
	var username, password string
	fmt.Fprint(w, "Username: ")
	fmt.Fscanln(r, &username)
	fmt.Fprint(w, "Password: ")
	fmt.Fscanln(r, &password)
	return username, password
}

func getHomePathname() string {
	homes := []string{
		"HOME",
		"USERPROFILE",
	}

	for _, home := range homes {
		pathname := os.Getenv(home)
		if pathname != "" {
			return pathname
		}
	}

	return ""
}

func makeConfigurationDirectory(configurationPathname string) {
	if e := os.MkdirAll(configurationPathname, 0755); e != nil {
		log.Fatal(e)
	}
	if e := os.MkdirAll(path.Join(configurationPathname, sessionsDirectoryName), 0755); e != nil {
		log.Fatal(e)
	}
}

// `port` is a string (not an integer) of the form ":1234".
func serveApp(root, port, configurationPathname string, c *Catalog) {
	addresses, e := net.InterfaceAddrs()
	if e != nil || len(addresses) == 0 {
		log.Fatal(e)
	}

	message := "Starting the web server. Point your browser to any of these addresses:"
	if len(addresses) == 1 {
		message = "Starting the web server. Point your browser to this address:"
	}
	log.Print(message)

	var hosts []string
	for _, address := range addresses {
		switch a := address.(type) {
		case *net.IPNet:
			if a.IP.To4() == nil {
				continue
			}
			names, e := net.LookupAddr(a.IP.String())
			if e != nil || len(names) == 0 {
				log.Printf("    https://%s%s/", a.IP, port)
				hosts = append(hosts, "%s", a.IP.String())
			} else {
				for _, name := range names {
					log.Printf("    https://%s%s/", name, port)
					hosts = append(hosts, name)
				}
			}
		}
	}

	handler := httpHandler{Root: root, ConfigurationPathname: configurationPathname, Catalog: c, Logger: log.Default()}

	minifier := minify.New()
	minifier.AddFunc("text/css", css.Minify)
	minifier.AddFunc("text/html", html.Minify)
	minifier.AddFunc("image/svg+xml", svg.Minify)
	minifier.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	minifier.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	minifier.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)

	mux := http.NewServeMux()
	mux.Handle("/", gziphandler.GzipHandler(minifier.Middleware(&handler)))
	certificatePathname, keyPathname := generateServerCredentials(hosts, configurationPathname)
	log.Fatal(http.ListenAndServeTLS(port, certificatePathname, keyPathname, mux))
}

func printHelp() {
	fmt.Println(`Usage:

  bean-machine -m music-directory serve
  bean-machine set-password

Here is what the commands do:

  serve
    Scans music-directory for music files, and creates a database of their
    metadata.

    Starts a web server rooted at music-directory, and prints out the URL(s)
    of the Bean Machine web app.

  set-password
    Prompts for a username and password, and sets the password for the given
    username.`)
	os.Exit(1)
}

func assertDirectory(pathname string) {
	info, e := os.Stat(pathname)
	if e != nil {
		log.Fatal(e)
	}
	if !info.IsDir() {
		log.Fatalf("%q is not a directory", pathname)
	}
}

func main() {
	log.SetFlags(log.Ldate | log.LUTC | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	needsHelp1 := flag.Bool("help", false, "Print the help message.")
	needsHelp2 := flag.Bool("h", false, "Print the help message.")
	rawRoot := flag.String("m", "", "Set the music directory.")
	port := flag.Int("p", 0, "Set the port the server listens on.")
	flag.Parse()

	root := strings.TrimRight(*rawRoot, string(os.PathSeparator))

	portString := ":1234"
	if *port > 0 && *port < 65536 {
		portString = fmt.Sprintf(":%d", *port)
	} else if *port != 0 {
		log.Fatal("The port number must be in the range 1 – 65535.")
	}

	if *needsHelp1 || *needsHelp2 || flag.NArg() == 0 {
		printHelp()
	}

	configurationPathname := path.Join(getHomePathname(), configurationBasename)
	makeConfigurationDirectory(configurationPathname)

	for i := 0; i < flag.NArg(); i++ {
		command := flag.Arg(i)
		switch command {
		case "catalog":
			assertDirectory(root)
			c, e := newCatalog(log.Default(), root)
			if e != nil {
				log.Fatal(e)
			}
			e = c.writeToFile(path.Join(root, catalogBasename))
			if e != nil {
				log.Fatal(e)
			}
		case "help":
			printHelp()
		case "serve":
			assertDirectory(root)
			catalog, e := readCatalogFromFile(path.Join(root, catalogBasename))
			if e != nil {
				log.Fatal(e)
			}
			serveApp(root, portString, configurationPathname, catalog)
		case "set-password":
			username, password := promptForCredentials(os.Stdin, os.Stdout)
			if e := setPassword(configurationPathname, username, password); e != nil {
				log.Fatal(e)
			}
		default:
			printHelp()
		}
	}
}
