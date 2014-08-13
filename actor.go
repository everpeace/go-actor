package actor

import (
	"runtime"
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
}

// Special Messages
type Become struct {
	newBehavior Receive
	discardOld  bool
}
type UnBecome struct{}
type Terminate struct{}

// actor starter
func Spawn(receive Receive) Actor {
	ActorImpl := ActorImpl{
		context: &actorContextt{
			behaviorStack:  []Receive{receive},
			mailbox:        make(chan Message),
			receiveTimeout: time.Duration(10) * time.Millisecond,
		},
	}
	ActorImpl.context.self = &ActorImpl
	go ActorImpl.context.loop()
	return ActorImpl
}

type ActorImpl struct {
	context *actorContextt
}

func (actor ActorImpl) Send(msg Message) {
	go func() { actor.context.mailbox <- msg }()
}
func (actor ActorImpl) Become(behavior Receive, discardOld bool) {
	go func() { actor.context.mailbox <- Message{Become{newBehavior: behavior, discardOld: discardOld}} }()
}
func (actor ActorImpl) Unbecome() {
	go func() { actor.context.mailbox <- Message{UnBecome{}} }()
}
func (actor ActorImpl) Terminate() {
	go func() { actor.context.mailbox <- Message{Terminate{}} }()
}

// Actor Context
// TODO actor hierarchy and supervision (this would introduce actor system)
type actorContextt struct {
	self           *ActorImpl
	behaviorStack  []Receive
	mailbox        chan Message
	receiveTimeout time.Duration
}

func (context *actorContextt) closeAll() {
	close(context.mailbox)
}

// Actor's main process
func (context *actorContextt) loop() {
	defer context.closeAll()
	forever(context.processOneMessage())
}

func (context *actorContextt) processOneMessage() func() bool {
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

func forever(fun func() bool) {
	for {
		// if fun() returns true, forever ends.
		if fun() {
			break
		}
		// force give back control to goroutine scheduler.
		runtime.Gosched()
	}
}
