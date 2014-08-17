package actor

import (
	"fmt"
	"sync"

	"strings"

	"github.com/dropbox/godropbox/container/set"
)

type actorSystem struct {
	//TODO top level supervisor
	name              string
	wg                sync.WaitGroup
	topLevelActors    set.Set
	monitorForwarders set.Set
	running           set.Set
	stopped           set.Set
}

func NewActorSystem(name string) *actorSystem {
	return &actorSystem{
		name:              name,
		topLevelActors:    set.NewSet(),
		monitorForwarders: set.NewSet(),
		running:           set.NewSet(),
		stopped:           set.NewSet(),
	}
}

func (system *actorSystem) Name() string {
	return system.name
}

// actor starter
func (system *actorSystem) Spawn(receive Receive) Actor {
	newName := system.canonicalName(fmt.Sprint(system.topLevelActors.Len()))
	startLatch, actor := system.spawnActor(system.newTopLevelActorImpl(newName, receive))
	startLatch <- true
	return actor
}

func (system *actorSystem) SpawnWithName(name string, receive Receive) Actor {
	newName := system.canonicalName(name)
	startLatch, actor := system.spawnActor(system.newTopLevelActorImpl(newName, receive))
	startLatch <- true
	return actor
}

func (system *actorSystem) SpawnWithLatch(receive Receive) (chan bool, Actor) {
	newName := system.canonicalName(fmt.Sprint(system.topLevelActors.Len()))
	startLatch, actor := system.spawnActor(system.newTopLevelActorImpl(newName, receive))
	return startLatch, actor
}

func (system *actorSystem) SpawnWithNameAndLatch(name string, receive Receive) (chan bool, Actor) {
	newName := system.canonicalName(name)
	startLatch, actor := system.spawnActor(system.newTopLevelActorImpl(newName, receive))
	return startLatch, actor
}

func (system *actorSystem) SpawnForwardActor(name string, actors ...Actor) ForwardingActor {
	s := set.NewSet()
	for _, actor := range actors {
		s.Add(actor)
	}
	forwardActor := &forwardingActor{
		system.newTopLevelActorImpl(system.canonicalName(name), forward(s)),
	}
	start := forwardActor.context.start()
	start <- true
	return forwardActor
}

func (system *actorSystem) WaitForAllActorsStopped() {
	system.shutdownMonitorForwarder()
	system.wg.Wait()
}

func (system *actorSystem) Shutdown() {
	system.topLevelActors.Subtract(system.monitorForwarders)
	system.topLevelActors.Subtract(system.stopped)
	for r := range system.topLevelActors.Iter() {
		if actor, ok := r.(*actorImpl); ok {
			actor.context.shutdown()
		}
	}
	system.shutdownMonitorForwarder()
	system.wg.Wait()
}

func (system *actorSystem) shutdownMonitorForwarder() {
	for r := range system.monitorForwarders.Iter() {
		if actor, ok := r.(*forwardingActor); ok {
			actor.Shutdown()
		}
	}
}
func (system *actorSystem) spawnMonitorForwarderFor(actor *actorImpl) ForwardingActor {
	name := strings.Replace(actor.Name(), "/"+system.name+"/", "", 1)
	forwarder := system.SpawnForwardActor(name)
	system.monitorForwarders.Add(forwarder)
	return forwarder
}

func (system *actorSystem) newTopLevelActorImpl(name string, receive Receive) *actorImpl {
	actor := &actorImpl{name: name}
	actor.context = newTopLevelActorContext(system, actor, receive)
	system.topLevelActors.Add(actor)
	return actor
}

func (system *actorSystem) spawnActor(actor *actorImpl) (chan bool, *actorImpl) {
	startLatch := actor.context.start()
	return startLatch, actor
}

func (system *actorSystem) canonicalName(name string) string {
	return "/" + system.name + "/" + name
}
