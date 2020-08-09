package main

import (
  "crypto/sha256"
  "flag"
  "fmt"
  "io"
  "log"
  "os"
  "path"
  "path/filepath"
  "strings"
)

func getFileHash(pathname string) string {
	f, e := os.Open(pathname)
	if e != nil {
		return ""
	}
	defer f.Close()

	h := sha256.New()
	if _, e = io.Copy(h, f); e != nil {
    return ""
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func main() {
  flag.Parse()
  filepath.Walk(flag.Arg(0),
    func(pathname string, info os.FileInfo, e error) error {
      basename := strings.ToLower(path.Base(pathname))
      if basename == "other.jpg" {
        coverPathname := path.Join(path.Dir(pathname), "cover.jpg")
        coverHash := getFileHash(coverPathname)

        if coverHash == "" {
          fmt.Printf("move %q to %q\n", pathname, coverPathname)
          e := os.Rename(pathname, coverPathname)
          if e != nil {
            log.Fatal(e)
          }
          return nil
        }

        otherHash := getFileHash(pathname)
        if coverHash == otherHash {
          fmt.Printf("remove %q\n", pathname)
          e := os.Remove(pathname)
          if e != nil {
            log.Fatal(e)
          }
        }
        return nil
      }
      return nil
    })
}
