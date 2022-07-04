package extendedChungLu

import (
	"github.com/gocql/gocql"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"
)

var (
	poolMutex sync.Mutex
)

type pools struct {
	pools    map[int]*pool //id -> pool
	CDFpools []*pool //sorted, by pool weight, array of pools
	nbEdges  int
	w        int
}

type pool struct {
	vertices map[int]*vertex // id -> vertex
	verIDs   []int
	cursorID int
	d        int
	w        float64
	wdel     float64
	cdf      float64
	delem    int
}

func initPools() *pools {
	return &pools{pools: make(map[int]*pool)}
}

func initPool(deg int) *pool {
	return &pool{
		vertices: make(map[int]*vertex),
		w:        0.0,
		delem:    deg,
		cursorID: 0,
	}
}

func (pools *pools) addPool(deg int) {
	if _, ok := pools.pools[deg]; !ok {
		pools.pools[deg] = initPool(deg)
		pools.CDFpools = append(pools.CDFpools, pools.pools[deg])
	}
}

func (pools *pools) addVertex(id int, deg int) (vertex *vertex) {
	pools.nbEdges += deg
	pools.pools[deg].vertices[id] = initVertex(id, deg)
	pools.pools[deg].verIDs = append(pools.pools[deg].verIDs, id)
	pools.pools[deg].w += float64(deg)
	return pools.pools[deg].vertices[id]
}
func (pools *pools) addExistVertex(vertex *vertex, w int) {
	pools.nbEdges += w
	id := vertex.id
	vertex.w = float64(w)
	pools.pools[w].vertices[id] = vertex
	pools.pools[w].verIDs = append(pools.pools[w].verIDs, id)
	pools.pools[w].w += float64(w)
}

func (pools *pools) addExistHasVertex(vertex *vertex, w int) {
	pools.nbEdges += w
	id := vertex.id
	pools.pools[w].vertices[id] = vertex
	pools.pools[w].verIDs = append(pools.pools[w].verIDs, id)
	pools.pools[w].w += float64(w)
}

func (pools *pools) computeCDF() {
	//sort pools by weight
	sumDeg := pools.nbEdges
	sort.Slice(pools.CDFpools, func(i, j int) bool {
		return pools.CDFpools[i].w < pools.CDFpools[j].w
	})
	for i, pool := range pools.CDFpools {
		if i != 0 {
			pool.cdf = pools.CDFpools[i-1].cdf + pool.w/float64(sumDeg)
		} else {
			pool.cdf = pool.w / float64(sumDeg)
		}
	}
}

func getSession() (*gocql.Session, error) {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "scikitmobility"
	return cluster.CreateSession()
}

func connectVerticesSciMob(pools *pools, startTime, endTime time.Time) (err error) {
	session, err := getSession()
	partition := 0
	communityID := 0
	//startTime:= time.Unix(1546329600, 0)
	//endTime:= time.Unix(1547452800, 0)
	batch := session.NewBatch(gocql.LoggedBatch)
	if err != nil {
		return
	}
	pools.w++
	for i := 0; i < pools.nbEdges/2; i++ {
		source, target := chooseVertex(pools), chooseVertex(pools)
		source.addNeighbor(target)
		timestamp := randate(startTime, endTime)
		event := "knows," + strconv.Itoa(communityID) + "," + strconv.Itoa(source.id) + "," + timeToString(timestamp) + "," + strconv.Itoa(target.id) + ", start"
		batch.Query("INSERT INTO events (partition, timestamp, event) values (?, ?, ?)", partition, timeToString(timestamp), event)
		if i%500 == 0 {
			partition++
			session.ExecuteBatch(batch)
			batch = session.NewBatch(gocql.LoggedBatch)
		}
	}
	session.ExecuteBatch(batch)
	return
}

func connectVertices(pools *pools) {
	pools.w++
	for i := 0; i < pools.nbEdges/2; i++ {
		source, target := chooseVertex(pools), chooseVertex(pools)
		source.addNeighbor(target)
	}

}

func connectVerticesMultiThreaded(conPools *pools, n int) {
	conPools.w++
	var waitGroup sync.WaitGroup
	nbEdges := conPools.nbEdges / (2 * n)
	for i := 0; i < n; i++ {
		waitGroup.Add(1)
		go func(n int, pools *pools, waitgroup *sync.WaitGroup) {
			for i := 0; i < nbEdges; i++ {
				source, target := chooseVertexMulti(pools), chooseVertexMulti(pools)
				source.addNeighbor(target)

			}
			waitgroup.Done()
		}(nbEdges, conPools, &waitGroup)
	}
	waitGroup.Wait()
}
func connectVerticesRepeat(pools *pools) {
	pools.w++
	for i := 0; i < pools.nbEdges/2; i++ {
		source, target := chooseVertex(pools), chooseVertex(pools)
		source.addNeighborRepeat(target)
	}

}

func addObjects(pools *pools, lastObjectID int, pMobile float64) {
	pools.w++
	for i := 0; i < pools.nbEdges; i++ {
		lastObjectID++
		user := chooseVertexHas(pools)
		x := rand.Float64()
		if x < pMobile {
			user.addMobileObject(lastObjectID)
		} else {
			user.addFixObject(lastObjectID)
		}
	}
}

func markTombstones(pools *pools) {
	for i := 0; i < pools.nbEdges; i++ {
		chosenVertex := chooseTombstone(pools)
		chosenVertex.tombstones++
	}
}

