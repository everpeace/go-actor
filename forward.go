package actor

import "gopkg.in/fatih/set.v0"

type forwardingActor struct {
	*actorImpl
}

// internal message used in forwardingActor
type addRecipient struct {
	recipient Actor
}
type removeRecipient struct {
	recipient Actor
}

func SpawnForwardActor(recipients ...Actor) ForwardingActor {
	recipientSet := set.New()
	for _, recipient := range recipients {
		recipientSet.Add(recipient)
	}
	forwardActor := &forwardingActor{
		newActorImpl(forward(recipientSet)),
	}
	latch := forwardActor.context.start()
	latch <- true
	return forwardActor
}

func (actor *forwardingActor) Add(recipient Actor) {
	go func() {
		actor.context.Self().Send(Message{addRecipient{
			recipient: recipient,
		}})
	}()
}

func (actor *forwardingActor) Remove(recipient Actor) {
	go func() {
		actor.context.Self().Send(Message{removeRecipient{
			recipient: recipient,
		}})
	}()
}

func forward(actors *set.Set) Receive {
	return func(msg Message, context ActorContext) {
		if len(msg) == 0 {
			for _, actor := range actors.List() {
				if a, ok := actor.(Actor); ok {
					a.Send(msg)
				}
			}
		} else if recipient, ok := msg[0].(addRecipient); ok {
			actors.Add(recipient)
		} else if recipient, ok := msg[0].(removeRecipient); ok {
			actors.Remove(recipient)
		} else {
			for _, actor := range actors.List() {
				if a, ok := actor.(Actor); ok {
					a.Send(msg)
				}
			}
		}
	}
}
