package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/otiai10/gosseract/v2"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	pdf "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

func main() {
	dir := "/root/development/leanfactory/extract/pdfs"
	search := "FORMULARIO DEL REGISTRO"

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".pdf" {
			fmt.Println("Processing:", path)
			outPath := strings.TrimSuffix(path, ".pdf") + "_formulario.pdf"
			extractPages(path, outPath, search)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func extractPages(inPath string, outPath string, search string) {
	// Extract pages as images

	err := api.ExtractImagesFile(inPath, ".", nil, nil)
	if err != nil {
		panic(err)
	}

	pagesToKeep := []int{}
	found := false
	client := gosseract.NewClient()
	client.SetLanguage("spa")
	defer client.Close()
	log.Println("starting ocr")
	fileName := strings.Split(filepath.Base(inPath), ".pdf")[0]
	numPages := pageCount(inPath)
	var numZeros int
	if numPages < 100 {
		numZeros = 2
	} else {
		numZeros = 3
	}
	for i := 1; i <= numPages; i++ {
		imagePath := fmt.Sprintf("%v_%0*d_Im0.jpg", fileName, numZeros, i)
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

	if len(pagesToKeep) == 0 {
		log.Println("No pages matched")
		return
	}

	// Use pdfcpu to copy the selected pages to a new PDF file
	pagesToCopy := []string{}
	for _, page := range pagesToKeep {
		pagesToCopy = append(pagesToCopy, fmt.Sprintf("%v_%0*d_Im0.jpg", fileName, numZeros, page))
	}
	imp := pdf.DefaultImportConfig()
	err = api.ImportImagesFile(pagesToCopy, outPath, imp, nil)
	if err != nil {
		panic(err)
	}

	// Delete the temporary files
	// cmd = exec.Command("rm", fmt.Sprintf("%v_*.png", fileName))
	// err = cmd.Run()
	// if err != nil {
	// 	panic(err)
	// }
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
