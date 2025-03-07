---
slug: /overview
sidebar_position: 2
---

# Overview
*An introduction to Apollo and its distributed computing model*\
This page serves as an introduction to how the Apollo framework works under the hood.\
The computation model that Apollo follows is the MapReduce model where a global computation is subdivided into two types of operations which are map operations and reduce operations. These operations are distributed on multiple kubernetes pods that perform their specific operations on the data chunks that are given to them as a responsibility.\
In addition to following the MapReduce model, Apollo is kubernetes native which means that it is directly grafted on top of the k8s abstractions without any added configuration or any customization effort.
## Architecture 
The Apollo architecture is based on 5 core components that interact with each other in order to have a computation happen on the cluster:
<p align="center" >
  <img src= {require('./assets/architecture.png').default} height="500" width="auto" />
</p>  

### Coordinator
The coordinator is the main component that the users and the other components interact with. It's goal is to coordinate/manage the different computation operations that happen while interfacing with the different kubernetes abstractions that are used such as pods, PVCs and services. It exposes an HTTP server that the users can interact with in order to start jobs (cf. Jobs & Tasks) and it also keeps track of all the binary executables that it gets sent in the form of artifacts that the workers execute.\
The coordinator has the responsibility of creating the worker pods and assigning them tasks and in order to do that it makes use of a gRPC server side streaming communication where the coordinator plays the role of a client that keeps receiving data from the fleet of workers it communicates with.\
In general the coordinator must be seen as the central brain of all the computations that happen in Apollo. 
### Worker
A worker plays the role of the component that works on atomic computation assigned to it. A worker can either be a mapper or a reducer. However, it is agnostic by nature because it takes the responsibility of an operation type till it receives it through the form of a gRPC request. Each worker exposes a gRPC server that interacts with the coordinator through a server side streaming communication. Workers use linux domain sockets internally to communicate with the executables received from the coordinator. In addition, they communicate with both an S3 object storage in order to fetch or write data (mappers fetch the data and reducers write it in this case) and a persistent volume to handle intermediate computation files.\

:::info

Intermediate computation files are the files that are created from a map computation. The reducers use them as an input.

:::

### Intermediate Files Persistent Volume
The intermediate files persistent volume is a volume that contains all of the intermediate files that computations generate. Every worker can read and write from it without exception. 

## Core Concepts
On this section, we will be exposing the different abstractions that apollo provides in order to perform a distributed map reduce computation.
### Jobs
A job is the primary abstraction that users deal when wanting to perform a distributed computation. It gets subsivided into tasks which are bound to workers.
### Tasks
A Task is bound to a job and is either a map or reduce task. Each task has also a 1 to 1 relationship to worker.
### Artifacts
An artifact represents a binary executable of a program that gets executed by a task.
