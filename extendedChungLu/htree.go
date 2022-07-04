package extendedChungLu

import (
	"math"

	"github.com/DzananGanic/numericalgo/root"
	"gonum.org/v1/gonum/mat"
)

type tree struct {
	id        int
	parent    *tree
	children  []*tree
	distances map[int]int //map id of tree with distance
}

//TODO test + combine i = 0
func generateTree(h int, b int) (leaves []*tree) {
	var id int
	var levels = make(map[int][]*tree) // Li -> slice of trees in Li
	for i := 0; i < h; i++ {
		if i == 0 {
			// init root
			id++
			htree := &tree{id: id} //root no parent neither distances with other siblings in the same level
			levels[i] = append(levels[i], htree)
			for j := 0; j < b; j++ {
				//create root children init distance to 1
				id++
				t := &tree{id: id,
					parent:    htree,
					distances: make(map[int]int)}
				htree.children = append(htree.children, t)
				levels[i+1] = append(levels[i+1], t)
			}

		} else {
			ancestors := levels[i]
			//pseudo code
			//1. create children for all ancestor
			//2. for all created children, compute distances toward other siblings
			for _, ancestor := range ancestors {
				for j := 0; j < b; j++ {
					id++
					t := &tree{
						id:        id,
						parent:    ancestor,
						distances: make(map[int]int)}
					ancestor.children = append(ancestor.children, t)
					levels[i+1] = append(levels[i+1], t)
				}
			}
		}

		for n, child := range levels[i+1] {
			for m, sibling := range levels[i+1] {
				//compute distances between siblings
				if n != m {
					if i == 0 {
						child.distances[sibling.id] = 1
					} 
					child.distances[sibling.id] =
						child.parent.distances[sibling.parent.id] + 1
				}

			}
		}
	}
	leaves = levels[h]
	return
}

func getSumWeights(leaves []*tree, pinMin, kf float64, c int) (sum float64) {
	cf := float64(c)
	for _, leave := range leaves {
		for _, d := range leave.distances {
			df := float64(d)
			sum += 0.5 / math.Pow(cf, df)

		}
		sum += pinMin + kf
	}
	return
}

func createHTreeComsT(h, b, c int, k float64, coms *coms) {
	cf := float64(c)
	kf := float64(k)
	leaves := generateTree(h, b)
	pinMin := estimatePin(h, b, cf)
	sum := getSumWeights(leaves, pinMin, kf, c)
	zeros := []float64{}
	dim := int(math.Pow(float64(b), float64(h)))
	for i := 0; i < dim*dim; i++ {
		zeros = append(zeros, 0.0)
	}
	coms.sourceCommunityMatrix = *mat.NewDense(dim, dim, zeros)
	//create coms
	for _, leave := range leaves {
		comID := leave.id
		coms.addCom(comID)
	}
	idCounter := 0
	idMap := make(map[int]int)
	for _, leave := range leaves {
		comAID := leave.id
		if _, ok := idMap[comAID]; !ok {
			idMap[comAID] = idCounter
			idCounter++
		}
		comA := coms.communitities[comAID] //coms.addCom(comAID)
		for comBID, d := range leave.distances {
			if _, ok := idMap[comBID]; !ok {
				idMap[comBID] = idCounter
				idCounter++
			}
			df := float64(d)
			comB := coms.communitities[comBID] //coms.addCom(comBID)
			ctc := &comsToConnect{
				comA: comA,
				comB: comB,
				nw:   0.5 / math.Pow(cf, df) / sum,
			}
			coms.addCDFCom(ctc)
			coms.sourceCommunityMatrix.Set(idMap[comAID], idMap[comBID], 0.5/math.Pow(cf, df)/sum)
		}

		ctc := &comsToConnect{
			comA: comA,
			comB: comA,
			nw:   (pinMin + kf) / sum,
		}
		coms.sourceCommunityMatrix.Set(idMap[comAID], idMap[comAID], (pinMin+kf)/sum)
		coms.addCDFCom(ctc)
	}
	coms.comIDsMap = idMap
	return
}

func createHTreeComs(h, b, c int, k float64) (coms *coms) {
	coms = initComs()
	cf := float64(c)
	kf := float64(k)
	leaves := generateTree(h, b)
	pinMin := estimatePin(h, b, cf)
	sum := getSumWeights(leaves, pinMin, kf, c)
	//create coms
	for _, leave := range leaves {
		comID := leave.id
		coms.addCom(comID)
	}
	for _, leave := range leaves {
		comAID := leave.id
		comA := coms.communitities[comAID] //coms.addCom(comAID)
		for comBID, d := range leave.distances {
			df := float64(d)
			comB := coms.communitities[comBID] //coms.addCom(comBID)
			ctc := &comsToConnect{
				comA: comA,
				comB: comB,
				nw:   0.5 / math.Pow(cf, df) / sum,
			}
			coms.addCDFCom(ctc)
		}
		ctc := &comsToConnect{
			comA: comA,
			comB: comA,
			nw:   (pinMin + kf) / sum,
		}
		coms.addCDFCom(ctc)
	}
	return
}

func polynomialFunc(h int, b int, c float64) float64 {
	hf := float64(h)
	bf := float64(h)
	var returned float64
	for i := 1; i < h; i++ {
		ifl := float64(i)
		if i == 1 {
			returned += float64(b-2) * math.Pow(c, hf-1.0)
		} else {
			returned += (math.Pow(bf, ifl) -
				math.Pow(bf, ifl-1.0)) *
				math.Pow(c, hf-ifl)
		}

	}
	return returned
}
func estimatePin(h int, b int, c float64) (result float64) {
	bf := float64(b)
	for i := 1; i < h; i++ {
		ifl := float64(i)
		result += (math.Pow(bf, ifl) -
			math.Pow(bf, ifl-1.0)) /
			math.Pow(c, ifl)
	}
	return
}
func cSolver(h int, b int, e int, exp float64, iterations int) (float64, error) {
	hf := float64(h)
	bf := float64(h)
	f := func(x float64) float64 {
		var returned float64
		for i := 1; i <= h; i++ {
			ifl := float64(i)
			if i == 1 {
				returned += float64(b-1-e) * math.Pow(x, hf-1.0)
			} else {
				returned += (math.Pow(bf, ifl) -
					math.Pow(bf, ifl-1.0)) *
					math.Pow(x, hf-ifl)
			}

		}
		return returned
	}
	return root.Newton(f, exp, iterations)

}
