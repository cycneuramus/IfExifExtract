package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

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

func find(rootDir string, fileExt []string) []string {
	var files []string

	walker := func(xpath string, xinfo fs.DirEntry, err error) error {
		check(err)
		for _, ext := range fileExt {
			if filepath.Ext(xinfo.Name()) == ext {
				files = append(files, xpath)
			}
		}
		return nil
	}

	filepath.WalkDir(rootDir, walker)
	return files
}

func exifGetVal(file, exifKey string, et *exiftool.Exiftool) string {
	if exists(filepath.Join(dstDir, filepath.Base(file))) {
		log.Printf("Skipping EXIF lookup (file already in dst): %v", filepath.Base(file))
		return ""
	}

	log.Printf("EXIF lookup: %v", filepath.Base(file))
	f := et.ExtractMetadata(file)
	val, _ := f[0].GetString(exifKey)

	return val
}

func main() {
	start := time.Now()
	log.Printf("Scanning %v...", srcDir)

	et, err := exiftool.NewExiftool()
	check(err)
	defer et.Close()

	for _, file := range find(srcDir, fileExts) {
		val := exifGetVal(file, exifKey, et)
		if contains(val, exifVal) {
			dst := filepath.Join(dstDir, filepath.Base(file))
			copyFile(file, dst)
		}
	}

	log.Printf("Scan complete in %v seconds", time.Since(start))
}
