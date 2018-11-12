// Copyright 2016 by Chris Palmer (https://noncombatant.org), and released under
// the terms of the GNU GPL3. See help.html for more information.

package main

import (
	"flag"
	"fmt"
	"id3"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
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
		"index.html",
		"bean-machine.js",

		"core.js",
		"util.js",
		"search.js",

		"favicon.ico",
		"clef-192.png",
		"clef-144.png",
		"clef-96.png",
		"clef-64.png",
		"pause.png",
		"play.png",
		"repeat.png",
		"search.png",
		"shuffle.png",
		"skip.png",
	}

	homePathname          = getHomePathname()
	configurationPathname = path.Join(homePathname, configurationBasename)
	bindToIPv4            = true
	bindToIPv6            = false

	musicRoot = ""

	// NOTE: These must be kept in sync with the format extensions arrays in the
	// JS code.
	audioFormatExtensions = []string{
		".flac",
		".m4a",
		".mid",
		".midi",
		".mp3",
		".ogg",
		".wav",
		".wave",
	}
	videoFormatExtensions = []string{
		".avi",
		".mkv",
		".mov",
		".mp4",
		".mpeg",
		".mpg",
		".ogv",
		".webm",
	}
)

func isAudioPathname(pathname string) bool {
	return isStringInStrings(getFileExtension(pathname), audioFormatExtensions)
}

func isVideoPathname(pathname string) bool {
	return isStringInStrings(getFileExtension(pathname), videoFormatExtensions)
}

func fileSizesToPathnames(root string) map[int64][]string {
	m := make(map[int64][]string)
	e := filepath.Walk(root, func(pathname string, info os.FileInfo, e error) error {
		if e == nil && info.Mode().IsRegular() && !shouldSkipFile(pathname, info) {
			s := info.Size()
			m[s] = append(m[s], pathname)
		}
		return nil
	})
	if e != nil {
		log.Printf("%q: %v\n", root, e)
	}
	return m
}

type ItemInfo struct {
	Pathname string
	Album    string
	Artist   string
	Name     string
	Disc     string
	Track    string
	Year     string
	Genre    string
	ModTime  time.Time
	File     *id3.File
}

func (i *ItemInfo) normalize() {
	if i.File != nil {
		if i.File.Album != "" {
			i.Album = i.File.Album
		}
		if i.File.Artist != "" {
			i.Artist = i.File.Artist
		}
		if i.File.Name != "" {
			i.Name = i.File.Name
		}
		if i.File.Disc != "" {
			i.Disc = i.File.Disc
		}
		if i.File.Track != "" {
			i.Track = i.File.Track
		}
		if i.File.Year != "" {
			i.Year = i.File.Year
		}
		if i.File.Genre != "" {
			i.Genre = i.File.Genre
		}
	}

	if i.Artist == "" || i.Album == "" || i.Name == "" {
		// Get info from pathname, assuming format:
		// "AC_DC/Back In Black/1-01 Hells Bells.m4a"
		//     performer/album/disc#-track# name
		parts := strings.Split(i.Pathname, string(filepath.Separator))
		if len(parts) != 3 {
			if i.Name == "" {
				i.Name = i.Pathname
			}
		} else {
			if i.Artist == "" {
				i.Artist = parts[0]
			}
			if i.Album == "" {
				i.Album = parts[1]
			}
			if i.Name == "" {
				i.Name = removeFileExtension(parts[2])
			}
		}
	}

	if i.Artist == "" {
		i.Artist = "Unknown Artist"
	}
	if i.Album == "" {
		i.Album = "Unknown Album"
	}
	if i.Name == "" {
		i.Name = "Unknown Item"
	}
	if i.Disc == "" {
		i.Disc = "1"
	}
	if i.Track == "" {
		i.Track = "1"
	}

	i.Disc = normalizeNumericString(i.Disc)
	i.Track = normalizeNumericString(i.Track)
	i.Year = normalizeNumericString(i.Year)
}

func (i *ItemInfo) ToJSON() string {
	i.normalize()
	return fmt.Sprintf("%q,%q,%q,%q,%q,%q,%q,%q",
		escapePathname(i.Pathname),
		i.Album,
		i.Artist,
		i.Name,
		i.Disc,
		i.Track,
		i.Year,
		i.Genre)
}

