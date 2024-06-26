package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	start := time.Now()

	src := flag.String("src", "", "Source file path")
	dst := flag.String("dst", "", "Destination file path")
	flag.Parse()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Error: Source file path and destination directory path must be specified. Use --src and --dst flags to specify them.\n")
		flag.PrintDefaults()
		os.Exit(2)
		log.Fatal("Source file path and destination directory path must be specified. Use --src and --dst flags to specify them.")
	}
	if *src == "" || *dst == "" {
		flag.Usage()
		os.Exit(1)
	}

	srcFile, err := os.Open(*src)
	if err != nil {
		log.Fatalf("Error opening file: %v\n", err)
	}
	defer srcFile.Close()

	scanner := bufio.NewScanner(srcFile)

	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url == "" {
			continue
		}

		fileName, err := getFileNameFromURL(url)

		if err != nil {
			log.Println(err)
			continue
		}
		if fileName == "" {
			continue
		}

		respBody, err := fetchUrl(url)
		if err != nil {
			log.Printf("Ошибка: %v\n", err)
			continue
		}
		if respBody == nil {
			continue
		}

		err = saveDst(fileName, *dst, respBody)
		if err != nil {
			log.Fatalf("err: %v\n", err)
		}

	}

	elapsed := time.Since(start)
	fmt.Printf("\nProgram execution time: %s\n", elapsed)
}

// fetchUrl() получает url и пытается выполнить get запрос по этому url
// и вернуть Body если получается вернуть resp с get запроса
func fetchUrl(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error fetching URL: %v", err))

	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("Unexpected status code for URL %s: %d\n", url, resp.StatusCode))
	}

	buf := make([]byte, 1024)
	body := bytes.NewBuffer(buf)
	_, err = io.Copy(body, resp.Body)

	return body.Bytes(), nil
}

// saveDst() получает fileName - имя файла в который необходимо записать respBody - тело запроса плученного с помощью get
// и сохранить respBody по пути dst если путь ./, то создается ./list в которую записывается файл
func saveDst(fileName, dst string, respBody []byte) error {

	fileName = fmt.Sprintf("%s.html", fileName)

	if dst == "./" || dst == "." {
		dst = "./list"
		err := os.MkdirAll(dst, 0755)
		if err != nil {
			return errors.New(fmt.Sprintf("Error creating folder %s: %v\n", dst, err))
		}
	}
	filePath := filepath.Join(dst, fileName)

	dstFile, err := os.Create(filePath)
	if err != nil {
		return errors.New(fmt.Sprintf("Error creating destination file %s: %v\n", dst, err))
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, bytes.NewBuffer(respBody))

	if err != nil {
		return errors.New(fmt.Sprintf("Error copying content: %v\n", err))
	}

	fmt.Printf("File copied successfully to %s\n\n", filePath)
	return nil
}

// getFileNameFromURL() функция получает url сайт
// и возвращает имя файла на основе доменного имени url
func getFileNameFromURL(siteURL string) (string, error) {
	parsedURL, err := url.Parse(siteURL)

	if err != nil {

		return "", err
	}
	domain := strings.TrimPrefix(parsedURL.Host, "www.")
	if domain == "" {
		return "", errors.New(fmt.Sprintf("Error: no such site with name: %s\n", siteURL))
	}
	if strings.Contains(domain, ":") {
		domain = strings.Split(domain, ":")[0]
	}

	parts := strings.SplitN(domain, ".", 2)
	if len(parts) < 2 {
		return domain, nil
	}

	return parts[0], nil
}
