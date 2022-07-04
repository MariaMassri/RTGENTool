package extendedChungLu

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"sync"
)

type vertex struct {
	id               int
	comID            int
	d                int
	w                float64
	neighbors        map[int]*vertex
	tombstones       int
	neighborsIDs     []int
	fixObjectsIDs    []int
	mobileObjectsIDs []int
	label            string
	properties       map[string]interface{}
	mutex            sync.RWMutex
}

type graph struct {
	nbEdges               int
	vertices              map[int]*vertex
	degreeDistribution    map[int]int
	degreeHasDistribution map[int]int
	givenDistribution     map[int]int
	degreeGraph           map[int][]*vertex //maps each degree to the set of vertices with the corresponding degree
	degreeHasGraph        map[int][]*vertex
}

func (graph *graph) getDegreeGraph() {
	graph.degreeGraph = make(map[int][]*vertex)
	for _, vertex := range graph.vertices {
		graph.degreeGraph[vertex.d] = append(graph.degreeGraph[vertex.d], vertex)
	}
}

func (graph *graph) getHasDegreeGraph() {
	graph.degreeHasGraph = make(map[int][]*vertex)
	for _, vertex := range graph.vertices {
		d := len(vertex.fixObjectsIDs) + len(vertex.mobileObjectsIDs)
		graph.degreeHasGraph[d] = append(graph.degreeHasGraph[d], vertex)
	}
}

func getTotDegree(graph *graph) (totNbEdges int) {
	for _, vertex := range graph.vertices {
		totNbEdges += vertex.d
	}
	return
}

func initVertex(id int, deg int) *vertex {
	return &vertex{
		id:        id,
		w:         float64(deg),
		d:         0,
		neighbors: make(map[int]*vertex),
		mutex:     sync.RWMutex{},
	}
}

func initGraph() *graph {
	return &graph{
		vertices:           make(map[int]*vertex),
		degreeDistribution: make(map[int]int),
		givenDistribution:  make(map[int]int),
		degreeGraph:        make(map[int][]*vertex),
	}
}
func (graph *graph) addVertex(vertex *vertex) {
	graph.vertices[vertex.id] = vertex
	graph.nbEdges += int(vertex.w)
}

func (source *vertex) addNeighbor(target *vertex) bool {
	source.neighbors[target.id] = target
	source.neighborsIDs = append(source.neighborsIDs, target.id)
	target.neighbors[source.id] = source
	target.neighborsIDs = append(target.neighborsIDs, source.id)
	return true
}

func (source *vertex) addNeighborRepeat(target *vertex) bool {
	source.neighbors[target.id] = target
	source.neighborsIDs = append(source.neighborsIDs, target.id)
	source.d++
	target.neighbors[source.id] = source
	target.neighborsIDs = append(target.neighborsIDs, source.id)
	target.d++
	return true
}

func (user *vertex) addMobileObject(id int) {
	user.mobileObjectsIDs = append(user.mobileObjectsIDs, id)
}

func (user *vertex) addFixObject(id int) {
	user.fixObjectsIDs = append(user.fixObjectsIDs, id)
}

func (graph *graph) getDistributionDeletions() {

	graph.degreeDistribution = make(map[int]int)
	for _, vertex := range graph.vertices {
		degree := vertex.d
		//degree := len(vertex.neighbors)
		//degree := len(vertex.neighborsIDs)
		graph.degreeDistribution[degree]++

	}
}

func (graph *graph) getDistribution() {

	graph.degreeDistribution = make(map[int]int)
	for _, vertex := range graph.vertices {
		//degree := vertex.d
		//degree := len(vertex.neighbors)
		degree := len(vertex.neighborsIDs)
		graph.degreeDistribution[degree]++

	}
}

func (graph *graph) getHasDistribution() {
	graph.degreeHasDistribution = make(map[int]int)
	for _, vertex := range graph.vertices {
		//degree := vertex.d
		//degree := len(vertex.neighbors)
		d := len(vertex.fixObjectsIDs) + len(vertex.mobileObjectsIDs)
		graph.degreeHasDistribution[d]++
	}
}

func (graph *graph) getDistributionRepeat() {

	graph.degreeDistribution = make(map[int]int)
	for _, vertex := range graph.vertices {
		degree := vertex.d
		//degree := len(vertex.neighbors)
		//degree := len(vertex.neighborsIDs)
		graph.degreeDistribution[degree]++

	}
}

func (graph *graph) getGivenDistribution() {
	for _, vertex := range graph.vertices {
		graph.givenDistribution[int(vertex.w)]++
	}
}

