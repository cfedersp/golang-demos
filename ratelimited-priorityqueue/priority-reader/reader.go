package main

// to run:
// go run reader.go ../jr-candidates.txt ../sr-candidates.txt

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
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
}

type CandidateRecommendation struct { // GoLang doesnt support inheritance - interfaces dont get declared beforehand so theres no need
	looker DataPacket
	JobRec string
	status int
}

//func controller(main chan DataPacket, priority chan DataPacket, outputChannel chan DataPacket, parallelization int)
func controller(main chan DataPacket, priority chan DataPacket, outputChannel chan CandidateRecommendation, adviceChannel chan int) {
	tick := time.Tick(1000 * time.Millisecond)

	workerCount := 0

	for {
		<-tick // run this loop once per second

		select {
		case workerAdvice := <-adviceChannel:
			workerCount = workerAdvice
		default:
		}

		// a WaitGroup is a cool feature but we dont really want to wait for workers to finish,
		// we want to limit the number of live workers by controlling the number of new workers.
		if(workerCount > 0) {
			dispatcher(main, priority, outputChannel, workerCount);
		}
	}
}
func dispatcher(main chan DataPacket, priority chan DataPacket, outputChannel chan CandidateRecommendation, workerCount int) {
	boom := time.After(500 * time.Millisecond)

	// process a few items simultaneously
PARALLELBATCH: for perTickCandidateCounter := 0; perTickCandidateCounter < workerCount; perTickCandidateCounter++ {
	select {
		case currCandidate := <-priority:

			go zEngineInt(currCandidate, outputChannel)

		case currCandidate := <-main:

			go zEngineInt(currCandidate, outputChannel)
		case <-boom: // if we receive a boom event before receiving workerCount records, break
			break PARALLELBATCH;
		}
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
func zEngineInt(inputCandidate DataPacket, outputChannel chan CandidateRecommendation) {
	var rec string;
	// we can only cast individual objects to their underlying original concrete type:
	var genericMap = inputCandidate.data.(map[string]interface{}) // cast the interface to a map<string,interface>
	var candidateName = genericMap["name"].(string) // cast the name to a string

	// we cant cast to map[string]string or we get: panic: interface conversion: interface {} is map[string]interface {}, not map[string]string
	// var name = inputCandidate.data.(map[string]string)["name"]

	if(strings.Compare(strings.ToLower(candidateName)[0:1], "e") < 1) {
		rec = "Data Scientist";
	} else {
		rec = "Java Developer";
	}
	if(strings.ToLower(candidateName) != candidateName) {
		rec += " 2"
	}

	outputChannel <- CandidateRecommendation{inputCandidate, rec, 200};
}

func saveToOutputBucket(outputChannel chan CandidateRecommendation) {
	for {
		select {
		case currCandidate := <-outputChannel:
			// fmt.Println(currCandidate.Name, ":", currCandidate.JobRec);
			fmt.Println("source: ", currCandidate.looker.source, ", data: ", currCandidate.looker.data);
		}
	}
}

func queueCandidates(candidates []DataPacket, queue chan DataPacket, mainWg *sync.WaitGroup) {
	mainWg.Add(1);
	defer mainWg.Done();

	for remainingOffset := 0; remainingOffset < len(candidates); remainingOffset++ {
		queue <- candidates[remainingOffset];
	}


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
	// An empty interface may hold values of any type.
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
		var currentPacket = DataPacket{"jr", "none for now", v};
		jrDataPackets = append(jrDataPackets, currentPacket);
		//fmt.Println("source: ", currentData.source, ", data: ", currentData.data);
	}

	for _,v := range srInputRecords {
		var currentPacket = DataPacket{"sr", "none for now", v};
		srDataPackets = append(srDataPackets, currentPacket);
		//fmt.Println("source: ", currentData.source, ", data: ", currentData.data);
	}


	// create main queue and priority queue
	mainChannel := make(chan DataPacket)
	priorityChannel := make(chan DataPacket);
	outputChannel := make(chan CandidateRecommendation);
	adviceChannel := make(chan int)


	// if a goRoutine has channel-consuming methods, go is smart enough to keep it running until the channels are drained.
	go controller(mainChannel, priorityChannel, outputChannel, adviceChannel);
	go saveToOutputBucket(outputChannel);
	// no workers will run until we tell them how many to run in parallel
	adviceChannel <- concurrency;
	// if a goRoutine only has synchronous logic pushing to a channel, we have to explicitly define a WaitGroup and wait for it.
	var mainAsyncMethods sync.WaitGroup
	go queueCandidates(jrDataPackets, mainChannel, &mainAsyncMethods);
	// Simulate secondary batch completion by waiting 2.5 seconds before processing the next file.
	time.Sleep(2500 * time.Millisecond)
	go queueCandidates(srDataPackets, priorityChannel, &mainAsyncMethods);

	mainAsyncMethods.Wait();


}