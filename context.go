package actor

import (
	"runtime"
	"time"
)

// Actor Context
//
// Actor Context will be passed to actor's message handler with message received.
// This data enables message handler to access myself(it's useful when you want
// to send a message to myself), and to change its behavior.
type ActorContext struct {
	Self             *Actor
	monitor          *ForwardingActor
	originalBehavior Receive
	currentBehavior  Receive
	behaviorStack    []Receive
	mailbox          chan Message
	killChan         chan kill
	attachMonChan    chan *Actor
	detachMonChan    chan *Actor
	addChildChan     chan *Actor
	receiveTimeout   time.Duration
	prePrecessHook   func()

}

// internal Messages accepted by actorContext
type terminate struct{}
type kill struct{}
type shutdown struct{}

// Become change actor's behavior.
//
// ActorContext can be accessed in message handler.
// For example,
//   echo := func(msg Message, context *ActorContext){
//     fmt.Println(msg)
//     context.Become(echoInUpper)
//   }
//   echoInUpper := func(msg Message, context *ActorContext){
//     fmt.Println(msg)
//     context.Unbecome()
//   }
//   alternate := system.Spawn(echo)
// "alternate" actor behaves echo and echoInUpper alternately.
// As you might notice, actor's behavior was stacked.  So, you can
// unbecome to set back to the previous behavior.
//
// If discardOld flag is true, this doesn't push a given behavior to
// its behavior stack but just overwrite the top of the stack.
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

// Unbecome set its behavior to the previous behavior in its behavior stack.
//
// Please see Become for details.
func (context *ActorContext) Unbecome() {
	l := len(context.behaviorStack)
	if l > 1 {
		context.behaviorStack = context.behaviorStack[:l-1]
		context.currentBehavior = context.behaviorStack[l-2]
	} else {
		// TODO raise error
	}
}

// Watch method attaches myself to given actor as monitor.
//
// This is equivalent with
//   actor.Monitor(context.Self)
func (context *ActorContext) Watch(actor *Actor){
	actor.Monitor(context.Self)
}

// Unwatch method detaches myself as monitor from given actor.
//
// This is equivalent with
//   actor.Demonitor(context.Self)
func (context *ActorContext) Unwatch(actor* Actor){
	actor.Demonitor(context.Self)
}

// constructor
func newActorContext(self *Actor, receive Receive) *ActorContext {
	// TODO make parameters configurable
	context := &ActorContext{
		Self:             self,
		originalBehavior: receive,
		currentBehavior:  receive,
		behaviorStack:    []Receive{receive},
		mailbox:          make(chan Message, 100),
		// buffer size for control message is 1 (cotrol method would block)
		attachMonChan:  make(chan *Actor),
		detachMonChan:  make(chan *Actor),
		killChan:       make(chan kill),
		addChildChan:   make(chan *Actor),
		receiveTimeout: time.Duration(10) * time.Millisecond,
	}
	return context
}

func (context *ActorContext) closeAllChan() {
	// TODO flush remained message to deadletter channel??
	go func() {
		defer logPanic(context.Self)
		close(context.mailbox)
		close(context.addChildChan)
		close(context.attachMonChan)
		close(context.detachMonChan)
		close(context.killChan)
	}()
}

func (context *ActorContext) start() chan bool {
	startLatch := make(chan bool)
	context.Self.System.wg.Add(1)
	go func() {
		defer func() {
			context.Self.System.running.Remove(context.Self)
			context.Self.System.wg.Done()
			context.closeAllChan()
			context.Self.System.stopped.Add(context.Self)
		}()
		context.Self.System.running.Add(context.Self)
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
	context.mailbox <- Message{PoisonPill{}}
}

// Actor's main loop which is executed in go routine
func (context *ActorContext) loop() {
	// TODO how monitor detect STARTED??
	// now monitor can't detect STARTED.
	// context.notifyMonitors(Message{STARTED})
	for {
		// TODO log
		receiveTerminated := false
		select {
		case <-context.killChan:
			context.notifyMonitors(Message{Down{
				Cause: "killed",
				Actor: context.Self,
			}})
			context.Self.children.Do(func(child *Actor){
				if child.IsRunning() {
					child.context.kill()
				}
			})
			return
		case mon := <-context.attachMonChan:
			if context.monitor == nil {
				context.monitor = context.Self.System.spawnMonitorForwarderFor(context.Self)
			}
			context.monitor.Add(mon)
		case mon := <-context.detachMonChan:
			context.monitor.Remove(mon)
		case child := <-context.addChildChan:
			context.Self.children.Add(child)
		default:
			if context.prePrecessHook!= nil {
				context.prePrecessHook()
			}
			receiveTerminated = context.processOneMessage()
		}
		if receiveTerminated {
			return
		}
		// force give back control to goroutine scheduler.
		runtime.Gosched()
	}
}

func (context *ActorContext) processOneMessage() bool{
	select {
	case msg := <-context.mailbox:
		if len(msg) == 1 {
			if _, ok := msg[0].(PoisonPill); ok {
				context.notifyMonitors(Message{Down{
					Cause: "terminated",
					Actor: context.Self,
				}})
				context.Self.children.Do(func(child *Actor){
					if child.IsRunning() {
						child.context.terminate()
					}
				})
				return true
			} else {
				context.currentBehavior(msg, context)
			}
		} else {
			context.currentBehavior(msg, context)
		}
	case <-time.After(context.receiveTimeout):
		// TODO log
		// timeout
	}
	return false
}

func (context *ActorContext) notifyMonitors(msg Message) {
	if context.monitor != nil {
		context.monitor.Send(msg)
	} else {
		//TODO log
	}
}
