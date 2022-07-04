package extendedChungLu

import (
	"bufio"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
)

type transformation struct {
	sd int
	td int
	p  float64
}
type transformations struct {
	ts map[int]map[int]transformation // sd -> td -> p
}

func initTs() *transformations {
	return &transformations{
		ts: make(map[int]map[int]transformation),
	}
}
func getEMD(path string) (sline string) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Could not open file: %v", err)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	line, err := r.ReadBytes('\n')
	if err != nil {
		log.Fatalf("Could not hin read line: %v", err)
	}
	sline = string(line)[0 : len(string(line))-4]
	return
}

func getTransformations(degreeSource []int, degreeTarget []int, path string) (transformations []transformation, ts *transformations) {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Could not open file: %v", err)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	ts = initTs()
	for i := range degreeSource {
		//Read Line for each portion != 0.0, add a transformation
		line, err := r.ReadBytes('\n')
		if err != nil {
			log.Fatalf("Could not read line: %v", err)
		}
		sline := string(line)[0 : len(string(line))-4]
		splittedLine := strings.Split(sline, " ")
		ts.ts[degreeSource[i]] = make(map[int]transformation)
		for j := 0; j < len(splittedLine); j++ {
			p := splittedLine[j]
			if p != "0.00000000000000000" && p != "0.000000000000000" {
				//transform i to j with portion = p
				p, err := strconv.ParseFloat(p, 64) //convert string to float64
				if err != nil {
					log.Fatalf("Could not parse string to float: %v", err)
				}
				t := transformation{sd: degreeSource[i],
					td: degreeTarget[j],
					p:  p} //convert string to float64
				transformations = append(transformations, t)
				ts.ts[degreeSource[i]][degreeTarget[j]] = t
			}

		}
	}
	return
}

func addVertices(graph *graph, addPools *pools, mu, sigma float64, n int) {
	vertices, _ := generateNormalVert(mu, sigma, n, len(graph.vertices))
	for degree, vertexSlice := range vertices {
		if _, ok := addPools.pools[degree]; !ok {
			addPools.addPool(degree)
		}
		for _, vertex := range vertexSlice {
			graph.addVertex(vertex)
			addPools.addExistVertex(vertex, degree)
		}
	}
}
func addVerticesComsZipf(graph *graph, coms *coms, n int, sRand int, s float64, v float64, imax int, nbEdges int) {
	nbVertices := coms.getTotNbVertices()
	//vertices, sumDeg := generateNormalVert(mu, sigma, n, nbVertices)
	vertices, sumDeg := generateZipfVert(n, sRand, s, v, imax, nbVertices)
	coms.dcAdd += sumDeg
	//nbEdgesOld := coms.nbEdges
	nbEdges += sumDeg
	verify := 0
	sumPC := 0.0
	sumW := 0
	sumdcAdd := 0
	for _, com := range coms.communitities {
		com.dc = int(com.pc*float64(nbEdges)) - com.w - com.dcAdd
		sumPC += com.pc
		sumW += com.w
		sumdcAdd += com.dcAdd
		verify += com.dc
	}
	for degree, vertexSlice := range vertices {
		for _, vertex := range vertexSlice {
			found := false
			nbRounds := 0
			for !found {
				nbRounds++
				com := chooseCom(coms)
				if _, ok := com.addPools.pools[degree]; !ok {
					com.addPools.addPool(degree)
				}
				if com.dc > 0 {

					com.addPools.addExistVertex(vertex, degree)
					com.graph.addVertex(vertex)
					graph.addVertex(vertex)
					com.dc -= degree
					found = true
				}
				if nbRounds == 100 {
					break
				}
			}

		}
	}
	for _, com := range coms.communitities {
		fmt.Printf("com %v dc: %v \n", com.id, com.dc)
	}
}
func addVerticesComs(graph *graph, coms *coms, mu, sigma float64, n int, nbEdges int) {
	nbVertices := coms.getTotNbVertices()
	vertices, sumDeg := generateNormalVert(mu, sigma, n, nbVertices)
	oldDcAdd := coms.dcAdd
	coms.dcAdd += sumDeg
	fmt.Printf("sumDeg %v, dc add %v \n", sumDeg, coms.dcAdd)
	nbEdgesOld := coms.nbEdges
	nbEdges += sumDeg
	fmt.Printf("number of edges new %v old %v \n", nbEdges, nbEdgesOld)
	verify := 0
	sumPC := 0.0
	sumW := 0

	sumdcAdd := 0
	for _, com := range coms.communitities {
		com.dc = int(com.pc*float64(nbEdges)) - com.w - com.dcAdd
		sumPC += com.pc
		sumW += com.w
		sumdcAdd += com.dcAdd
		fmt.Printf("com id: %v, dc: %v , w: %v, dcAdd: %v \n", com.id, com.dc, com.w, com.dcAdd)
		fmt.Printf("com id: %v, dc: %v \n", com.id, com.dc)
		verify += com.dc
	}
	fmt.Printf("dcAdd exp: %v obt: %v \n", oldDcAdd, sumdcAdd)
	fmt.Printf("nbEgdes exp: %v obt: %v \n", nbEdgesOld, sumW)
	fmt.Printf("pc exp: 1 obt: %v \n", sumPC)
	fmt.Printf("expected: %v obtained: %v \n", sumDeg, verify)
	for degree, vertexSlice := range vertices {
		for _, vertex := range vertexSlice {
			found := false
			nbRounds := 0
			for !found {
				nbRounds++
				com := chooseCom(coms)
				if _, ok := com.addPools.pools[degree]; !ok {
					com.addPools.addPool(degree)
				}
				if com.dc > 0 {

					com.addPools.addExistVertex(vertex, degree)
					com.graph.addVertex(vertex)
					graph.addVertex(vertex)
					com.dc -= degree
					found = true
				}
				if nbRounds == 100 {
					break
				}
			}

		}
	}
	for _, com := range coms.communitities {
		fmt.Printf("com %v dc: %v \n", com.id, com.dc)
	}
}

