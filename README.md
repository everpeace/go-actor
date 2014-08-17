go-actor
========

This is far far incomplete actor implementation in golang. This is only for my golang learning.

go-actor now supports:
* become/unbecome
* actor hierarchy (actor has children. but supervisor is comming soon.)
* forwarding actor (this actor forwards all messages other actors.)

If you tried to run this, please just hit.

```
$ cd example
$ go run ping-pong/main.go
==========================================================
== Ping-Pong example.
/ping-pong/pinger sends : Ping
/ping-pong/ponger received: Ping
/ping-pong/ponger sends : Pong
/ping-pong/pinger receives: Pong.
Ping-Pong finished.
==========================================================

$ go run become-unbecome/main.go
==========================================================
== Become/UnBecome example
== An actor's behavior changes:
== echo --(become)-->echo in upper case --(unbecome)--> echo.
/become-unbecome/0: this should be echoed..  become echoInUpper.
/become-unbecome/0: THIS SHOULD BE ECHOED IN UPPER CASE..  unbecome
/become-unbecome/0: become terminate!!.  become terminate
==========================================================

$ go run forward/main.go
==========================================================
== Forwarding actor example
== It is sometimes useful that an actor which just forward
== messages other actors.  In this example,
== "forward" actor forwards to "echo1" and "echo2"
Sent [hello] to "forward"
/forwad/echo1 : [hello]
/forwad/echo2 : [hello]
==========================================================


$ go run monitor/main.go
==========================================================
== Monitor example
== Monitor can detect actor's termination.
== In this example, spawn "traget" actor and it is monitored
== by "monitor"
/monitor-system/target receive: [hello]
/monitor-system/monitor detects: /monitor-system/target terminated
==========================================================

$ go run parent-child/main.go
==========================================================
== Actor hierarchy example
== Actor has children.  Children's termination is propagated
== to parent.
child receive: [hello]
actorA receive: [hello]
terminate child. you will see parent's termination too.
/parent-child/monitor detects: /parent-child/actorA/child terminated
/parent-child/monitor detects: /parent-child/actorA terminated
==========================================================
```
