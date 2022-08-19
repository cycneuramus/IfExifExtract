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

type Image struct {
	file         string
	exifKey      string
	exifValWant  string
	exifValFound string
}

func newImg(file, exifKey, exifValWant, exifValFound string) Image {
	return Image{
		file:         file,
		exifKey:      exifKey,
		exifValWant:  exifValWant,
		exifValFound: exifValFound,
	}
}

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

func exifGetVal(img Image, dstDir string, et *exiftool.Exiftool, imgChan chan<- Image, wg *sync.WaitGroup) {
	defer wg.Done()

	if exists(pathJoin(dstDir, pathBase(img.file))) {
		log.Printf("Skipping EXIF lookup (file already in dst): %v", pathBase(img.file))
		imgChan <- img
		return
	}

	f := et.ExtractMetadata(img.file)
	val, err := f[0].GetString(img.exifKey)
	if err != nil {
		log.Printf("exifGetVal: %v: %v", pathBase(img.file), err)
	}
	img.exifValFound = val

	imgChan <- img
}

func extractOnMatch(exifValFound, dstDir string, imgChan <-chan Image, wg *sync.WaitGroup) {
	defer wg.Done()

	img := <-imgChan
	if img.exifValFound != "" && contains(img.exifValFound, exifValFound) {
		dst := pathJoin(dstDir, pathBase(img.file))
		go copyFile(img.file, dst)
	}
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

	start := time.Now()
	log.Printf("Scanning %v...", srcDir)

	et, err := exiftool.NewExiftool()
	check(err)
	defer et.Close()

	for _, file := range find(srcDir, fileExts) {
		wg.Add(2)
		img := newImg(file, exifKey, exifValWant, "")
		go exifGetVal(img, dstDir, et, imgChan, &wg)
		go extractOnMatch(exifValWant, dstDir, imgChan, &wg)
		wg.Wait()
	}

	log.Printf("Scan complete in %v seconds", time.Since(start))
}