//creates transformation pools and fill them with references towards corresponding vertices
func createTransPools(graph *graph, ts []transformation) (addPools, delPools *pools) {
	addPools = initPools()
	delPools = initPools()
	nbVertices := len(graph.vertices)
	for _, t := range ts {
		nbMod := int(math.Ceil(float64(nbVertices) * t.p))
		if t.sd < t.td {
			//addPools <- initPool with w = t.td - t.sd
			delta := t.td - t.sd
			if _, ok := addPools.pools[delta]; !ok {
				addPools.addPool(delta)
			}
			for i := 0; i < nbMod; i++ {
				//choose a vertex with degree t.sd and insert it in pool
				//then remove vertex from candidate vertices slice
				if len(graph.degreeGraph[t.sd]) != 0 {
					vertex := graph.degreeGraph[t.sd][0]
					addPools.addExistVertex(vertex, delta)
					graph.degreeGraph[t.sd] = graph.degreeGraph[t.sd][1:len(graph.degreeGraph[t.sd])]
				}

			}
		} else if t.sd > t.td {
			delta := t.sd - t.td
			if _, ok := delPools.pools[delta]; !ok {
				delPools.addPool(delta)
				fmt.Printf("Pool delta: %v and degree source: %v, p: %v, nbMod: %v \n", delta, t.sd, t.p, nbMod)
			}
			for i := 0; i < nbMod; i++ {
				if len(graph.degreeGraph[t.sd]) != 0 {
					vertex := graph.degreeGraph[t.sd][0]
					delPools.addExistVertex(vertex, delta)
					graph.degreeGraph[t.sd] = graph.degreeGraph[t.sd][1:len(graph.degreeGraph[t.sd])]
				}
			}
		}
	}
	fmt.Printf("expected deletions: %v \n", delPools.nbEdges)
	return
}

//creates transformation pools and fill them with references towards corresponding vertices
func createHasTransPools(graph *graph, ts []transformation) (addPools, delPools *pools) {
	addPools = initPools()
	delPools = initPools()
	nbVertices := len(graph.vertices)
	for _, t := range ts {
		nbMod := int(math.Ceil(float64(nbVertices) * t.p))
		if t.sd < t.td {
			//addPools <- initPool with w = t.td - t.sd
			delta := t.td - t.sd
			if _, ok := addPools.pools[delta]; !ok {
				addPools.addPool(delta)
			}
			for i := 0; i < nbMod; i++ {
				//choose a vertex with degree t.sd and insert it in pool
				//then remove vertex from candidate vertices slice
				if len(graph.degreeHasGraph[t.sd]) != 0 {
					vertex := graph.degreeHasGraph[t.sd][0]
					addPools.addExistHasVertex(vertex, delta)
					graph.degreeHasGraph[t.sd] = graph.degreeHasGraph[t.sd][1:len(graph.degreeHasGraph[t.sd])]
				}

			}
		} else if t.sd > t.td {
			delta := t.sd - t.td
			if _, ok := delPools.pools[delta]; !ok {
				delPools.addPool(delta)
				fmt.Printf("Pool delta: %v and degree source: %v, p: %v, nbMod: %v \n", delta, t.sd, t.p, nbMod)
			}
			for i := 0; i < nbMod; i++ {
				if len(graph.degreeHasGraph[t.sd]) != 0 {
					vertex := graph.degreeHasGraph[t.sd][0]
					delPools.addExistVertex(vertex, delta)
					graph.degreeHasGraph[t.sd] = graph.degreeHasGraph[t.sd][1:len(graph.degreeHasGraph[t.sd])]
				}
			}
		}
	}
	fmt.Printf("expected deletions: %v \n", delPools.nbEdges)
	return
}

