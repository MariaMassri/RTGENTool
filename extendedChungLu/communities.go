package extendedChungLu

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gonum/matrix/mat64"
	"gonum.org/v1/gonum/mat"
)

//permits to choose where to place the edge
type coms struct {
	communitities         map[int]*com
	CDFCommunities        []*comsToConnect
	comIDs                []int
	comIDsMap             map[int]int
	cursorID              int
	nbEdges               int
	nbVertices            int
	dcAdd                 int
	dcDel                 int
	clock                 time.Time
	timeUnit              time.Duration
	vertexWriter          *bufio.Writer
	edgeWriter            *bufio.Writer
	sourceCommunityMatrix mat.Dense
}

type comsToConnect struct {
	comA    *com
	comB    *com
	cdf     float64
	nw      float64
	nbEdges int
}

type com struct {
	id       int
	dc       int //dc = coms.nbEdges*pc
	pc       float64
	w        int
	pools    *pools
	graph    *graph
	addPools *pools
	delPools *pools
	dcAdd    int //number of edges to add
	dcDel    int
}

func initComs() *coms {
	return &coms{
		communitities: make(map[int]*com),
	}
}
func initComsT(startTime time.Time, timeUnit time.Duration, vertexPath, edgePath string) *coms {
	vertexFile, err := os.Create(vertexPath)
	if err != nil {
		log.Fatalf("Could not open file: %v", err)
	}
	edgeFile, err := os.Create(edgePath)
	if err != nil {
		log.Fatalf("Could not open file: %v", err)
	}
	vertexWriter := bufio.NewWriter(vertexFile)
	edgeWriter := bufio.NewWriter(edgeFile)
	vertexWriter.WriteString("id;comID;startTime\n")
	edgeWriter.WriteString("Source;Target;startTime\n")
	return &coms{
		communitities: make(map[int]*com),
		clock:         startTime,
		timeUnit:      timeUnit,
		vertexWriter:  vertexWriter,
		edgeWriter:    edgeWriter,
	}
}
func (coms *coms) addCDFComs(cdfComs []*comsToConnect) {
	coms.CDFCommunities = cdfComs
}
func (coms *coms) addCDFCom(cdfCom *comsToConnect) {
	coms.CDFCommunities = append(coms.CDFCommunities, cdfCom)
}

func (graph *graph) computeTargetCommunityMatrix(r, c int, coms *coms) mat.Dense {
	nbEdges := 0

	zeros := []float64{}
	for i := 0; i < r*c; i++ {
		zeros = append(zeros, 0)
	}
	targetCommunityMatrix := *mat.NewDense(r, c, zeros)

	for _, vertex := range graph.vertices {
		comAID := vertex.comID
		for _, neighbor := range vertex.neighbors {
			nbEdges++
			comBID := neighbor.comID
			targetCommunityMatrix.Set(coms.comIDsMap[comAID], coms.comIDsMap[comBID], targetCommunityMatrix.At(coms.comIDsMap[comAID], coms.comIDsMap[comBID])+float64(1))
		}
	}
	//diviser chqaque element de targetCommunityMatrix par le nombre total de relations
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			targetCommunityMatrix.Set(i, j, targetCommunityMatrix.At(i, j)/float64(nbEdges))
		}
	}

	sum := 0.0
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			sum += targetCommunityMatrix.At(i, j)
		}
	}
	return targetCommunityMatrix
}

func (coms *coms) addCom(id int) (c *com) {
	pls := initPools()
	addPls := initPools()
	delPls := initPools()
	g := initGraph()
	c = &com{
		id:       id,
		pools:    pls,
		graph:    g,
		addPools: addPls,
		delPools: delPls,
	}
	coms.communitities[id] = c
	coms.comIDs = append(coms.comIDs, c.id)
	return
}

