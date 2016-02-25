package main

import (
	"os"
	"log"
	"bufio"
	"fmt"
	"sync"
	_ "github.com/mattn/go-sqlite3"
	"database/sql"
	"runtime"
	"flag"
	"net/http"
)

func requestWorker(domains chan string, wg *sync.WaitGroup, db *sql.DB) {
	defer wg.Done()

	httpClient := &http.Client{}

	for domain := range domains {
		response, err := httpClient.Get(domain)
		var statusCode int;

		if (err == nil) {
			statusCode = response.StatusCode
			log.Println(fmt.Sprintf("[%s] %d", domain, response.StatusCode))
		} else {
			statusCode = -1
			log.Println(fmt.Sprintf("[%s] HTTP Error %s", domain, err.Error()))
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

	taskChannel := make(chan string, 10000)
	go func() {
		for scanner.Scan() {
			domain := fmt.Sprintf("http://%s/wso.php", scanner.Text())
			taskChannel <- domain
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}

		close(taskChannel)
	}()

	db := initDataBase()
	wg := new(sync.WaitGroup)

	for i := 0; i < threadsCount; i++ {
		wg.Add(1)
		go requestWorker(taskChannel, wg, db)
	}

	wg.Wait()
}
