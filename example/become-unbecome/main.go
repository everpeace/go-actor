package main

import (
	"fmt"
	"strings"

	actor "github.com/everpeace/go-actor"
)

func main() {
	fmt.Println("==========================================================")
	fmt.Println("== Become/UnBecome example")
	fmt.Println("== An actor's behavior changes:")
	fmt.Println("== echo --(become)-->echo in upper case --(unbecome)--> echo.")

	// The latch is to ensure the completion of all examples' execution.
	latch := make(chan bool)

	actorSystem := actor.NewActorSystem("become-unbecome")
	terminate := func(msg actor.Message, context actor.ActorContext) {
		context.Self().Terminate()
		latch <- true
	}

	echoInUpper := func(msg actor.Message, context actor.ActorContext) {
		fmt.Printf("%s: %s.  unbecome\n", context.Self().Name(), strings.ToUpper(msg[0].(string)))
		context.Unbecome()
	}

	echo := func(msg actor.Message, context actor.ActorContext) {
		fmt.Printf("%s: %s.  ", context.Self().Name(), msg[0].(string))
		if msg[0] == "become terminate!!" {
			fmt.Println("become terminate")
			context.Become(terminate, false)
		} else {
			fmt.Println("become echoInUpper.")
			context.Become(echoInUpper, false)
		}
	}

	a := actorSystem.Spawn(echo)
	a.Send(actor.Message{"this should be echoed."})
	a.Send(actor.Message{"this should be echoed in upper case."})
	a.Send(actor.Message{"become terminate!!"})
	a.Send(actor.Message{})

	<-latch

	// Shutdown method shutdown all actors.
	actorSystem.Shutdown()
	fmt.Println("==========================================================")

}
