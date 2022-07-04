# RTGEN
RTGEN is Relative Temporal Graph GENerator which produces a sequence of graph snapshots given the parameters of an evolutionary degree distribution and community structure.

That is, we propose a relative generation mechanism that is able to control the evolution of the underlying graph's degree distribution. In particular, we propose to generate new graphs with a desired degree distribution out of existing ones while minimizing the efforts to transform our source graph to target. Our proposed relative graph generation method relies on optimal transport methods. We extend our method to also deal with the community structure of the generated graphs that is of high importance for a number of applications.  Our generation model extends the concepts proposed in the Chung-Lu model with a temporal and community-aware support.

----


# HOW TO USE RTGEN?

**Prerequisites:**

Go version = 1.17.5

----


**Steps to use the tool:**

1. Clone this repo:
```
> git clone https://github.com/MariaMassri/RTGEN.git
```
2. In the project directory, launch the following command to import all go modules:
```
> go mod tidy
```
3. Build:
```
> go build
```
4. Execute:
```
> ./RTGEN [args]
```

| args                            | description                                   |
| ------------------------------- | --------------------------------------------- |
| -outputPath                     | Path of the output graphs                     |
| -task                           | Option of the generation procedure            |
| -nbVertices                     | Initial number of vertices                    |
| -mu, -sigma, -dMu, -dSigma      | Parameters of the Gaussian distribution       |
| -s, -v, -iMax, -dS, -dV, -dImax | Parameters of the Zipfian distribution        |
| -h, -b, -k                      | parameters of the hierarchical community tree |

----

**Static graph generation with a given degree distribution**

**Output files format:**
Using this generation option, an edgeFile is generated.  The edgeFile is a **.csv** file with the format **sourceVertexID;targetVertexID**.

To generate a static graph with a Gaussian distribution, you can use the following commands:
```
> ./RTGen -task="GenStatGauss" -nbVertices=1000000 -mu=30 -sigma=5 -outputPath="outputDatasets\statGauss\graph.csv"
```

To generate a static graph with a zipfian distribution, you can use the following commands:
```
> ./RTGen -task="GenStatZipf" -outputPath="outputDatasets\statZipf\" -nbVertices=1000000 -s='1.5' -v='10.0' -iMax=30
```
----


**Static graph generation with a given degree distribution and community structure**

***Static graph generation with a given degree distribution and community structure


**Output files format:**
Using this generation option, two output files are generated: vertexFile and edgeFile.  The vertex file is a **.csv** file with format **vertexID;communityID**, that is, each vertex is assigned with the graph community to which it belongs. Now, edgeFile is a **.csv** file with format **sourceVertexID;targetVertexID**.

To generate a static graph with a given Gaussian degree distribution and hierarchical community structure use the following commands:

```
> ./RTGen -task="GenStatGaussCom" -nbVertices=1000000 -mu=50 -h=4 b=2

```

To generate a static graph with a given Zipfian degree distribution and hierarchical community structure use the following commands:

```
> ./RTGen -task="GenStatZipfCom" -nbVertices=1000000 -iMax=30 -h=4 b=2 -outputPath="outputDatasets\statZipfComs\"

```
----

**Temporal relative graph generation with a given evolutionary degree distribution and community structure**


**Output files format:**
We propose two formats for the generated graphs:

1. Events Format: It returns the edge addition operations as a series of events. Indeed, it combines all the events that occurried between all the generated files in a single file (edgeFile.csv having the format **sourceID;targetID;timestamp**). Besides, a vertex file is also generated (vertexFile.csv having the format **vertexID;comID**).
2. Snapshots Format: Consider N the number of snapshots, It returns N edgeFile.csv such that each file corresponds to a graph snapshot. These files have the format 'sourceID;targetID'. Besides, a vertex file is also generated (vertexFile.csv having the format **vertexID;comID**).

To generate a temporal graph with an evolutionary sequence of Gaussian degree distributions and a hierarchical community structure, use the following command:
```
> ./RTGen -task="GenDynGaussCom" -nbVertices=1000000 -mu=50 -h=4 b=2 -dMu=5 -outputPath="outputDatasets\dynGaussComs\"

```
The parameter dMu=5 implies that the average edge degree will increase by 5 between each pair of consecutive snapshots.  
Note that, by default the returned graphs are generated according to the events format.

Now, to generate a temporal graph with an evolutionary sequence of Zipfian degree distributions and a hierarchical community structure, use the following command:

```
> ./RTGen -task="GenDynZipfCom" -nbVertices=10000 -nbSnapshots=5 -iMax=30 -h=4 b=2 -dImax=5 -outputPath="outputDatasets\statZipfComs\" -outputFormat="snapshots"
```
The parameter dImax=5 implies that the maximum edge degree will increases by a value of 5 between each pair of successive snapshots.

