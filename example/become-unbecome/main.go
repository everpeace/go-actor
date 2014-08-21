package main

import (
	"fmt"
	"strings"

	actor "github.com/everpeace/go-actor"
	"time"
)

func main() {
	fmt.Println("==========================================================")
	fmt.Println("== Become/UnBecome example")
	fmt.Println("== An actor's behavior changes:")
	fmt.Println("== echo --(become)-->echo in upper case --(unbecome)--> echo.")

	system := actor.NewActorSystem("become-unbecome")
	terminate := func(msg actor.Message, context *actor.ActorContext) {
		context.Self.Terminate()
	}

	echoInUpper := func(msg actor.Message, context *actor.ActorContext) {
		fmt.Printf("%s: %s.  unbecome\n", context.Self.Name, strings.ToUpper(msg[0].(string)))
		context.Unbecome()
	}

	echo := func(msg actor.Message, context *actor.ActorContext) {
		fmt.Printf("%s: %s.  ", context.Self.Name, msg[0].(string))
		if msg[0] == "become terminate!!" {
			fmt.Println("become terminate")
			context.Become(terminate, false)
		} else {
			fmt.Println("become echoInUpper.")
			context.Become(echoInUpper, false)
		}
	}

	a := system.Spawn(echo)
	a.Send(actor.Message{"this should be echoed."})
	a.Send(actor.Message{"this should be echoed in upper case."})
	a.Send(actor.Message{"become terminate!!"})
	a.Send(actor.Message{})

	// Shutdown method shutdown all actors.
	system.GracefulShutdownIn(time.Duration(1) * time.Second)
	fmt.Println("==========================================================")

}
