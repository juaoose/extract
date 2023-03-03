package main

import (
	"errors"
	"fmt"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/otiai10/gosseract/v2"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	pdf "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func main() {
	dir := "/root/development/leanfactory/extract/pdfs"
	search := "FORMULARIO DEL REGISTRO"

	var wg sync.WaitGroup // create a wait group

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".pdf" {
			fmt.Println("Processing:", path)
			fileName := strings.TrimSuffix(path, ".pdf")
			outPath := fileName + "_formulario.pdf"

			// We start goroutines per extraction job and make sure to wait
			wg.Add(1)
			go func() {
				extractPages(path, outPath, search)
				wg.Done()
			}()
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	// Wait for all the goroutines to end
	wg.Wait()
}

func extractPages(inPath string, outPath string, search string) {
	// Extract pages as images
	file, err := os.Open(inPath)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	fileName := strings.TrimSuffix(inPath, ".pdf")

	// We extract and convert every image, this because every document
	// could have different image types
	err = api.ExtractImages(file, nil, digestImage(fileName), nil)
	if err != nil {
		panic(err)
	}

	pagesToKeep := []int{}
	client := gosseract.NewClient()
	client.SetLanguage("spa")
	defer client.Close()
	log.Println("starting ocr")

	// Just so that we can end early
	found := false

	numPages := pageCount(inPath)
	for i := 1; i <= numPages; i++ {
		imagePath := fmt.Sprintf("%v_%d_Im0.jpg", fileName, i)
		client.SetImage(imagePath)
		text, err := client.Text()
		if err != nil {
			panic(err)
		}

		if strings.Contains(text, search) {
			pagesToKeep = append(pagesToKeep, i)
			found = true
		} else if found {
			log.Printf("Processed %d lines to find document", i)
			break
		}

	}

	log.Println("finished ocr")

	// Dont generate am empty pdf
	if len(pagesToKeep) == 0 {
		log.Println("No pages matched")
		return
	}

	// Use pdfcpu to copy the selected pages to a new PDF file
	pagesToCopy := []string{}
	for _, page := range pagesToKeep {
		pagesToCopy = append(pagesToCopy, fmt.Sprintf("%v_%d_Im0.jpg", fileName, page))
	}
	imp := pdf.DefaultImportConfig()
	err = api.ImportImagesFile(pagesToCopy, outPath, imp, nil)
	if err != nil {
		panic(err)
	}

	// TODO Clean after
	// Delete the temporary files
	// cmd = exec.Command("rm", fmt.Sprintf("%v_*.png", fileName))
	// err = cmd.Run()
	// if err != nil {
	// 	panic(err)
	// }
}

func digestImage(docName string) func(model.Image, bool, int) error {
	return func(img model.Image, singleImgPerPage bool, maxPageDigits int) error {
		// docname_pageNr_Im0
		f, err := os.Create(fmt.Sprintf("%v_%v_%v.jpg", docName, img.PageNr, img.Name))
		if err != nil {
			panic(err)
		}
		defer f.Close()

		imageOut, err := imaging.Decode(img)
		if err != nil {
			fmt.Println(err)
			return errors.New("imaging.Decode() Error")
		}
		err = jpeg.Encode(f, imageOut, &jpeg.Options{Quality: 100})
		if err != nil {
			return errors.New("digestImage jpeg.Encode( Error")
		}
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
	log.Printf("This document has %d pages", pageCount)
	return pageCount
}