func initCommunities(comMatrix *mat.Dense, pools *pools) (returnedComs *coms, err error) {
	//matrix
	// | 0.34 0.04 0.0.4 |
	// | 0.0.4 0.22 0.04 |
	// | 0.04 0.04 0.32  |
	r, c := comMatrix.Dims()
	if r != c {
		return nil, fmt.Errorf("The community matrix should be a square matrix")
	}
	coms := initComs()
	coms.sourceCommunityMatrix = *comMatrix
	zeros := []float64{}
	for i := 0; i < r*c; i++ {
		zeros = append(zeros, 0)
	}
	var comSlice []*com
	for i := 0; i < r; i++ {
		ci := coms.addCom(i)
		comSlice = append(comSlice, ci)
	}

	coms.distributePools(pools)
	mapComputed := make(map[string]int)
	for i := 0; i < r; i++ {
		for j := 0; j < r; j++ {
			if _, okij := mapComputed[strconv.Itoa(i)+" "+strconv.Itoa(j)]; okij {
				continue
			}
			if _, okji := mapComputed[strconv.Itoa(j)+" "+strconv.Itoa(i)]; okji {
				continue
			}
			mapComputed[strconv.Itoa(i)+" "+strconv.Itoa(j)] = 0
			if i == j {
				ctcij := &comsToConnect{
					comA: comSlice[i],
					comB: comSlice[j],
					nw:   comMatrix.At(i, j),
				}
				coms.addCDFCom(ctcij)
			} else {
				ctcij := &comsToConnect{
					comA: comSlice[i],
					comB: comSlice[j],
					nw:   comMatrix.At(i, j) * 2,
				}
				coms.addCDFCom(ctcij)
			}
		}
	}
	coms.computeCDF()
	coms.computeComNbEgdes()
	coms.distributeVertices(pools)
	coms.computeCDFPools()
	return coms, nil
}

func generateGraphWithNormalCom(mu float64, sigma float64, n int, communityMatrix mat.Dense, path string) (coms *coms, graph *graph) {
	pools, graph, _ := generateNormalDegDist(mu, sigma, n)
	//pools, graph := generateZipfianDegDist(n, 23, 1.5, 1, 80)
	pools.computeCDF()
	graph.getGivenDistribution()
	distFirst := graph.givenDistribution
	plotDistribution(distFirst, path+"sourceCom.png")
	comMatrix := mat.NewDense(3, 3, []float64{0.5, 0.1, 0.025, 0.1, 0.1,
		0.025, 0.025, 0.025, 0.1})
	coms, err := initCommunities(comMatrix, pools)
	if err != nil {
		fmt.Printf("Cannot create communities!")
	}
	connectVerticesComs(pools.nbEdges/2, coms)
	return
}

func (coms *coms) addExistingCom(c *com) {
	coms.communitities[c.id] = c
	coms.comIDs = append(coms.comIDs, c.id)
}

func (com *com) addPool(deg int) {
	com.pools.addPool(deg)
}

func (coms *coms) distributePools(pools *pools) {
	coms.nbEdges = pools.nbEdges
	for deg := range pools.pools {
		for _, com := range coms.communitities {
			com.addPool(deg)
		}
	}
}

func chooseCom(coms *coms) (com *com) {
	id := coms.comIDs[coms.cursorID]
	com = coms.communitities[id]
	//com = coms.communitities[coms.cursorID]
	coms.cursorID++
	if coms.cursorID == len(coms.comIDs) {
		coms.cursorID = 0
	}
	return
}

func (com *com) addVertex(deg int, vertex *vertex) {
	com.pools.addExistVertex(vertex, deg)
	com.graph.addVertex(vertex)
}
func (com *com) addVertexG(vertex *vertex) {
	com.graph.addVertex(vertex)
}
func (coms *coms) computeComNbEgdes() {
	for _, com := range coms.communitities {
		if com.pc == 0 {
			for _, ctc := range coms.CDFCommunities {
				comA := ctc.comA
				comB := ctc.comB
				if com.id == comA.id && com.id == comB.id {
					com.pc += ctc.nw
				} else if com.id == comA.id || com.id == comB.id {
					com.pc += ctc.nw / 2
				}
			}
		} else {
			break
		}
	}
	for _, com := range coms.communitities {
		com.dc = int(float64(coms.nbEdges) * com.pc)
	}
}

func (coms *coms) getCom(comID int) (com *com) {
	return coms.communitities[comID]
}

func (coms *coms) setPc(comID int, pc float64) {
	coms.getCom(comID).pc = pc
}

