package main

import (
	"fmt"

	"time"

	actor "github.com/everpeace/go-actor"
)

func main() {
	fmt.Println("==========================================================")
	fmt.Println("== Actor hierarchy example")
	fmt.Println("== Actor has children.  Children's termination is propagated")
	fmt.Println("== to parent.")


	system := actor.NewActorSystem("parent-child")
	monitor := system.SpawnWithName("monitor", func(msg actor.Message, context *actor.ActorContext) {
		if down, ok := msg[0].(actor.Down); ok {
			fmt.Printf("%s detects: %s %s\n", context.Self.Name, down.Actor.Name, down.Cause)
		}
	})

	parent := system.SpawnWithName("actorA", func(msg actor.Message, context *actor.ActorContext) {
		fmt.Printf("actorA receive: %s\n", msg)
	})
	parent.Monitor(monitor)

	child := parent.SpawnWithName("child", func(msg actor.Message, context *actor.ActorContext) {
		fmt.Printf("child receive: %s\n", msg)
	})
	child.Monitor(monitor)

	parent.Send(actor.Message{"hello"})
	child.Send(actor.Message{"hello"})

	<-time.After(time.Duration(1) * time.Second)
	fmt.Println("terminate parent. you will see children's termination too.")
	parent.Terminate()

	<-time.After(time.Duration(1) * time.Second)
	monitor.Terminate()
	system.WaitForAllActorsStopped()
	fmt.Println("==========================================================")
}
