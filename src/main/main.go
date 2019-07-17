package main

import (
	"fmt"
	"time"
)

/*
ch <- v    // Send v to channel ch.		('output su canale)
v := <-ch  // Receive from ch, and		(input su canale)
           // assign value to v.
(The data flows in the direction of the arrow.)
*/

type Signal struct{}

type SantaSem struct{
	Wait chan Signal
	Signal chan Signal
}

func (self *SantaSem) Run() {
	fmt.Println("Santa semaphore goroutine started")
	for {
		<- self.Wait
		fmt.Println("Santa semaphore put on wait")
		<- self.Signal
		fmt.Printf("Santa sempahore signaled open")
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
		fmt.Println("Mutex semaphore put on wait")
		<- self.Signal
		fmt.Println("Mutex sempahore signaled open")
	}
}

type DeerCounter struct{
	deerNum 		int
	Return 			chan Signal
	Reset 			chan Signal
	CheckDeers 		chan Signal
	PrepareSleigh 	chan Signal
	//santaSignal chan Signal
}

func (self *DeerCounter) Run() {
	var curVal int
	fmt.Println("DeerCounter goroutine started")
	for {
		curVal = self.deerNum
		if curVal == 9 {
			// TODO: signal Santa
			// fmt.Printf("I'm in Pog\n")
			<- self.PrepareSleigh
			curVal = 0	// reset counter

		} else if curVal < 9 {
			select {
				case <- self.Return: 		// TODO: check, ma ha senso che lo riceva
					curVal = curVal + 1
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

func santa (santSem *SantaSem, mutexSem *Sem, deerCount *DeerCounter){
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
				fmt.Printf("There are 9 deer, C H R I S T M A S T I M E\n")
				// TODO: signal the 9 deers
				mutexSem.Signal <- Signal{}
			// checkDeers.SantaElves
			case <- deerCount.CheckDeers:
				fmt.Printf("Deer are not 9; help elves\n")
				// TODO: signal/help the 3 elves
				mutexSem.Signal <- Signal{}
			//Timeout if no choice is made
			//case <-time.After(5*time.Second):
			//	fmt.Println("Santa timed out after 5s")
		}

	}
}


func main(){
	fmt.Println("Start")
	ss := SantaSem{ make(chan Signal), make(chan Signal)}
	go ss.Run()
	ms := Sem{ make(chan Signal), make(chan Signal)}
	go ms.Run()
	dc := DeerCounter{9, make(chan Signal), make(chan Signal),
	                make(chan Signal), make(chan Signal)}
	go dc.Run()
	//	fmt.Printf("Bro?")
	go santa(&ss, &ms, &dc)

	// All goroutines are killed when main ends
	//for {
	time.Sleep(time.Duration(10*time.Second))
	//}


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