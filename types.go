package actor

// Exported Types.
type Message []interface{} // Message is just slice (instead of tuple.)

type Receive func(msg Message, context ActorContext)

type ActorContext interface {
	Self() Actor
	Become(behavior Receive, discardOld bool)
	Unbecome()
}

type Actor interface {
	Send(msg Message)
	Terminate()
	Kill()
	// TODOs
	// Restart()
	// Monitor(actor *Actor)
	// Demonitor(actor *Actor)
}

type ForwardingActor interface {
	Actor
	Add(recipient Actor)
	Remove(recipient Actor)
}
