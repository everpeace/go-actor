package main

import (
	"fmt"
	"strings"

	"github.com/everpeace/go-actor"
)

func main() {
	latch := make(chan int)
	// Become/Unbecome
	echo := func(msg actor.Message, self actor.Actor) {
		fmt.Println(msg[0])
	}
	echoInUpper := func(msg actor.Message, self actor.Actor) {
		fmt.Println(strings.ToUpper(msg[0].(string)))
	}
	terminate := func(msg actor.Message, self actor.Actor) {
		self.Terminate()
		latch <- 0
	}

	a := actor.Spawn(echo)
	a.Send(actor.Message{"this should be echoed."})
	a.Become(echoInUpper, false)
	a.Send(actor.Message{"this should be echoed in upper case."})
	a.Unbecome()
	a.Send(actor.Message{"this should be echoed."})
	a.Become(terminate, true)
	a.Send(actor.Message{})

	<-latch
}
