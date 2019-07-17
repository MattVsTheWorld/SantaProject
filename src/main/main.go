package main

import (
	"fmt"
	"math/rand"
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
		<- self.Wait
		fmt.Println("Mutex semaphore put on wait")
		<- self.Signal
		fmt.Printf("Mutex sempahore signaled open")
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

func santa (santSem *SantaSem, mutexSem *Sem){
	for {
		santSem.Wait <- Signal{}
		mutexSem.Wait <- Signal{}
	}
}


func main(){
	fmt.Println("Start")
	ss := SantaSem{ make(chan Signal), make(chan Signal)}
	go ss.Run()
	ms := Sem{ make(chan Signal), make(chan Signal)}
	go ms.Run()
	go santa(&ss, &ms)

	time.Sleep(time.Duration(rand.Int63n(2*1e9)))
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