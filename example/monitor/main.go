package main

import (
	"fmt"
	"time"

	actor "github.com/everpeace/go-actor"
)

func main() {
	fmt.Println("==========================================================")
	fmt.Println("== Monitor example")
	fmt.Println("== Monitor can detect actor's termination.")
	fmt.Println("== In this example, spawn \"traget\" actor and it is monitored")
	fmt.Println("== by \"monitor\" ")

	// The latch is to ensure the completion of all examples' execution.
	latch := make(chan bool)

	system := actor.NewActorSystem("monitor-system")
	echo := func() actor.Receive {
		return func(msg actor.Message, context *actor.ActorContext) {
			if m, ok := msg[0].(actor.Down); ok {
				fmt.Printf("%s detects: %s %s\n", context.Self.Name, m.Actor.Name, m.Cause)
				latch <- true
			} else {
				fmt.Printf("%s receive: %s\n", context.Self.Name, msg)
			}
		}
	}

	monitor := system.SpawnWithName("monitor", echo())
	target := system.SpawnWithName("target", echo())
	target.Monitor(monitor)

	<-time.After(time.Duration(1) * time.Second)

	target.Send(actor.Message{"hello"})
	target.Terminate()

	<-latch

	// Shutdown method shutdown all actors.
	system.Shutdown()
	fmt.Println("==========================================================")
}
