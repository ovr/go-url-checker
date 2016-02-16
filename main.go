package main

import (
	"github.com/ddliu/go-httpclient"
	"os"
	"log"
	"bufio"
	"fmt"
	"sync"
)

func main() {
	httpclient.Defaults(httpclient.Map{
		httpclient.OPT_USERAGENT: "my awsome httpclient",
		httpclient.OPT_TIMEOUT: 1,
		"Accept-Language": "en-us",
	})

	file, err := os.Open("ru_domains_200_ok")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	taskChannel := make(chan string, 75)

	go func() {
		var wg sync.WaitGroup

		for {
			select {
			case domain := <-taskChannel:
				wg.Add(1)

				go func() {
					print("1234")
					println(domain)
					res, err := httpclient.Get(domain, map[string]string{})

					if (err == nil) {
						println(res.StatusCode, err)
					} else {
						fmt.Sprintf("HTTP Error %s", err.Error())
					}

					wg.Done()
				}()
			default:
				fmt.Println("no message received")
				break
			}

			wg.Wait()
		}
	}()

	for scanner.Scan() {
		domain := fmt.Sprintf("http://%s/wso.php", scanner.Text())
		println(domain)
		taskChannel <- domain
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
