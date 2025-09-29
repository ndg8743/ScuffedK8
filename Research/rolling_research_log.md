--9-2-25--
“Kubernetes” also known as K8s is an open-source container orchestration system for automating software deployment, scaling, and management

Or in simple terms, a group of computers that each handle a single job

Cool fact, “Kubernetes” is the Greek word for a ship’s captain. So if you think about Docker as a shipping container the idea kind of works. 

For right now I get k8's as this:

```
┌───────────────────────────────────┐
│         Kubernetes Cluster        │
├───────────────────────────────────┤
│           Control Plane           │
│    (API Server, Scheduler, etc.)  │
├───────────────────────────────────┤
│                                   │
│           Worker Node 1           │
│  ┌─────────────────────────────┐  │
│  │       Kubelet, Proxy        │  │
│  ├─────────────────────────────┤  │
│  │    Docker (Containerd)      │  │
│  ├─────────────────────────────┤  │
│  │      Docker Image           │  │
│  │   (Application + Deps)      │  │
│  └─────────────────────────────┘  │
├───────────────────────────────────┤
│                                   │
│           Worker Node 2           │
│  ┌─────────────────────────────┐  │
│  │       Kubelet, Proxy        │  │
│  ├─────────────────────────────┤  │
│  │    Docker (Containerd)      │  │
│  ├─────────────────────────────┤  │
│  │      Docker Image           │  │
│  │   (Application + Deps)      │  │
│  └─────────────────────────────┘  │
└───────────────────────────────────┘
```



But there is a lot more pieces:

![Kubernetes 101 Architecture Diagram](https://www.aquasec.com/wp-content/uploads/2020/11/Kubernetes-101-Architecture-Diagram.jpg)

api: Stores interface protocols
build: Code related to building applications
cmd: main entry points for each application
pkg: Main implementation of each component
staging: Temporarily stores code that is interdependent among components

For right now I am thinking about the project structure for how I think it should be then how it really is implemented. I am unsure of which Go libraries I will or won't use.

```
src
├── cmd/
│   ├── apiserver/main.go
│   ├── scheduler/main.go
│   ├── kubelet/main.go
│   └── kubectl-lite/main.go
├── pkg/
│   ├── api/
│   │   ├── types.go
│   │   └── client.go
│   └── store/
│       ├── store.go
│       └── memory.go
└── Makefile
```

Also was reading some of Designing Data-Intensive Applications, I might try and read through more of it or at least a section every few days because its entirely about Distributed Systems.

Oh and I refreshed myself on go a tiny bit and did some network calls

--9-13-25--

Was really finding it hard to just read documentation alone, needed to make something basic and also dive into implementing the scheduler algo. I might make it a cli based ASCII bowling thing to strive for a goal that isn't the end goal. Idea being that the Pinsetter is the master node with the pins being the worker nodes that you simulate a node going down.

I will probably implement Greedy for now as I can optimize this a bunch later.

https://arxiv.org/abs/2203.00671
https://arxiv.org/pdf/2504.17033
https://en.wikipedia.org/wiki/Minimum-cost_flow_problem
https://en.wikipedia.org/wiki/Fibonacci_heap
https://moderndata101.substack.com/p/the-new-dijkstras-algorithm-shortest

In terms of understanding go, there will be pieces I miss or don't get, and coming from Java and C its pretty similar. But the majority of my examples I keep referring back to were great from here:
https://gobyexample.com/

Anyway I should just start

--9-14-25--

Last night I did a basic implementation of Greedy:
    Sort pods by dominant resource (biggest first), then by priority (highest first)

Referenced docs and also this yt series I found for more Go clarification:

https://youtube.com/playlist?list=PLmD8u-IFdreyh6EUfevBcbiuCKzFk0EW_&si=RlCccO4JSxeZC-7Y

You could also do round robin but also trying to understand http scheduling as well which I think that might use round robin

So to me this is how I understand it:

master node/contol plane runs an api server that has http requests controlled by mux's (simlar idea to electronics in idea but imagine for http)

kubectl is the command-line tool that talks to the api server and lets you control the cluster 

scheduler decides which worker node a new pod should run on

controller manager watches the state of the cluster and makes sure it matches what you declared (e.g., restart pods if they crash).

etcd is the key-value database that stores all cluster state.

worker nodes actually run your containers inside pods. Each worker has a kubelet agent that talks to the API server, and a kube-proxy that manages networking.

Node<Pod<Container (usually)

what trips me up is that, what the health api would consider (network, ram, disk, etc), and all the toggles that real k8's has.

I will add my full Kub setup to here to test features as well as further develop mine.

--9-27-25--
https://minikube.sigs.k8s.io/docs/tutorials/

Great tutorial I went through for just how to setup minikube and it and found some info about how they still use TCP and UDP services with the nginx.

Will experiment more with the other options inside it as well, https://minikube.sigs.k8s.io/docs/tutorials/nvidia/


--9-28-25--
I got the basic greedy scheduler working. Just sorts pods by resource requirements and finds the first node that can fit them. 

Added Gorilla mux and WebSocket connections so nodes can actually talk to the scheduler. Right now it's just basic connection tracking but the foundation is there for real-time communication between the scheduler and worker nodes.

--9-29-25--

```
Client (ws_demo.go)          Server (main.go)
      |                            |
      |--- WebSocket Handshake  -->|  (HTTP Upgrade)
      |<-- 101 Switching Protocols-|
      |                            |
      |                            |--- Store in nodeConnections map
      |                            |--- Log "Node worker-1 connected"
      |                            |
      |--- Heartbeat JSON -------->|  (Every 2 seconds)
      |                            |--- Log "[WS] worker-1 -> server..."
      |                            |
      |<-- Pod Assignment ---------|  (Future feature)
      |                            |
      |--- Close Connection ------>|
      |                            |--- Remove from map
```

Updates:
- Added `/nodes` endpoint that returns JSON of connected worker nodes
- Added mux routing logs with `[MUX]` prefix to see request handling
- Added WebSocket message logging with `[WS]` prefix to show heartbeat data