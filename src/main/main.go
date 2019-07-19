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

type Sem struct{
	name 	string
	Wait 	chan Signal
	Signal 	chan Signal
}

func (self *Sem) Run() {
	fmt.Printf("> %s semaphore goroutine started\n", self.name)
	for {
		// aspetta il segnale su wait, lo butta
		<- self.Wait
		//fmt.Printf("> Wait called: someone accessed %s ...\n", self.name)
		<- self.Signal
		//fmt.Println("> Signal called: %s released...\n-------------\n", self.name)
	}
}


type ElfCounter struct{
	elfNum			int
	problem 		chan int
	solveProblem	chan int
}

func (self *ElfCounter) Run(elfTex *Sem, santaSem *Sem){
	var curVal int
	fmt.Printf("> elfCounter goroutine started")
	curVal = self.elfNum
	for {
		if curVal == 3{
			fmt.Printf("Three elves have problems! Waking up santa...\n")
			// Signal santa
			santaSem.Signal <- Signal{}
			// Last elf asks for problem to be solved, but does not release elfTex
			elfId := <-self.solveProblem
			fmt.Printf("Solved elf n°%d's problem!\n", elfId)
		} else if curVal < 3{
		select {
			case elfId := <-self.problem:
				fmt.Printf("Elf n°%d has a problem!\n", elfId)
				elfTex.Signal <- Signal{}
				curVal++
			case elfId := <-self.solveProblem:
				fmt.Printf("Solved elf n°%d's problem!\n", elfId)
				// The last elf having its problem solved releases ElfTex
				if curVal == 1 { elfTex.Signal <- Signal{} }
				curVal--
			}
		}
	}
}
type DeerCounter struct{
	deerNum 		int
	Return 			chan int
	CheckDeers 		chan Signal
	PrepareSleigh 	chan Signal
	//santaSignal chan Signal
}

func (self *DeerCounter) Run(santaSem *Sem) {
	var curVal int
	fmt.Println("> deerCounter goroutine started")
	curVal = self.deerNum
	for {

		if curVal == 9 {
			fmt.Printf("All the deers have returned; waking up santa...\n")
			// Signal santa
			santaSem.Signal <- Signal{}
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

func santa (santaSem *Sem, mutexSem *Sem, deerSem *Sem, deerCount *DeerCounter){
	for {
		// invia il segnale su wait
		santaSem.Wait <- Signal{}
		mutexSem.Wait <- Signal{}
		// ++ Wait a second ++
		time.Sleep(time.Duration(time.Second))
		fmt.Println("Santa is awake, choosing who to help...")
		select {
		//SantaDeers
		case deerCount.PrepareSleigh <- Signal{}:
			fmt.Printf("* prepareSleigh *\n")
			// Signal deers
			for i:=1; i<=9; i++ {
				deerSem.Signal <- Signal{}
			}
			/* sleep for at most 2 seconds */
			time.Sleep(time.Duration(rand.Int63n(2*1e9)))
			fmt.Printf("Christmas time!\n")

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

func reindeer (mutexSem *Sem, deerSem *Sem, deerCount *DeerCounter, deerNo int){


	//for {	// TODO: endless loop?
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
	//}

}
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

func elf (mutexSem *Sem, elfTex *Sem, deerNo int){
	for {
		elfTex.Wait <- Signal{}
		mutexSem.Wait <- Signal{}
		// TODO: aspetta, lavora, fai problemi
	}
}


func main(){
	fmt.Println("> Starting program...")
	ss := Sem{"santaSem", make(chan Signal), make(chan Signal)}
	go ss.Run()
	// put santa to sleep
	ss.Wait <- Signal{}
	ms := Sem{"mutexSem", make(chan Signal), make(chan Signal)}
	go ms.Run()
	ds := Sem{"deerSem", make(chan Signal), make(chan Signal)}
	go ds.Run()
	// Lock the warming hut (so deers don't go out prematurely)
	ds.Wait <- Signal{}
	dc := DeerCounter{0, make(chan int),
		make(chan Signal), make(chan Signal)}
	go dc.Run(&ss)

	/* sleep for at most 1 second */
	time.Sleep(time.Duration(rand.Int63n(1*1e9)))
	for i:= 1; i<= 9; i++{
		go reindeer(&ms,&ds,&dc,i)
	}
	go santa(&ss, &ms, &ds, &dc)

	// All goroutines are killed when main ends
	time.Sleep(time.Duration(30*time.Second))
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