package actor

// actor starter
func Spawn(receive Receive) Actor {
	actor := newActorImpl(receive)
	actor.context.self = actor
	startLatch := actor.context.start()
	startLatch <- true
	return actor
}

type actorImpl struct {
	context *actorContext
}

func (actor *actorImpl) Send(msg Message) {
	go func() {
		actor.context.mailbox <- msg
	}()
}

func (actor *actorImpl) Terminate() {
	go func() {
		actor.context.terminateChan <- terminate{}
	}()
}

func (actor *actorImpl) Kill() {
	go func() {
		actor.context.killChan <- kill{}
	}()
}

// private constructors
func newActorImpl(receive Receive) *actorImpl {
	return &actorImpl{
		context: newActorContext(receive),
	}
}
