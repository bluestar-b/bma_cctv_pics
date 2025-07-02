package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	baseURL    = "http://110.170.214.36/images/clips/images/"
	minVal     = 40
	maxVal     = 170
	userAgent  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	numWorkers = 256
)

var (
	stopRequested = false
	client        = &http.Client{
		Timeout: 1 * time.Second,
	}
)

type result struct {
	url string
	ok  bool
}

func makeURL(x, y, z int) string {
	return fmt.Sprintf("%s10_%d_%d_%d.jpg", baseURL, x, y, z)
}

func headRequest(url string) bool {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func worker(jobs <-chan string, results chan<- result, wg *sync.WaitGroup) {
	defer wg.Done()
	for url := range jobs {
		ok := headRequest(url)
		results <- result{url: url, ok: ok}
		if stopRequested {
			break
		}
	}
}

func main() {
	found := []string{}
	processed := 0
	startTime := time.Now()

	jobs := make(chan string, 1000)
	results := make(chan result, 1000)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived Ctrl+C. Stopping after current batch...")
		stopRequested = true
	}()

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(jobs, results, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	go func() {
		for x := minVal; x <= maxVal && !stopRequested; x++ {
			for y := minVal; y <= maxVal && !stopRequested; y++ {
				for z := minVal; z <= maxVal && !stopRequested; z++ {
					url := makeURL(x, y, z)
					jobs <- url
				}
			}
		}
		close(jobs)
	}()

	for res := range results {
		processed++
		if res.ok {
			fmt.Printf("FOUND: %s\n", res.url)
			found = append(found, res.url)
		}
		if processed%5000 == 0 {
			elapsed := time.Since(startTime).Seconds()
			fmt.Printf("Processed: %d - Found: %d - Rate: %.1f/s\n", processed, len(found), float64(processed)/elapsed)
		}
	}

	outFile, err := os.Create("found_urls.txt")
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer outFile.Close()

	for _, url := range found {
		_, err := outFile.WriteString(url + "\n")
		if err != nil {
			fmt.Printf("Error writing to output file: %v\n", err)
			break
		}
	}

	fmt.Printf("Saved found URLs to found_urls.txt\n")

	elapsed := time.Since(startTime).Seconds()
	fmt.Printf("\nTotal found URLs: %d\n", len(found))
	fmt.Printf("Processed: %d URLs in %.1fs\n", processed, elapsed)
}
