// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See web/index.html for more information.

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pkg/xattr"
)

const (
	configurationBasename     = ".bean-machine"
	serverKeyBasename         = "server-key.pem"
	serverCertificateBasename = "server-certificate.pem"
	passwordsBasename         = "passwords"
	sessionsDirectoryName     = "sessions"
)

var (
	combiningCharacterReplacements = map[string]string{
		"á": "á",
		"à": "à",
		"ä": "ä",
		"â": "â",
		"ç": "ç",
		"è": "è",
		"É": "É",
		"È": "È",
		"ë": "ë",
		"é": "é",
		"ê": "ê",
		"í": "í",
		"ì": "ì",
		"ï": "ï",
		"ñ": "ñ",
		"Ó": "Ó",
		"ó": "ó",
		"ò": "ò",
		"ö": "ö",
		"ô": "ô",
		"ř": "ř",
		"ú": "ú",
		"ù": "ù",
		"ü": "ü",
		"ž": "ž",
	}
)

func Lint(root string) {
	e := filepath.Walk(root,
		func(pathname string, info os.FileInfo, e error) error {
			if e != nil {
				log.Print(e)
				return e
			}

			basename := path.Base(pathname)
			if basename == ".AppleFileInfo" && info.IsDir() {
				return os.RemoveAll(pathname)
			} else if basename == ".DS_Store" && !info.IsDir() {
				return os.Remove(pathname)
			} else if basename[0] == '.' {
				log.Printf("Hidden: %q", pathname)
			}

			file, e := os.OpenFile(pathname, os.O_RDONLY, 0755)
			if e != nil {
				log.Print(e)
				return e
			}
			defer file.Close()

			if info.IsDir() {
				empty, e := IsDirectoryEmpty(pathname)
				if e != nil {
					log.Printf("%q: %v", pathname, e)
					//return e
				}
				if empty {
					e = os.Remove(pathname)
					if e != nil {
						log.Printf("%q: %v", pathname, e)
					}
					return e
				}

				if info.Mode().Perm() != 0755 {
					e = file.Chmod(0755)
					if e != nil {
						log.Print(e)
						return e
					}
				}
			} else if info.Mode().IsRegular() {
				if info.Size() == 0 {
					return os.Remove(pathname)
				}

				if info.Mode().Perm() != 0644 {
					e = file.Chmod(0644)
					if e != nil {
						log.Print(e)
						return e
					}
				}
			}

			xattrs, e := xattr.FList(file)
			if e != nil {
				log.Print(e)
				return e
			}
			for _, name := range xattrs {
				e = xattr.FRemove(file, name)
				if e != nil {
					log.Print(e)
					return e
				}
			}

			savedPathname := pathname
			for k, v := range combiningCharacterReplacements {
				pathname = strings.ReplaceAll(pathname, k, v)
			}
			if savedPathname != pathname {
				e := os.Rename(savedPathname, pathname)
				if e != nil {
					log.Print(e)
				}
			}

			return nil
		})

	if e != nil {
		log.Print(e)
	}
}

func generateServerCredentials(hosts []string, configurationPathname string) (string, string) {
	certificatePathname := path.Join(configurationPathname, serverCertificateBasename)
	keyPathname := path.Join(configurationPathname, serverKeyBasename)

	_, e1 := os.Stat(certificatePathname)
	_, e2 := os.Stat(keyPathname)
	if e1 == nil && e2 == nil {
		return certificatePathname, keyPathname
	}

	certificateFile, e := os.Create(certificatePathname)
	if e != nil {
		log.Fatal(e)
	}

	keyFile, e := os.OpenFile(keyPathname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if e != nil {
		log.Fatal(e)
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

	certificatePathname, keyPathname := generateServerCredentials(hosts, configurationPathname)
	handler := HTTPHandler{Root: root, ConfigurationPathname: configurationPathname, Catalog: c, Logger: log.Default()}
	log.Fatal(http.ListenAndServeTLS(port, certificatePathname, keyPathname, &handler))
}

func printHelp() {
	fmt.Println(`Usage:

  bean-machine -m music-directory serve
  bean-machine -m music-directory lint
  bean-machine set-password

Here is what the commands do:

  serve
    Scans music-directory for music files, and creates a database of their
    metadata.

    Starts a web server rooted at music-directory, and prints out the URL(s)
    of the Bean Machine web app.

  lint
    Scans music-directory for junk and empty files and removes them. Sets
    file and directory permissions, and removes POSIX extended attributes.

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
			c, e := BuildCatalog(log.Default(), root)
			if e != nil {
				log.Fatal(e)
			}
			e = WriteCatalogByPathname(path.Join(root, catalogFile), c)
			if e != nil {
				log.Fatal(e)
			}
		case "help":
			printHelp()
		case "lint":
			assertDirectory(root)
			Lint(root)
		case "serve":
			assertDirectory(root)
			catalog, e := ReadCatalogByPathname(path.Join(root, catalogFile))
			if e != nil {
				log.Fatal(e)
			}
			serveApp(root, portString, configurationPathname, catalog)
		case "set-password":
			setPassword(configurationPathname)
		default:
			printHelp()
		}
	}
}