func (coms *coms) computeComNbEdges(ts []transformation) {
	//we alrady have com.pc
	//we need to compute com.dc
	//com.dcAdd = com.pc * coms.dcAdd
	//com.dcDel = com.pc * coms.dcDel
	//TODO: compute coms.dcAdd and coms.dcDel
	var dcAdd, dcDel int
	nbVertices := coms.getTotNbVertices()
	for _, t := range ts {
		nbMod := int(math.Ceil(float64(nbVertices) * t.p)) //graph.nbVertices TODO
		delta := 0
		if t.sd < t.td {
			delta = t.td - t.sd
			dcAdd += delta * nbMod
		} else if t.sd > t.td {
			delta = t.sd - t.td
			dcDel += delta * nbMod
		}
	}
	coms.dcAdd = dcAdd
	coms.dcDel = dcDel
	sumPC := 0.0
	for _, com := range coms.communitities {
		//com.dcAdd = dcAdd*com.pc
		//com.dcDel = dcDel*com.pc
		com.graph.getDegreeGraph()
		sumPC += com.pc
		com.dcAdd = int(com.pc * float64(dcAdd))
		com.dcDel = int(com.pc * float64(dcDel))
	}
}
func testComTrans(coms *coms, ts []transformation) {
	var lmafrudDegree = make(map[int]int)
	var li3ena = make(map[int]int)
	nbVertices := coms.getTotNbVertices()
	for _, t := range ts {
		nbMod := int(math.Ceil(float64(nbVertices) * t.p))
		lmafrudDegree[t.sd] += nbMod
	}
	for _, com := range coms.communitities {
		com.graph.getDistribution()
		for deg, nbMod := range com.graph.degreeDistribution {
			li3ena[deg] += nbMod
		}
	}

}

func (coms *coms) getCtc(comAID int, comBID int) (ctc *comsToConnect) {
	for _, ctc := range coms.CDFCommunities {
		if ctc.comA.id == comAID && ctc.comB.id == comBID {
			return ctc
		} else if ctc.comB.id == comAID && ctc.comA.id == comBID {
			return ctc
		}
	}
	return
}

func (coms *coms) generateTransPools1(ts []transformation) {
	nbVertices := coms.getTotNbVertices()
	n := len(ts)
	for i := 0; i < coms.dcAdd; i++ {
		found := false
		com := chooseCom(coms)
		if i%nbVertices == 0 {
			fmt.Printf("not blocking!!!\n")
		}
		m := 0
		if com.dcAdd > 0 {
			//choose a transformation at random
			for !found {
				m++
				j := rand.Intn(n)
				t := ts[j]
				delta := t.td - t.sd

				if t.p > 0 && len(com.graph.degreeGraph[t.sd]) != 0 {
					found = true
					ts[j].p = t.p - float64(1/nbVertices)
					com.dcAdd -= delta //to change later
					vertex := com.graph.degreeGraph[t.sd][0]
					if _, ok := com.addPools.pools[delta]; !ok {
						com.addPools.addPool(delta)
					}
					com.addPools.addExistVertex(vertex, delta)
					com.graph.degreeGraph[t.sd] = com.graph.degreeGraph[t.sd][1:len(com.graph.degreeGraph[t.sd])] //TODO transform this to a function

				}
			}

		}
	}

}
func (coms *coms) initTransPools() {
	coms.dcAdd = 0
	coms.dcDel = 0
	for _, com := range coms.communitities {
		com.dcAdd = 0
		com.dcDel = 0
		com.addPools = initPools()
		com.delPools = initPools()
	}
}
func (coms *coms) generateTransPools(ts []transformation) {
	nbVertices := coms.getTotNbVertices()
	for _, t := range ts {
		nbMod := int(math.Ceil(float64(nbVertices) * t.p))
		if t.sd < t.td {
			delta := t.td - t.sd
			//add
			//choose com if com.dcAdd >0 add it to the corresponding addPool of this com
			for i := 0; i < nbMod; i++ {

				found := false
				y := 0
				for !found && y <= len(coms.communitities)+5 {
					com := chooseCom(coms)
					if com.dcAdd > 0 {
						//search for a vertex in this com with deg == t.sd
						if len(com.graph.degreeGraph[t.sd]) != 0 {
							found = true
							com.dcAdd -= delta
							vertex := com.graph.degreeGraph[t.sd][0]
							if _, ok := com.addPools.pools[delta]; !ok {
								com.addPools.addPool(delta)
							}
							com.addPools.addExistVertex(vertex, delta)
							com.graph.degreeGraph[t.sd] = com.graph.degreeGraph[t.sd][1:len(com.graph.degreeGraph[t.sd])] //TODO transform this to a function

						}
					}
					y++
				}

			}

		} else if t.sd > t.td {
			delta := t.sd - t.td
			for i := 0; i < nbMod; i++ {
				found := false
				for !found {
					com := chooseCom(coms)
					if com.dcDel > 0 {
						if len(com.graph.degreeGraph[t.sd]) != 0 {
							found = true
							com.dcDel -= delta
							vertex := com.graph.degreeGraph[t.sd][0]
							if _, ok := com.delPools.pools[delta]; !ok {
								com.delPools.addPool(delta)
							}
							com.delPools.addExistVertex(vertex, delta)
							com.graph.degreeGraph[t.sd] = com.graph.degreeGraph[t.sd][1:len(com.graph.degreeGraph[t.sd])]
						}
					}
				}
			}
		}
	}
}