func (i *ItemInfo) ToTSV() string {
	i.normalize()
	year, month, day := i.ModTime.Date()
	modTime := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s",
		replaceTSVMetacharacters(escapePathname(i.Pathname)),
		replaceTSVMetacharacters(i.Album),
		replaceTSVMetacharacters(i.Artist),
		replaceTSVMetacharacters(i.Name),
		replaceTSVMetacharacters(i.Disc),
		replaceTSVMetacharacters(i.Track),
		replaceTSVMetacharacters(i.Year),
		replaceTSVMetacharacters(i.Genre),
		modTime)
}

func assertValidRootPathname(root string) {
	if "" == root {
		log.Fatal("Cannot continue without a valid music-directory.")
	}
}

func buildCatalog(root string) {
	assertValidRootPathname(root)
	log.Printf("Building catalog in %q.\n", root)

	if os.PathSeparator == root[len(root)-1] {
		root = root[:len(root)-1]
	}
	pathname := path.Join(root, "catalog.tsv")
	output, e := os.Create(pathname)
	if e != nil {
		log.Fatalf("Could not create %q: %s\n", pathname, e)
	}
	defer output.Close()

	count := 0
	e = filepath.Walk(root,
		func(pathname string, info os.FileInfo, e error) error {
			if e != nil {
				log.Printf("%q: %s\n", pathname, e)
				return e
			}
			if shouldSkipFile(pathname, info) {
				log.Printf("Skipping %q\n", pathname)
				return nil
			}
			if !info.Mode().IsRegular() {
				return nil
			}

			input, e := os.Open(pathname)
			count++
			if e != nil {
				log.Printf("\n%q: %s\n", pathname, e)
				return nil
			}
			defer input.Close()

			if 0 == count%1000 {
				log.Printf("%v items", count)
			}

			webPathname := pathname[len(root)+1:]
			if isAudioPathname(pathname) || isVideoPathname(pathname) {
				itemInfo := ItemInfo{Pathname: webPathname}
				if isAudioPathname(pathname) {
					itemInfo.File, e = id3.Read(input)
					if e != nil {
						log.Printf("%q: %v", pathname, e)
					}
				}
				fileInfo, e := os.Stat(pathname)
				if e == nil {
					itemInfo.ModTime = fileInfo.ModTime()
				}
				fmt.Fprintf(output, "%s\n", itemInfo.ToTSV())
			}

			return nil
		})

	if e != nil {
		log.Printf("Problem walking %q: %s\n", root, e)
	}

	log.Println("Finished building catalog.")
}

func printDuplicates(root string) error {
	assertValidRootPathname(root)
	for _, pathnames := range fileSizesToPathnames(root) {
		if len(pathnames) < 2 {
			continue
		}

		hashes := make(map[string][]string)
		for _, pathname := range pathnames {
			hash, e := computeMD5(pathname)
			if e != nil {
				log.Printf("%q: %v\n", pathname, e)
				continue
			}
			hashes[hash] = append(hashes[hash], pathname)
		}
		for _, pathnames := range hashes {
			if len(pathnames) < 2 {
				continue
			}
			printStringArray(pathnames)
			fmt.Println()
		}
	}

	return nil
}

func printEmpties(root string) error {
	assertValidRootPathname(root)
	e := filepath.Walk(root, func(pathname string, info os.FileInfo, e error) error {
		if e != nil {
			log.Printf("%q: %v\n", pathname, e)
			return nil
		}
		if info.Mode().IsRegular() && info.Size() == 0 {
			fmt.Printf("%q\n", pathname)
		}
		if info.Mode().IsDir() {
			infos, e := ioutil.ReadDir(pathname)
			if e != nil {
				log.Printf("%q: %v\n", pathname, e)
			}
			if len(infos) == 0 {
				fmt.Printf("%q\n", pathname)
			}
		}
		return nil
	})
	return e
}

func installFrontEndFiles(root string) {
	assertValidRootPathname(root)
	for _, f := range frontEndFiles {
		copyFile(f, path.Join(root, string(os.PathSeparator), f))
	}
	log.Printf("Installed web front-end files in %q.\n", root)
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
		log.Fatalf("Failed to open %q for writing: %s", certificatePathname, err)
	}

	keyFile, err := os.OpenFile(keyPathname, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to open %q for writing: %s", keyPathname, err)
	}

	generateCertificate(hosts, false, keyFile, certificateFile)
	certificateFile.Close()
	keyFile.Close()
	log.Printf("Generated key for X.509 certificate. %q", keyPathname)
	log.Printf("Generated X.509 certificate. %q", certificatePathname)
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

	log.Fatal("No HOME environment variable is set.")
	return ""
}

