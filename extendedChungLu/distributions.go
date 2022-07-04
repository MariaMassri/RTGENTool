package extendedChungLu

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/stat/distuv"
)

func generateZipfianDegDist(n int, sRand int, s float64, v float64, imax int) (pools *pools, graph *graph, sumDeg int) {
	lastID := 0
	r := rand.New(rand.NewSource(int64(sRand)))
	max := uint64(imax)
	zipfDist := rand.NewZipf(r, s, v, max)
	pools = initPools()
	graph = initGraph()
	initGraph()
	var degrees []int
	for i := 0; i < n; i++ {
		deg := int(zipfDist.Uint64())
		if deg < 0 {
			fmt.Printf("%v degree is negative!", deg)
		}
		degrees = append(degrees, deg)
	}
	for _, deg := range degrees {
		sumDeg += deg
		if _, ok := pools.pools[deg]; !ok {
			pools.addPool(deg)
		}
		graph.addVertex(pools.addVertex(lastID, deg))
		lastID++
	}
	pools.computeCDF()
	return
}

func generateNormalVert(mu, sigma float64, n, lastID int) (vertices map[int][]*vertex, sumDeg int) {
	vertices = make(map[int][]*vertex)
	distLaw := distuv.Normal{
		Mu:    mu,
		Sigma: sigma,
	}
	for i := 0; i < n; i++ {
		deg := int(math.Floor(distLaw.Rand() + 0.5))
		sumDeg += deg
		vertex := initVertex(lastID, deg)
		lastID++
		vertices[deg] = append(vertices[deg], vertex)
	}
	return
}

func generateZipfVert(n int, sRand int, s float64, v float64, imax int, lastID int) (vertices map[int][]*vertex, sumDeg int) {
	vertices = make(map[int][]*vertex)
	r := rand.New(rand.NewSource(int64(sRand)))
	max := uint64(imax)
	zipfDist := rand.NewZipf(r, s, v, max)

	for i := 0; i < n; i++ {
		deg := int(zipfDist.Uint64())
		sumDeg += deg
		vertex := initVertex(lastID, deg)
		lastID++
		vertices[deg] = append(vertices[deg], vertex)
	}
	return
}

func generateNormalDegDist(mu, sigma float64, n int) (pools *pools, graph *graph, sumDeg int) {
	lastID := 0
	distLaw := distuv.Normal{
		Mu:    mu,
		Sigma: sigma,
	}
	pools = initPools()
	graph = initGraph()
	for i := 0; i < n; i++ {
		deg := int(math.Floor(distLaw.Rand() + 0.5))
		sumDeg += deg
		if _, ok := pools.pools[deg]; !ok {
			pools.addPool(deg)
		}
		graph.addVertex(pools.addVertex(lastID, deg))
		lastID++
	}
	pools.computeCDF()
	return
}

func extractDegreeDist(distribution map[int]int, path string) (degrees []int) {
	degs := ""
	nbOccs := ""
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Could not create file: %v", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for deg, nbOcc := range distribution {
		degs += strconv.Itoa(deg) + ", "
		nbOccs += strconv.Itoa(nbOcc) + ", "
		degrees = append(degrees, deg)
	}
	strings.TrimRight(degs, ",")
	strings.TrimRight(nbOccs, ",")
	_, err = w.WriteString(degs + "\n")
	_, err = w.WriteString(nbOccs + "\n")
	w.Flush()
	return

}
