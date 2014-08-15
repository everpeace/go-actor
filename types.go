package actor

// Exported Types.
type Message []interface{} // Message is just slice (instead of tuple.)

// materialized event messaged sent to Monitor
type ActorEvent int
const (
	_ ActorEvent = iota
	KILLED
	TERMINATED
	STARTED
)
func (e ActorEvent) String() string {
	switch e {
	case KILLED:
		return "KILLED"
	case TERMINATED:
		return "TERMINATED"
	case STARTED:
		return "STARTED"
	}
	return ""
}

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
	Monitor(actor Actor)
	Demonitor(actor Actor)
	Name() string
	// TODOs
	// Restart()
}

type ForwardingActor interface {
	Actor
	Add(recipient Actor)
	Remove(recipient Actor)
}
