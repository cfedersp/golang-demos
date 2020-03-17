package main

// to run:
// go run int-framework.go jr-candidates.txt sr-candidates.txt

import (
	"encoding/json"
    "fmt"
    "io/ioutil"
    "os"
	"sync"
	"time"
	"strings"
	// "strconv"
)


type Candidate struct {
	Id string
	Name string
	JobRec string
}
type CandidateRecommendation struct { // GoLang doesnt support inheritance - interfaces dont get declared beforehand so theres no need
	looker Candidate
	JobRec string
	status int
}

func controller(main chan Candidate, priority chan Candidate, outputChannel chan CandidateRecommendation, adviceChannel chan int) {
	tick := time.Tick(1000 * time.Millisecond)
	boom := time.After(500 * time.Millisecond)

	workerCount := 0

	for {
		<-tick // run this loop once per second

		select {
		case workerAdvice := <-adviceChannel:
			workerCount = workerAdvice
		default:
		}

		var wg sync.WaitGroup
		// process a few items simultaneously
		PARALLELBATCH: for perTickCandidateCounter := 0; perTickCandidateCounter < workerCount; perTickCandidateCounter++ {

			select {
			case currCandidate := <-priority:
				wg.Add(1);
				go zEngineInt(currCandidate, outputChannel, &wg)

			case currCandidate := <-main:
				wg.Add(1);
				go zEngineInt(currCandidate, outputChannel, &wg)
			case <-boom: // if we receive a boom event before receiving workerCount records, break
				break PARALLELBATCH;
			}
		}
		wg.Wait()
	}
}

func zEngineInt(inputCandidate Candidate, outputChannel chan CandidateRecommendation, wg *sync.WaitGroup) {
	var rec string;
	defer wg.Done()

	if(strings.Compare(strings.ToLower(inputCandidate.Name)[0:1], "e") < 1) {
		rec = "Data Scientist";
	} else {
		rec = "Java Developer";
	}
	if(strings.ToLower(inputCandidate.Name) != inputCandidate.Name) {
		rec += " 2"
	}

	outputChannel <- CandidateRecommendation{inputCandidate, rec, 200};
}
func saveToOutputBucket(outputChannel chan CandidateRecommendation) {
	for {
		select {
		case currCandidate := <-outputChannel:
			fmt.Println(currCandidate.looker.Name, ":", currCandidate.JobRec);
		}
	}
}
func queueCandidates(candidates []Candidate, queue chan Candidate, mainWg *sync.WaitGroup) {
	mainWg.Add(1);
	
	for remainingOffset := 0; remainingOffset < len(candidates); remainingOffset++ {
		queue <- candidates[remainingOffset];
	}
	
	defer mainWg.Done();
}

func main() {

	concurrency := 3
	fmt.Println("Process all developers at a rate of ", concurrency, " per second. Process all Senior devs when they become available 2.5 seconds after processing begins.")
	fmt.Println("Hint: Senior devs are the video game characters.");
	fmt.Println("This pipeline calls a recommendation service that simply suggests high level jobs to people that capitalize their name :)");
	
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
	outputChannel := make(chan CandidateRecommendation);
	adviceChannel := make(chan int)
	
	// if a goRoutine has channel-consuming methods, go is smart enough to keep it running until the channels are drained.
	go controller(mainChannel, priorityChannel, outputChannel, adviceChannel);
	go saveToOutputBucket(outputChannel);
	adviceChannel <- concurrency;
	
	// if a goRoutine only has synchronous logic pushing to a channel, we have to explicitly define a WaitGroup and wait for it.
	var mainAsyncMethods sync.WaitGroup
	go queueCandidates(jrCandidates, mainChannel, &mainAsyncMethods);
	// Simulate secondary batch completion by waiting 2.5 seconds before processing the next file.
	time.Sleep(2500 * time.Millisecond)
	go queueCandidates(srCandidates, priorityChannel, &mainAsyncMethods);
	
	mainAsyncMethods.Wait();

	
}