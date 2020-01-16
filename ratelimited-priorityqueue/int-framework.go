package main

import (
	"encoding/json"
    "fmt"
    "io/ioutil"
    "os"
	"sync"
	"time"
	"strings"
)


type Candidate struct {
	Id string
	Name string
	JobRec string
}
func controller(main chan Candidate, priority chan Candidate, outputChannel chan Candidate, parallelization int) {


	tick := time.Tick(1000 * time.Millisecond)
	boom := time.After(500 * time.Millisecond)

	for {
		<-tick // run this loop once per second
		var wg sync.WaitGroup
		// process a few items simultaneously
		PARALLELBATCH: for perTickCandidateCounter := 0; perTickCandidateCounter < parallelization; perTickCandidateCounter++ {
			select {
			case currCandidate := <-priority:

				go zEngineInt(currCandidate, outputChannel)

			case currCandidate := <-main:

				go zEngineInt(currCandidate, outputChannel)
			case <-boom:
				break PARALLELBATCH;
			}
		}
		wg.Wait()
	}
}
func zEngineInt(inputCandidate Candidate, outputChannel chan Candidate) {

	if(strings.Compare(strings.ToLower(inputCandidate.Name)[0:1], "e") < 1) {
		inputCandidate.JobRec = "Data Scientist";
	} else {
		inputCandidate.JobRec = "Java Developer";
	}
	if(strings.ToLower(inputCandidate.Name) != inputCandidate.Name) {
		inputCandidate.JobRec += " 2"
	}

	outputChannel <- inputCandidate;
}
func saveToOutputBucket(outputChannel chan Candidate) {
	for {
		select {
		case currCandidate := <-outputChannel:
			fmt.Println(currCandidate.Name, ":", currCandidate.JobRec);
		}
	}
}
func main() {

	pageSize := 10
	concurrency := 3
	fmt.Println("Process all developers at a rate of ", concurrency, " per second. Process all Senior devs when they become available 2.5 seconds after processing begins.")
	fmt.Println("Senior devs are the video game characters.");
	
	jrData, jrErr := ioutil.ReadFile(os.Args[1])
	if(jrErr != nil) {
		panic(jrErr);
	}
	srData, srErr := ioutil.ReadFile(os.Args[2])
	if(srErr != nil) {
		panic(srErr);
	}
	var jrCandidates []Candidate;
	var srCandidates []Candidate;
	_ = json.Unmarshal([]byte(jrData), &jrCandidates)
	json.Unmarshal([]byte(srData), &srCandidates)
	
	fmt.Println("jr records: ", len(jrCandidates))
	fmt.Println("sr records: ", len(srCandidates))

	

	// create main queue and priority queue
	mainChannel := make(chan Candidate)
	priorityChannel := make(chan Candidate);
	outputChannel := make(chan Candidate);
	
	// instantiate go routines. These will run continuously. Go is smart enough to stop these when all channels are drained.
	go controller(mainChannel, priorityChannel, outputChannel, concurrency);
	go saveToOutputBucket(outputChannel);
	
	// put first page onto the main channel
	for firstPageOffset := 0; firstPageOffset < pageSize; firstPageOffset++ {
		mainChannel <- jrCandidates[firstPageOffset];
	}
	
	// Simulate secondary batch completion by waiting 2.5 seconds before processing the next file.
	time.Sleep(2500 * time.Millisecond)
	
	// Now put priority items onto priority channel
	for srOffset := 0; srOffset < len(srCandidates); srOffset++ {
		priorityChannel <- srCandidates[srOffset];
	}
	// put remaining page(s) onto main channel
	for remainingOffset := pageSize; remainingOffset < len(jrCandidates); remainingOffset++ {
		mainChannel <- jrCandidates[remainingOffset];
	}


	
}