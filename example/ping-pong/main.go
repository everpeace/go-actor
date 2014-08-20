package main

import (
	"fmt"
	"time"

	actor "github.com/everpeace/go-actor"
)

func main() {
	fmt.Println("==========================================================")
	fmt.Println("== Ping-Pong example.")

	system := actor.NewActorSystem("ping-pong")
	pong := func(msg actor.Message, context *actor.ActorContext) {
		fmt.Printf("%s received: %s\n", context.Self.Name, msg[1])
		fmt.Printf("%s sends : Pong\n", context.Self.Name)
		msg[0].(*actor.Actor).Send(actor.Message{"Pong"})
	}
	ping := func(_ponger *actor.Actor) actor.Receive {
		return func(msg actor.Message, context *actor.ActorContext) {
			if msg[0] == "Pong" {
				fmt.Printf("%s receives: Pong.\nPing-Pong finished.\n", context.Self.Name)
			} else {
				fmt.Printf("%s sends : Ping\n", context.Self.Name)
				_ponger.Send(actor.Message{context.Self, "Ping"})
			}
		}
	}

	ponger := system.SpawnWithName("ponger", pong)
	pinger := system.SpawnWithName("pinger", ping(ponger))
	pinger.Send(actor.Message{"_"})

	<-time.After(time.Duration(1)*time.Second)

	system.GracefulShutdown()
	fmt.Println("==========================================================")
}
