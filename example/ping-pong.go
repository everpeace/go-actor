package main

import (
	"fmt"

	"github.com/everpeace/go-actor"
)

func main() {
	latch := make(chan int)
	pong := func(msg actor.Message, self actor.Actor) {
		fmt.Printf("ponger received: %s\n", msg[1])
		msg[0].(actor.Actor).Send(actor.Message{"Pong"})
		self.Terminate()
	}
	ping := func(_ponger actor.Actor) func(msg actor.Message, self actor.Actor) {
		return func(msg actor.Message, self actor.Actor) {
			if msg[0] == "Pong" {
				fmt.Printf("pinger received: Pong.\nPing-Pong finished.\n")
				self.Terminate()
				latch <- 0
			} else {
				fmt.Printf("pinger sends: Ping\n")
				_ponger.Send(actor.Message{self, "Ping"})
			}
		}
	}

	ponger := actor.Spawn(pong)
	pinger := actor.Spawn(ping(ponger))
	pinger.Send(actor.Message{"_"})

	<-latch
}