func (graph *graph) purgeEdges() {
	n := 0
	nbNotDeleted := 0
	for _, vertex := range graph.vertices {
		t := vertex.tombstones
		for i := 0; i < t; i++ {

			for _, neighborID := range vertex.neighborsIDs {
				neighbor := graph.vertices[neighborID]
				if neighbor.tombstones > 0 {
					//delete edge
					if _, ok := vertex.neighbors[neighborID]; ok {
						n++
						delete(vertex.neighbors, neighborID)
						vertex.d--
						delete(neighbor.neighbors, vertex.id)
						neighbor.tombstones--
						neighbor.d--
						vertex.tombstones--
						break
					}

				}
			}
		}
		if vertex.tombstones != 0 {
			nbNotDeleted += vertex.tombstones
		}
		vertex.tombstones = 0
	}
	fmt.Printf("deleted: %v \n", n)
	fmt.Printf("Not Deleted: %v \n", nbNotDeleted)
}

func addConnections(graph1, graph2 *graph) {
	graph2.getGivenDistribution()

	degreeSource := extractDegreeDist(graph1.degreeDistribution, "sourceDistribution.txt")
	degreeTarget := extractDegreeDist(graph2.givenDistribution, "targetDistribution.txt")

	cmd := exec.Command("python",
		"OTSolver.py",
		"OTSolver.py")
	_, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	_, err = cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	cmd.Wait()
	ts, _ := getTransformations(degreeSource, degreeTarget, "OTMATRIX.txt")
	addPools, _ := createTransPools(graph1, ts)
	addPools.computeCDF()
	connectVertices(addPools)
	graph1.getDistributionRepeat()
	graph1.getDegreeGraph()
}

func removeConnections(graph1, graph2 *graph) {
	firstDist := graph1.degreeDistribution
	graph2.getGivenDistribution()

	degreeSource := extractDegreeDist(graph1.degreeDistribution, "sourceDistribution.txt")
	degreeTarget := extractDegreeDist(graph2.givenDistribution, "targetDistribution.txt")

	cmd := exec.Command("python",
		"OTSolver.py",
		"OTSolver.py")
	_, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	_, err = cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	cmd.Wait()
	ts, _ := getTransformations(degreeSource, degreeTarget, "OTMATRIX.txt")
	addPools, delPools := createTransPools(graph1, ts)
	addPools.computeCDF()
	delPools.computeCDF()
	connectVertices(addPools)
	markTombstones(delPools)
	graph1.purgeEdges()
	graph1.getDistributionRepeat()
	graph1.getDegreeGraph()
	var emd float64
	emd = 1
	for emd > 0.005 {
		fmt.Println("Repeat!!!")
		graph1.getDistributionRepeat()
		emd = repeat(firstDist, graph1.degreeDistribution, graph2.givenDistribution, graph1)
		fmt.Printf("EMD: %v\n", emd)
	}
}

func getEMDValue(srcDist map[int]int, tgtDist map[int]int) (EMD float64) {

	degreeSource := extractDegreeDist(srcDist, "sourceDistribution.txt")
	degreeTarget := extractDegreeDist(tgtDist, "targetDistribution.txt")

	cmd := exec.Command("python",
		"OTSolver.py",
		"OTSolver.py")
	_, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	_, err = cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	getTransformations(degreeSource, degreeTarget, "OTMATRIX.txt")
	EMD, err = strconv.ParseFloat(getEMD("EMD.txt"), 64)
	if err != nil {
		fmt.Printf("Could  not parse Float, ERROR: %v \n", err)
	}
	return EMD
}

func repeat(firstDist map[int]int, srcDist map[int]int, tgtDist map[int]int, graph *graph) (EMD float64) {
	degreeSource := extractDegreeDist(srcDist, "sourceDistribution.txt")
	degreeTarget := extractDegreeDist(tgtDist, "targetDistribution.txt")

	cmd := exec.Command("python",
		"OTSolver.py",
		"OTSolver.py")
	_, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	_, err = cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	cmd.Wait()
	ts, _ := getTransformations(degreeSource, degreeTarget, "OTMATRIX.txt")
	EMD, err = strconv.ParseFloat(getEMD("EMD.txt"), 64)
	if err != nil {
		fmt.Printf("Could  not parse Float, ERROR: %v \n", err)
	}
	addPools, delPools := createTransPools(graph, ts)
	addPools.computeCDF()
	delPools.computeCDF()
	connectVertices(addPools)
	markTombstones(delPools)

	graph.purgeEdges()
	graph.getDistributionRepeat()
	graph.getDegreeGraph()

	plotDistributions(graph.degreeDistribution, firstDist, "targetWithDeletions.png")
	return
}
