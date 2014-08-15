package main

import (
	"fmt"

	actor "github.com/everpeace/go-actor"
)
func main() {
	exitLatch := make(chan bool)
	echo := func(exit chan bool) actor.Receive {
		return func(msg actor.Message, context actor.ActorContext) {
			fmt.Printf("%s receive: %s\n", context.Self().Name(), msg[0])
			exit <- true
		}
	}

	mon := actor.SpawnWithName("monitor", echo(exitLatch))
	targetStart, target := actor.SpawnWithNameAndLatch("target", echo(exitLatch))

	target.Monitor(mon)

	target.Send(actor.Message{"hello"})
	targetStart <- true

	target.Terminate()
	mon.Terminate()
	<-exitLatch
	<-exitLatch
}
