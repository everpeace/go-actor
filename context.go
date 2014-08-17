package actor

import (
	"runtime"
	"time"

	"github.com/dropbox/godropbox/container/set"
)

// Actor Context
// TODO actor hierarchy and supervision (this would introduce actor system)
type actorContext struct {
	system           *actorSystem
	parent           *actorContext
	self             *actorImpl
	children         set.Set
	monitor          ForwardingActor
	originalBehavior Receive
	currentBehavior  Receive
	behaviorStack    []Receive
	mailbox          chan Message
	killChan         chan kill
	terminateChan    chan terminate
	attachMonChan    chan Actor
	detachMonChan    chan Actor
	addChildChan     chan *actorContext
	shutdownChan     chan shutdown
	receiveTimeout   time.Duration
}

// internal Messages accepted by actorContext
type terminate struct{}
type kill struct{}
type shutdown struct{}

// ActorContext Interfaces
func (context *actorContext) Self() Actor {
	return context.self
}

func (context *actorContext) Become(behavior Receive, discardOld bool) {
	if discardOld {
		l := len(context.behaviorStack)
		context.behaviorStack[l-1] = behavior
	} else {
		newStack := append(context.behaviorStack, behavior)
		context.behaviorStack = newStack
	}
	context.currentBehavior = behavior
}

func (context *actorContext) Unbecome() {
	l := len(context.behaviorStack)
	if l > 1 {
		context.behaviorStack = context.behaviorStack[:l-1]
		context.currentBehavior = context.behaviorStack[l-2]
	} else {
		// TODO raise error
	}
}

// constructor
func newTopLevelActorContext(system *actorSystem, self *actorImpl, receive Receive) *actorContext {
	// TODO make parameters configurable
	return &actorContext{
		system:           system,
		self:             self,
		children:         set.NewSet(),
		originalBehavior: receive,
		currentBehavior:  receive,
		behaviorStack:    []Receive{receive},
		mailbox:          make(chan Message, 100),
		// buffer size for control message is 1 (cotrol method would block)
		attachMonChan:    make(chan Actor),
		detachMonChan:    make(chan Actor),
		terminateChan:    make(chan terminate),
		killChan:         make(chan kill),
		addChildChan:     make(chan *actorContext),
		shutdownChan:     make(chan shutdown),
		receiveTimeout:   time.Duration(10) * time.Millisecond,
	}
}

func newActorContext(parent *actorContext, self *actorImpl, receive Receive) *actorContext {
	// TODO make parameters configurable
	context :=  &actorContext{
		system:           parent.system,
		parent:           parent,
		self:             self,
		children:         set.NewSet(),
		originalBehavior: receive,
		currentBehavior:  receive,
		behaviorStack:    []Receive{receive},
		mailbox:          make(chan Message, 100),
		// buffer size for control message is 1 (cotrol method would block)
		attachMonChan:    make(chan Actor),
		detachMonChan:    make(chan Actor),
		terminateChan:    make(chan terminate),
		killChan:         make(chan kill),
		addChildChan:     make(chan *actorContext),
		shutdownChan:     make(chan shutdown),
		receiveTimeout:   time.Duration(10) * time.Millisecond,
	}
	parent.addChildChan <- context
	return context
}

func (context *actorContext) closeAllChan() {
	// TODO flush remained message to deadletter channel??
	go func() {
		defer logPanic(context.self)
		close(context.mailbox)
		close(context.terminateChan)
		close(context.killChan)
	}()
}

func (context *actorContext) start() chan bool {
	startLatch := make(chan bool)
	context.system.wg.Add(1)
	go func() {
		defer func(){
			context.system.running.Remove(context.Self())
			context.system.stopped.Add(context.Self())
			context.system.wg.Done()
			context.closeAllChan()
		}()
		context.system.running.Add(context.Self())
		<-startLatch
		close(startLatch)
		context.loop()
	}()
	return startLatch
}

func (context *actorContext) attachMonitor(mon Actor) {
	context.attachMonChan <- mon
}

func (context *actorContext) detachMonitor(mon Actor) {
	context.detachMonChan <- mon
}

func (context *actorContext) kill() {
	context.killChan <- kill{}
}

func (context *actorContext) terminate() {
	context.terminateChan <- terminate{}
}

func (context *actorContext) shutdown() {
	context.shutdownChan <- shutdown{}
}


// Actor's main loop which is executed in go routine
func (context *actorContext) loop() {
	// TODO how monitor detect STARTED??
	// now monitor can't detect STARTED.
	// context.notifyMonitors(Message{STARTED})
	for {
		// TODO log
		select {
		case <-context.killChan:
			context.notifyMonitors(Message{Down{
				Cause: "killed",
				Actor: context.Self(),
			}})
			if context.parent != nil {
				context.parent.kill()
			}
			return
		case <-context.terminateChan:
			context.notifyMonitors(Message{Down{
				Cause: "terminated",
				Actor: context.Self(),
			}})
			if context.parent != nil {
				context.parent.terminate()
			}
			return
		case mon := <-context.attachMonChan:
			if context.monitor == nil {
				context.monitor = context.system.spawnMonitorForwarderFor(context.self)
			}
			context.monitor.Add(mon)
		case mon := <-context.detachMonChan:
			context.monitor.Remove(mon)
		case child := <-context.addChildChan:
			context.children.Add(child)
		case <-context.shutdownChan:
			for child := range context.children.Iter() {
				if childContext, ok := child.(*actorContext); ok {
					childContext.shutdown()
				}
			}
			context.notifyMonitors(Message{Down{
				Cause: "shutdowned",
				Actor: context.Self(),
			}})
			return
		default:
			context.processOneMessage()
		}
		// force give back control to goroutine scheduler.
		runtime.Gosched()
	}
}

func (context *actorContext) processOneMessage() {
	select {
	case msg := <-context.mailbox:
		context.currentBehavior(msg, context)
	case <-time.After(context.receiveTimeout):
		// TODO log
		// timeout
	}
}

func (context *actorContext) notifyMonitors(msg Message){
	if context.monitor != nil {
		context.monitor.Send(msg)
	} else {
		//TODO log
	}
}
