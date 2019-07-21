package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

/*
ch <- v    // Send v to channel ch.		(~ 'output on channel)
v := <-ch  // Receive from ch, and		(~   input on channel)
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
	// Basic semaphore structure
	// First attempt communication on the the "Wait" channel
	// Then, when the user is done, communicate on the "Signal" channel and repeat
	fmt.Printf("> %s semaphore goroutine started\n", self.name)
	for {
		<- self.Wait
		<- self.Signal
		// "Verbose" version
		/*
		whoWait := <- self.Wait
		fmt.Printf("> Wait called: someone accessed %s (%s) ...\n", self.name, whoWait)
		whoSignal := <- self.Signal
		fmt.Printf("> Signal called: %s released (%s)...\n-------------\n", self.name, whoSignal)
		*/
	}
}

// ElfCounter struct and behaviour
type ElfCounter struct{
	elfNum			int					// Tracks the initial number of troubled elves
	problem 		chan int			// Elf communicates he has a problem
	solveProblem	chan Signal			// TODO: check
	helpElf			chan Signal			// Expose a channel that Santa will use to help the elves
}

func (self *ElfCounter) Run(elfTex *Sem, santaSem *Sem){
	// Structure that keeps track of the number of elves that have problems
	// Uses a semaphore (elfTex) to do the following
	//	- If <3 elves have problems, allow more ask for help
	//	- If ==3 elves have problems, signal Santa's semaphore (santaSem)
	//	- If the elves are being helped, signal elfTex only when they are all done
	var curVal int
	fmt.Printf("> elfCounter goroutine started\n")
	curVal = self.elfNum
	for {
		select {
			// An elf is signaling a problem
			case elfId := <- self.problem:
				fmt.Printf("⛄ Elf %d has a problem!\n", elfId)
				curVal++
				// Three elves have problems; wake up Santa
				// Don't allow any more elves in until the current 3 have been helped
				if curVal == 3 {
					fmt.Printf("Three elves have problems! Waking up santa...\n")
					santaSem.Signal <- "elfCounter"
				} else {
				// Less than 3 elves have asked for help; more are allowed in before Santa is woken up
					elfTex.Signal <- "elfCounter"
				}
			case self.solveProblem <- Signal{}:
				curVal--
				// The last elf having its problem solved releases ElfTex
				// This allows more troubled elves to ask for help
				if curVal == 0 {
					elfTex.Signal <- "elfCounter"
				}
			}
		}
}

// DeerCounter struct and behaviour
type DeerCounter struct{
	deerNum 		int				// Tracks the initial number of returned reindeers
	Return 			chan int		// Reindeer communicates they have returned
	CheckDeers 		chan Signal		// This channel allows Santa to ask if all deers have returned
	PrepareSleigh 	chan Signal		// Dual counterpart, the reindeer let Santa know that the sleigh is ready to be prepared
}

