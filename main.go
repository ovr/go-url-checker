package main

import (
	"github.com/ddliu/go-httpclient"
	"os"
	"log"
	"bufio"
	"fmt"
)

func main()  {
	httpclient.Defaults(httpclient.Map {
		httpclient.OPT_USERAGENT: "my awsome httpclient",
		"Accept-Language": "en-us",
	})

	file, err := os.Open("ru_domains_200_ok")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)


	taskChannel := make(chan string, 10)


	go func() {
		for  {
			select {
				case domain := <-taskChannel:
					go func() {
						print("1234")
						println(domain)
						res, err := httpclient.Get(domain, map[string]string{})

						if (err == nil) {
							println(res.StatusCode, err)
						} else {
							fmt.Sprintf("HTTP Error %s", err.Error())
						}
					}()
				default:
					fmt.Println("no message received")
					break
			}
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