func (coms *coms) distributeVertices(pools *pools) {
	for deg, pool := range pools.pools {
		for _, vertex := range pool.vertices {
			found := false
			for !found {
				com := chooseCom(coms)
				if com.dc >= 0 {
					found = true
					com.addVertex(deg, vertex)
					vertex.comID = com.id
					com.dc -= deg
				}
			}
		}
	}
}

func (coms *coms) getNbVertices(id int) (n int) {
	com := coms.communitities[id]
	n = len(com.graph.vertices)
	/*for _, pool := range com.pools.pools {
		n += len(pool.vertices)
	}*/
	return
}
func (coms *coms) getTotNbVertices() (n int) {
	for id := range coms.communitities {
		n += coms.getNbVertices(id)
	}
	return
}

func (coms *coms) computeCDFPools() {
	for _, com := range coms.communitities {
		com.pools.computeCDF()
		com.addPools.computeCDF()
		com.delPools.computeCDF()
	}
}

//init coms (done)
//add all com (done)
//distribute pools into coms (done)
//distribute vertices (done)
//insert comToConnect to coms
//commputeCDF
//connectVerticesWithCommunitiy

func (coms *coms) computeCDF() {
	//sort communities by weight
	sort.Slice(coms.CDFCommunities, func(i, j int) bool {
		return coms.CDFCommunities[i].nw < coms.CDFCommunities[j].nw
	})
	for i, comToCon := range coms.CDFCommunities {
		if i != 0 {
			comToCon.cdf = coms.CDFCommunities[i-1].cdf + comToCon.nw
		} else {
			comToCon.cdf = comToCon.nw
		}
	}
}

func chooseCommunity(coms *coms) (comsToConnect *comsToConnect) {
	var ro = rand.Float64()
	var choose bool
	l := len(coms.CDFCommunities) - 1
	for j := 0; j < len(coms.CDFCommunities); j++ {
		choose = false
		if j == 0 {
			if 0 < ro && ro <= coms.CDFCommunities[0].cdf {

				comsToConnect = coms.CDFCommunities[0]
				choose = true
			} else {
				continue
			}
		} else if coms.CDFCommunities[j-1].cdf < ro && ro <= coms.CDFCommunities[j].cdf {
			comsToConnect = coms.CDFCommunities[j]
			choose = true
		} else if ro > coms.CDFCommunities[l].cdf {
			comsToConnect = coms.CDFCommunities[l]
			choose = true
		}
		if choose {
			break
		}
	}
	return
}

func connectVerticesT(nbEdges int, coms *coms) {
	coms.nbEdges = nbEdges * 2
	for i := 0; i < nbEdges; i++ {
		ctc := chooseCommunity(coms)
		source := chooseVertex(ctc.comA.pools)
		target := chooseVertex(ctc.comB.pools)
		timestamp := coms.clock
		timestampS := timestamp.Format("2006-01-02 15:04:05")
		coms.clock = coms.clock.Add(coms.timeUnit)
		if len(source.neighborsIDs) == 0 {
			coms.vertexWriter.WriteString(strconv.Itoa(source.id) + ";" + strconv.Itoa(source.comID) + ";" + timestampS + "\n")
		}
		if len(target.neighborsIDs) == 0 {
			coms.vertexWriter.WriteString(strconv.Itoa(target.id) + ";" + strconv.Itoa(source.comID) + ";" + timestampS + "\n")
		}
		coms.edgeWriter.WriteString(strconv.Itoa(source.id) + ";" + strconv.Itoa(target.id) + ";" + timestampS + "\n")
		source.addNeighbor(target)
		ctc.comA.w++
		ctc.comB.w++
		ctc.nbEdges += 2

	}
}

