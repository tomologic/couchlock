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

// Version to return in "couchlock version" command
var Version = "0.0.0"

type response struct {
	Ok bool   `json:"ok"`
	ID string `json:"id"`
}

type queue struct {
	TotalRows int        `json:"total_rows,omitempty"`
	Rows      []queueRow `json:"rows,omitempty"`
}
type queueRow struct {
	Lock lock `json:"value,omitempty"`
}

type lock struct {
	ID      string `json:"_id,omitempty"`
	Lock    string `json:"Lock"`
	Name    string `json:"Name"`
	Status  string `json:"Status,omitempty"`
	Created uint64 `json:"Created,omitempty"`
}

type couchLockConfig struct {
	lock     string
	name     string
	couchdb  string
	interval int
}

var config couchLockConfig

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
		fmt.Printf("\tlock\t\tAquire lock\n")
		fmt.Printf("\tunlock\t\tUnlock lock\n")
		fmt.Printf("\tlist-queue\tList queue for lock\n")
		fmt.Printf("\tversion\t\tPrint current version\n\n")

	}

	flag.Parse()
	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	command := flag.Args()[0]

	if command == "version" {
		fmt.Println(Version)
		os.Exit(0)
	}

	if config.name == "" && (command == "lock" || command == "unlock") {
		fmt.Println("ERROR: Unique 'name' identifier for lock session required.")
		flag.Usage()
		os.Exit(1)
	}

	verifyDesignUpdate()

	if command == "lock" {
		lock := createLock()
		waitForLock(lock)
		lockLock(lock)
	} else if command == "unlock" {
		unlockLock()
	} else if command == "list-queue" {
		listQueue()
	} else {
		fmt.Printf("ERROR: Unknown command '%s'.\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

func verifyDesignUpdate() {
	client := &http.Client{}
	designLocksURL := config.couchdb + "/_design/locks"

	resp, err := http.Get(designLocksURL)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		os.Exit(1)
	}
	if resp.StatusCode == 404 {
		// View document probably doesn't exist

		// Get design document
		designDocument, err := Asset("data/designs/locks.json")
		if err != nil {
			panic(err)
		}

		// Create design document in couchdb
		buf := bytes.NewBuffer(designDocument)
		req, err := http.NewRequest("PUT", designLocksURL, buf)
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

func createLock() *lock {
	client := &http.Client{}

	// create lock
	lock := &lock{Lock: config.lock, Name: config.name}

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
	fmt.Printf("INFO: lock '%s' created.\n", config.lock)

	buf = new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	var res response
	err = json.Unmarshal(buf.Bytes(), &res)
	if err != nil {
		panic(err)
	}

	// save id for couchdb lock document
	lock.ID = res.ID

	return lock
}

func lockLock(lock *lock) {
	client := &http.Client{}

	// change status of lock in couchdb to 'locked'
	req, err := http.NewRequest("POST", config.couchdb+"/_design/locks/_update/lock/"+lock.ID, nil)
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

func unlockLock() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.couchdb+"/_design/locks/_view/queue/?startkey=[\""+config.lock+"\"]&endkey=[\""+config.lock+"\",{}]", nil)
	if err != nil {
		panic(err)
	}

	// wait until our lock is top of list
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	var queue queue
	err = json.Unmarshal(buf.Bytes(), &queue)
	if err != nil {
		panic(err)
	}

	if len(queue.Rows) == 0 {
		fmt.Printf("INFO: Lock '%s' is not locked. Doing nothing.\n", config.lock)
		os.Exit(0)
	}

	// Get first lock in list
	lock := queue.Rows[0].Lock
	if lock.Name != config.name {
		fmt.Printf("ERROR: Could not unlock '%s' since it's owned by '%s'.\n",
			config.lock,
			queue.Rows[0].Lock.Name)
		os.Exit(1)
	}

	req, err = http.NewRequest("POST", config.couchdb+"/_design/locks/_update/unlock/"+lock.ID, nil)
	resp, err = client.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != 201 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		fmt.Printf("ERROR: %d %s\n", resp.StatusCode, buf.String())
		os.Exit(1)
	}
	fmt.Printf("INFO: Lock '%s' unlocked.\n", config.lock)
}

func waitForLock(lock *lock) bool {
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

		var queue queue
		err = json.Unmarshal(buf.Bytes(), &queue)
		if err != nil {
			panic(err)
		}

		if len(queue.Rows) > 0 {
			if queue.Rows[0].Lock.ID == lock.ID {
				return true
			}
		}

		time.Sleep(time.Duration(config.interval) * 1000 * time.Millisecond)
	}

	return false
}

func listQueue() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.couchdb+"/_design/locks/_view/queue/?startkey=[\""+config.lock+"\"]&endkey=[\""+config.lock+"\",{}]", nil)
	if err != nil {
		panic(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	var queue queue
	err = json.Unmarshal(buf.Bytes(), &queue)
	if err != nil {
		panic(err)
	}

	for _, row := range queue.Rows {
		fmt.Printf("%d %s %s\n",
			row.Lock.Created,
			row.Lock.Name,
			row.Lock.Status)
	}
}
