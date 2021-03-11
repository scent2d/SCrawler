// EXAMPLE	: go run main.go -host=https://www.youtube.com -file=/home/scent2d/output.txt
// EXAMPLE	: go run main.go -host="https://www.youtube.com" -file="./output.txt" -session="JSESSIONID=abcd" -header="Authorazation: abcd"
// EXAMPLE	: go run main.go -host="https://www.youtube.com" -file="./output.txt" -session="JSESSIONID=abcd" -header="Authorazation: abcd" -proxy=http://127.0.0.1:8888
// 프록시 기능 더 예쁘게 구현해야함 - Update 필요

package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/steelx/extractlinks"
)

var (
	// proxy, _ = url.Parse("127.0.0.1:8888")

	// config = &tls.Config{
	// 	InsecureSkipVerify: true,
	// }
	// transport = &http.Transport{
	// 	TLSClientConfig: config,
	// 	Proxy:           http.ProxyURL(proxy),
	// }
	// netClient = &http.Client{
	// 	Transport: transport,
	// }
	queue      = make(chan string)
	hasVisited = make(map[string]bool)
)

// 크롤링 함수
func crawlURL(href string, session string, header string, proxyURL string) {
	// func crawlURL(href string, session string, header string) {
	hasVisited[href] = true
	fmt.Printf("Crawling URL --> %v \n", href)

	// proxyURL 어규먼츠 값에 http:// 로 시작되면 제대로 파싱되고, 그냥 일반 값이 넘어오면 파싱되지 않아 proxy 설정되지 않음
	// proxyURL: http://127.0.0.1:8888 	--> Proxy On
	// proxyURL: 127.0.0.1:8888 		--> Proxy Off
	proxy, _ := url.Parse(proxyURL)

	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	// 프록시 설정 여부 확인 후 변수 대입
	// if proxyURL != "" {
	// 	transport = &http.Transport{
	// 		TLSClientConfig: config,
	// 		Proxy:           http.ProxyURL(proxy),
	// 	}
	// } else {
	// 	transport = &http.Transport{
	// 		TLSClientConfig: config,
	// 	}
	// }
	transport := &http.Transport{
		TLSClientConfig: config,
		Proxy:           http.ProxyURL(proxy),
	}

	netClient := &http.Client{
		Transport: transport,
	}

	req, err := http.NewRequest("GET", href, nil)
	checkError(err)

	// 세션값 추가
	if session != "" {
		req.Header.Set("Cookie", session)
	}

	// 헤더 추가
	headerSlice := strings.Split(header, ":")
	req.Header.Add(headerSlice[0], headerSlice[1])

	response, err := netClient.Do(req)
	checkError(err)
	defer response.Body.Close()

	// response, err := netClient.Get(href)
	// checkError(err)
	// defer response.Body.Close()

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

	// 아규먼츠 변수에 할당
	baseURL := flag.String("host", "http://localhost", "Input host, ex) https://www.youtube.com")
	fileName := flag.String("file", "~/output.txt", "Input Output filename, ex) /home/scent2d/output.txt")
	session := flag.String("session", "", "Input session, ex) session=session_value")
	header := flag.String("header", "User-Agent: SCrawler 1.0.0", "Input Header, ex) Authorazation: abcdef")
	proxyURL := flag.String("proxy", "127.0.0.1:8888", "Input Proxy, ex) Proxy On -> http://127.0.0.1:8888")

	flag.Parse()
	file, err := os.OpenFile(*fileName, os.O_CREATE|os.O_RDWR, os.FileMode(0777))
	checkError(err)
	defer file.Close()

	w := bufio.NewWriter(file)

	// baseURL := arguments[0]

	go func() {
		queue <- *baseURL
	}()

	for href := range queue {
		if !hasVisited[href] && isSameDomain(href, *baseURL) {
			crawlURL(href, *session, *header, *proxyURL)
			// crawlURL(href, *session, *header)
			w.WriteString(href + "\r\n")
		}
		w.Flush()
	}
}