func connectVerticesComs(nbEdges int, coms *coms) {
	coms.nbEdges = nbEdges * 2
	for i := 0; i < nbEdges; i++ {
		ctc := chooseCommunity(coms)
		source := chooseVertex(ctc.comA.pools)
		target := chooseVertex(ctc.comB.pools)
		source.addNeighbor(target)
		ctc.comA.w++
		ctc.comB.w++
		ctc.nbEdges += 2
	}
}

func (coms *coms) getEdgeDensitiesBlockMatrix(nbEdges int) (edbm map[int]map[int]float64) {
	edbm = make(map[int]map[int]float64)
	for sComID, sCom := range coms.communitities {
		edbm[sComID] = make(map[int]float64)
		for _, source := range sCom.graph.vertices {
			for _, tID := range source.neighborsIDs {
				tComID := coms.findVertexCom(tID)
				edbm[sComID][tComID] += 1.0
			}
		}
	}
	return
}

func connectVerticesComsAdd(nbEdges int, coms *coms) {
	coms.nbEdges += nbEdges * 2
	for i := 0; i < nbEdges; i++ {
		ctc := chooseCommunity(coms)
		source := chooseVertex(ctc.comA.addPools)
		target := chooseVertex(ctc.comB.addPools)
		source.addNeighbor(target)
		ctc.nbEdges += 2
		ctc.comA.w++
		ctc.comB.w++
	}
}

func connectVerticesComsSplit(nbEdges int, splittedComs *coms, coms *coms) {
	coms.nbEdges += nbEdges * 2
	for i := 0; i < nbEdges; i++ {
		ctc := chooseCommunity(splittedComs)
		comA := ctc.comA
		comB := ctc.comB
		ctcComs := coms.getCtc(comA.id, comB.id)
		source := chooseVertex(comA.addPools)
		target := chooseVertex(comB.addPools)
		source.addNeighbor(target)
		ctc.nbEdges += 2
		ctcComs.nbEdges += 2
		comA.w++
		comB.w++
	}
}

func connectVerticesComsT(coms *coms) {
	nbEdges := coms.dcAdd / 2
	coms.nbEdges += nbEdges * 2
	for i := 0; i < nbEdges; i++ {
		ctc := chooseCommunity(coms)
		ctc.comA.w++
		ctc.comB.w++
		ctc.nbEdges += 2
		source := chooseVertex(ctc.comA.addPools)
		target := chooseVertex(ctc.comB.addPools)
		source.addNeighbor(target)
	}

}
func (coms *coms) clearPools() {
	for _, com := range coms.communitities {
		com.addPools.pools = make(map[int]*pool)
		com.addPools.nbEdges = 0
		com.addPools.w = 0
		com.addPools.CDFpools = nil
	}
}
func (coms *coms) generateDegreeGraphs() {
	for _, com := range coms.communitities {
		com.graph.getDegreeGraph()
	}
}
func kronecker(a, b mat64.Matrix) *mat64.Dense {
	ar, ac := a.Dims()
	br, bc := b.Dims()
	k := mat64.NewDense(ar*br, ac*bc, nil)
	for i := 0; i < ar; i++ {
		for j := 0; j < ac; j++ {
			s := k.Slice(i*br, (i+1)*br, j*bc, (j+1)*bc).(*mat64.Dense)
			s.Scale(a.At(i, j), b)
		}
	}
	return k
}
func (coms *coms) clearCDFComs() {
	coms.CDFCommunities = nil
}

func (coms *coms) getVertex(comID int, vertexID int) (vertex *vertex, ok bool) {
	vertex, ok = coms.communitities[comID].graph.vertices[vertexID]
	return
}

func (coms *coms) findVertexCom(vertexID int) (comID int) {
	for comID := range coms.communitities {
		if _, ok := coms.getVertex(comID, vertexID); ok {
			return comID
		}
	}
	return
}

func getTotDeg(ctcs []*comsToConnect) (totNbDeg int) {
	for _, ctc := range ctcs {
		totNbDeg += ctc.nbEdges
	}
	return
}

