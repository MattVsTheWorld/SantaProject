package toDelete

import (
	"fmt"
	"math"
	"math/rand"
	"time")

// -------------------// -------------------// -------------------
// ------------------- Channels/Concurrency // -------------------
// -------------------// -------------------// -------------------

// empty structs for signaling on channels
type Signal struct{}
var sigVal = struct{}{}

// Number of philosophers
const NPhil = 5

// Start value of the counters
const MaxToken = 5

type Fork struct {
	name string // fork name

	// (public) channels	// TODO: ccould make it private? see slides
	Take chan Signal
	Leave chan Signal
}

func (self *Fork) Run(){
	for{
		<-self.Take
		fmt.Printf("Fork %s taken\n", self.name)
		<-self.Leave
		fmt.Printf("Fork %s released\n", self.name)
	}
}

type Counter struct{
	top int		// top value

	// channels
	Dec chan Signal		// use a token
	Res chan Signal		// restore
}

func (self *Counter) Run () {
	var curVal int
	for {
		curVal = self.top
		for curVal > 0 {
			<- self.Dec
			curVal = curVal -1
		}
		<- self.Res
	}
}

func Resetter (counters []Counter) {
	for{
		for _, counter := range counters {
			counter.Res <- sigVal
		}
	}
}

func Phil (id int, left *Fork, right *Fork, counter *Counter){
	// Number of tokens disregarded consecutively
	disregarded := 0

	for {
		counter.Dec <- sigVal

		if (disregarded < MaxToken) && (rand.Intn(disregarded+1) == 0){
			disregarded++
			fmt.Printf("Philosopher %d, token disregarded ... %d times\n", id, disregarded)
		} else {
			fmt.Printf("Phil %d is thinking\n", id)
			// SLEEP 2 secs at most
			time.Sleep(time.Duration(rand.Int63n(2*1e9)))
			// takes left fork
			left.Take <- sigVal
			time.Sleep(time.Duration(rand.Int63n(2*1e9)))
			// takes right fork
			right.Take <- sigVal
			fmt.Printf("MO MAGNO %d\n", id)
			left.Leave <- sigVal
			right.Leave <- sigVal

			// reset disregard
			disregarded = 0
		}
	}

}
// -------------------// -------------------// -------------------
// ------------------- Structs / Interfaces // -------------------
// -------------------// -------------------// -------------------

type Santa struct{
	// values
	X, Y float64
	// can use functions as first class values
	ThreePointTwo func()float64
}

// functions
func (self *Santa) Abs() float64 {
	return math.Sqrt(self.X*self.X + self.Y*self.Y)
}

// "Personalised" New
func New (x, y float64) *Santa {
	// 1
	//var s *Santa = new(Santa)	// allocate memory for Santa, return pointer
	//s.X, s.Y = x, y
	//return s
	// 2
	return &Santa{x,
		y,
		func()float64{ return 3.2 }}
}

// Interfaces
// They are satisfied/implemented implicitly
// A type can satisfy multiple interfaces
type Abser interface {
	Abs() float64
}

func DoSomething (a Abser){
	fmt.Println(a.Abs())
}	// structs that implement Abs may use DoSomething