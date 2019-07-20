package main

import (
	"fmt"
	"math/rand"
	"strconv"
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
	Wait 	chan string
	Signal 	chan string
}

func (self *Sem) Run() {
	fmt.Printf("> %s semaphore goroutine started\n", self.name)
	for {
		// aspetta il segnale su wait, lo butta
		whoWait := <- self.Wait
		fmt.Printf("> Wait called: someone accessed %s (%s) ...\n", self.name, whoWait)
		whoSignal := <- self.Signal
		fmt.Printf("> Signal called: %s released (%s)...\n-------------\n", self.name, whoSignal)
	}
}

type ElfCounter struct{
	elfNum			int
	problem 		chan int
	solveProblem	chan Signal
	helpElf			chan Signal
}

func (self *ElfCounter) Run(elfTex *Sem, santaSem *Sem){
	var curVal int
	fmt.Printf("> elfCounter goroutine started\n")
	curVal = self.elfNum
	for {
		//if curVal == 3{
		//	fmt.Printf("Three elves have problems! Waking up santa...\n")
		//	// Signal santa
		//	santaSem.Signal <- "elfCounter"
		//	// Last elf asks for problem to be solved, but does not release elfTex
		//	//elfId := <- self.solveProblem
		//	self.solveProblem <- Signal{}
		//	fmt.Printf("Solved and elf's problem!\n")
		//	curVal--
		//	fmt.Printf(" > > > (curVal = 3) Number of elves: %d\n", curVal)
		//	//fmt.Printf("Solved elf n°%d's problem!\n", elfId)
		//} else {
		select {
			case elfId := <- self.problem:
				fmt.Printf("(ELFCOUNTER): Signal elf %d has a problem\n", elfId)
				curVal++
				//elfTex.Signal <- "elfCounter"	// let someone else in
				if curVal == 3 {
					fmt.Printf("Three elves have problems! Waking up santa...\n")
					santaSem.Signal <- "elfCounter"
				} else {
					elfTex.Signal <- "elfCounter"
				}
				fmt.Printf("> > > (problem) Number of elves: %d\n", curVal)
			case self.solveProblem <- Signal{}:
				fmt.Printf("Solved an elf's problem!\n")
				curVal--
				fmt.Printf(" > > > (solveProblem) Number of elves: %d\n", curVal)
				// The last elf having its problem solved releases ElfTex
				if curVal == 0 {
					elfTex.Signal <- "elfCounter"
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
			santaSem.Signal <- "deerCounter"
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
		santaSem.Wait <- "Santa"
		mutexSem.Wait <- "Santa"
		// ++ Wait a second ++
		time.Sleep(time.Duration(time.Second))
		fmt.Println("Santa is awake, choosing who to help...")
		select {
		//SantaDeers
		case deerCounter.PrepareSleigh <- Signal{}:
			fmt.Printf("* prepareSleigh *\n")
			// Signal deers
			for i:=1; i<=9; i++ {
				deerSem.Signal <- "Santa"
			}
			/* sleep for at most 2 seconds */
			time.Sleep(time.Duration(rand.Int63n(2*1e9)))
			fmt.Printf("Christmas time!\n------------\n")

		// checkDeers.SantaElves
		case <- deerCounter.CheckDeers:
			/* sleep for at most 2 seconds */
			time.Sleep(time.Duration(rand.Int63n(2*1e9)))

			fmt.Printf("* help elves *\n")
			for i:= 1; i<=3; i++ {
				elfCounter.helpElf <- Signal{}
			}// TODO: signal/help the 3 elves; correct?
			/* sleep for at most 2 seconds */
			time.Sleep(time.Duration(rand.Int63n(2*1e9)))
			fmt.Printf("Helped three elves!\n------------\n")

			//Timeout if no choice is made
			//case <-time.After(5*time.Second):
			//	fmt.Println("Santa timed out after 5s")
		}
		mutexSem.Signal <- "Santa"

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
		// Deer stays on vacation for 4 to 8 seconds.
		time.Sleep(time.Duration(rand.Int63n(4*1e9) + 4*1e9))

		/* sleep for at most 2 seconds */
		time.Sleep(time.Duration(rand.Int63n(2*1e9)))
		mutexSem.Wait <- "Reindeer " + strconv.Itoa(deerNo)

		deerCount.Return <- deerNo

		/* sleep for at most 2 seconds */
		time.Sleep(time.Duration(rand.Int63n(2*1e9)))
		mutexSem.Signal <- "Reindeer " + strconv.Itoa(deerNo)
		// TODO: getHitched? Not good! They go on vacation!
		// TODO: fine if not in loop... but if they are?
		deerSem.Wait <- "Reindeer " + strconv.Itoa(deerNo)
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
	//helpElf := make(chan Signal)

	for {
		// Elf works for 2 to 10 seconds before he has a problem. Yikes!
		time.Sleep(time.Duration(rand.Int63n(8*1e9) + 2*1e9))
		elfTex.Wait 		<- "Elf " + strconv.Itoa(elfId)
		mutexSem.Wait 		<- "Elf " + strconv.Itoa(elfId)
		//fmt.Printf("It is I, brutus.\n")
		// The elf that wants its problem solves communicates its ID
		elfCount.problem 	<- elfId						// Increments number of elves
		mutexSem.Signal 	<- "Elf " + strconv.Itoa(elfId)
		<- elfCount.helpElf 									// TODO: check helpElf
		fmt.Printf("elf %d has gotten help\n", elfId)
		mutexSem.Wait 		<- "Elf " + strconv.Itoa(elfId)		// TODO: issue, can't do this without helpElf
		fmt.Printf("VICTORY.\n")
		<- elfCount.solveProblem								// Elf was helped, solve your own problems!
		//fmt.Printf("I'M ACTUALLY A GOD, PROBLEM SOLVED\n")
		mutexSem.Signal 	<- "Elf " + strconv.Itoa(elfId)


	}
}

func main(){
	fmt.Println("> Starting program...")
	const numElves = 5
	// ------- SEMAPHORES -------
	ss := Sem{"santaSem", make(chan string), make(chan string)}
	go ss.Run()
	ms := Sem{"mutexSem", make(chan string), make(chan string)}
	go ms.Run()
	ds := Sem{"deerSem", make(chan string), make(chan string)}
	go ds.Run()
	es := Sem{"elfTex", make(chan string), make(chan string)}
	go es.Run()
	// Put santa to sleep
	ss.Wait <- "inital sleep"
	// Lock the warming hut (so deers don't go out prematurely)
	ds.Wait <- "hut locker"
	// ------- COUNTERS -------
	dc := DeerCounter{0, make(chan int),
		make(chan Signal), make(chan Signal)}
	go dc.Run(&ss)
	ec := ElfCounter{0, make(chan int), make(chan Signal), make(chan Signal) }
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
	time.Sleep(time.Duration(120*time.Second))
}