package actor

// Message envelope for actor
// All message sent to actors should be wrapped with this envelope.
// for example:
//   import actor "github.com/everpeace/go-actor"
//   someActor.Send(actor.Message{"hello"})
type Message []interface{}

// Actor's message handler type.
type Receive func(msg Message, context *ActorContext)

// Down message sent to Monitor
type Down struct {
	Cause string
	Actor *Actor
}
