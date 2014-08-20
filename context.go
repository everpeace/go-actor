package actor

import (
	"runtime"
	"time"

	"github.com/dropbox/godropbox/container/set"
)

// Actor Context
// TODO actor hierarchy and supervision (this would introduce actor system)
type ActorContext struct {
	System           *ActorSystem
	Parent           *ActorContext
	Self             *Actor
	Children         set.Set
	monitor          *ForwardingActor
	originalBehavior Receive
	currentBehavior  Receive
	behaviorStack    []Receive
	mailbox          chan Message
	killChan         chan kill
	terminateChan    chan terminate
	attachMonChan    chan *Actor
	detachMonChan    chan *Actor
	addChildChan     chan *ActorContext
	shutdownChan     chan shutdown
	receiveTimeout   time.Duration
}

// internal Messages accepted by actorContext
type terminate struct{}
type kill struct{}
type shutdown struct{}

func (context *ActorContext) Become(behavior Receive, discardOld bool) {
	if discardOld {
		l := len(context.behaviorStack)
		context.behaviorStack[l-1] = behavior
	} else {
		newStack := append(context.behaviorStack, behavior)
		context.behaviorStack = newStack
	}
	context.currentBehavior = behavior
}

func (context *ActorContext) Unbecome() {
	l := len(context.behaviorStack)
	if l > 1 {
		context.behaviorStack = context.behaviorStack[:l-1]
		context.currentBehavior = context.behaviorStack[l-2]
	} else {
		// TODO raise error
	}
}

// constructor
func newTopLevelActorContext(system *ActorSystem, self *Actor, receive Receive) *ActorContext {
	// TODO make parameters configurable
	return &ActorContext{
		System:           system,
		Self:             self,
		Children:         set.NewSet(),
		originalBehavior: receive,
		currentBehavior:  receive,
		behaviorStack:    []Receive{receive},
		mailbox:          make(chan Message, 100),
		// buffer size for control message is 1 (cotrol method would block)
		attachMonChan:  make(chan *Actor),
		detachMonChan:  make(chan *Actor),
		terminateChan:  make(chan terminate),
		killChan:       make(chan kill),
		addChildChan:   make(chan *ActorContext),
		shutdownChan:   make(chan shutdown),
		receiveTimeout: time.Duration(10) * time.Millisecond,
	}
}

func newActorContext(parent *ActorContext, self *Actor, receive Receive) *ActorContext {
	// TODO make parameters configurable
	context := &ActorContext{
		System:           parent.System,
		Parent:           parent,
		Self:             self,
		Children:         set.NewSet(),
		originalBehavior: receive,
		currentBehavior:  receive,
		behaviorStack:    []Receive{receive},
		mailbox:          make(chan Message, 100),
		// buffer size for control message is 1 (cotrol method would block)
		attachMonChan:  make(chan *Actor),
		detachMonChan:  make(chan *Actor),
		terminateChan:  make(chan terminate),
		killChan:       make(chan kill),
		addChildChan:   make(chan *ActorContext),
		shutdownChan:   make(chan shutdown),
		receiveTimeout: time.Duration(10) * time.Millisecond,
	}
	parent.addChildChan <- context
	return context
}

func (context *ActorContext) closeAllChan() {
	// TODO flush remained message to deadletter channel??
	go func() {
		defer logPanic(context.Self)
		close(context.mailbox)
		close(context.terminateChan)
		close(context.killChan)
	}()
}

func (context *ActorContext) start() chan bool {
	startLatch := make(chan bool)
	context.System.wg.Add(1)
	go func() {
		defer func() {
			context.System.wg.Done()
			context.closeAllChan()
		}()
		context.System.running.Add(context.Self)
		<-startLatch
		close(startLatch)
		context.loop()
	}()
	return startLatch
}

func (context *ActorContext) attachMonitor(mon *Actor) {
	context.attachMonChan <- mon
}

func (context *ActorContext) detachMonitor(mon *Actor) {
	context.detachMonChan <- mon
}

func (context *ActorContext) kill() {
	context.killChan <- kill{}
}

func (context *ActorContext) terminate() {
	context.terminateChan <- terminate{}
}

func (context *ActorContext) shutdown() {
	context.shutdownChan <- shutdown{}
}

// Actor's main loop which is executed in go routine
func (context *ActorContext) loop() {
	// TODO how monitor detect STARTED??
	// now monitor can't detect STARTED.
	// context.notifyMonitors(Message{STARTED})
	for {
		// TODO log
		select {
		case <-context.killChan:
			context.notifyMonitors(Message{Down{
				Cause: "killed",
				Actor: context.Self,
			}})
			if context.Parent != nil {
				context.Parent.kill()
			}
			return
		case <-context.terminateChan:
			context.notifyMonitors(Message{Down{
				Cause: "terminated",
				Actor: context.Self,
			}})
			if context.Parent != nil {
				context.Parent.terminate()
			}
			return
		case mon := <-context.attachMonChan:
			if context.monitor == nil {
				context.monitor = context.System.spawnMonitorForwarderFor(context.Self)
			}
			context.monitor.Add(mon)
		case mon := <-context.detachMonChan:
			context.monitor.Remove(mon)
		case child := <-context.addChildChan:
			context.Children.Add(child)
		case <-context.shutdownChan:
			for child := range context.Children.Iter() {
				if childContext, ok := child.(*ActorContext); ok {
					childContext.shutdown()
				}
			}
			context.notifyMonitors(Message{Down{
				Cause: "shutdowned",
				Actor: context.Self,
			}})
			return
		default:
			context.processOneMessage()
		}
		// force give back control to goroutine scheduler.
		runtime.Gosched()
	}
}

func (context *ActorContext) processOneMessage() {
	select {
	case msg := <-context.mailbox:
		context.currentBehavior(msg, context)
	case <-time.After(context.receiveTimeout):
		// TODO log
		// timeout
	}
}

func (context *ActorContext) notifyMonitors(msg Message) {
	if context.monitor != nil {
		context.monitor.Send(msg)
	} else {
		//TODO log
	}
}
