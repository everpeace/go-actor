package actor

import (
	"runtime"
	"time"
)

// Actor Context
// TODO actor hierarchy and supervision (this would introduce actor system)
type actorContext struct {
	self             *actorImpl
	monitors         map[string]*actorImpl
	originalBehavior Receive
	currentBehavior  Receive
	behaviorStack    []Receive
	mailbox          chan Message
	killChan         chan kill
	terminateChan    chan terminate
	receiveTimeout   time.Duration
}

// internal Messages accepted by actorContext
type terminate struct{}
type kill struct{}

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
func newActorContext(receive Receive) *actorContext {
	// TODO make parameters configurable
	return &actorContext{
		originalBehavior: receive,
		currentBehavior:  receive,
		behaviorStack:    []Receive{receive},
		mailbox:          make(chan Message, 100),
		// buffer size for control message is 1 (cotrol method would block)
		terminateChan:  make(chan terminate),
		killChan:       make(chan kill),
		receiveTimeout: time.Duration(10) * time.Millisecond,
	}
}

func (context *actorContext) closeAll() {
	// TODO deadletter channel??
	close(context.mailbox)
	close(context.terminateChan)
	close(context.killChan)
}

func (context *actorContext) restart() {
	// force restart
	context.killChan <- kill{}
	context.start()
}

func (context *actorContext) start() chan bool {
	startLatch := make(chan bool)
	go func() {
		<-startLatch
		close(startLatch)
		context.loop()
	}()
	return startLatch
}

// Actor's main loop which is executed in go routine
func (context *actorContext) loop() {
	defer context.closeAll()
	for {
		// TODO log
		select {
		case <-context.killChan:
			break
		case <-context.terminateChan:
			break
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