func getDoubleEdges(graph *graph) (m map[int]map[int]int) {
	m = make(map[int]map[int]int)
	for sourceID, source := range graph.vertices {
		m[sourceID] = make(map[int]int)
		for _, neighborID := range source.neighborsIDs {
			m[sourceID][neighborID]++
		}
	}
	return
}

func (coms *coms) splitCom(comID int, graph *graph) {
	nbEdges := getTotDeg(coms.CDFCommunities)
	var ctcs []*comsToConnect
	m := getDoubleEdges(graph)
	splitCom, ok := coms.communitities[comID]
	if !ok {
		fmt.Printf("com %v does not exist\n", comID)
	}
	comD := make(map[int]float64)
	comWijDesired := make(map[int]map[int]float64) //comA -> comB -> wABdesiree= 0.2
	comWijObtained := make(map[int]map[int]int)    //comA -> comB -> wABdesiree= 0.2
	comWijToDel := make(map[int]map[int]int)
	comWijToAdd := make(map[int]map[int]int)
	comPc := make(map[int]float64)
	lastID := coms.comIDs[len(coms.comIDs)-1]
	comNID := lastID + 1
	comMID := lastID + 2
	splittedComs := initComs()
	comN := coms.addCom(comNID)
	comM := coms.addCom(comMID)
	splittedComs.addExistingCom(comN)
	splittedComs.addExistingCom(comM)
	wnn := 0.66
	wmm := 0.33
	wnm := 0.005
	wnj := 0.5
	wmj := 0.5
	/*wnn := 0.6 wmm := 0.6 wnj = 0.65 wmj = 0.45*/
	dDel := 0
	dAdd := 0
	//compute desired wij
	comWijDesired[comNID] = make(map[int]float64)
	comWijDesired[comMID] = make(map[int]float64)
	comWijObtained[comNID] = make(map[int]int)
	comWijObtained[comMID] = make(map[int]int)
	comWijToDel[comNID] = make(map[int]int)
	comWijToDel[comMID] = make(map[int]int)
	comWijToAdd[comNID] = make(map[int]int)
	comWijToAdd[comMID] = make(map[int]int)

	for _, comsToConnect := range coms.CDFCommunities {
		sourceCom := comsToConnect.comA.id
		targetCom := comsToConnect.comB.id
		nwST := float64(comsToConnect.nbEdges) / float64(nbEdges)
		if sourceCom != splitCom.id && targetCom != splitCom.id {
			if sourceCom != targetCom {
				comPc[sourceCom] += nwST * 0.5
				comPc[targetCom] += nwST * 0.5
			} else {
				comPc[sourceCom] += nwST
			}
			comsToConnect.nw = nwST
			ctcs = append(ctcs, comsToConnect)
		}

		if sourceCom == splitCom.id && targetCom != sourceCom {
			comWijDesired[comNID][targetCom] = wnj * nwST
			comWijDesired[comMID][targetCom] = wmj * nwST
		}
		if sourceCom != targetCom && targetCom == splitCom.id {
			comWijDesired[comNID][sourceCom] = wnj * nwST
			comWijDesired[comMID][sourceCom] = wmj * nwST
		}
		if sourceCom == splitCom.id && targetCom == splitCom.id {
			comWijDesired[comNID][comNID] = wnn * nwST
			comWijDesired[comMID][comMID] = wmm * nwST
			comWijDesired[comNID][comMID] = wnm * nwST
			comWijDesired[comMID][comNID] = wnm * nwST
		}
	}
	for comID, pc := range comPc {
		coms.getCom(comID).pc = pc
	}
	coms.clearCDFComs()
	for sourceComID, targetComs := range comWijDesired {
		p := 0.0
		for targetComID, wij := range targetComs {
			if _, ok := comWijDesired[targetComID]; !ok {
				p += wij * 0.5
			} else {
				p += wij
			}
		}
		comD[sourceComID] = p
		coms.setPc(sourceComID, p)
	}

	comNbEdges := make(map[int]int)
	nbLeftVertices := 0
	for _, vertex := range splitCom.graph.vertices {
		found := false
		n := 0
		for !found {
			n++
			com := chooseCom(splittedComs)
			for _, com = range splittedComs.communitities {
				break
			}
			if comNbEdges[com.id] <= int(math.Ceil(comD[com.id]*float64(nbEdges))) {
				com.addVertexG(vertex)
				comNbEdges[com.id] += len(vertex.neighborsIDs)
				found = true
			}
			if n == 1000 {
				nbLeftVertices++
				continue
			}
		}
	}

	//compute obtained wij after vertex distribution
	coms.removeCom(splitCom.id)
	inEdges, outEdges := breakEdges(coms, splittedComs)
	total := 0
	for comID, nbInEdges := range inEdges {
		comWijObtained[comID][comID] = nbInEdges
		total += nbInEdges
	}
	for comSourceID, comTargetIDs := range outEdges {
		for targetComID, nbOutEdges := range comTargetIDs {
			comWijObtained[comSourceID][targetComID] = nbOutEdges
			total += nbOutEdges
		}
	}

	//compute the number of edges to remove
	for comSourceID, comsTargetID := range comWijDesired {
		for comTargetID, wijDesired := range comsTargetID {
			//to delete
			wijObtained := comWijObtained[comSourceID][comTargetID]
			delta := wijObtained - int(math.Ceil(wijDesired*float64(nbEdges)))
			if delta < 0 {
				comWijToAdd[comSourceID][comTargetID] = -delta
				dAdd += -delta
			}
			if delta > 0 {
				comWijToDel[comSourceID][comTargetID] = delta
				dDel += delta
			}
		}
	}

	for sourceComID, targetComs := range comWijDesired {
		nw := 0.0
		n := 0
		for targetComID, wij := range targetComs {

			if sourceComID == comNID && targetComID == comMID {
				nw = 2.0 * wij
				n = 2 * comWijObtained[sourceComID][targetComID]
			} else if sourceComID == comMID && targetComID == comNID {
				continue
			} else {
				nw = wij
				n = comWijObtained[sourceComID][targetComID]
			}
			ctc := &comsToConnect{
				comA:    coms.communitities[sourceComID],
				comB:    coms.communitities[targetComID],
				nw:      nw,
				nbEdges: n,
			}
			ctcs = append(ctcs, ctc)
		}
	}
	coms.addCDFComs(ctcs)
	coms.computeCDF()
	for comSourceID, comsTargetID := range comWijToAdd {
		for comTargetID, wijToAdd := range comsTargetID {
			ctc := &comsToConnect{
				comA: coms.communitities[comSourceID],
				comB: coms.communitities[comTargetID],
				nw:   float64(wijToAdd) / float64(dAdd),
			}
			splittedComs.addCDFCom(ctc)
		}

	}

	for _, com := range splittedComs.communitities {
		com.graph.getDegreeGraph()
		for d, vertices := range com.graph.degreeGraph {
			com.addPool(d)
			for _, vertex := range vertices {
				com.pools.addExistVertex(vertex, d)
			}
		}
	}
	for id, com := range coms.communitities {
		if id != comID && id != comNID && id != comMID {
			splittedComs.addExistingCom(com)
		}
	}
	splittedComs.computeCDF()

	var tombstones = make(map[int]int) //vertex -> tombstones
	//TODO init CDFCOMMUNITIES with new weights BECAUSE ALL nws HAVE BEEN CHANGED!
	//toDel := 0
	nbLeftVertices = 0
	for comSourceID, comsTargetID := range comWijToDel {
		for comTargetID, edgesToDelete := range comsTargetID {
			ctc := coms.getCtc(comSourceID, comTargetID)

			for i := 0; i < edgesToDelete/2; i++ {
				coms.getCom(comSourceID).w--
				coms.getCom(comTargetID).w--
				ctc.nbEdges -= 2
				coms.nbEdges -= 2
				found := false
				n := 0
				for !found {
					n++
					var source *vertex
					var sourceID int

					for sourceID, source = range splittedComs.communitities[comSourceID].graph.vertices {
						break
					}
					for i, targetID := range source.neighborsIDs {
						if target, ok := coms.getVertex(comTargetID, targetID); ok {
							tombstones[sourceID]++
							tombstones[targetID]++

							source.neighborsIDs = append(source.neighborsIDs[:i], source.neighborsIDs[i+1:]...)
							target.neighborsIDs = deleteFromSlice(target.neighborsIDs, sourceID)

							m[sourceID][targetID]--
							m[targetID][sourceID]--

							if m[sourceID][targetID] == 0 {
								delete(source.neighbors, targetID)
							}

							if m[targetID][sourceID] == 0 {
								delete(target.neighbors, sourceID)
							}

							source.d--
							target.d--
							found = true
							break
						}
					}
					if n == 1000 {
						nbLeftVertices++
						break
					}
				}
			}
		}
	}
	var tombstonesInversé = make(map[int][]int) //tombstones -> []vertexIDs
	sumToAdd := 0
	for vertexID, t := range tombstones {
		tombstonesInversé[t] = append(tombstonesInversé[t], vertexID)
	}
	for t, vertexIDs := range tombstonesInversé {
		sumToAdd += t * len(vertexIDs)
		for _, vertexID := range vertexIDs {
			comID := splittedComs.findVertexCom(vertexID)
			splittedComs.communitities[comID].addPools.addPool(t)
			splittedComs.communitities[comID].delPools.addPool(t)

			vertex := coms.communitities[comID].graph.vertices[vertexID]

			splittedComs.communitities[comID].addPools.addExistVertex(vertex, t)
			splittedComs.communitities[comID].delPools.addExistVertex(vertex, t)
		}
	}
	splittedComs.computeCDFPools()
	connectVerticesComsSplit(dAdd/2, splittedComs, coms)
	coms.clearAddPools()
	coms.updatePc()
}
func (coms *coms) clearAddPools() {
	for _, com := range coms.communitities {
		com.addPools.CDFpools = nil
		com.addPools.pools = make(map[int]*pool)
		com.addPools.nbEdges = 0
		com.addPools.w = 0
	}
}
func (coms *coms) removeCom(comID int) {
	//remove comID from comIDs, communities, CDFCommunities
	delete(coms.communitities, comID)
	for i, id := range coms.comIDs {
		if comID == id {
			coms.comIDs = append(coms.comIDs[:i], coms.comIDs[i+1:]...)
		}
	}
}

