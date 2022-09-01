package main

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/barasher/go-exiftool"
)

type Image struct {
	file         string
	exifKey      string
	exifValWant  string
	exifValFound string
}

func newImage(file, exifKey, exifValWant string) Image {
	return Image{
		file:         file,
		exifKey:      exifKey,
		exifValWant:  exifValWant,
		exifValFound: "",
	}
}

func joinPath(p1, p2 string) string {
	return filepath.Join(p1, p2)
}

func getPathBase(p string) string {
	return filepath.Base(p)
}

func getFileExt(p string) string {
	return filepath.Ext(p)
}

func contains(str, substr string) bool {
	return strings.Contains(str, substr)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

func validateExists(paths ...string) {
	for _, p := range paths {
		if !exists(p) {
			log.Fatalf("Not found: %v", p)
		}
	}
}

func validateIsSet(vars ...string) {
	for _, v := range vars {
		if v == "" {
			log.Fatalf("Not set: %v", v)
		}
	}
}

func find(rootDir string, fileExts []string) []string {
	var files []string

	walker := func(xpath string, xinfo fs.DirEntry, err error) error {
		check(err)
		for _, ext := range fileExts {
			if getFileExt(xinfo.Name()) == ext {
				files = append(files, xpath)
			}
		}
		return nil
	}

	filepath.WalkDir(rootDir, walker)
	return files
}

func exifGetVal(img Image, dstDir string, et *exiftool.Exiftool, imgChan chan<- Image, wg *sync.WaitGroup) {
	defer wg.Done()

	if exists(joinPath(dstDir, getPathBase(img.file))) {
		log.Printf("exifGetVal: Skipping EXIF lookup (file already in dst): %v", getPathBase(img.file))
		imgChan <- img
		return
	}

	exif := et.ExtractMetadata(img.file)
	val, err := exif[0].GetString(img.exifKey)
	if err != nil {
		log.Printf("exifGetVal: %v: %v", getPathBase(img.file), err)
	}

	img.exifValFound = val
	imgChan <- img
}

func extractMatch(dstDir string, imgChan <-chan Image, wg *sync.WaitGroup) {
	defer wg.Done()

	img := <-imgChan
	if !contains(img.exifValFound, img.exifValWant) {
		return
	}

	dst := joinPath(dstDir, getPathBase(img.file))
	if exists(dst) {
		log.Printf("extractMatch: Skipping (already in dst): %v", getPathBase(img.file))
		return
	}

	log.Printf("extractMatch: Extracting %v", getPathBase(img.file))

	r, err := os.Open(img.file)
	check(err)
	defer r.Close()

	w, err := os.Create(dst)
	check(err)
	defer w.Close()

	w.ReadFrom(r)
}

func main() {
	var (
		srcDir      = os.Getenv("SRC_DIR")
		dstDir      = os.Getenv("DST_DIR")
		exifKey     = os.Getenv("EXIF_KEY")
		exifValWant = os.Getenv("EXIF_VAL")
		fileExts    = []string{".jpg", ".jpeg"}

		imgChan = make(chan Image)
		wg      sync.WaitGroup
	)

	validateIsSet(srcDir, dstDir, exifKey, exifValWant)
	validateExists(srcDir, dstDir)

	et, err := exiftool.NewExiftool()
	check(err)
	defer et.Close()

	start := time.Now()
	log.Printf("Scanning %v...", srcDir)

	for _, file := range find(srcDir, fileExts) {
		img := newImage(file, exifKey, exifValWant)

		wg.Add(2)
		go exifGetVal(img, dstDir, et, imgChan, &wg)
		go extractMatch(dstDir, imgChan, &wg)
	}

	wg.Wait()
	log.Printf("Scan completed in %v", time.Since(start))
}
