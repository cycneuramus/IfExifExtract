package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

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

func copyFile(src string, dst string) {
	if exists(dst) {
		log.Printf("Skipping copy (file already in dst): %v", filepath.Base(src))
		return
	}

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

func exifGetVal(file string, exifKey string) (string, error) {
	et, err := exiftool.NewExiftool()
	check(err)
	defer et.Close()

	f := et.ExtractMetadata(file)
	val, err := f[0].GetString(exifKey)
	if err != nil {
		return "", err
	}

	return val, nil
}

func exifIsMatch(file string, exifKey string, exifVal string) bool {
	if exists(filepath.Join(dstDir, filepath.Base(file))) {
		log.Printf("Skipping exif lookup (file already in dst): %v", filepath.Base(file))
		return false
	}

	val, err := exifGetVal(file, exifKey)
	if err != nil {
		log.Printf("%v: %v", filepath.Base(file), err)
	}

	return val == exifVal
}

func main() {
	for _, ext := range fileExts {
		for _, file := range find(srcDir, ext) {
			if exifIsMatch(file, exifKey, exifVal) {
				dst := filepath.Join(dstDir, filepath.Base(file))
				copyFile(file, dst)
			}
		}
	}
}
