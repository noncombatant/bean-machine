// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See web/index.html for more information.

package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

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
	// TODO: Make this map more complete.
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

func installFrontEndFiles(root string) {
	webDirectoryPathname := "web"
	webDirectory, e := os.Open(webDirectoryPathname)
	if e != nil {
		Logger.Fatal(e)
	}
	defer webDirectory.Close()

	files, e := webDirectory.Readdirnames(0)
	if e != nil {
		Logger.Fatal(e)
	}

	for _, f := range files {
		MustCopyFileByName(path.Join(root, f), path.Join(webDirectoryPathname, f))
	}
}

func Lint(root string) {
	e := filepath.Walk(root,
		func(pathname string, info os.FileInfo, e error) error {
			if e != nil {
				Logger.Print(e)
				return e
			}

			basename := path.Base(pathname)
			if basename == ".AppleFileInfo" && info.IsDir() {
				return os.RemoveAll(pathname)
			} else if basename == ".DS_Store" && !info.IsDir() {
				return os.Remove(pathname)
			} else if basename[0] == '.' {
				Logger.Printf("Hidden: %q", pathname)
			}

			file, e := os.OpenFile(pathname, os.O_RDONLY, 0755)
			if e != nil {
				Logger.Print(e)
				return e
			}
			defer file.Close()

			if info.IsDir() {
				empty, e := IsDirectoryEmpty(pathname)
				if e != nil {
					Logger.Print(e)
					return e
				}
				if empty {
					return os.Remove(pathname)
				}

				if info.Mode().Perm() != 0755 {
					e = file.Chmod(0755)
					if e != nil {
						Logger.Print(e)
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
						Logger.Print(e)
						return e
					}
				}
			}

			xattrs, e := xattr.FList(file)
			if e != nil {
				Logger.Print(e)
				return e
			}
			for _, name := range xattrs {
				e = xattr.FRemove(file, name)
				if e != nil {
					Logger.Print(e)
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
					Logger.Print(e)
				}
			}

			return nil
		})

	if e != nil {
		Logger.Print(e)
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
		Logger.Fatal(e)
	}

	keyFile, e := os.OpenFile(keyPathname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if e != nil {
		Logger.Fatal(e)
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
		Logger.Fatal(e)
	}
	if e := os.MkdirAll(path.Join(configurationPathname, sessionsDirectoryName), 0755); e != nil {
		Logger.Fatal(e)
	}
}

func monitorCatalogForUpdates(root string) {
	for {
		time.Sleep(2 * time.Minute)
		catalog.BuildCatalog(root)
	}
}

// `port` is a string (not an integer) of the form ":1234".
func serveApp(root, port, configurationPathname string) {
	addresses, e := net.InterfaceAddrs()
	if e != nil || len(addresses) == 0 {
		Logger.Fatal(e)
	}

	message := "Starting the web server. Point your browser to any of these addresses:"
	if len(addresses) == 1 {
		message = "Starting the web server. Point your browser to this address:"
	}
	Logger.Print(message)

	var hosts []string
	for _, address := range addresses {
		switch a := address.(type) {
		case *net.IPNet:
			// Skip non-IPv4 (i.e. IPv6) addresses, because reverse DNS is rarely
			// configured properly (I guess) and the lookup timeouts slow down server
			// startup. TODO: Maybe fix this someday.
			if a.IP.To4() == nil {
				continue
			}
			names, e := net.LookupAddr(a.IP.String())
			if e != nil || len(names) == 0 {
				Logger.Printf("    https://%s%s/", a.IP, port)
				hosts = append(hosts, "%s", a.IP.String())
			} else {
				for _, name := range names {
					Logger.Printf("    https://%s%s/", name, port)
					hosts = append(hosts, name)
				}
			}
		}
	}

	certificatePathname, keyPathname := generateServerCredentials(hosts, configurationPathname)
	go monitorCatalogForUpdates(root)
	handler := HTTPHandler{Root: root, ConfigurationPathname: configurationPathname}
	Logger.Fatal(http.ListenAndServeTLS(port, certificatePathname, keyPathname, &handler))
}

func printHelp() {
	fmt.Println(`Usage:

  bean-machine -m music-directory serve
  bean-machine -m music-directory lint
  bean-machine set-password

Here is what the commands do:

  serve
    Installs the web app front-end files in music-directory.

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

func assertValidRootPathname(root string) {
	info, e := os.Stat(root)
	if e != nil || !info.IsDir() {
		Logger.Fatal("Cannot continue without a valid music-directory.")
	}
}

func main() {
	needsHelp1 := flag.Bool("help", false, "Print the help message.")
	needsHelp2 := flag.Bool("h", false, "Print the help message.")
	root := flag.String("m", "", "Set the music directory.")
	port := flag.Int("p", 0, "Set the port the server listens on.")
	flag.Parse()

	cleanedRoot := strings.TrimRight(*root, string(os.PathSeparator))

	portString := ":1234"
	if *port > 0 && *port < 65536 {
		portString = fmt.Sprintf(":%d", *port)
	} else if *port != 0 {
		Logger.Fatal("The port number must be in the range 1 – 65535.")
	}

	if *needsHelp1 || *needsHelp2 || flag.NArg() == 0 {
		printHelp()
	}

	configurationPathname := path.Join(getHomePathname(), configurationBasename)
	makeConfigurationDirectory(configurationPathname)

	Logger.Printf("Music: %q, configuration: %q", cleanedRoot, configurationPathname)

	for i := 0; i < flag.NArg(); i++ {
		command := flag.Arg(i)
		switch command {
		case "lint":
			assertValidRootPathname(cleanedRoot)
			Lint(cleanedRoot)
		case "help":
			printHelp()
		case "serve":
			assertValidRootPathname(cleanedRoot)
			installFrontEndFiles(cleanedRoot)
			catalog.BuildCatalog(cleanedRoot)
			serveApp(cleanedRoot, portString, configurationPathname)
		case "set-password":
			setPassword(configurationPathname)
		default:
			printHelp()
		}
	}
	os.Exit(0)
}
