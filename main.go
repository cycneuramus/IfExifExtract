package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/barasher/go-exiftool"
)

var (
	srcDir   = os.Getenv("SRC_DIR")
	dstDir   = os.Getenv("DST_DIR")
	exifKey  = os.Getenv("EXIF_KEY")
	exifVal  = os.Getenv("EXIF_VAL")
	fileExts = []string{".jpg", ".jpeg"}
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func exists(file string) bool {
	_, err := os.Open(file)
	return err == nil
}

func contains(str, substr string) bool {
	return strings.Contains(str, substr)
}

func copyFile(src, dst string) {
	if exists(dst) {
		log.Printf("Skipping copy (file already in dst): %v", filepath.Base(src))
		return
	}

	log.Printf("Copying: %v", filepath.Base(src))

	r, err := os.Open(src)
	check(err)
	defer r.Close()

	w, err := os.Create(dst)
	check(err)
	defer w.Close()

	w.ReadFrom(r)
}

func find(rootDir, fileExt string) []string {
	var files []string

	filepath.WalkDir(rootDir, func(s string, d fs.DirEntry, err error) error {
		check(err)
		if filepath.Ext(d.Name()) == fileExt {
			files = append(files, s)
		}

		return nil
	})

	return files
}

func exifGetVal(file, exifKey string) string {
	et, err := exiftool.NewExiftool()
	check(err)

	f := et.ExtractMetadata(file)
	val, _ := f[0].GetString(exifKey)

	return val
}

func exifIsMatch(file, exifKey, exifVal string) bool {
	if exists(filepath.Join(dstDir, filepath.Base(file))) {
		log.Printf("Skipping exif lookup (file already in dst): %v", filepath.Base(file))
		return false
	}

	val := exifGetVal(file, exifKey)
	return contains(val, exifVal)
}

func main() {
	log.Printf("Scanning %v...\n", srcDir)

	for _, ext := range fileExts {
		for _, file := range find(srcDir, ext) {
			if exifIsMatch(file, exifKey, exifVal) {
				dst := filepath.Join(dstDir, filepath.Base(file))
				copyFile(file, dst)
			}
		}
	}

	log.Printf("Scan complete")
}
