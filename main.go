package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/otiai10/gosseract/v2"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// export TESSDATA_PREFIX=/root/development/leanfactory/extract/tess
func main() {

	search := "FORMULARIO DEL REGISTRO"

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Result directory
		createDirectories()

		if d.IsDir() && d.Name() == "result" {
			return filepath.SkipDir
		}
		client := gosseract.NewClient()
		client.SetLanguage("spa")
		defer client.Close()

		if !d.IsDir() && filepath.Ext(path) == ".pdf" {
			fileName := strings.TrimSuffix(path, ".pdf")
			outPath := "result/" + fileName + "_formulario.pdf"

			extractPages(client, path, outPath, search)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

}

func createDirectories() {
	resultPath := filepath.Join(".", "/result")
	err := os.MkdirAll(resultPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

func extractPages(client *gosseract.Client, inPath string, outPath string, search string) {
	defer timer(inPath)()

	// Extract pages as images
	file, err := os.Open(inPath)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	pagesToKeep := []int{}

	// Just so that we can end early
	found := false

	numPages := pageCount(inPath)
	for i := 1; i <= numPages; i++ {
		// Extract the current image to a buffer and pass it to tesseract
		buffer := &bytes.Buffer{}
		err = api.ExtractImages(file, []string{strconv.Itoa(i)}, toBuffer(buffer), nil)
		if err != nil {
			panic(err)
		}

		client.SetImageFromBytes(buffer.Bytes())
		text, err := client.Text()
		if err != nil {
			// TODO(juaoose) not panicking so that we just move on
			log.Printf("Fatal error performing OCR %v", err)
			return
		}

		if strings.Contains(text, search) {
			pagesToKeep = append(pagesToKeep, i)
			found = true
		} else if found {
			break
		}

	}

	// Dont generate an empty or non compliant PDF
	// The ones we want have at least 2 pages, if not, OCR was not succesful
	if len(pagesToKeep) < 2 {
		return
	}

	pagesToCopy := []string{}
	for _, page := range pagesToKeep {
		pagesToCopy = append(pagesToCopy, strconv.Itoa(page))
	}

	//Use pdfcpu to generate a trimmed file
	err = api.TrimFile(inPath, outPath, pagesToCopy, nil)
	if err != nil {
		panic(err)
	}
}

// TODO(juaoose) can i improve this? maybe use ExtractImagesRaw?
func toBuffer(buff *bytes.Buffer) func(model.Image, bool, int) error {
	return func(img model.Image, singleImgPerPage bool, maxPageDigits int) error {
		buff.ReadFrom(img.Reader)
		return nil
	}
}

func pageCount(pdfPath string) int {

	info, err := api.InfoFile(pdfPath, nil, nil)
	if err != nil {
		panic(err)
	}

	countString := strings.TrimSpace(strings.Split(info[1], ":")[1])

	pageCount, err := strconv.Atoi(countString)
	if err != nil {
		panic(err)
	}
	return pageCount
}

func timer(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %v\n", name, time.Since(start))
	}
}
