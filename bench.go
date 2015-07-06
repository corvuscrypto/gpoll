package gpoll

import (
		"fmt"
		"errors"
		"time"
		)

//###### TESTING FUNCTIONS ######

func (b *Broadcaster) benchmark(size int, load int, quant int) error{
	if size <= 0 {
		return errors.New("'size' variable in *Broadcaster.benchmark(size, load, quant) is not allowed to be negative or 0")
	}
	if load <= 0 {
		return errors.New("'load' variable in *Broadcaster.benchmark(size, load, quant) is not allowed to be negative or 0")
	}
	if quant <= 0 {
		return errors.New("'quant' variable in *Broadcaster.benchmark(size, load, quant) is not allowed to be negative or 0")
	}
	if quant > b.CLIENT_BUFFER_SIZE {
		return errors.New("The number of messages ('quant') you are benchmarking with must be less than or equal to client buffer size.")
	}
	fmt.Println("Benchmarking server...")
	b.isBenchmarking = true
	var testMessage string = string(make([]byte, size))
	nT := new(Routine)
	nT.Pipe = make(chan string,b.ROUTINE_BUFFER_SIZE)
	b.routines = append(b.routines, *nT)
	go b.handleMessage(nT.Pipe, len(b.routines))
	fmt.Printf("\nTesting with %d clients...\n", load)

	for i := 0; i < load; i++ {
		b.addClient()
	}
	testTime := time.Now()
	for i:=0; i<quant;i++{
		for _,t := range b.routines{
			b.wg.Add(1)
			t.Pipe <- testMessage
		}
			b.wg.Wait()
	}


	finishTime := time.Since(testTime).Nanoseconds()
	b.isBenchmarking = false
	fmt.Printf("Broadcast of %d messages to %d clients finished in %d ns!\n", quant, load, finishTime)
	if b.CSalt == "" {
		b.CSalt = UUID()
	}
	for _,t := range b.routines {
		b.wg.Add(1)
		t.Pipe <- "quit" + b.CSalt
	}
	b.wg.Wait()
	b.routines = nil


	var averageLag int64 = 0
	var maxLag int64 = 0
	var minLag int64 = 0
	for _,d := range b.benchData {

		averageLag += d.Ns
		if d.Ns > maxLag || maxLag == 0 {
			maxLag = d.Ns
		}
		if d.Ns < minLag || minLag == 0 {
			minLag = d.Ns
		}

	}
	averageLag /= int64(len(b.benchData))
	fmt.Printf("_________________\n\n Test Summary\n_________________"+
				"\n\n   Min. Latency: %d ns\n   Average Latency: %d ns\n   Max Latency: %d ns\n\n",
				minLag, averageLag, maxLag)

	return nil
}
