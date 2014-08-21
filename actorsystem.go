package actor

import (
	"fmt"
	"sync"
	"time"

	"strings"

	"github.com/dropbox/godropbox/container/set"
)

// ActorSystem is an umbrella which maintains actor hierarchy.
type ActorSystem struct {
	//TODO top level supervisor
	Name              string
	wg                sync.WaitGroup
	guardian          *Actor
	topLevelActors    *actorSet
	monitorForwarders set.Set
	running           set.Set
	stopped           set.Set
}

// NewActorSystem creates an ActorSystem instance.
func NewActorSystem(name string) *ActorSystem {
	actorSystem :=  &ActorSystem{
		Name:              name,
		topLevelActors:    newActorSet(set.NewSet()),
		monitorForwarders: set.NewSet(),
		running:           set.NewSet(),
		stopped:           set.NewSet(),
	}
	actorSystem.guardian = newGuardian(actorSystem)
	return actorSystem
}

// root guardian
func newGuardian(system *ActorSystem) *Actor{
	actor := &Actor{
		Name: "ROOT_GUARDIAN",
		System: system,
		parent: nil,
		children: newActorSet(set.NewSet()),
	}
	actor.context = newActorContext(actor, func(msg Message, context *ActorContext){
		//nop
	})
	latch := actor.context.start()
	latch <- true
	return actor
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
	name := fmt.Sprint(system.topLevelActors.Len())
	return system.SpawnWithName(name, receive)
}

// SpawnWithName is the same as Spawn except that you can name it.
func (system *ActorSystem) SpawnWithName(name string, receive Receive) *Actor {
	startLatch, actor := system.spawnActor(system.newTopLevelActor(name, receive))
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
	forwardActor.Actor = system.newTopLevelActor(name, forwardActor.receive())
	forwardActor.context.prePrecessHook = forwardActor.preProcessHook()
	start := forwardActor.context.start()
	start <- true
	return forwardActor
}

// WaitForAllActorsStopped waits for all the actors in the actor system stopped(terminated or killed).
func (system *ActorSystem) WaitForAllActorsStopped() {
	system.internalShutdown()
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
	system.topLevelActors.Do(func (actor *Actor) {
			actor.context.kill()
	})
	system.WaitForAllActorsStopped()
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
	system.topLevelActors.Do(func (actor *Actor) {
		actor.context.terminate()
	})
	system.WaitForAllActorsStopped()
}

func (system *ActorSystem) internalShutdown(){
	system.terminateMonitorForwarder()
	system.guardian.Terminate()
}

func (system *ActorSystem) terminateMonitorForwarder() {
	system.monitorForwarders.Do(func(r interface{}) {
		if actor, ok := r.(*ForwardingActor); ok {
			actor.context.terminate()
		}
	})
}
func (system *ActorSystem) spawnMonitorForwarderFor(actor *Actor) *ForwardingActor {
	name := strings.Replace(actor.CanonicalName(), system.Name+":/", "", 1)
	name = strings.Replace(name, "/", "-", 1)
	forwarder := system.SpawnForwardActor(name+"_MonitorForwarder")
	system.topLevelActors.Remove(forwarder.Actor)
	system.monitorForwarders.Add(forwarder)
	return forwarder
}

func (system *ActorSystem) newTopLevelActor(name string, receive Receive) *Actor {
	actor := system.guardian.newChildActor(name, receive)
	system.topLevelActors.Add(actor)
	return actor
}

func (system *ActorSystem) spawnActor(actor *Actor) (chan bool, *Actor) {
	startLatch := actor.context.start()
	return startLatch, actor
}