func chooseTombstone(pools *pools) (vertex *vertex) {
	var ro = rand.Float64()
	var p int //index of the chosen pool
	var choose bool
	for j := 0; j < len(pools.CDFpools); j++ {
		choose = false
		if j == 0 {
			if 0 < ro && ro <= pools.CDFpools[0].cdf {
				choose = true
				p = 0
			} else {
				continue
			}
		} else if pools.CDFpools[j-1].cdf < ro && ro <= pools.CDFpools[j].cdf {
			choose = true
			p = j
		}
		if choose {
			cursorID := pools.CDFpools[p].cursorID
			vertex = pools.CDFpools[p].vertices[pools.CDFpools[p].verIDs[cursorID]]
			pools.CDFpools[p].d++
			pools.CDFpools[p].cursorID++
			if pools.CDFpools[p].cursorID == len(pools.CDFpools[p].vertices) {
				pools.CDFpools[p].cursorID = 0
			}
			break
		}
	}
	return
}

func chooseVertex(pools *pools) (vertex *vertex) {
	var ro = rand.Float64()
	var p int //index of the chosen pool
	var choose bool
	l := len(pools.CDFpools) - 1
	for j := 0; j < len(pools.CDFpools); j++ {
		choose = false
		if j == 0 {
			if 0 < ro && ro <= pools.CDFpools[0].cdf {
				choose = true
				p = 0
			} else {
				continue
			}
		} else if pools.CDFpools[j-1].cdf < ro && ro <= pools.CDFpools[j].cdf {
			choose = true
			p = j
		} else if ro > pools.CDFpools[l].cdf {
			choose = true
			p = l
		}
		if choose {
			cursorID := pools.CDFpools[p].cursorID
			vertex = pools.CDFpools[p].vertices[pools.CDFpools[p].verIDs[cursorID]]
			vertex.d++
			pools.CDFpools[p].d++
			pools.CDFpools[p].cursorID++
			if pools.CDFpools[p].cursorID == len(pools.CDFpools[p].verIDs) {
				pools.CDFpools[p].cursorID = 0
			}
			break
		}
	}
	return
}

func chooseVertexMulti(pools *pools) (vertex *vertex) {
	var ro = rand.Float64()
	var p int //index of the chosen pool
	var choose bool
	l := len(pools.CDFpools) - 1
	for j := 0; j < len(pools.CDFpools); j++ {
		choose = false
		if j == 0 {
			if 0 < ro && ro <= pools.CDFpools[0].cdf {
				choose = true
				p = 0
			} else {
				continue
			}
		} else if pools.CDFpools[j-1].cdf < ro && ro <= pools.CDFpools[j].cdf {
			choose = true
			p = j
		} else if ro > pools.CDFpools[l].cdf {
			choose = true
			p = l
		}
		if choose {
			//cursorID := pools.CDFpools[p].cursorID
			max := len(pools.CDFpools[p].verIDs)
			id := rand.Intn(max)
			vertex = pools.CDFpools[p].vertices[pools.CDFpools[p].verIDs[id]]
			vertex.d++
			/*pools.CDFpools[p].d++
			pools.CDFpools[p].cursorID++
			if pools.CDFpools[p].cursorID == len(pools.CDFpools[p].verIDs) {
				pools.CDFpools[p].cursorID = 0
			}*/
			break
		}
	}
	return
}

func chooseVertexHas(pools *pools) (vertex *vertex) {
	var ro = rand.Float64()
	var p int //index of the chosen pool
	var choose bool
	l := len(pools.CDFpools) - 1
	for j := 0; j < len(pools.CDFpools); j++ {
		choose = false
		if j == 0 {
			if 0 < ro && ro <= pools.CDFpools[0].cdf {
				choose = true
				p = 0
			} else {
				continue
			}
		} else if pools.CDFpools[j-1].cdf < ro && ro <= pools.CDFpools[j].cdf {
			choose = true
			p = j
		} else if ro > pools.CDFpools[l].cdf {
			choose = true
			p = l
		}
		if choose {
			cursorID := pools.CDFpools[p].cursorID
			vertex = pools.CDFpools[p].vertices[pools.CDFpools[p].verIDs[cursorID]]
			//vertex.d++
			pools.CDFpools[p].d++
			pools.CDFpools[p].cursorID++
			if pools.CDFpools[p].cursorID == len(pools.CDFpools[p].verIDs) {
				pools.CDFpools[p].cursorID = 0
			}
			break
		}
	}
	return
}

func chooseVertexDetectError(pools *pools) (vertex *vertex, p int) {
	var ro = rand.Float64()
	var pi int
	var choose bool
	for j := 0; j < len(pools.CDFpools); j++ {
		choose = false
		if j == 0 {
			if 0 < ro && ro <= pools.CDFpools[0].cdf {
				choose = true
				pi = 0
			} else {
				continue
			}
		} else if pools.CDFpools[j-1].cdf < ro && ro <= pools.CDFpools[j].cdf {
			choose = true
			pi = j
		}
		if choose {
			cursorID := pools.CDFpools[pi].cursorID
			if cursorID >= len(pools.CDFpools[pi].verIDs) {
				cursorID = 0
			}
			vertex = pools.CDFpools[pi].vertices[pools.CDFpools[pi].verIDs[cursorID]]

			break
		}
	}
	return vertex, pi
}
