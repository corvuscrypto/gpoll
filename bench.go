package gpoll

import (
		"fmt"
		"errors"
		"time"
		)

// Benchmark registers virtual clients and tests the server by sending messages of size 'msgSize'
// to all of them. The data is collected automatically and some basic statistics are output.
// This only looks at time, but may look at memory usage in the future.
func (b *Broadcaster) Benchmark(msgSize int, clients int, numMsgs int) error{
	if msgSize <= 0 {
		return errors.New("'msgSize' is not allowed to be negative or 0")
	}
	if clients <= 0 {
		return errors.New("'clients' is not allowed to be negative or 0")
	}
	if numMsgs <= 0 {
		return errors.New("'numMsgs' is not allowed to be negative or 0")
	}
	if numMsgs > b.ClientBufferSize {
		return errors.New("numMsgs must be less than or equal to client buffer size.")
	}
	fmt.Println("Benchmarking server...")
	b.IsBenchmarking = true
	var testMessage string = string(make([]byte, msgSize))
	nT := new(Routine)
	nT.Pipe = make(chan string,b.RoutineBufferSize)
	b.Routines = append(b.Routines, *nT)
	go b.HandleMessage(nT.Pipe, len(b.Routines))
	fmt.Printf("\nTesting with %d clients...\n", clients)

	for i := 0; i < clients; i++ {
		b.AddClient()
	}
	testTime := time.Now()
	for i:=0; i<numMsgs;i++{
		for _,t := range b.Routines{
			b.WG.Add(1)
			t.Pipe <- testMessage
		}
			b.WG.Wait()
	}


	finishTime := time.Since(testTime).Nanoseconds()
	b.IsBenchmarking = false
	fmt.Printf("Broadcast of %d messages to %d clients finished in %d ns!\n", numMsgs, clients, finishTime)
	if b.CSalt == "" {
		b.CSalt = UUID()
	}
	for _,t := range b.Routines {
		b.WG.Add(1)
		t.Pipe <- "quit" + b.CSalt
	}
	b.WG.Wait()
	b.Routines = nil


	var averageLag int64 = 0
	var maxLag int64 = 0
	var minLag int64 = 0
	for _,d := range b.BenchData {

		averageLag += d.Ns
		if d.Ns > maxLag || maxLag == 0 {
			maxLag = d.Ns
		}
		if d.Ns < minLag || minLag == 0 {
			minLag = d.Ns
		}

	}
	averageLag /= int64(len(b.BenchData))
	fmt.Printf("_________________\n\n Test Summary\n_________________"+
				"\n\n   Min. Latency: %d ns\n   Average Latency: %d ns\n   Max Latency: %d ns\n\n",
				minLag, averageLag, maxLag)

	return nil
}
