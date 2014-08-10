package actor

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
		context: &ActorContext{
			behaviorStack: []Receive{receive},
			mailbox:       make(chan Message),
		},
	}
	ActorImpl.context.self = &ActorImpl
	go ActorImpl.context.loop()
	return ActorImpl
}

type ActorImpl struct {
	context *ActorContext
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
type ActorContext struct {
	self          *ActorImpl
	behaviorStack []Receive
	mailbox       chan Message
}

func (context *ActorContext) closeAll() {
	close(context.mailbox)
}

// Actor's main process
func (context *ActorContext) loop() {
	for {
		select {
		case msg := <-context.mailbox:
			if len(msg) == 0 {
				context.behaviorStack[len(context.behaviorStack)-1](msg, context.self)
			} else if _, ok := msg[0].(Terminate); ok {
				context.closeAll()
				return
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
		}
	}
}
