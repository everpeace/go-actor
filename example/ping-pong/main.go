package main

import (
	"fmt"

	"github.com/everpeace/go-actor"
)

func main() {
	latch := make(chan int)
	pong := func(msg actor.Message, context actor.ActorContext) {
		fmt.Printf("ponger received: %s\n", msg[1])
		msg[0].(actor.Actor).Send(actor.Message{"Pong"})
		context.Self().Terminate()
	}
	ping := func(_ponger actor.Actor) actor.Receive {
		return func(msg actor.Message, context actor.ActorContext) {
			if msg[0] == "Pong" {
				fmt.Printf("pinger received: Pong.\nPing-Pong finished.\n")
				context.Self().Terminate()
				latch <- 0
			} else {
				fmt.Printf("pinger sends: Ping\n")
				_ponger.Send(actor.Message{context.Self(), "Ping"})
			}
		}
	}

	ponger := actor.Spawn(pong)
	pinger := actor.Spawn(ping(ponger))
	pinger.Send(actor.Message{"_"})

	<-latch
}
