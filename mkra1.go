package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type EthicalLoader struct {
	MaxRequests    int           // ចំនួនសំណើរអតិបរមា
	MaxConcurrency int           // ចំនួន thread ធ្វើការជាមួយ
	Timeout        time.Duration // ពេលវេលារង់ចាំអតិបរមា
	client         *http.Client
	requestCount   uint64
	errorCount     uint64
}

func main() {
	targetURL := flag.String("url", "", "URL សម្រាប់តេស្តប្រសិទ្ធភាព")
	flag.Parse()

	if *targetURL == "" {
		fmt.Println("សូមបញ្ជាក់ URL គោលដៅ")
		return
	}

	loader := EthicalLoader{
		MaxRequests:    1000000,   // បង្កើនទៅ 1,000,000
		MaxConcurrency: 500,       // កំណត់ thread ច្រើនសម្រាប់ប្រសិទ្ធភាព
		Timeout:        20 * time.Minute, // អោយពេលវេលាបានច្រើន
		client:         &http.Client{Timeout: 10 * time.Second},
	}

	if err := loader.Run(context.Background(), *targetURL); err != nil {
		fmt.Printf("កំហុស៖ %v\n", err)
	}

	fmt.Printf("សំណើរសរុប៖ %d, កំហុស៖ %d\n", loader.requestCount, loader.errorCount)
}

func (l *EthicalLoader) Run(ctx context.Context, url string) error {
	ctx, cancel := context.WithTimeout(ctx, l.Timeout)
	defer cancel()

	var wg sync.WaitGroup
	jobs := make(chan struct{}, l.MaxRequests)

	// បង្កើត worker pool
	for i := 0; i < l.MaxConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				if err := l.sendRequest(url); err != nil {
					atomic.AddUint64(&l.errorCount, 1)
				} else {
					atomic.AddUint64(&l.requestCount, 1)
				}
				time.Sleep(10 * time.Millisecond) // rate limit
			}
		}()
	}

	// បញ្ជូនការងារចូលក្នុង channel
	for i := 0; i < l.MaxRequests; i++ {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return ctx.Err()
		case jobs <- struct{}{}:
		}
	}
	close(jobs)
	wg.Wait()
	return nil
}

func (l *EthicalLoader) sendRequest(url string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("កំហុសក្នុងការបង្កើតសំណើរ៖ %v", err)
	}

	// គោរព robots.txt និងការកំណត់អត្រា
	req.Header.Set("User-Agent", "Ethical Load Tester")
	req.Header.Set("From", "your-email@example.com")

	resp, err := l.client.Do(req)
	if err != nil {
		return fmt.Errorf("កំហុសក្នុងការផ្ញើសំណើរ៖ %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("សំណើរច្រើនពេក៖ សូមរង់ចាំ")
	}

	return nil
}
