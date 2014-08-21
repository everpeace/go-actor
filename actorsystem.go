package actor

import (
	"fmt"
	"sync"

	"strings"

	"github.com/dropbox/godropbox/container/set"
	"time"
)

// ActorSystem is an umbrella which maintains actor hierarchy.
type ActorSystem struct {
	//TODO top level supervisor
	Name              string
	wg                sync.WaitGroup
	topLevelActors    set.Set
	monitorForwarders set.Set
	running           set.Set
	stopped           set.Set
}

// NewActorSystem creates an ActorSystem instance.
func NewActorSystem(name string) *ActorSystem {
	return &ActorSystem{
		Name:              name,
		topLevelActors:    set.NewSet(),
		monitorForwarders: set.NewSet(),
		running:           set.NewSet(),
		stopped:           set.NewSet(),
	}
}


// Spawn creates and starts an actor in the actor system.
//
// This takes Receive(type ailis for actor's message handler) returns the pointer to started actor.
//
// For example, simple echo actor would be:
//  actorSystem.Spawn(func(msg Message, context *ActorContext){
//     fmt.Println(msg)
//  })
func (system *ActorSystem) Spawn(receive Receive) *Actor {
	newName := system.canonicalName(fmt.Sprint(system.topLevelActors.Len()))
	startLatch, actor := system.spawnActor(system.newTopLevelActor(newName, receive))
	startLatch <- true
	return actor
}

// SpawnWithName is the same as Spawn except that you can name it.
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

// WaitForAllActorsStopped waits for all the actors in the actor system stopped(terminated or killed).
func (system *ActorSystem) WaitForAllActorsStopped() {
	system.shutdownMonitorForwarder()
	system.wg.Wait()
}

// Shutdown the actor system.
//
// It sends kill signal(Kill() method) to all the actors in the actor system.
func (system *ActorSystem) Shutdown() {
	system.ShutdownIn(time.Duration(0))
}

// ShutdownIn shutdowns the actor system after waiting a given duration.
//
// It sends kill signal(Kill() method) to all the actors in the actor system after waiting for a given duration.
func (system *ActorSystem) ShutdownIn(duration time.Duration){
	<-time.After(duration)
	system.topLevelActors.Subtract(system.stopped)
	system.topLevelActors.Do(func (r interface{}) {
		if actor, ok := r.(*Actor); ok {
			actor.context.kill()
		}
	})
	system.shutdownMonitorForwarder()
	system.wg.Wait()
}

// GracefulShutdown shutdowns the actor system in graceful manner.
//
// It sends terminate signal(Terminate() method) to all the actors in the actor system.
func (system *ActorSystem) GracefulShutdown() {
	system.GracefulShutdownIn(time.Duration(0))
}

// GracefulShutdownIn shutdowns the actor system in graceful manner after waiting a given duration.
//
// It sends terminate signal(Terminate() method) to all the actors in the actor system after waiting for a given duration.
func (system *ActorSystem) GracefulShutdownIn(duration time.Duration) {
	<-time.After(duration)
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
