package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Response struct {
	Ok bool   `json:"ok"`
	Id string `json:"id"`
}

type Queue struct {
	Total_rows int        `json:"total_rows,omitempty"`
	Rows       []QueueRow `json:"rows,omitempty"`
}
type QueueRow struct {
	Lock Lock `json:"value,omitempty"`
}

type Lock struct {
	Id      string `json:"_id,omitempty"`
	Lock    string `json:"Lock"`
	Name    string `json:"Name"`
	Status  string `json:"Status,omitempty"`
	Created int    `json:"Created,omitempty"`
}

func NewLock(lock string, name string) *Lock {
	return &Lock{Lock: lock, Name: name}
}

type Config struct {
	lock     string
	name     string
	couchdb  string
	interval int
}

var config Config

func main() {
	flag.StringVar(&config.lock, "lock", "default", "Lock name - default to 'default'.")
	flag.StringVar(&config.name, "name", "", "Unique identifier for lock session.")
	flag.StringVar(&config.couchdb, "couchdb", "http://couchdb/couchlock", "Couchdb - default to 'http://couchdb/couchlock'.")
	flag.IntVar(&config.interval, "interval", 5, "Polling interval in seconds - default to 5.")

	flag.Usage = func() {
		fmt.Printf("Usage: couchlock [options] command\n\n")

		fmt.Printf("Options:\n")
		flag.PrintDefaults()
		fmt.Printf("\n")

		fmt.Printf("Commands:\n")
		fmt.Printf("\tlock\tAquire lock\n")
		fmt.Printf("\tunlock\tUnlock lock\n")
		fmt.Printf("\tqueue\tList queue for lock\n\n")

	}

	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	if config.name == "" {
		fmt.Println("ERROR: Unique 'name' identifier for lock session required.")
		flag.Usage()
		os.Exit(1)
	}

	command := flag.Args()[0]
	if command == "lock" {
		verifyDesignUpdate()
		lock := createLock()
		waitForLock(lock)
		lockLock(lock)
	} else if command == "unlock" {
		fmt.Println("INFO: Not implemented.")
		os.Exit(1)
	} else if command == "queue" {
		fmt.Println("INFO: Not implemented.")
		os.Exit(1)
	} else {
		fmt.Printf("ERROR: Unknown command '%s'.\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

func verifyDesignUpdate() {
	client := &http.Client{}
	design_locks_url := config.couchdb + "/_design/locks"

	resp, err := http.Get(design_locks_url)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
	if resp.StatusCode == 404 {
		// View document probably doesn't exist

		// Get design document
		design_document, err := Asset("data/designs/locks.json")
		if err != nil {
			panic(err)
		}

		// Create design document in couchdb
		buf := bytes.NewBuffer(design_document)
		req, err := http.NewRequest("PUT", design_locks_url, buf)
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		if resp.StatusCode != 201 {
			buf := new(bytes.Buffer)
			buf.ReadFrom(resp.Body)
			fmt.Printf("ERROR: %d %s\n", resp.StatusCode, buf.String())
			os.Exit(1)
		}
		fmt.Printf("INFO: %s\n", "Locks design created.")
	} else {
		fmt.Printf("INFO: %s\n", "Locks design exists.")
	}
}

func createLock() *Lock {
	client := &http.Client{}

	// create lock
	lock := NewLock(config.lock, config.name)

	// create lock in couchdb
	json1, _ := json.Marshal(lock)
	buf := bytes.NewBuffer(json1)
	req, err := http.NewRequest("POST", config.couchdb+"/_design/locks/_update/create/", buf)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 201 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		fmt.Printf("ERROR: %d %s\n", resp.StatusCode, buf.String())
		os.Exit(1)
	}
	fmt.Printf("INFO: Lock '%s' created.\n", config.lock)

	buf = new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	var res Response
	err = json.Unmarshal(buf.Bytes(), &res)
	if err != nil {
		panic(err)
	}

	// save id for couchdb lock document
	lock.Id = res.Id

	return lock
}

func lockLock(lock *Lock) {
	client := &http.Client{}

	// change status of lock in couchdb to 'locked'
	req, err := http.NewRequest("POST", config.couchdb+"/_design/locks/_update/lock/"+lock.Id, nil)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 201 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		fmt.Printf("ERROR: %d %s\n", resp.StatusCode, buf.String())
		os.Exit(1)
	}
	fmt.Printf("INFO: Lock '%s' aquired.\n", config.lock)
}

func waitForLock(lock *Lock) bool {
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.couchdb+"/_design/locks/_view/queue/?startkey=[\""+config.lock+"\"]&endkey=[\""+config.lock+"\",{}]", nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("INFO: Waiting for lock '%s' to be available.\n", config.lock)

	// wait until our lock is top of list
	for true {
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)

		var queue Queue
		err = json.Unmarshal(buf.Bytes(), &queue)
		if err != nil {
			panic(err)
		}

		if queue.Rows[0].Lock.Id == lock.Id {
			return true
		}

		time.Sleep(time.Duration(config.interval) * 1000 * time.Millisecond)
	}

	return false
}
