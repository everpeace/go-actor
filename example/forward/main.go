package main

import (
	"fmt"
	"time"

	actor "github.com/everpeace/go-actor"
)

func main() {
	fmt.Println("==========================================================")
	fmt.Println("== Forwarding actor example")
	fmt.Println("== It is sometimes useful that an actor which just forward")
	fmt.Println("== messages other actors.  In this example, ")
	fmt.Println("== \"forward\" actor forwards to \"echo1\" and \"echo2\"")

	system := actor.NewActorSystem("forwad")
	echo := func() actor.Receive {
		return func(msg actor.Message, context *actor.ActorContext) {
			fmt.Printf("%s : %s\n", context.Self.Name, msg)
		}
	}

	echo1 := system.SpawnWithName("echo1", echo())
	echo2 := system.SpawnWithName("echo2", echo())
	forward := system.SpawnForwardActor("forward", echo1)
	forward.Add(echo2)

	<-time.After(time.Duration(1) * time.Second)
	fmt.Println("Sent [hello] to \"forward\"")
	forward.Send(actor.Message{"hello"})

	<-time.After(time.Duration(1) * time.Second)
	system.GracefulShutdown()
	fmt.Println("==========================================================")
}
