package actor

import "github.com/dropbox/godropbox/container/set"

type ForwardingActor struct {
	*Actor
	addRecipientChan chan addRecipient
	delRecipientChan chan removeRecipient
	recipients set.Set
}

// internal message used in ForwardingActor
type addRecipient struct {
	recipient *Actor
}
type removeRecipient struct {
	recipient *Actor
}

func (actor *ForwardingActor) Add(recipient *Actor) {
	actor.addRecipientChan <- addRecipient{recipient: recipient}
}

func (actor *ForwardingActor) Remove(recipient *Actor) {
	actor.delRecipientChan <- removeRecipient{recipient: recipient}
}

func (actor *ForwardingActor) preProcessHook() func(){
	return func() {
		select{
		case m := <-actor.addRecipientChan:
			actor.recipients.Add(m.recipient)
		case m := <-actor.delRecipientChan:
			actor.recipients.Remove(m.recipient)
		default:
		}
	}
}

func (actor *ForwardingActor) receive() Receive {
	return func(msg Message, context *ActorContext) {
		actor.recipients.Do(func(e interface{}) {
			if a, ok := e.(*Actor); ok {
				a.Send(msg)
			}
			if a, ok := e.(*ForwardingActor); ok {
				a.Send(msg)
			}
		})
	}
}