func establishConfiguration() {
	if homePathname == "" {
		log.Fatal("No HOME environment variable is set.")
	}

	if e := os.MkdirAll(configurationPathname, 0755); e != nil {
		log.Fatalf("Could not create %q: %v", configurationPathname, e)
	}
}

func serveApp(root string) {
	assertValidRootPathname(root)
	addresses, e := net.InterfaceAddrs()
	if e != nil || 0 == len(addresses) {
		log.Println("Hmm, I can't find any network interfaces to run the web server on. I have to give up.")
		os.Exit(1)
	}

	message := "Starting the web server. Point your browser to any of these addresses:"
	if 1 == len(addresses) {
		message = "Starting the web server. Point your browser to this address:"
	}
	log.Println(message)

	var hosts []string
	for _, address := range addresses {
		switch a := address.(type) {
		case *net.IPNet:
			if !bindToIPv4 && a.IP.To4() != nil || !bindToIPv6 && a.IP.To4() == nil {
				continue
			}
			names, e := net.LookupAddr(a.IP.String())
			if e != nil || 0 == len(names) {
				log.Printf("    https://%s%s/\n", a.IP, httpPort)
				hosts = append(hosts, fmt.Sprintf("%s", a.IP))
			} else {
				for _, name := range names {
					log.Printf("    https://%s%s/\n", name, httpPort)
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
  bean-machine [-m music-directory] command [command...]

Invoking bean-machine with no command is equivalent to invoking it with the
"run" command (see below). The commands are:

  catalog
    Scans music-directory for music files, and creates a database of their
    metadata in music-directory/catalog.tsv.

  duplicate
    Scans music-directory for duplicate files, and prints out a list of any.

  empty
    Scans music-directory for empty files and directories, and prints out a
    list of any found.

  install
    Installs the web front-end files in music-directory.

  run
    Equivalent to "bean-machine music-directory catalog install serve".

  serve
    Starts a web server rooted at music-directory, and prints out the URL(s)
    of the Bean Machine web app.

There are 2 additional commands for managing the password authentication for the
web app:

  set-password
    Prompts for a username and password, and sets the password for the given
    username.

  check-password
    Prompts for a username and password, and checks the password for the given
    username. Exits with status 0 if the username is in the database and the
    password is correct; 1 otherwise.`)
	os.Exit(1)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	establishConfiguration()

	if os.Getenv("IPV6") != "" {
		bindToIPv6 = true
	}

	needsHelp1 := flag.Bool("help", false, "Print the help message.")
	needsHelp2 := flag.Bool("h", false, "Print the help message.")
	root := flag.String("m", "", "Set the music directory.")
	flag.Parse()
	musicRoot = *root

	if *needsHelp1 || *needsHelp2 {
		printHelp()
		os.Exit(1)
	}

	if flag.NArg() == 0 {
		if "" != musicRoot {
			buildCatalog(musicRoot)
			installFrontEndFiles(musicRoot)
			serveApp(musicRoot)
		} else {
			printHelp()
			os.Exit(1)
		}
	}

	status := 0
	for i := 0; i < flag.NArg(); i++ {
		command := flag.Arg(i)
		switch command {
		case "catalog":
			buildCatalog(musicRoot)
		case "check-password":
			username, password := promptForCredentials()
			stored := readPasswordDatabase(path.Join(configurationPathname, passwordsBasename))
			ok := checkPassword(stored, username, password)
			log.Printf("Password check for %q: %v\n", username, ok)
			if !ok {
				status = 1
			}
		case "duplicate":
			printDuplicates(musicRoot)
		case "empty":
			printEmpties(musicRoot)
		case "help":
			printHelp()
		case "install":
			installFrontEndFiles(musicRoot)
		case "run":
			buildCatalog(musicRoot)
			installFrontEndFiles(musicRoot)
			serveApp(musicRoot)
		case "serve":
			serveApp(musicRoot)
		case "set-password":
			setPassword()
		default:
			printHelp()
			status = 1
		}
	}
	os.Exit(status)
}
