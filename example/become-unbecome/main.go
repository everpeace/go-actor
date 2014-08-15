package main

import (
	"fmt"
	"strings"

	actor "github.com/everpeace/go-actor"
)

func main() {
	latch := make(chan int)
	// Become/Unbecome

	terminate := func(msg actor.Message, context actor.ActorContext) {
		context.Self().Terminate()
		latch <- 0
	}

	echoInUpper := func(msg actor.Message, context actor.ActorContext) {
		fmt.Println(strings.ToUpper(msg[0].(string)))
		context.Unbecome()
	}

	echo := func(msg actor.Message, context actor.ActorContext) {
		fmt.Println(msg[0])
		if msg[0] == "go terminate" {
			context.Become(terminate, false)
		} else {
			context.Become(echoInUpper, false)
		}
	}

	a := actor.Spawn(echo)
	a.Send(actor.Message{"this should be echoed."})
	a.Send(actor.Message{"this should be echoed in upper case."})
	a.Send(actor.Message{"this should be echoed."})
	a.Send(actor.Message{"this should be echoed in upper case."})
	a.Send(actor.Message{"go terminate"})
	a.Send(actor.Message{})

	<-latch
}
