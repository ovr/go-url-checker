package main

import (
	"github.com/ddliu/go-httpclient"
	"os"
	"log"
	"bufio"
	"fmt"
	"sync"
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	db, err := sql.Open("sqlite3", "./sites.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
	DELETE FROM sites;
	DROP TABLE sites;
	CREATE TABLE sites (domain TEXT NOT NULL PRIMARY KEY, code INTEGER);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}

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

	taskChannel := make(chan string, 1000)

	go func() {
		var wg sync.WaitGroup

		for {
			select {
			case domain := <-taskChannel:
				wg.Add(1)

				go func() {
					println(domain)
					res, err := httpclient.Get(domain, map[string]string{})

					if (err == nil) {
						db.Exec(fmt.Sprintf("INSERT INTO sites(domain, code) values(\"%s\", %d)", domain, res.StatusCode))
						println(res.StatusCode, err)
					} else {
						db.Exec(fmt.Sprintf("INSERT INTO sites(domain, code) values(\"%s\", %d)", domain, -1))
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
		taskChannel <- domain
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