func breakEdges(coms *coms, splittedComs *coms) (inEdges map[int]int, outEdges map[int]map[int]int) {
	inEdges = make(map[int]int)
	outEdges = make(map[int]map[int]int)
	nbleftVertices := 0
	for _, comSource := range splittedComs.communitities {
		outEdges[comSource.id] = make(map[int]int)
		for _, vertex := range comSource.graph.vertices {
			for _, neighborID := range vertex.neighborsIDs {
				comSource.w++
				neighbor := vertex.neighbors[neighborID]
				//find the community of the neighbor
				found := false
				n := 0
				for !found {
					n++
					for _, comTarget := range coms.communitities {
						if _, ok := comTarget.graph.vertices[neighbor.id]; ok {
							found = true
							if comSource.id == comTarget.id {
								inEdges[comSource.id]++
							} else {
								_, secondCond := splittedComs.communitities[comTarget.id]
								if secondCond {
									outEdges[comSource.id][comTarget.id]++

								} else {
									outEdges[comSource.id][comTarget.id] += 2
								}
							}
							break
						}
					}
					if n == 1000 {
						nbleftVertices++
						break
					}
				}
			}
		}
	}
	return
}

func (coms *coms) updatePc() {
	ctcs := coms.CDFCommunities
	nbEdges := coms.nbEdges
	comPcsNew := make(map[int]float64)
	comPcsOld := make(map[int]float64)
	for comID, com := range coms.communitities {
		comPcsOld[comID] = com.pc
	}
	for _, ctc := range ctcs {
		comSource := ctc.comA
		comTarget := ctc.comB
		nw := float64(ctc.nbEdges) / float64(nbEdges)
		if compareComs(comSource, comTarget) {
			comPcsNew[comSource.id] += nw
		} else {
			comPcsNew[comSource.id] += nw * 0.5
			comPcsNew[comTarget.id] += nw * 0.5
			//comPcsNew[comSource.id] += nw
			//comPcsNew[comTarget.id] += nw
		}
	}
	sumpc := 0.0
	for _, com := range coms.communitities {
		//com.pc = comPcsNew[comID]
		com.pc = float64(com.w) / float64(nbEdges)
		sumpc += com.pc
	}
}
func compareComs(comA *com, comB *com) bool {
	return comA.id == comB.id
}
