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
	"flag"
)

func requestWorker(domains chan string, wg *sync.WaitGroup, db *sql.DB) {
	defer wg.Done()

	httpclient.Defaults(httpclient.Map{
		httpclient.OPT_USERAGENT: "my awsome httpclient",
		httpclient.OPT_TIMEOUT: 1,
		"Accept-Language": "en-us",
	})

	for domain := range domains {
		println(domain)
		res, err := httpclient.Get(domain, map[string]string{})

		var statusCode int;

		if (err == nil) {
			statusCode = res.StatusCode
			fmt.Println(res.StatusCode, err)
		} else {
			statusCode = -1
			fmt.Sprintf("HTTP Error %s", err.Error())
		}

		go func() {
			db.Exec(fmt.Sprintf("INSERT INTO sites(domain, code) values(\"%s\", %d)", domain, statusCode))
		}()
	}
}

func initDataBase() *sql.DB {
	db, err := sql.Open("sqlite3", "./sites.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	sqlStmt := `
	DELETE FROM sites;
	DROP TABLE sites;
	`
	db.Exec(sqlStmt)
	if err != nil {
		log.Panic("%q: %s\n", err, sqlStmt)
		os.Exit(1)
	}

	sqlStmt = `
	CREATE TABLE sites (domain TEXT NOT NULL PRIMARY KEY, code INTEGER);
	`
	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Panic("%q: %s\n", err, sqlStmt)
		os.Exit(1)
	}


	return db
}
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var threadsCount int;
	flag.IntVar(&threadsCount, "threads", 15, "Count of used threads (goroutines)")

	file, err := os.Open("ru_domains_200_ok")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	db := initDataBase()


	taskChannel := make(chan string, 10000)

	wg := new(sync.WaitGroup)

	for i := 0; i < threadsCount; i++ {
		wg.Add(1)
		go requestWorker(taskChannel, wg, db)
	}

	for scanner.Scan() {
		domain := fmt.Sprintf("http://%s/wso.php", scanner.Text())
		taskChannel <- domain
	}

	close(taskChannel)
	wg.Wait()

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
