package main

// to run:
// go run reader.go jr-candidates.txt sr-candidates.txt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	// "strconv"
)

// generalized data structure
type DataPacket struct {
	source string
	// priority string
	config string
	data interface{}
	JobRec string
}

func controller(main chan DataPacket, priority chan DataPacket, outputChannel chan DataPacket, parallelization int) {


	tick := time.Tick(1000 * time.Millisecond)
	boom := time.After(500 * time.Millisecond)

	for {
		<-tick // run this loop once per second
		var wg sync.WaitGroup
		// process a few items simultaneously
		// this "just-in-time" scheduling+prioritizing for loop is not necessary - go channels can handle delegation for us
		// // How would we rate-limit in that case? number of workers would be tuned and the channel size would be low or zero
	PARALLELBATCH: for perTickCandidateCounter := 0; perTickCandidateCounter < parallelization; perTickCandidateCounter++ {
		select {
		case currCandidate := <-priority:

			wg.Add(1);
			go zEngineInt(currCandidate, outputChannel, &wg);

		case currCandidate := <-main:
			wg.Add(1);
			go zEngineInt(currCandidate, outputChannel, &wg);
		case <-boom:
			break PARALLELBATCH;
		}
	}
		wg.Wait()
		//fmt.Println("------", loopNum);
	}
}


// this is my 'scheduler'
// This loop runs as a goroutine, forever pulling work to do and delegating it.
/*
func schedulerLoop() {

	for {
		if rateIsBlocked() {
			time.Sleep((250 * time.Millisecond))
			continue
			}

		candidate, err := getNextCandidate() // getNext is our prioritizer
		if err != nil {
			log.Printf("Error getting candidate: %v\n", err)
			time.Sleep((250 * time.Millisecond))
			continue
		}
		if candidate == nil {
			time.Sleep((250 * time.Millisecond))
			continue
		}

		workChannel <- candidate
	}
}
 */
func zEngineInt(inputCandidate DataPacket, outputChannel chan DataPacket, wg *sync.WaitGroup) {

	defer wg.Done()

	//if(strings.Compare(strings.ToLower(inputCandidate.data.Name)[0:1], "e") < 1) {

	inputCandidate.JobRec = "Java Developer";
	if(inputCandidate.source == "sr") {
		inputCandidate.JobRec += " 2"
	}
	// inputCandidate.JobRec += " " + strconv.FormatInt(loopCount, 10);

	outputChannel <- inputCandidate;
}
func saveToOutputBucket(outputChannel chan DataPacket) {
	for {
		select {
		case currCandidate := <-outputChannel:
			// fmt.Println(currCandidate.Name, ":", currCandidate.JobRec);
			fmt.Println("source: ", currCandidate.source, ", data: ", currCandidate.data);
		}
	}
}

func queueCandidates(candidates []DataPacket, queue chan DataPacket, mainWg *sync.WaitGroup) {
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

	// load the files into interface objects
	var rawJrRecords interface{}
	jrLoadErr := json.Unmarshal(jrData, &rawJrRecords);
	if(jrLoadErr != nil) {
		panic(jrLoadErr);
	}
	var rawSrRecords interface{}
	srLoadErr := json.Unmarshal(srData, &rawSrRecords);
	if(srLoadErr != nil) {
		panic(srLoadErr);
	}

	type foo struct {
		Id int `json:"id"`
		CandidateName string `json:"candidate_name"`
		JobPositionName string `json:"job_position_name"`

	}

	// cast the children of the interface to an array of interface objects
	var jrInputRecords = rawJrRecords.([]interface{})
	var srInputRecords = rawSrRecords.([]interface{})

	var jrDataPackets []DataPacket;
	var srDataPackets []DataPacket;

	// put the interfaces into our standard structure, including src, config, JobRec
	for _,v := range jrInputRecords {
		var currentPacket = DataPacket{"jr", "none for now", v, ""};
		jrDataPackets = append(jrDataPackets, currentPacket);
		//fmt.Println("source: ", currentData.source, ", data: ", currentData.data);
	}

	for _,v := range srInputRecords {
		var currentPacket = DataPacket{"sr", "none for now", v, ""};
		srDataPackets = append(srDataPackets, currentPacket);
		//fmt.Println("source: ", currentData.source, ", data: ", currentData.data);
	}


	// create main queue and priority queue
	mainChannel := make(chan DataPacket)
	priorityChannel := make(chan DataPacket);
	outputChannel := make(chan DataPacket);


	// if a goRoutine has channel-consuming methods, go is smart enough to keep it running until the channels are drained.
	go controller(mainChannel, priorityChannel, outputChannel, concurrency);
	go saveToOutputBucket(outputChannel);

	// if a goRoutine only has synchronous logic pushing to a channel, we have to explicitly define a WaitGroup and wait for it.
	var mainAsyncMethods sync.WaitGroup
	go queueCandidates(jrDataPackets, mainChannel, &mainAsyncMethods);
	// Simulate secondary batch completion by waiting 2.5 seconds before processing the next file.
	time.Sleep(2500 * time.Millisecond)
	go queueCandidates(srDataPackets, priorityChannel, &mainAsyncMethods);

	mainAsyncMethods.Wait();


}