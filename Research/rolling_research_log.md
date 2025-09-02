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


Also was reading some of Designing Data-Intensive Applications, I might try and read through more of it or at least a section every few days because its entirely about Distributed Systems.

Oh and I refreshed myself on go a tiny bit and did some network calls

--------