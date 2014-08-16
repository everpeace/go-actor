package actor

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
)

// actor starter
func Spawn(receive Receive) Actor {
	startLatch, actor := spawn(nil, uuid.New(), receive)
	startLatch <- true
	return actor
}

func SpawnWithName(name string, receive Receive) Actor {
	startLatch, actor := spawn(nil, name, receive)
	startLatch <- true
	return actor
}

func SpawnWithLatch(receive Receive) (chan bool, Actor) {
	return spawn(nil, uuid.New(), receive)
}

func SpawnWithNameAndLatch(name string, receive Receive) (chan bool, Actor) {
	return 	spawn(nil, name, receive)
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
	actor.context.attachMonitor(mon)
}

func (actor *actorImpl) Demonitor(mon Actor) {
	actor.context.detachMonitor(mon)
}

func (actor *actorImpl) Spawn(receive Receive) Actor{
	child_name := actor.Name()+"/"+fmt.Sprint(len(actor.context.children))
	latch, child := spawn(actor,child_name,receive)
	latch <- true
	return child
}

func (actor *actorImpl) SpawnWithName(name string, receive Receive) Actor{
	latch, child := spawn(actor, actor.Name()+"/"+name, receive)
	latch <- true
	return child
}

func (actor *actorImpl) SpawnWithLatch(receive Receive) (chan bool, Actor){
	child_name := actor.Name()+"/"+string(len(actor.context.children))
	latch, child := spawn(actor,child_name,receive)
	return latch, child
}

func (actor *actorImpl) SpawnWithNameAndLatch(name string, receive Receive) (chan bool, Actor){
	latch, child := spawn(actor, actor.Name()+"/"+name, receive)
	return latch, child
}

func (actor *actorImpl) Name() string{
	return actor.name
}

// private constructors
func newActorImpl(parent *actorImpl, name string, receive Receive) *actorImpl {
	if parent == nil {
		actor := &actorImpl{
			name:    name,
			context: newActorContext(nil, receive),
		}
		actor.context.self = actor
		return actor
	} else {
		actor := &actorImpl{
			name:    name,
			context: newActorContext(parent.context, receive),
		}
		actor.context.self = actor
		return actor
	}
}

func spawn(parent *actorImpl, name string, receive Receive) (chan bool, Actor){
	actor := newActorImpl(parent, name, receive)
	startLatch := actor.context.start()
	return startLatch, actor
}
