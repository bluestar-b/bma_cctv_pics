package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	inputFile  = "found_urls.txt"
	outputDir  = "images"
	numWorkers = 32
	userAgent  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
)

func main() {
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		fmt.Printf("Failed to create directory %s: %v\n", outputDir, err)
		return
	}

	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Printf("Error opening %s: %v\n", inputFile, err)
		return
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if url != "" {
			urls = append(urls, url)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading %s: %v\n", inputFile, err)
		return
	}

	jobs := make(chan string, len(urls))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(jobs, &wg)
	}

	for _, url := range urls {
		jobs <- url
	}
	close(jobs)

	wg.Wait()
	fmt.Println("All downloads complete.")
}

func worker(jobs <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for url := range jobs {
		err := downloadImage(url)
		if err != nil {
			fmt.Printf("Failed: %s (%v)\n", url, err)
		} else {
			fmt.Printf("Downloaded: %s\n", url)
		}
	}
}

func downloadImage(url string) error {
	const maxRetries = 3
	var lastErr error

	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	filePath := filepath.Join(outputDir, fileName)

	if _, err := os.Stat(filePath); err == nil {
		return nil
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", userAgent)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(attempt))
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			resp.Body.Close()
			time.Sleep(time.Second * time.Duration(attempt))
			continue
		}

		outFile, err := os.Create(filePath)
		if err != nil {
			resp.Body.Close()
			return err
		}

		_, err = io.Copy(outFile, resp.Body)
		resp.Body.Close()
		outFile.Close()

		if err != nil {
			lastErr = err
			time.Sleep(time.Second * time.Duration(attempt))
			continue
		}

		return nil
	}

	return fmt.Errorf("failed after %d attempts: %v", maxRetries, lastErr)
}
