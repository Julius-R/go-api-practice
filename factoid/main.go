package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type fact struct {
	Id        string `json:"id"`
	Text      string `json:"text"`
	Source    string `json:"source"`
	SourceUrl string `json:"source_url"`
}

type factoids struct {
	Facts []fact `json:"facts"`
}

func handleError(err error) {
	logFile, errFi := os.OpenFile("./errors.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errFi != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.Printf("Error: %v\n", err)
	os.Exit(1)
}

func getRandomFact(wg *sync.WaitGroup) (fact, error) {
	defer wg.Done()
	url := "https://uselessfacts.jsph.pl/api/v2/facts/random?language=en"
	client := &http.Client{Timeout: 10 * time.Second}

	res, err := client.Get(url)
	if err != nil {
		return fact{}, fmt.Errorf("error making request: %v", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fact{}, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var r fact
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return fact{}, fmt.Errorf("error decoding response body: %v", err)
	}

	return r, nil
}

func main() {
	startTime := time.Now()
	var facts factoids
	runs := 6
	wg := &sync.WaitGroup{}
	ch := make(chan fact, runs)

	for i := runs; i > 0; i-- {
		wg.Add(1)
		go func() {
			fact, err := getRandomFact(wg)
			if err != nil {
				handleError(err)
				return
			}
			ch <- fact
		}()
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for factoid := range ch {
		facts.Facts = append(facts.Facts, factoid)
	}

	data, err := json.MarshalIndent(facts, "", " ")

	if err != nil {
		handleError(err)
	}

	err = os.WriteFile("./data.json", data, 0666)

	if err != nil {
		handleError(err)
	}

	fmt.Println(fmt.Sprintf("Code finished running in %s", time.Since(startTime).String()))
}
