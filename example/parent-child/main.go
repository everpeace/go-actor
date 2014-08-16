package main

import (
	"fmt"

	actor "github.com/everpeace/go-actor"
	"time"
)

func main() {
	exitLatch := make(chan bool, 2)
	exitGardian := actor.Spawn(func(msg actor.Message, context actor.ActorContext){
		if down, ok := msg[0].(actor.Down); ok {
			fmt.Printf("%s %s\n", down.Actor.Name(), down.Cause)
			exitLatch <- true
		}
	})

	parent := actor.SpawnWithName("parent", func(msg actor.Message, context actor.ActorContext){
			fmt.Printf("parent receive: %s\n", msg)
		})
	parent.Monitor(exitGardian)

	child := parent.Spawn(func(msg actor.Message, context actor.ActorContext){
		fmt.Printf("child receive: %s\n", msg)
	})
	child.Monitor(exitGardian)

	parent.Send(actor.Message{"hello"})
	child.Send(actor.Message{"hello"})

	<- time.After(time.Duration(1)*time.Second)
	child.Terminate()

	<-exitLatch
	<-exitLatch
}
