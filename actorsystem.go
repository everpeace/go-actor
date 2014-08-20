package actor

import (
	"fmt"
	"sync"

	"strings"

	"github.com/dropbox/godropbox/container/set"
)

type ActorSystem struct {
	//TODO top level supervisor
	Name              string
	wg                sync.WaitGroup
	topLevelActors    set.Set
	monitorForwarders set.Set
	running           set.Set
	stopped           set.Set
}

// ActorSystem contstructor.
func NewActorSystem(name string) *ActorSystem {
	return &ActorSystem{
		Name:              name,
		topLevelActors:    set.NewSet(),
		monitorForwarders: set.NewSet(),
		running:           set.NewSet(),
		stopped:           set.NewSet(),
	}
}


// actor starter
func (system *ActorSystem) Spawn(receive Receive) *Actor {
	newName := system.canonicalName(fmt.Sprint(system.topLevelActors.Len()))
	startLatch, actor := system.spawnActor(system.newTopLevelActor(newName, receive))
	startLatch <- true
	return actor
}

func (system *ActorSystem) SpawnWithName(name string, receive Receive) *Actor {
	newName := system.canonicalName(name)
	startLatch, actor := system.spawnActor(system.newTopLevelActor(newName, receive))
	startLatch <- true
	return actor
}

func (system *ActorSystem) SpawnForwardActor(name string, actors ...*Actor) *ForwardingActor {
	s := set.NewSet()
	for _, actor := range actors {
		s.Add(actor)
	}
	forwardActor := &ForwardingActor{
		addRecipientChan: make(chan addRecipient),
		delRecipientChan: make(chan removeRecipient),
		recipients: s,
	}
	forwardActor.Actor = system.newTopLevelActor(system.canonicalName(name), forwardActor.receive())
	forwardActor.context.prePrecessHook = forwardActor.preProcessHook()
	start := forwardActor.context.start()
	start <- true
	return forwardActor
}

func (system *ActorSystem) WaitForAllActorsTerminated() {
	system.shutdownMonitorForwarder()
	system.wg.Wait()
}

func (system *ActorSystem) ShutdownNow() {
	system.topLevelActors.Subtract(system.stopped)
	system.topLevelActors.Do(func (r interface{}) {
		if actor, ok := r.(*Actor); ok {
			actor.context.kill()
		}
	})
	system.shutdownMonitorForwarder()
	system.wg.Wait()
}

func (system *ActorSystem) GracefulShutdown() {
	system.topLevelActors.Subtract(system.stopped)
	system.topLevelActors.Do(func (r interface{}) {
		if actor, ok := r.(*Actor); ok {
			actor.context.terminate()
		}
	})
	system.shutdownMonitorForwarder()
	system.wg.Wait()
}

func (system *ActorSystem) shutdownMonitorForwarder() {
	system.monitorForwarders.Do(func(r interface{}) {
		if actor, ok := r.(*ForwardingActor); ok {
			actor.context.terminate()
		}
	})
}
func (system *ActorSystem) spawnMonitorForwarderFor(actor *Actor) *ForwardingActor {
	name := strings.Replace(actor.Name, "/"+system.Name+"/", "", 1)
	forwarder := system.SpawnForwardActor(name+"-MonitorForwarder")
	system.topLevelActors.Remove(forwarder.Actor)
	system.monitorForwarders.Add(forwarder)
	return forwarder
}

func (system *ActorSystem) newTopLevelActor(name string, receive Receive) *Actor {
	actor := &Actor{Name: name, System: system }
	actor.context = newActorContext(nil, actor, receive)
	system.topLevelActors.Add(actor)
	return actor
}

func (system *ActorSystem) spawnActor(actor *Actor) (chan bool, *Actor) {
	startLatch := actor.context.start()
	return startLatch, actor
}

func (system *ActorSystem) canonicalName(name string) string {
	return "/" + system.Name + "/" + name
}
