package actor

import "github.com/dropbox/godropbox/container/set"

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

func forward(recipients set.Set) Receive {
	return func(msg Message, context ActorContext) {
		if len(msg) == 0 {
			for actor := range recipients.Iter() {
				if a, ok := actor.(Actor); ok {
					a.Send(msg)
				}
			}
		} else if m, ok := msg[0].(addRecipient); ok {
			recipients.Add(m.recipient)
		} else if m, ok := msg[0].(removeRecipient); ok {
			recipients.Remove(m.recipient)
		} else {
			for actor := range recipients.Iter() {
				if a, ok := actor.(Actor); ok {
					a.Send(msg)
				}
			}
		}
	}
}