func (self *DeerCounter) Run(santaSem *Sem) {
	// Structure that keeps track of the number of reindeers that have returned
	// Wakes up Santa if all 9 reindeer have returned
	// Else offer to either:
	// 	- receive a returning deer or
	// 	- communicate to santa that they are not 9
	var curVal int
	fmt.Println("> deerCounter goroutine started")
	curVal = self.deerNum
	for {
		if curVal == 9 {
			fmt.Printf("All the reindeer have returned; waking up santa...\n")
			// All reindeers have returned; wake up (signal) Santa
			santaSem.Signal <- "deerCounter"
			// Ask Santa to prepare the sleigh
			<- self.PrepareSleigh
			// Reset counter
			curVal = 0

		} else if curVal < 9 {
			select {
				// Allow reindeers to return (receive)
				case i:= <- self.Return:
					fmt.Printf("♞ Reindeer n°%d has returned!\n",i)
					curVal = curVal + 1
				// Offer to communicate to Santa that there are still reindeers missing
				case self.CheckDeers <- Signal{}:
					fmt.Printf("> Reindeer are still on vacation: Santa can help elves\n")
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
	// Santa process
	// Sleeps until woken up by 3 troubled elves or 9 reindeer back from vacation
	for {
		// Santa sleeps until the Wait channel is available
		santaSem.Wait <- "Santa"
		// Close the shared mutex
		mutexSem.Wait <- "Santa"
		fmt.Println("🎅 Santa is awake, choosing who to help...")
		select {
			// Offer to prepare the sleigh if all deers have returned
			case deerCounter.PrepareSleigh <- Signal{}:
				/* sleep for at most 2 seconds */
				time.Sleep(time.Duration(rand.Int63n(2*1e9)))
				/* --------------------------- */
				fmt.Printf("*** Prepare Sleigh ***\n")
				// Signal deers
				for i:=1; i<=9; i++ {
					deerSem.Signal <- "Santa"
				}
				/* sleep for at most 2 seconds */
				time.Sleep(time.Duration(rand.Int63n(2*1e9)))
				/* --------------------------- */
				fmt.Printf("✨✨✨ Reindeers hitched; Christmas time! ✨✨✨\n")
				// It takes a whole 2 seconds to go around the globe!
				time.Sleep(2*time.Second)

			// Ask to be told whether all reindeer have returned
			// A message on this channel implies they have not, and elves must be helped
			case <- deerCounter.CheckDeers:
				/* sleep for at most 2 seconds */
				time.Sleep(time.Duration(rand.Int63n(2*1e9)))
				/* --------------------------- */
				fmt.Printf("*** Help Elves ***\n")
				for i:= 1; i<=3; i++ {
					elfCounter.helpElf <- Signal{}
				}
				/* sleep for at most 2 seconds */
				time.Sleep(time.Duration(rand.Int63n(2*1e9)))
				/* --------------------------- */
				fmt.Printf("🎁🎁🎁 Helped three elves! 🎁🎁🎁\n")

		}
		// Release the shared mutex
		mutexSem.Signal <- "Santa"
		fmt.Printf("Santa is done; back to sleep...\n------------\n")
		// Could add a minimum sleep time for poor Santa...
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
	// Reindeer process
	// Goes on vacation for 4 to 10 seconds, then comes back to prepare for Christmas
	for {
		fmt.Printf("Reindeer %d is going on vacation... Will be back eventually!\n", deerNo)
		// Deer stays on vacation for 4 to 10 seconds.
		time.Sleep(time.Duration(rand.Int63n(4*1e9) + 6*1e9))
		// Lock the shared mutex
		mutexSem.Wait <- "Reindeer " + strconv.Itoa(deerNo)

		// The reindeer sends a message to the counter, telling it it has returned
		deerCount.Return <- deerNo

		// Release the shared mutex
		mutexSem.Signal <- "Reindeer " + strconv.Itoa(deerNo)
		// Reindeer waits until Santa is ready to hitch them
		deerSem.Wait 	<- "Reindeer " + strconv.Itoa(deerNo)
		fmt.Printf("Reindeer %d is being hitched...\n", deerNo)
		// It takes a whole 2 seconds to go around the globe!
		time.Sleep(2*time.Second)
	}
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
		// Elf works for 4 to 10 seconds before he has a problem. Yikes!
		time.Sleep(time.Duration(rand.Int63n(6*1e9) + 4*1e9))
		elfTex.Wait 		<- "Elf " + strconv.Itoa(elfId)
		mutexSem.Wait 		<- "Elf " + strconv.Itoa(elfId)
		// The elf that wants its problem solves communicates its ID
		elfCount.problem 	<- elfId								// Increments number of elves
		mutexSem.Signal 	<- "Elf " + strconv.Itoa(elfId)
		<- elfCount.helpElf
		fmt.Printf("Elf %d has gotten help\n", elfId)
		mutexSem.Wait 		<- "Elf " + strconv.Itoa(elfId)
		<- elfCount.solveProblem									// Elf was helped, solve your own problems!	// TODO: check
		mutexSem.Signal 	<- "Elf " + strconv.Itoa(elfId)
		fmt.Printf("Elf %d is happy and goes back to work ...\n", elfId)

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
	ss.Wait <- "initial sleep"
	// Lock the warming hut (so reindeer don't go out prematurely)
	ds.Wait <- "hut locker"
	// (These Wait messages simulate the semaphore starting from 0)
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
	// Run program for at least two minutes
	time.Sleep(time.Duration(2*time.Minute))
}