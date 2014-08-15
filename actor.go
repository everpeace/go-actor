package actor

import "code.google.com/p/go-uuid/uuid"

// actor starter
func Spawn(receive Receive) Actor {
	actor := newActorImpl(receive)
	startLatch := actor.context.start()
	startLatch <- true
	return actor
}

func SpawnWithName(name string, receive Receive) Actor {
	actor := newActorImplWithName(name, receive)
	startLatch := actor.context.start()
	startLatch <- true
	return actor
}

func SpawnWithLatch(receive Receive) (chan bool, Actor) {
	actor := newActorImpl(receive)
	startLatch := actor.context.start()
	return startLatch, actor
}

func SpawnWithNameAndLatch(name string, receive Receive) (chan bool, Actor) {
	actor := newActorImplWithName(name, receive)
	startLatch := actor.context.start()
	return startLatch, actor
}
type actorImpl struct {
	name    string
	context *actorContext
}

func (actor *actorImpl) Send(msg Message) {
	go func() {
		actor.context.mailbox <- msg
	}()
}

func (actor *actorImpl) Terminate() {
	go func() {
		actor.context.terminateChan <- terminate{}
	}()
}

func (actor *actorImpl) Kill() {
	go func() {
		actor.context.killChan <- kill{}
	}()
}

func (actor *actorImpl) Monitor(mon Actor) {
	go func() {
		actor.context.attachMonitor(mon)
	}()
}

func (actor *actorImpl) Demonitor(mon Actor) {
	go func() {
		actor.context.detachMonitor(mon)
	}()
}

func (actor *actorImpl) Name() string{
	return actor.name
}

// private constructors
func newActorImpl(receive Receive) *actorImpl {
	return &actorImpl{
		name:    uuid.New(),
		context: newActorContext(receive),
	}
}

func newActorImplWithName(name string, receive Receive) *actorImpl {
	actor := &actorImpl{
		name:    name,
		context: newActorContext(receive),
	}
	actor.context.self = actor
	return actor
}
