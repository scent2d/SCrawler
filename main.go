// USAGE	: go run main.go [target]
// EXAMPLE	: go run main.go https://youtube.com

package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/steelx/extractlinks"
)

var (
	config = &tls.Config{
		InsecureSkipVerify: true,
	}
	transport = &http.Transport{
		TLSClientConfig: config,
	}
	netClient = &http.Client{
		Transport: transport,
	}
	queue      = make(chan string)
	hasVisited = make(map[string]bool)
)

// 크롤링 함수
func crawlURL(href string) {
	hasVisited[href] = true
	fmt.Printf("Crawling URL --> %v \n", href)

	response, err := netClient.Get(href)
	checkError(err)
	defer response.Body.Close()

	links, err := extractlinks.All(response.Body)
	checkError(err)

	for _, link := range links {
		absoluteURL := toFixedURL(link.Href, href)
		go func() {
			queue <- absoluteURL
		}()
	}
}

// 도메인 비교함수
func isSameDomain(href string, baseURL string) bool {
	url, err := url.Parse(href)
	if err != nil {
		return false
	}
	parentURL, err := url.Parse(baseURL)
	if err != nil {
		return false
	}
	if url.Host != parentURL.Host {
		return false
	}

	return true
}

// URL 통일함수
func toFixedURL(href string, baseURL string) string {
	uri, err := url.Parse(href)
	if err != nil {
		return ""
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return ""
	}

	fixedURL := base.ResolveReference(uri)

	return fixedURL.String()
}

// 예외처리 함수
func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	// 실행 어규먼츠 배열 선언
	arguments := os.Args[1:]
	if len(arguments) == 0 {
		fmt.Println("Arguments Missing !")
		os.Exit(1)
	}

	file, err := os.OpenFile("/Users/scent2d/golang/src/SCrawler/output.txt", os.O_CREATE|os.O_RDWR, os.FileMode(0777))
	checkError(err)
	defer file.Close()

	w := bufio.NewWriter(file)

	baseURL := arguments[0]

	go func() {
		queue <- baseURL
	}()

	for href := range queue {
		if !hasVisited[href] && isSameDomain(href, baseURL) {
			crawlURL(href)
			w.WriteString(href + "\r\n")
		}
		w.Flush()
	}
}
