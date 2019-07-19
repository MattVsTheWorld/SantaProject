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

// Empty signal
type Signal struct{}

// Semaphore struct and behaviour
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
	solveProblem	chan Signal
}

func (self *ElfCounter) Run(elfTex *Sem, santaSem *Sem){
	var curVal int
	fmt.Printf("> elfCounter goroutine started\n")
	curVal = self.elfNum
	for {
		if curVal == 3{
			fmt.Printf("Three elves have problems! Waking up santa...\n")
			// Signal santa
			santaSem.Signal <- Signal{}
			// Last elf asks for problem to be solved, but does not release elfTex
			//elfId := <- self.solveProblem
			<- self.solveProblem
			fmt.Printf("Solved and elf's problem!\n")
			curVal--
			//fmt.Printf("Solved elf n°%d's problem!\n", elfId)
		} else if curVal < 3{
		select {
			case elfId := <- self.problem:
				fmt.Printf("Elf n°%d has a problem!\n", elfId)
				elfTex.Signal <- Signal{}
				curVal++
			case <- self.solveProblem:
				//fmt.Printf("Solved elf n°%d's problem!\n", elfId)
				fmt.Printf("Solved and elf's problem!\n")
				// The last elf having its problem solved releases ElfTex
				if curVal == 1 {
					elfTex.Signal <- Signal{}
				}
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
					fmt.Println("> Deers are still on vacation: Best help elves.")
					//Timeout if no choice is made
					//case <-time.After(5*time.Second):
					//	fmt.Println("Deer Counter timed out after 5s")
			}

		}
	}
}

// Santa
/*
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
func santa (santaSem *Sem, mutexSem *Sem, deerSem *Sem, deerCounter *DeerCounter, elfCounter *ElfCounter){
	for {
		// invia il segnale su wait
		santaSem.Wait <- Signal{}
		mutexSem.Wait <- Signal{}
		// ++ Wait a second ++
		time.Sleep(time.Duration(time.Second))
		fmt.Println("Santa is awake, choosing who to help...")
		select {
		//SantaDeers
		case deerCounter.PrepareSleigh <- Signal{}:
			fmt.Printf("* prepareSleigh *\n")
			// Signal deers
			for i:=1; i<=9; i++ {
				deerSem.Signal <- Signal{}
			}
			/* sleep for at most 2 seconds */
			time.Sleep(time.Duration(rand.Int63n(2*1e9)))
			fmt.Printf("Christmas time!\n------------\n")

		// checkDeers.SantaElves
		case <- deerCounter.CheckDeers:
			fmt.Printf("* help elves *\n")
			for i:= 1; i<=3; i++ {
				elfCounter.solveProblem <- Signal{}
			}// TODO: signal/help the 3 elves; correct?
			/* sleep for at most 2 seconds */
			time.Sleep(time.Duration(rand.Int63n(2*1e9)))
			fmt.Printf("Helped three elves!\n------------\n")

			//Timeout if no choice is made
			//case <-time.After(5*time.Second):
			//	fmt.Println("Santa timed out after 5s")
		}
		mutexSem.Signal <- Signal{}

	}
}

// Reindeer
/*
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
		// Deer stays on vacation for 4 to 12 seconds.
		time.Sleep(time.Duration(rand.Int63n(4*1e9) + 8*1e9))

		/* sleep for at most 2 seconds */
		time.Sleep(time.Duration(rand.Int63n(2*1e9)))
		mutexSem.Wait <- Signal{}

		deerCount.Return <- deerNo

		/* sleep for at most 2 seconds */
		time.Sleep(time.Duration(rand.Int63n(2*1e9)))
		mutexSem.Signal <- Signal{}
		// TODO: getHitched? Not good! They go on vacation!
		// TODO: fine if not in loop... but if they are?
		deerSem.Wait <- Signal{}
		fmt.Printf("Deer %d is being hitched...\n", deerNo)
	//}
}

// Elf
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
func elf(mutexSem *Sem, elfTex *Sem, elfCount *ElfCounter, elfId int){
	fmt.Printf("Elf %d has gone to work ...\n", elfId)
	for {
		// Elf works for 2 to 10 seconds before he has a problem. Yikes!
		time.Sleep(time.Duration(rand.Int63n(8*1e9) + 2*1e9))
		elfTex.Wait 		<- Signal{}
		mutexSem.Wait 		<- Signal{}
		// The elf that wants its problem solves communicates its ID
		elfCount.problem 	<- elfId
		mutexSem.Signal 	<- Signal{}
		//<- self.helpElf 	// TODO: check helpElf

		//mutexSem.Wait <- Signal{}		// TODO: issue, can't do this without helpElf
		<- elfCount.solveProblem
		//mutexSem.Signal <- Signal{}


	}
}

func main(){
	fmt.Println("> Starting program...")
	const numElves = 20
	// ------- SEMAPHORES -------
	ss := Sem{"santaSem", make(chan Signal), make(chan Signal)}
	go ss.Run()
	ms := Sem{"mutexSem", make(chan Signal), make(chan Signal)}
	go ms.Run()
	ds := Sem{"deerSem", make(chan Signal), make(chan Signal)}
	go ds.Run()
	es := Sem{"elfTex", make(chan Signal), make(chan Signal)}
	go es.Run()
	// Put santa to sleep
	ss.Wait <- Signal{}
	// Lock the warming hut (so deers don't go out prematurely)
	ds.Wait <- Signal{}
	// ------- COUNTERS -------
	dc := DeerCounter{0, make(chan int),
		make(chan Signal), make(chan Signal)}
	go dc.Run(&ss)
	ec := ElfCounter{0, make(chan int), make(chan Signal) }
	go ec.Run(&es, &ss)

	// ------- ENTITIES -------
	/* sleep for at most 1 second */
	time.Sleep(time.Duration(rand.Int63n(1*1e9)))
	// ---- Start Reindeers ----
	for i:= 1; i<= 9; i++{
		go reindeer(&ms,&ds,&dc,i)
	}
	// ---- Start Elves ----
	for i:= 1; i<=numElves; i++{
		go elf(&ms, &es, &ec, i)
	}
	// ---- Start Santa ----
	go santa(&ss, &ms, &ds, &dc, &ec)

	// All goroutines are killed when main ends
	time.Sleep(time.Duration(30*time.Second))
}