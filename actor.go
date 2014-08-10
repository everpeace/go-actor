package main

import (
	"fmt"
	"strings"
)

// Message is just slice.  I use this instead of tuple.
type Message []interface{}

type Receive func(msg Message, self Actor)

func spawn(receive Receive) Actor {
	ActorImpl := ActorImpl{
		context: &ActorContext{
			behaviorStack: []Receive{receive},
			mailbox:       make(chan Message),
			terminateChan: make(chan string),
			becomeChan:    make(chan Become),
			unbecomeChan:  make(chan string),
		},
	}
	ActorImpl.context.self = &ActorImpl
	go ActorImpl.context.loop()
	return ActorImpl
}

type Actor interface {
	send(msg Message)
	become(behavior Receive, discardOld bool)
	unbecome()
	terminate()
}

type ActorImpl struct {
	context *ActorContext
}

func (actor ActorImpl) send(msg Message) {
	actor.context.mailbox <- msg
}
func (actor ActorImpl) become(behavior Receive, discardOld bool) {
	actor.context.becomeChan <- Become{newBehavior: behavior, discardOld: discardOld}
}
func (actor ActorImpl) unbecome() {
	actor.context.unbecomeChan <- "unbecome"
}
func (actor ActorImpl) terminate() {
	actor.context.terminateChan <- "terminate"
}

type Become struct {
	newBehavior Receive
	discardOld  bool
}

// Actor Context
type ActorContext struct {
	self          *ActorImpl
	behaviorStack []Receive
	mailbox       chan Message
	terminateChan chan string
	becomeChan    chan Become
	unbecomeChan  chan string
}

func (context *ActorContext) closeAll() {
	close(context.mailbox)
	close(context.becomeChan)
	close(context.unbecomeChan)
	close(context.terminateChan)
}

func (context *ActorContext) loop() {
	for {
		select {
		// These select clauses are evaluated concurrently??
		// If so, this would not be safe.
		case <-context.terminateChan:
			context.closeAll()
			return
		case <-context.unbecomeChan:
			l := len(context.behaviorStack)
			if l > 1 {
				context.behaviorStack = context.behaviorStack[:l-1]
			} else {
				// TODO raise error
			}
		case become := <-context.becomeChan:
			if become.discardOld {
				l := len(context.behaviorStack)
				context.behaviorStack[l-1] = become.newBehavior
			} else {
				newStack := append(context.behaviorStack, become.newBehavior)
				context.behaviorStack = newStack
			}
		case msg := <-context.mailbox:
			context.behaviorStack[len(context.behaviorStack)-1](msg, context.self)
		}
	}
}

func main() {
	latch := make(chan int, 1)

	stage := func(task func()) {
		_task := func() {
			task()
			latch <- 0
		}
		go _task()
	}

	stage(func() {
		// Ping-Pong
		ponger := spawn(func(msg Message, self Actor) {
			fmt.Printf("ponger received: %s\n", msg[1])
			msg[0].(Actor).send(Message{"Pong"})
			self.terminate()
		})
		pinger := spawn(func(msg Message, self Actor) {
			if msg[0] == "Pong" {
				fmt.Printf("pinger received: Pong.\nPing-Pong finished.\n")
				self.terminate()
			} else {
				fmt.Printf("pinger sends: Ping\n")
				ponger.send(Message{self, "Ping"})
			}
		})
		pinger.send(Message{"_"})
	})

	<-latch

	stage(func() {
		// Become/Unbecome
		echo := func(msg Message, self Actor) {
			fmt.Println(msg[0])
		}
		echoInUpper := func(msg Message, self Actor) {
			fmt.Println(strings.ToUpper(msg[0].(string)))
		}
		actor := spawn(echo)
		actor.send(Message{"this should be echoed."})
		actor.become(echoInUpper, false)
		actor.send(Message{"this should be echoed in upper case."})
		actor.unbecome()
		actor.send(Message{"this should be echoed."})
		actor.terminate()
	})

	<-latch
	close(latch)
}