func plotEMDs(emds []float64, path string) {

	p := plot.New()
	p.Title.Text = "EMD values versus the number of iterations"
	p.X.Label.Text = "Number of iteration"
	p.Y.Label.Text = "EMD"

	pxysRes := make(plotter.XYs, len(emds))
	i := 0
	for t, emd := range emds {
		pxysRes[i].X = float64(t)
		pxysRes[i].Y = emd
		i++
	}
	sRes, err := plotter.NewScatter(pxysRes)
	if err != nil {
		log.Fatalf("Could not create scatter: %v", err)
	}
	sRes.GlyphStyle.Shape = draw.CrossGlyph{}
	sRes.Color = color.RGBA{R: 255, A: 255}

	// Make a line plotter with points and set its style.
	lpLine, lpPoints, err := plotter.NewLinePoints(pxysRes)
	if err != nil {
		panic(err)
	}
	lpLine.Color = color.RGBA{G: 255, A: 255}
	lpPoints.Shape = draw.PyramidGlyph{}
	lpPoints.Color = color.RGBA{R: 255, A: 255}

	p.Add(sRes, lpLine, lpPoints)

	wt, err := p.WriterTo(512, 256, "png")
	if err != nil {
		log.Fatalf("Could not create writer: %v", err)
	}
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Could not create file: %v", err)
	}
	defer f.Close()

	_, err = wt.WriteTo(f)
	if err != nil {
		log.Fatalf("could not write to out.png: %v", err)
	}
}

