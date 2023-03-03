package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/otiai10/gosseract/v2"
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
	cmd := exec.Command("pdfcpu", "extract", "-m", "i", inPath, ".")
	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	pagesToKeep := []int{}
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
		log.Println(imagePath)
		client.SetImage(imagePath)
		text, err := client.Text()
		if err != nil {
			panic(err)
		}
		if strings.Contains(text, search) {
			pagesToKeep = append(pagesToKeep, i)
		} else {
			// TODO we can stop with the document here
		}
	}

	log.Println("finished ocr")

	// Use pdfcpu to copy the selected pages to a new PDF file
	pagesToCopy := []string{}
	for _, page := range pagesToKeep {
		pagesToCopy = append(pagesToCopy, fmt.Sprintf("%v_%0*d_Im0.jpg", fileName, numZeros, page))
	}
	// pagesArg := strings.Join(pagesToCopy, " ")
	cmd = exec.Command("pdfcpu", "import", outPath)
	cmd.Args = append(cmd.Args, pagesToCopy...)
	log.Printf("command to run %v", cmd.String())
	err = cmd.Run()
	if err != nil {
		log.Printf("%v", err)
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
	// Get page count, probably easier using the API instead of the CLI
	cmd := exec.Command("pdfcpu", "info", pdfPath)
	grepCmd := exec.Command("grep", "Page count:")
	awkCmd := exec.Command("awk", "{print $3}")
	grepCmd.Stdin, _ = cmd.StdoutPipe()
	awkCmd.Stdin, _ = grepCmd.StdoutPipe()

	var output bytes.Buffer
	awkCmd.Stdout = &output

	_ = awkCmd.Start()
	_ = grepCmd.Start()
	_ = cmd.Run()
	_ = grepCmd.Wait()
	_ = awkCmd.Wait()

	countString := strings.TrimSpace(output.String())

	pageCount, err := strconv.Atoi(countString)
	if err != nil {
		panic(err)
	}
	return pageCount
}
