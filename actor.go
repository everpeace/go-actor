package actor

import (
	"fmt"
	"os"

	"github.com/dropbox/godropbox/container/set"
)

// Actor.
//
// An actor has its name and belongs to exactly one actor system.
type Actor struct {
	Name    string
	System  *ActorSystem
	context *ActorContext
}

// Send sends message to the actor asynchronously.
//
// Please note that message should be wrapped in actor.Message.
// example:
//   actor.Send(Message{"hello"})
func (actor *Actor) Send(msg Message) {
	go func() {
		defer logPanic(actor)
		actor.context.mailbox <- msg
	}()
}

// Terminate sends "Terminate" signal to the actor asynchronously.
//
// "Terminate" signal stop the actor in graceful manner.
// This method will just post "Message{PoisonPill{}}" to their mailbox.
// Thus, the actor will stop after processing remained messages in their mailbox.
func (actor *Actor) Terminate() {
	actor.System.running.Remove(actor)
	actor.System.stopped.Add(actor)
	go func() {
		defer logPanic(actor)
		actor.context.terminate()
	}()
}

// Kill sends "Kill" signal to the actor asynchronously.
//
// If the actor receives the signal, the actor stops immediately.
// However, the timing of receipt of the signal depends on goroutine scheduler.  Therefore, the number of messages will be processed before its stop is undefined.
func (actor *Actor) Kill() {
	actor.System.running.Remove(actor)
	actor.System.stopped.Add(actor)
	go func() {
		defer logPanic(actor)
		actor.context.kill()
	}()
}

// Monitor attaches another actor(mon) as its monitor asynchronously.
//
// Attached monitors will be notified its stop (terminate and kill) event with actor.Down message.
func (actor *Actor) Monitor(mon *Actor) {
	go func() {
		defer logPanic(actor)
		actor.context.attachMonitor(mon)
	}()
}

// Demonitor detaches a given monitor asynchronously.
//
// Detached monitors will not be notified its stop (terminate and kill) event anymore.
func (actor *Actor) Demonitor(mon *Actor) {
	go func() {
		defer logPanic(actor)
		actor.context.detachMonitor(mon)
	}()
}

func (actor *Actor) IsRunning() bool {
	return actor.System.running.Contains(actor)
}

// Spawn creates and starts a child actor of the actor.
//
// This takes Receive(an type alias for actor's message handler) returns the pointer to started actor.
// The difference between top level actors which are created directly from actor system is that created child actor will be terminated or killed when the actor is terminated or killed.
//
// For example:
//  child := actor.Spawn(func(msg Message, context *ActorContext){
//     fmt.Println(msg)
//  })
//  actor.Terminates()
//  // then child will also terminate.
func (actor *Actor) Spawn(receive Receive) *Actor {
	system := actor.System
	latch, child := system.spawnActor(actor.newActor(fmt.Sprint(actor.context.Children.Len()), receive))
	latch <- true
	return child
}

// SpawnWithName is the same as Spawn except that you can name it.
func (actor *Actor) SpawnWithName(name string, receive Receive) *Actor {
	system := actor.System
	latch, child := system.spawnActor(actor.newActor(name, receive))
	latch <- true
	return child
}

func (actor *Actor) SpawnForwardActor(name string, actors ...*Actor) *ForwardingActor {
	s := set.NewSet()
	for _, actor := range actors {
		s.Add(actor)
	}
	forwardActor := &ForwardingActor{
		addRecipientChan: make(chan addRecipient),
		delRecipientChan: make(chan removeRecipient),
		recipients: s,
	}
	forwardActor.Actor = actor.newActor(name, forwardActor.receive())
	forwardActor.context.prePrecessHook = forwardActor.preProcessHook()
	start := forwardActor.context.start()
	start <- true
	return forwardActor
}

func (actor *Actor) newActor(name string, receive Receive) *Actor {
	child := &Actor{Name: actor.canonicalName(name), System: actor.System }
	child.context = newActorContext(actor.context, child, receive)
	return child
}

func (actor *Actor) canonicalName(name string) string {
	return actor.Name + "/" + name
}

func logPanic(actor *Actor) {
	if r := recover(); r != nil {
		fmt.Fprintf(os.Stderr, "[%s] %s", actor.Name, r)
	}
}
