package actor

// Message envelope for actor
// All message sent to actors should be wrapped with this envelope.
// example:
//   someActor.Send(actor.Message{"hello"})
type Message []interface{}

// This pill terminates actors
// example:
//  someActor.Send(actor.Message{actor.PoisonPill{}}
type PoisonPill struct{}

// Actor's message handler type.
type Receive func(msg Message, context *ActorContext)

// Down message sent to Monitor
type Down struct {
	Cause string
	Actor *Actor
}
