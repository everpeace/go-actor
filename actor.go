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
	go func() { actor.context.mailbox <- msg }()
}
func (actor ActorImpl) become(behavior Receive, discardOld bool) {
	go func() { actor.context.mailbox <- Message{Become{newBehavior: behavior, discardOld: discardOld}} }()
}
func (actor ActorImpl) unbecome() {
	go func() { actor.context.mailbox <- Message{UnBecome{}} }()
}
func (actor ActorImpl) terminate() {
	go func() { actor.context.mailbox <- Message{Terminate{}} }()
}

type Become struct {
	newBehavior Receive
	discardOld  bool
}
type UnBecome struct{}
type Terminate struct{}

// Actor Context
type ActorContext struct {
	self          *ActorImpl
	behaviorStack []Receive
	mailbox       chan Message
}

func (context *ActorContext) closeAll() {
	close(context.mailbox)
}

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

func main() {
	staged := func(task func(fin chan int)) chan int {
		barrier := make(chan int)
		go func() {
			_fin := make(chan int)
			<-barrier           // waiting start
			go task(_fin)       // task should handle their finish
			barrier <- (<-_fin) // waiting for task finishing and report it.
			close(_fin)
		}()
		return barrier
	}

	st1 := staged(func(ch chan int) {
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
				ch <- 0
			} else {
				fmt.Printf("pinger sends: Ping\n")
				ponger.send(Message{self, "Ping"})
			}
		})

		pinger.send(Message{"_"})
	})

	st2 := staged(func(ch chan int) {
		// Become/Unbecome
		echo := func(msg Message, self Actor) {
			fmt.Println(msg[0])
		}
		echoInUpper := func(msg Message, self Actor) {
			fmt.Println(strings.ToUpper(msg[0].(string)))
		}
		terminate := func(msg Message, self Actor) {
			self.terminate()
			ch <- 0
		}
		actor := spawn(echo)
		actor.send(Message{"this should be echoed."})
		actor.become(echoInUpper, false)
		actor.send(Message{"this should be echoed in upper case."})
		actor.unbecome()
		actor.send(Message{"this should be echoed."})
		actor.become(terminate, true)
		actor.send(Message{})
	})

	// control stages
	// start first stage
	fmt.Println("=== Ping/Pont Test ===")
	st1 <- 0
	<-st1
	close(st1)

	// start second stage
	fmt.Println("=== Become/Unbecome Test ===")
	st2 <- 0
	<-st2
	close(st2)
}
