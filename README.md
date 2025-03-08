 # Apollo
<p align="center" >
  <img src= "assets/apollo-social-card.png" height="342" width="auto" />
</p>
<pre align="center">
A lightweight modern map reduce framework brought to k8s
</pre>

Apollo is a lightweight modern kubernetes native map reduce framework based on the original [Google MapReduce paper](https://research.google.com/archive/mapreduce-osdi04.pdf).\
Apollo provides a distributed computation framework grafted on top of the kubernetes orchestrator while requiring minimal configuration and staying lightweight. It mainly relies on S3 based object storages as input sources instead of bulky distributed filesystems such as HDFS or GFS.

The computation model that Apollo follows is the MapReduce model where a global computation is subdivided into two types of operations which are map operations and reduce operations. These operations are distributed on multiple kubernetes pods that perform their specific operations on the data chunks that are given to them as a responsibility.
In addition to following the MapReduce model, Apollo is kubernetes native which means that it is directly grafted on top of the k8s abstractions without any added configuration or any customization effort.

For more details on how Apollo works and how to get started with it check our [docs](https://assifar-karim.github.io/apollo).

<pre align="center">
Made with ❤️ by your friendly neighborhood software engineer Karim Assifar
</pre>

