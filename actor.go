package actor

import (
	"fmt"
	"os"

	"github.com/dropbox/godropbox/container/set"
)

type actorImpl struct {
	name    string
	context *actorContext
}

func (actor *actorImpl) Send(msg Message) {
	go func() {
		defer logPanic(actor)
		actor.context.mailbox <- msg
	}()
}

func (actor *actorImpl) Terminate() {
	actor.context.system.running.Remove(actor)
	go func() {
		defer logPanic(actor)
		actor.context.terminate()
	}()
}

func (actor *actorImpl) Kill() {
	actor.context.system.running.Remove(actor)
	go func() {
		defer logPanic(actor)
		actor.context.kill()
	}()
}

func (actor *actorImpl) Shutdown() {
	actor.context.system.running.Remove(actor)
	go func() {
		defer logPanic(actor)
		actor.context.shutdown()
	}()
}

func (actor *actorImpl) Monitor(mon Actor) {
	go func() {
		defer logPanic(actor)
		actor.context.attachMonitor(mon)
	}()
}

func (actor *actorImpl) Demonitor(mon Actor) {
	go func() {
		defer logPanic(actor)
		actor.context.detachMonitor(mon)
	}()
}

func (actor *actorImpl) Spawn(receive Receive) Actor{
	system := actor.context.system
	latch, child := system.spawnActor(actor.newActorImpl(fmt.Sprint(actor.context.children.Len()), receive))
	latch <- true
	return child
}

func (actor *actorImpl) SpawnWithName(name string, receive Receive) Actor{
	system := actor.context.system
	latch, child := system.spawnActor(actor.newActorImpl(name, receive))
	latch <- true
	return child
}

func (actor *actorImpl) SpawnWithLatch(receive Receive) (chan bool, Actor){
	system := actor.context.system
	latch, child := system.spawnActor(actor.newActorImpl(fmt.Sprint(actor.context.children.Len()), receive))
	return latch, child
}

func (actor *actorImpl) SpawnWithNameAndLatch(name string, receive Receive) (chan bool, Actor){
	system := actor.context.system
	latch, child := system.spawnActor(actor.newActorImpl(name, receive))
	return latch, child
}

func (actor *actorImpl) SpawnForwardActor(name string, actors...Actor) ForwardingActor {
	s := set.NewSet()
	for _, actor := range actors {
		s.Add(actor)
	}
	forwardActor := &forwardingActor{
		actor.newActorImpl(name, forward(s)),
	}
	start := forwardActor.context.start()
	start <- true
	return forwardActor
}

func (actor *actorImpl) newActorImpl(name string, receive Receive) *actorImpl {
	child := &actorImpl{ name: actor.canonicalName(name) }
	child.context = newActorContext(actor.context, child, receive)
	return child
}

func (actor *actorImpl) Name() string{
	return actor.name
}

func (actor *actorImpl) canonicalName(name string) string{
	return actor.Name()+"/"+name
}

func logPanic(actor *actorImpl) {
	if r := recover(); r != nil {
		fmt.Fprintf(os.Stderr, "[%s] %s", actor.Name(), r)
	}
}
