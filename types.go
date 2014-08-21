package actor

// Message envelope for actor
// All message sent to actors should be wrapped with this envelope.
// example:
//   someActor.Send(actor.Message{"hello"})
type Message []interface{}

// PoisonPill terminates actors.
// Sending the PoisonPill message is nearly equivalent with Terminate() method.
// example:
//  someActor.Send(actor.Message{actor.PoisonPill{}})
type PoisonPill struct{}

// Receive is a type for Actor's message handler.
// It is just an alias for func(msg Message, context *ActorContext).
// For example, simple echo actor would be:
//   actorSystem.Spawn(func(msg Message, context *ActorContext){
//     fmt.Println(msg)
//   })
type Receive func(msg Message, context *ActorContext)

// Down is a message sent to Monitor
// If monitored actor was terminates, monitor will receive
//   Message{Down{
//     Cause: "terminated",
//     Actor: <pointer to the actor>
//   }}
// If monitored actor was killed, monitor will receive
//   Message{Down{
//     Cause: "killed",
//     Actor: <pointer to the actor>
//   }}
type Down struct {
	Cause string
	Actor *Actor
}
