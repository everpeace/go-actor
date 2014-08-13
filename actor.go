package actor

import (
	"time"
)

// Types
type Message []interface{} // Message is just slice (instead of tuple.)
type Receive func(msg Message, self Actor)
type Actor interface {
	Send(msg Message)
	Become(behavior Receive, discardOld bool)
	Unbecome()
	Terminate()
	// child actor starter
	Spawn(receive Receive) Actor
}

// Special Messages
type Become struct {
	newBehavior Receive
	discardOld  bool
}
type UnBecome struct{}
type Terminate struct{}

// orphan actor starter
func Spawn(receive Receive) Actor {
	return spawn(nil, receive)
}

type ActorImpl struct {
	context *actorContext
}

func (actor *ActorImpl) Send(msg Message) {
	go func() { actor.context.mailbox <- msg }()
}
func (actor *ActorImpl) Become(behavior Receive, discardOld bool) {
	go func() { actor.context.mailbox <- Message{Become{newBehavior: behavior, discardOld: discardOld}} }()
}
func (actor *ActorImpl) Unbecome() {
	go func() { actor.context.mailbox <- Message{UnBecome{}} }()
}
func (actor *ActorImpl) Terminate() {
	go func() { actor.context.mailbox <- Message{Terminate{}} }()
}
func (actor *ActorImpl) Spawn(receive Receive) Actor {
	spawnedActor := spawn(actor, receive)
	actor.context.children = append(actor.context.children, actor)
	return spawnedActor
}

// Actor Context
// TODO actor hierarchy and supervision (this would introduce actor system)
type actorContext struct {
	parent         *ActorImpl
	self           *ActorImpl
	children       []*ActorImpl
	behaviorStack  []Receive
	mailbox        chan Message
	receiveTimeout time.Duration
}

func newActorContext(parent *ActorImpl, receive Receive) *actorContext {
	return &actorContext{
		parent:         parent,
		children:       make([]*ActorImpl, 0, 100),
		behaviorStack:  []Receive{receive},
		mailbox:        make(chan Message),
		receiveTimeout: time.Duration(10) * time.Millisecond,
	}
}

func (context *actorContext) closeAll() {
	close(context.mailbox)
}

// Actor's main process
func (context *actorContext) loop() {
	defer context.closeAll()
	forever(context.processOneMessage())
}

func (context *actorContext) processOneMessage() func() bool {
	return func() bool {
		select {
		case msg := <-context.mailbox:
			if len(msg) == 0 {
				context.behaviorStack[len(context.behaviorStack)-1](msg, context.self)
			} else if _, ok := msg[0].(Terminate); ok {
				return true
			} else if become, ok := msg[0].(Become); ok {
				if become.discardOld {
					l := len(context.behaviorStack)
					context.behaviorStack[l-1] = become.newBehavior
				} else {
					newStack := append(context.behaviorStack, become.newBehavior)
					context.behaviorStack = newStack
				}
			} else if _, ok := msg[0].(UnBecome); ok {
				l := len(context.behaviorStack)
				if l > 1 {
					context.behaviorStack = context.behaviorStack[:l-1]
				} else {
					// TODO raise error
				}
			} else {
				context.behaviorStack[len(context.behaviorStack)-1](msg, context.self)
			}
		case <-time.After(context.receiveTimeout):
			// timeout
		}
		return false
	}
}

func spawn(parent *ActorImpl, receive Receive) *ActorImpl {
	actor := &ActorImpl{
		context: newActorContext(parent, receive),
	}
	actor.context.self = actor
	go actor.context.loop()
	return actor
}
