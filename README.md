# onedari

Onedari is simple service discovery.

# Data Model #

The data model is simple. There are three types of entities: _nodes_, _instances_,
and _services_.

- _nodes_ a "server"
- _instances_ represent a single instance of an application. An
  instance has an address and a port  associated with it. "belongs" to a node
- _services_ represent a collection of instances. instances are
  associated with a service via label queries.

 _instances_ and _services_ have _labels_. Labels are just
key/value pairs (strings only) and have no inherent meaning to the
system.  A service has a query defined which is also key/value
pairs. These are used to query the instances.  Note: all query keys
and values must match.

Label queries were inspired by Kubernetes.

# Installation #

```
git clone https://github.com/bakins/onedari.git
cd cmd/onedari
go get .
go build
```

`onedari` will now be in `$GOPATH/bin`


# Modes #
onedari is a single binary that implements the following "modes":

- _server_ A simple, stateless HTTP API server. It uses etcd as its
  backing store. All other "modes" use this API - nothing else
  communicates with etcd
- _dns_ A simple DNS server for a single (sub)domain.
- _announce_ A helper used to "announce" a single instance.

All of these are lightweight and can be ran on every node.  It is
suggested to run an etcd proxy on every node as well.


# Example #

A quick example.

Assuming you have etcd running on localhost.

Start the server:
```
$ onedari server
time="2015-04-14T16:42:39-04:00" level=warning msg="truncating hostname leoben.local to leoben"
```

Now, use curl to see the node:
```
$ curl -s http://127.0.0.1:63412/v0/nodes
[{"id":"leoben","ip":"10.191.63.72"}]
```

Now, announce an instance of an app. While leaving the server running,
do:
```
$ onedari announce foo
time="2015-04-14T16:44:44-04:00" level=warning msg="truncating hostname leoben.local to leoben"
```

Now, use curl to see the instance:
```
 curl -s http://127.0.0.1:63412/v0/instances
 [{"id":"leoben-foo","node":"leoben","labels":{"app":"foo"},"ip":"10.191.63.72","port":0,"target":"","up":true}]
 ```

Notice that it automatically added the label `{"app":"foo"}`.

Ctrl+C the announce. Now, let's add some additional labels:
```
$ onedari announce foo track=dev
time="2015-04-14T16:47:15-04:00" level=warning msg="truncating hostname leoben.local to leoben"
```

Now, curl the instance:
```
$ curl -s http://127.0.0.1:63412/v0/node/instances
[{"id":"leoben-foo","node":"leoben","labels":{"app":"foo","track":"dev"},"ip":"10.191.63.72","port":0,"target":"","up":true}]
```

Notice the additional label.

Now, add a service:
```
$ curl  -H "Content-Type: application/json" -svo /dev/null -X PUT --data-binary '{"labels":{"hello":"world"}, "query": {"app":"foo"}}'  http://127.0.0.1:63412/v0/services/foo
```

Using curl on `/services` will give us a summary view:
```
$ curl http://127.0.0.1:63412/v0/services
[{"id":"foo","labels":{"hello":"world"},"query":{"app":"foo"}}]
```

While, getting the actual service will give us the instances:
```
$ curl http://127.0.0.1:63412/v0/services/foo
{"id":"foo","labels":{"hello":"world"},"query":{"app":"foo"},"instances":[{"id":"leoben-foo","node":"leoben","labels":{"app":"foo","track":"dev"},"ip":"10.191.63.72","port":0,"target":"","up":true}]}
```

Notice that the instance we created earlier is associated with the
service via the label `{"app":"foo"}}` that we set as the query for
the service.

We can also query instances by label as well:
```
$ curl -s http://127.0.0.1:63412/v0/node/instances?track=dev
[{"id":"leoben-foo","node":"leoben","labels":{"app":"foo","track":"dev"},"ip":"10.191.63.72","port":0,"target":"","up":true}]
```
```
 $ curl -s http://127.0.0.1:63412/v0/node/instances?track=prod
 []
 ```

We can also search "globally." `/v0/node/` only gets/sets for
instances associated with the local node. `/v0/instances` is across
all nodes - this is effectively what the services use internally.


## Server ##

## DNS ##

## Announce ##

