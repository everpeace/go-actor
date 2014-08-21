package actor

import "github.com/dropbox/godropbox/container/set"

// this maintains set and map simultaneously.
type actorSet struct{
	s set.Set
	m map[string]*Actor
}

func newActorSet(s set.Set) *actorSet{
	actorS := set.NewSet()
	s.Do(func(v interface{}){
		if actor, ok := v.(*Actor); ok{
			actorS.Add(actor)
		}
	})
	return &actorSet{
		s: actorS,
		m: actorSet2map(actorS),
	}
}

func actorSet2map(s set.Set) map[string]*Actor {
	m := make(map[string]*Actor)
	s.Do(func(e interface{}){
		if actor, ok := e.(*Actor); ok {
			m[actor.Name] = actor
		}
	})
	return m
}

func (as *actorSet) Len() int{
	return as.s.Len()
}


func (as *actorSet) Add(a *Actor){
	as.s.Add(a)
	as.m[a.Name] = a
}

func (as *actorSet) Remove(a *Actor) bool{
	r := as.s.Remove(a)
	delete(as.m, a.Name)
	return r
}

func (as *actorSet) Do(f func(actor *Actor)){
	as.s.Do(func(v interface{}){
		if a, ok := v.(*Actor); ok{
			f(a)
		}
	})
}

func (as *actorSet) Subtract(s set.Set){
	s.Do(func(v interface{}){
		if a, ok := v.(*Actor); ok{
			as.Remove(a)
		}
	})
}


