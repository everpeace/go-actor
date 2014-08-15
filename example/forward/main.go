package main

import (
	"fmt"

	actor "github.com/everpeace/go-actor"
)

func main() {
	latch := make(chan int)

	echo := func(name string) actor.Receive {
		return func(msg actor.Message, context actor.ActorContext) {
			fmt.Printf("%s : %s\n", name, msg)
			latch <- 0
		}
	}

	echo1 := actor.Spawn(echo("echo1"))
	echo2 := actor.Spawn(echo("echo2"))
	forward := actor.SpawnForwardActor(echo1, echo2)

	forward.Send(actor.Message{"hello"})
	<-latch
	<-latch
}
