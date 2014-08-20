package actor

import (
	"fmt"
	"os"

	"github.com/dropbox/godropbox/container/set"
)

type Actor struct {
	Name    string
	System  *ActorSystem
	context *ActorContext
}

func (actor *Actor) Send(msg Message) {
	go func() {
		defer logPanic(actor)
		actor.context.mailbox <- msg
	}()
}

// Terminates the actor in graceful manner
// That means actor will stop after processing all messages in their mailbox.
func (actor *Actor) Terminate() {
	actor.System.running.Remove(actor)
	actor.System.stopped.Add(actor)
	go func() {
		defer logPanic(actor)
		actor.context.terminate()
	}()
}

// Kills the actor immediately.
func (actor *Actor) Kill() {
	actor.System.running.Remove(actor)
	actor.System.stopped.Add(actor)
	go func() {
		defer logPanic(actor)
		actor.context.kill()
	}()
}

func (actor *Actor) Monitor(mon *Actor) {
	go func() {
		defer logPanic(actor)
		actor.context.attachMonitor(mon)
	}()
}

func (actor *Actor) Demonitor(mon *Actor) {
	go func() {
		defer logPanic(actor)
		actor.context.detachMonitor(mon)
	}()
}

func (actor *Actor) isRunning() bool {
	return actor.System.running.Contains(actor)
}

func (actor *Actor) Spawn(receive Receive) *Actor {
	system := actor.System
	latch, child := system.spawnActor(actor.newActor(fmt.Sprint(actor.context.Children.Len()), receive))
	latch <- true
	return child
}

func (actor *Actor) SpawnWithName(name string, receive Receive) *Actor {
	system := actor.System
	latch, child := system.spawnActor(actor.newActor(name, receive))
	latch <- true
	return child
}

func (actor *Actor) SpawnWithLatch(receive Receive) (chan bool, *Actor) {
	system := actor.System
	latch, child := system.spawnActor(actor.newActor(fmt.Sprint(actor.context.Children.Len()), receive))
	return latch, child
}

func (actor *Actor) SpawnWithNameAndLatch(name string, receive Receive) (chan bool, *Actor) {
	system := actor.System
	latch, child := system.spawnActor(actor.newActor(name, receive))
	return latch, child
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
