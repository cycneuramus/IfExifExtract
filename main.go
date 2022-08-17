package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

func pathBase(p string) string {
	return filepath.Base(p)
}

func pathJoin(p1, p2 string) string {
	return filepath.Join(p1, p2)
}

func pathExt(p string) string {
	return filepath.Ext(p)
}

func contains(str, substr string) bool {
	return strings.Contains(str, substr)
}

func exists(file string) bool {
	_, err := os.Open(file)
	return err == nil
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func copyFile(src, dst string) {
	if exists(dst) {
		log.Printf("Skipping copy (file already in dst): %v", pathBase(src))
		return
	}

	log.Printf("Copying: %v", pathBase(src))

	r, err := os.Open(src)
	check(err)
	defer r.Close()

	w, err := os.Create(dst)
	check(err)
	defer w.Close()

	w.ReadFrom(r)
}

func find(rootDir string, fileExts []string) []string {
	var files []string

	walker := func(xpath string, xinfo fs.DirEntry, err error) error {
		check(err)
		for _, ext := range fileExts {
			if pathExt(xinfo.Name()) == ext {
				files = append(files, xpath)
			}
		}
		return nil
	}

	filepath.WalkDir(rootDir, walker)
	return files
}

func exifGetVal(file, exifKey string, et *exiftool.Exiftool) string {
	if exists(pathJoin(dstDir, pathBase(file))) {
		log.Printf("Skipping EXIF lookup (file already in dst): %v", pathBase(file))
		return ""
	}

	f := et.ExtractMetadata(file)

	val, err := f[0].GetString(exifKey)
	if err != nil {
		log.Printf("exifGetVal: %v: %v", pathBase(file), err)
	}

	return val
}

func extract(file, dstDir string, et *exiftool.Exiftool, wg *sync.WaitGroup) {
	val := exifGetVal(file, exifKey, et)

	if contains(val, exifVal) {
		dst := pathJoin(dstDir, pathBase(file))
		copyFile(file, dst)
	}

	wg.Done()
}

func main() {
	start := time.Now()
	log.Printf("Scanning %v...", srcDir)

	et, err := exiftool.NewExiftool()
	check(err)
	defer et.Close()

	var wg sync.WaitGroup

	for _, file := range find(srcDir, fileExts) {
		wg.Add(1)
		go extract(file, dstDir, et, &wg)
	}
	wg.Wait()

	log.Printf("Scan complete in %v seconds", time.Since(start))
}
