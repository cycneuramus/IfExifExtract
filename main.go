package main

import (
	"errors"
	"flag"
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
	file        string
	exifKey     string
	exifValue   string
	exifQueries []string
}

func newImage(file, exifKey string, exifQueries []string) Image {
	return Image{
		file:        file,
		exifKey:     exifKey,
		exifValue:   "",
		exifQueries: exifQueries,
	}
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

func isSet(vars ...string) bool {
	for _, v := range vars {
		if v == "" {
			return false
		}
	}
	return true
}

func validateExists(paths ...string) {
	for _, p := range paths {
		if !exists(p) {
			log.Fatalf("Not found: %v", p)
		}
	}
}

func parse(queryFlag string) []string {
	exifQueries := strings.Split(queryFlag, ",")
	var res []string
	for _, str := range exifQueries {
		res = append(res, strings.TrimSpace(str))
	}

	return res
}

func find(rootDir string, fileExts []string) []string {
	var files []string

	walker := func(xpath string, xinfo fs.DirEntry, err error) error {
		check(err)
		for _, ext := range fileExts {
			if filepath.Ext(xinfo.Name()) == ext {
				files = append(files, xpath)
			}
		}
		return nil
	}

	filepath.WalkDir(rootDir, walker)
	return files
}

func isMatch(value string, queries []string) bool {
	match := 0
	for _, q := range queries {
		if strings.Contains(value, q) {
			match++
		}
	}

	return match > 0
}

func copyFile(src, dst string) {
	r, err := os.Open(src)
	check(err)
	defer r.Close()

	w, err := os.Create(dst)
	check(err)
	defer w.Close()

	w.ReadFrom(r)
}

func exifGetVal(img Image, dstDir string, et *exiftool.Exiftool, imgChan chan<- Image, wg *sync.WaitGroup) {
	defer wg.Done()

	if exists(filepath.Join(dstDir, filepath.Base(img.file))) {
		log.Printf("exifGetVal: Skipping EXIF lookup (file already in dst): %v", filepath.Base(img.file))
		imgChan <- img
		return
	}

	exif := et.ExtractMetadata(img.file)
	val, _ := exif[0].GetString(img.exifKey)

	img.exifValue = val
	imgChan <- img
}

func extractMatch(dstDir string, imgChan <-chan Image, wg *sync.WaitGroup) {
	defer wg.Done()
	img := <-imgChan

	if !isMatch(img.exifValue, img.exifQueries) {
		return
	}

	dst := filepath.Join(dstDir, filepath.Base(img.file))
	if exists(dst) {
		log.Printf("extractMatch: Skipping (already in dst): %v", filepath.Base(img.file))
		return
	}

	log.Printf("extractMatch: Extracting %v", filepath.Base(img.file))
	copyFile(img.file, dst)
}

func main() {
	var (
		srcDir    string
		dstDir    string
		exifKey   string
		exifQuery string
		fileExts  = []string{".jpg", ".jpeg"}
		imgChan   = make(chan Image)
		wg        sync.WaitGroup
	)

	flag.StringVar(&srcDir, "srcDir", "", "Directory to scan")
	flag.StringVar(&dstDir, "dstDir", "", "Directory to receive matching files")
	flag.StringVar(&exifKey, "exifKey", "", "EXIF key to query")
	flag.StringVar(&exifQuery, "exifQuery", "", "EXIF values to find (comma-separated)")
	flag.Parse()

	if !isSet(srcDir, dstDir, exifKey, exifQuery) {
		flag.Usage()
		os.Exit(1)
	}
	validateExists(srcDir, dstDir)
	exifQueries := parse(exifQuery)

	et, err := exiftool.NewExiftool()
	check(err)
	defer et.Close()

	start := time.Now()
	log.Printf("Scanning %v...", srcDir)

	for _, file := range find(srcDir, fileExts) {
		img := newImage(file, exifKey, exifQueries)

		wg.Add(2)
		go exifGetVal(img, dstDir, et, imgChan, &wg)
		go extractMatch(dstDir, imgChan, &wg)
	}

	wg.Wait()
	log.Printf("Scan completed in %v", time.Since(start))
}
