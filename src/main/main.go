package main

import (
	"fmt"
	"math/rand"
	"time"
)

/*
ch <- v    // Send v to channel ch.		('output su canale)
v := <-ch  // Receive from ch, and		(input su canale)
           // assign value to v.
(The data flows in the direction of the arrow.)
*/

type Signal struct{}

type Sleigh struct{
	DeerReady chan Signal
	SantaReady chan Signal
}

func (self *Sleigh) Run(){
	for{
		<- self.SantaReady
		<- self.DeerReady
		fmt.Printf("MERRY CHRISTMAS\n")
	}
}

type SantaSem struct{
	Wait chan Signal
	Signal chan Signal
}

func (self *SantaSem) Run() {
	fmt.Println("Santa semaphore goroutine started")
	for {
		<- self.Wait
		//fmt.Println("> Wait called: someone accessed SantaSem...")
		<- self.Signal
		//fmt.Println("> Signal called: SantaSem released...\n-------------")
	}
}

type Sem struct{
	Wait chan Signal
	Signal chan Signal
}

func (self *Sem) Run() {
	fmt.Println("Mutex semaphore goroutine started")
	for {
		// aspetta il segnale su wait, lo butta
		<- self.Wait
		//fmt.Println("> Wait called: someone accessed Sem...")
		<- self.Signal
		//fmt.Println("> Signal called: Sem released...\n-------------")
	}
}

type DeerSem struct{
	Wait chan Signal
	Signal chan Signal
}

func (self *DeerSem) Run() {
	fmt.Println("Deer semaphore goroutine started")
	for {
		// aspetta il segnale su wait, lo butta
		<- self.Wait
		//fmt.Println("> Wait called: someone accessed DeerSem...")
		<- self.Signal
		//fmt.Println("> Signal called: DeerSem released...\n-------------")
	}
}

type DeerCounter struct{
	deerNum 		int
	Return 			chan int
	Reset 			chan Signal	// TODO: check: using all?
	CheckDeers 		chan Signal
	PrepareSleigh 	chan Signal
	//santaSignal chan Signal
}

func (self *DeerCounter) Run(ss *SantaSem) {
	var curVal int
	fmt.Println("DeerCounter goroutine started")
	curVal = self.deerNum
	for {

		if curVal == 9 {
			fmt.Printf("All the deers have returned; waking up santa...\n")
			// Signal santa
			ss.Signal <- Signal{}
			// Ask to prepare the sleigh
			<- self.PrepareSleigh
			curVal = 0	// reset counter

		} else if curVal < 9 {
			select {
				case i:= <- self.Return: 		// TODO: check, ma ha senso che lo riceva
					fmt.Printf("> Deer n°%d has returned.\n",i)
					curVal = curVal + 1
					fmt.Printf("Number of deers: %d\n", curVal)
				case self.CheckDeers <- Signal{}:
					fmt.Println("No 9 deers are available: go help elves")
				//Timeout if no choice is made
				//case <-time.After(5*time.Second):
				//	fmt.Println("Deer Counter timed out after 5s")
		}

		}
	}
}

/* SANTA
santaSem . wait ()
mutex . wait ()
	if reindeer >= 9:
		prepareSleigh ()
		reindeerSem . signal (9)
		reindeer -= 9
	else if elves == 3:
		helpElves ()
mutex . signal ()
*/

func santa (santSem *SantaSem, mutexSem *Sem, deerSem *DeerSem, deerCount *DeerCounter, sleigh *Sleigh){
	for {
		// invia il segnale su wait
		santSem.Wait <- Signal{}
		mutexSem.Wait <- Signal{}
		// ++ Wait a second ++
		time.Sleep(time.Duration(time.Second))
		fmt.Println("Santa is awake, choosing who to help...")
		select {
			//SantaDeers
			case deerCount.PrepareSleigh <- Signal{}:
				fmt.Printf("* prepareSleigh *")
				// Signal deers
				for i:=1; i<=9; i++ {
					deerSem.Signal <- Signal{}
				}
				// ++ Wait a second ++
				time.Sleep(time.Duration(time.Second))
				// Hop on the sleigh
				sleigh.SantaReady <- Signal{}

			// checkDeers.SantaElves
			case <- deerCount.CheckDeers:
				fmt.Printf("Deer are not 9; help elves\n")
				// TODO: signal/help the 3 elves
			//Timeout if no choice is made
			//case <-time.After(5*time.Second):
			//	fmt.Println("Santa timed out after 5s")
		}
		mutexSem.Signal <- Signal{}

	}
}

/* REINDEER
mutex . wait ()
	reindeer += 1
	if reindeer == 9:
		santaSem . signal ()
mutex . signal ()

reindeerSem . wait ()
getHitched ()
*/

func reindeer (mutexSem *Sem, deerSem *DeerSem, deerCount *DeerCounter, deerNo int, sleigh *Sleigh){


	for {
		fmt.Printf("Reindeer %d is going on vacation... Will be back eventually!\n", deerNo)
		// Deer stays on vacation for 2 to 6 seconds.
		time.Sleep(time.Duration(rand.Int63n(2*1e9) + 4*1e9))

		/* sleep for at most 2 seconds */
		time.Sleep(time.Duration(rand.Int63n(2*1e9)))
		mutexSem.Wait <- Signal{}

		deerCount.Return <- deerNo

		/* sleep for at most 2 seconds */
		time.Sleep(time.Duration(rand.Int63n(2*1e9)))
		mutexSem.Signal <- Signal{}
		// TODO: getHitched? Not good! They go on vacation!
		deerSem.Wait <- Signal{}
		fmt.Printf("Deer %d is being hitched...\n", deerNo)
		sleigh.DeerReady <- Signal{}
	}

}

func main(){
	fmt.Println("Start")
	ss := SantaSem{ make(chan Signal), make(chan Signal)}
	go ss.Run()
	// put santa to sleep
	ss.Wait <- Signal{}
	ms := Sem{ make(chan Signal), make(chan Signal)}
	go ms.Run()
	ds := DeerSem{ make(chan Signal), make(chan Signal)}
	go ds.Run()
	// Lock the warming hut (so deers don't go out prematurely)
	ds.Wait <- Signal{}
	dc := DeerCounter{0, make(chan int), make(chan Signal),
	                make(chan Signal), make(chan Signal)}
	go dc.Run(&ss)

	sleigh := Sleigh{make(chan Signal, 9), make(chan Signal)}

	for i:= 1; i<= 9; i++{
		go reindeer(&ms,&ds,&dc,i, &sleigh)
	}
	go santa(&ss, &ms, &ds, &dc, &sleigh)

	// All goroutines are killed when main ends
	time.Sleep(time.Duration(60*time.Second))
}



/* SANTA
santaSem . wait ()
mutex . wait ()
	if reindeer >= 9:
		prepareSleigh ()
		reindeerSem . signal (9)
		reindeer -= 9
	else if elves == 3:
		helpElves ()
mutex . signal ()
*/

/* REINDEER
mutex . wait ()
	reindeer += 1
	if reindeer == 9:
		santaSem . signal ()
mutex . signal ()

reindeerSem . wait ()
getHitched ()
*/

/*
elfTex . wait ()
mutex . wait ()
	elves += 1
	if elves == 3:
		santaSem . signal ()
	else
		elfTex . signal ()
mutex . signal ()

getHelp ()

mutex . wait ()
	elves -= 1
	if elves == 0:
		elfTex . signal ()
mutex . signal ()
*/