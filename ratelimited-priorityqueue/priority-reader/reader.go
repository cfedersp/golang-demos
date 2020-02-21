package main

// to run:
// go run reader.go jr-candidates.txt sr-candidates.txt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	// "strconv"
)

/*
type Candidate struct {
	Id string
	Name string
	JobRec string
}
func controller(main chan Candidate, priority chan Candidate, outputChannel chan Candidate, parallelization int) {


	tick := time.Tick(1000 * time.Millisecond)
	boom := time.After(500 * time.Millisecond)

	loopNum := int64(0)

	for {
		loopNum++;
		//fmt.Println("+++++", loopNum);
		<-tick // run this loop once per second
		var wg sync.WaitGroup
		// process a few items simultaneously
	PARALLELBATCH: for perTickCandidateCounter := 0; perTickCandidateCounter < parallelization; perTickCandidateCounter++ {
		select {
		case currCandidate := <-priority:

			wg.Add(1);
			go zEngineInt(loopNum, currCandidate, outputChannel, &wg);

		case currCandidate := <-main:
			wg.Add(1);
			go zEngineInt(loopNum, currCandidate, outputChannel, &wg);
		case <-boom:
			//fmt.Println("######", loopNum);
			break PARALLELBATCH;
		}
	}
		wg.Wait()
		//fmt.Println("------", loopNum);
	}
}

func saveToOutputBucket(outputChannel chan Candidate) {
	for {
		select {
		case currCandidate := <-outputChannel:
			fmt.Println(currCandidate.Name, ":", currCandidate.JobRec);
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
*/

func main() {

	concurrency := 3
	fmt.Println("Process all developers at a rate of ", concurrency, " per second. Process all Senior devs when they become available 2.5 seconds after processing begins.")
	fmt.Println("Hint: Senior devs are the video game characters.");
	fmt.Println("This pipeline calls a recommendation service that simply suggests high level jobs to people that capitalize their name :)");

	jrData, jrErr := ioutil.ReadFile(os.Args[1])
	if(jrErr != nil) {
		panic(jrErr);
	}
	/*
	srData, srErr := ioutil.ReadFile(os.Args[2])
	if(srErr != nil) {
		panic(srErr);
	}
	*/

	var rawBucketRecords interface{}
	loadErr := json.Unmarshal(jrData, &rawBucketRecords);
	if(loadErr != nil) {
		panic(loadErr);
	}

	// := means you are allocating new memory
	// = means you are assigning an address or primitive to an L-Value

	var inputRecords = rawBucketRecords.([]interface{})
	for k,v := range inputRecords {
		fmt.Println("key: " , k, "value: ", v);
	}



}