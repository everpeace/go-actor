package actor

import "github.com/dropbox/godropbox/container/set"

type ForwardingActor struct {
	*Actor
}

// internal message used in ForwardingActor
type addRecipient struct {
	recipient *Actor
}
type removeRecipient struct {
	recipient *Actor
}

func (actor *ForwardingActor) Add(recipient *Actor) {
	go func() {
		actor.context.Self.Send(Message{addRecipient{
			recipient: recipient,
		}})
	}()
}

func (actor *ForwardingActor) Remove(recipient *Actor) {
	go func() {
		actor.context.Self.Send(Message{removeRecipient{
			recipient: recipient,
		}})
	}()
}

func forward(recipients set.Set) Receive {
	return func(msg Message, context *ActorContext) {
		if len(msg) == 0 {
			recipients.Do(func(actor interface{}) {
				if a, ok := actor.(*Actor); ok {
					a.Send(msg)
				}
			})
		} else if m, ok := msg[0].(addRecipient); ok {
			recipients.Add(m.recipient)
		} else if m, ok := msg[0].(removeRecipient); ok {
			recipients.Remove(m.recipient)
		} else {
			recipients.Do(func(actor interface{}) {
				if a, ok := actor.(*Actor); ok {
					a.Send(msg)
				}
			})
		}
	}
}
