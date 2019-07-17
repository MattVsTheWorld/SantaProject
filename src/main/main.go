package main

import (
	"fmt"
	"time"
)

type Signal struct{}

type SantaSem struct{
	Wait chan Signal
	Signal chan Signal
}

func (self *SantaSem) Run() {
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
	for {
		curVal = self.deerNum
		if curVal == 9 {
			// TODO: signal Santa
			self.PrepareSleigh <- Signal{}
			curVal = 0	// reset counter

		} else if curVal < 9 {
			select {
				case self.Return <- Signal{}: 	// TODO: In realta' sara' una renna a dirlo
					curVal = curVal + 1
				case self.CheckDeers <- Signal{}:
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
		time.Sleep()
		fmt.Println("Santa is awake, choosing who to help...")
		select {
			//SantaDeers
			case deerCount.PrepareSleigh <- Signal{}:
				fmt.Printf("There are 9 deer, C H R I S T M A S T I M E\n")
			// checkDeers.SantaElves
			case deerCount.CheckDeers <- Signal{}:
				fmt.Printf("Deer are not 9; help elves\n")
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
	go santa(&ss, &ms, &dc)

	time.Sleep(time.Duration(5*time.Second))
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