func getEquations(ts transformations, fromDist []int, toDist []int, fromDistMap map[int]int, toDistMap map[int]int, coms *coms) {
	//n = len of fromDist
	//m = len of toDist
	//nc = len of coms
	nbVertices := coms.getTotNbVertices()
	n := len(fromDist)
	m := len(toDist)
	nc := len(coms.communitities)
	//fmt.Printf("n: %v, m: %v, n*m: %v \n", n, m, n*m)
	var diff []int
	for _, ds := range fromDist {
		for _, dt := range toDist {
			//ds -> dt
			diff = append(diff, (dt - ds))
		}
	}
	var eqsOneA [][]int   // [nc, nc*n*m]
	var eqsOneB []float64 // [nc, 1]

	var eqsTwoA [][]int   // [n*m, nc*n*m]
	var eqsTwoB []float64 // [n*m, 1]

	var eqsThreeA [][]int   //[nc*n, nc*n*m]
	var eqsThreeB []float64 //[nc*n, 1]

	var eqFourA []int
	var eqFourB int
	totNb := coms.dcAdd

	//fmt.Printf("Première étape \n")
	for i := 0; i < nc; i++ {
		comID := coms.comIDs[i]
		k := 0
		compute := false
		var vec []int
		for j := 0; j < nc*n*m; j++ {
			if j == i*(n*m) {
				compute = true
			}
			if k == n*m {
				compute = false
			}
			if compute {
				vec = append(vec, diff[k])
				k++
			} else {
				vec = append(vec, 0)
			}
		}
		eqsOneA = append(eqsOneA, vec)
		eqsOneB = append(eqsOneB, float64(totNb)*coms.communitities[comID].pc/float64(nbVertices))
		//nbOp := float64(totNb) * coms.communitities[comID].pc / float64(nbVertices)
		//fmt.Printf("for community %v, number of operations: %v \n", i, nbOp)
	}
	//deuxieme etape somme(dij)
	//fmt.Printf("deuxieme étape \n")
	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			var vec []int
			for r := 0; r < nc; r++ {

				var vec1 []int
				for k := 0; k < n*m; k++ {
					if k == i*m+j {
						vec1 = append(vec1, 1)
					} else {
						vec1 = append(vec1, 0)
					}

				}
				vec = append(vec, vec1...)

			}
			eqsTwoA = append(eqsTwoA, vec)
			ds := fromDist[i]
			dt := toDist[j]
			t := ts.ts[ds][dt]
			//fmt.Printf("from d[%v] to d[%v] is equal to %v\n", ds, dt, t.p)
			eqsTwoB = append(eqsTwoB, t.p)
		}

	}
	//troisieme etape
	//fmt.Printf("Troisieme étape \n")
	for i := 0; i < nc; i++ {
		comID := coms.comIDs[i]
		for j := 0; j < n; j++ {
			k := 0
			compute := false
			var vec []int
			for r := 0; r < nc*n*m; r++ {
				if r == i*(n*m)+j*m {
					compute = true
				}
				if compute {
					vec = append(vec, 1)
					k++
					if k == m {
						compute = false
					}
				} else {
					vec = append(vec, 0)
				}
			}
			eqsThreeA = append(eqsThreeA, vec)
			p := float64(len(coms.communitities[comID].graph.degreeGraph[fromDist[j]])) / float64(nbVertices)
			//fmt.Printf("for pij %v we got %v in com: %v\n", fromDist[j], p, comID)
			eqsThreeB = append(eqsThreeB, p)
		}
	}

	//Quatrieme etape
	for i := 0; i < nc*n*m; i++ {
		eqFourA = append(eqFourA, 1)
	}
	eqFourB = 1

	path := "extendedChungLu/"
	f, err := os.Create(path + "A.txt")
	if err != nil {
		log.Fatalf("Could not create file: %v", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, eq := range eqsOneA {
		for i, a := range eq {
			if i == len(eq)-1 {
				_, err = w.WriteString(strconv.FormatInt(int64(a), 10))
			} else {

				_, err = w.WriteString(strconv.FormatInt(int64(a), 10) + " ")
			}
		}
		_, err = w.WriteString("\n")

	}
	for _, eq := range eqsTwoA {
		for i, a := range eq {
			if i == len(eq)-1 {
				_, err = w.WriteString(strconv.FormatInt(int64(a), 10))
			} else {

				_, err = w.WriteString(strconv.FormatInt(int64(a), 10) + " ")
			}
		}
		_, err = w.WriteString("\n")

	}
	for _, eq := range eqsThreeA {
		for i, a := range eq {
			if i == len(eq)-1 {
				_, err = w.WriteString(strconv.FormatInt(int64(a), 10))
			} else {

				_, err = w.WriteString(strconv.FormatInt(int64(a), 10) + " ")
			}
		}
		_, err = w.WriteString("\n")

	}

	for i, a := range eqFourA {
		if i == len(eqFourA)-1 {
			_, err = w.WriteString(strconv.FormatInt(int64(a), 10))
		} else {

			_, err = w.WriteString(strconv.FormatInt(int64(a), 10) + " ")
		}
	}
	w.Flush()

	f1, err := os.Create(path + "B.txt")
	if err != nil {
		log.Fatalf("Could not create file: %v", err)
	}
	defer f1.Close()
	w = bufio.NewWriter(f1)
	for _, b := range eqsOneB {
		_, err = w.WriteString(fmt.Sprintf("%f", b) + "\n")
	}
	for _, b := range eqsTwoB {
		_, err = w.WriteString(fmt.Sprintf("%f", b) + "\n")
	}
	for _, b := range eqsThreeB {
		_, err = w.WriteString(fmt.Sprintf("%f", b) + "\n")
	}
	_, err = w.WriteString(fmt.Sprintf("%f", float64(eqFourB)) + "\n")
	w.Flush()
	//write matrices into a file
	//solve problem in python
}

func createSystemPools(coms *coms, firstDist []int, finalDist []int, ts transformations) {
	nbVertices := coms.getTotNbVertices()
	nc := len(coms.communitities)
	n := len(firstDist)
	m := len(finalDist)
	var pijcms []float64
	//totNb := coms.dcAdd
	var fakeComs = make(map[int]map[int]int) //comsID -> degree -> nbOcc
	for comID, com := range coms.communitities {
		fakeComs[comID] = make(map[int]int)
		for degree, vertices := range com.graph.degreeGraph {
			fakeComs[comID][degree] = len(vertices)
		}
		com.addPools = initPools()
	}

	path := "extendedChungLu/"
	f, err := os.Open(path + "pijcm.txt")
	if err != nil {
		log.Fatalf("Could not open file: %v", err)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for i := 0; i < nc; i++ {
		comID := coms.comIDs[i]
		com := coms.communitities[comID]
		for j := 0; j < n; j++ {
			for k := 0; k < m; k++ {
				line, err := r.ReadBytes('\n')
				if err != nil {
					log.Fatalf("Could not read line: %v", err)
				}
				pString := string(line)[0 : len(string(line))-2]
				if pString != "0.00000000e+00" {
					pijcm, err := strconv.ParseFloat(pString, 64) //transform to float64
					if err != nil {
						fmt.Printf("unable to parse float64, ERROR: %v \n", err)
					}
					pijcms = append(pijcms, pijcm)
					if pijcm >= (float64(1)/float64(nbVertices)+0.1/float64(nbVertices))/float64(2) {
						//createAdd Pool
						delta := finalDist[k] - firstDist[j]
						com.addPools.addPool(delta)
						l := int(math.Ceil(float64(nbVertices) * pijcm))
						for x := 0; x < l; x++ {
							//fmt.Printf("len of degree: %v com %v is %v left: %v \n", firstDist[j], i, len(coms.communitities[i].graph.degreeGraph[firstDist[j]]), l-x)
							if len(com.graph.degreeGraph[firstDist[j]]) != 0 {
								vertex := com.graph.degreeGraph[firstDist[j]][0]
								if _, ok := com.addPools.pools[delta]; !ok {
									com.addPools.addPool(delta)
								}
								com.addPools.addExistVertex(vertex, delta)
								//should add some exception
								com.graph.degreeGraph[firstDist[j]] = com.graph.degreeGraph[firstDist[j]][1:len(com.graph.degreeGraph[firstDist[j]])] //TODO transform this to a function
							}
						}
					}
				}
			}
		}
	}
	var diff []int
	for _, ds := range firstDist {
		for _, dt := range finalDist {
			//ds -> dt
			diff = append(diff, (dt - ds))
		}
	}
	//fmt.Printf("Première étape \n")
	for i := 0; i < nc; i++ {
		//comID := coms.comIDs[i]
		//com := coms.communitities[comID]
		k := 0
		compute := false
		var vec []int
		var sommeCm float64
		for j := 0; j < nc*n*m; j++ {
			if j == i*(n*m) {
				compute = true
			}
			if k == n*m {
				compute = false
			}
			if compute {
				vec = append(vec, diff[k])
				sommeCm += float64(diff[k]) * pijcms[j]
				k++
			} else {
				vec = append(vec, 0)
			}
		}
		//nbOp := float64(totNb) * com.pc / float64(nbVertices)
		//fmt.Printf("expected value: %v obtained somme %v \n", sommeCm, nbOp)
	}
	//deuxieme etape somme(pij)
	//fmt.Printf("deuxieme étape \n")
	sommel := 0
	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			var somme float64
			var vec []int
			sommel = 0
			for r := 0; r < nc; r++ {
				var vec1 []int
				for k := 0; k < n*m; k++ {
					if k == i*m+j {
						if pijcms[r*(n*m)+k] > float64(1)/float64(nbVertices) {
							l := int(math.Ceil(float64(nbVertices) * pijcms[r*(n*m)+k]))
							sommel += l
							//fmt.Printf("pijcms: %v\n", pijcms[r*(n*m)+k])
							//fmt.Printf("com: %v, l: %v, ts: %v ,td: %v \n", coms.comIDs[r], l, firstDist[i], finalDist[j])
						}
						somme += pijcms[r*(n*m)+k]
						vec1 = append(vec1, 1)
					} else {
						vec1 = append(vec1, 0)
					}

				}
				vec = append(vec, vec1...)

			}
			//make sure that if we take math.Ceil of pijcms and then eliminate the picijm < 1/nbVertices andprint for each pijcm the number of vertices aplying the modification
			//then perform the sum
			/*	ds := firstDist[i]
				dt := finalDist[j]
				t := ts.ts[ds][dt]*/
			//fmt.Printf("expected %v obtained %v \n", t.p, somme)
			/*if int(math.Ceil(t.p*float64(nbVertices))) != sommel {

				fmt.Printf("expected %v obtained %v \n", int(math.Ceil(t.p*float64(nbVertices))), sommel)
			}*/

		}

	}
	//troisieme etape
	//fmt.Printf("Troisieme étape \n")
	for i := 0; i < nc; i++ {
		for j := 0; j < n; j++ {
			var somme float64
			k := 0
			compute := false
			var vec []int
			for r := 0; r < nc*n*m; r++ {
				if r == i*(n*m)+j*m {
					compute = true
				}
				if compute {
					vec = append(vec, 1)
					somme += pijcms[r]
					k++
					if k == m {
						compute = false
					}
				} else {
					vec = append(vec, 0)
				}
			}
			//p := float64(fakeComs[i][firstDist[j]]) / float64(nbVertices)
		}
	}

}
