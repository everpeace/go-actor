package actor

// Exported Types.
type Message []interface{} // Message is just slice (instead of tuple.)

type Receive func(msg Message, context ActorContext)

type ActorSystem interface {
	actorFactory
	Name()
	WaitForAllActorsStopped()
	Shutdown()
}

type Actor interface {
	actorFactory
	Name() string
	Send(msg Message)
	Terminate()
	Kill()
	Shutdown()
	Monitor(actor Actor)
	Demonitor(actor Actor)
	// Restart()
}

type ForwardingActor interface {
	Actor
	Add(recipient Actor)
	Remove(recipient Actor)
}

type ActorContext interface {
	Self() Actor
	Become(behavior Receive, discardOld bool)
	Unbecome()
}

// Down message sent to Monitor
type Down struct {
	Cause string
	Actor Actor
}

type actorFactory interface {
	Spawn(receive Receive) Actor
	SpawnWithName(name string, receive Receive) Actor
	SpawnWithLatch(receive Receive) (chan bool, Actor)
	SpawnWithNameAndLatch(name string, receive Receive) (chan bool, Actor)
	SpawnForwardActor(name string, actors ...Actor) ForwardingActor
}
