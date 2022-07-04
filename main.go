package main

import (
	"flag"

	"github.com/MariaMassri/RTGen/extendedChungLu"
)

func main() {
	fOutputPath := flag.String("outputPath", "outputDatasets\\statGauss\\", "Path of the output graphs")

	fTask := flag.String("task", "GenStatGauss", "GenStaticGauss, GenStatZipf, GenStatGaussCom, GenStatZipfCom, GenDynGaussCom, GenDynZipfCom")

	fNbVertices := flag.Int("nbVertices", 10000, "Initial number of vertices")

	fMu := flag.Float64("mu", 30, "Parameter mu of the Gaussian distribution")

	fSigma := flag.Float64("sigma", 2.0, "Parameter sigma of the Gaussian distribution")

	fS := flag.Float64("s", 1.1, "Parameter s of the Zifpian degree distribution")

	fV := flag.Float64("v", 10.0, "Parameter v of the Zifpian degree distribution")

	fImax := flag.Int("iMax", 10, "Parameter iMax of the Zifpian degree distribution")

	fH := flag.Int("h", 2, "Height of the community hierarchical tree")

	fB := flag.Int("b", 2, "Branching factor of the community hierarchical tree")

	fK := flag.Float64("k", 20, "Parameter k of the community hierarchical tree controlling the within and between linkage probabilities")

	fNbSnapshots := flag.Int("nbSnapshots", 5, "Total number of snapshots")

	fDeltaMu := flag.Float64("dMu", 10.0, "Variation of the parameter Mu between successive snapshots")

	fDeltaSigma := flag.Float64("dSigma", 0, "Variation of the parameter Sigma between successive snapshots")
	//s,v, imax

	fDeltaS := flag.Float64("dS", 0.0, "Variation of the parameter s between successive snapshots")

	fDeltaV := flag.Float64("dV", 0.0, "Variation of the parameter v between successive snapshots")

	fDeltaImax := flag.Int("dImax", 5, "Variation of the parameter imax between successive snapshots")

	fOuputFormat := flag.String("outputFormat", "events", "Format of the returned temporal graphs: events or snapshots")

	flag.Parse()

	outputPath := *fOutputPath
	task := *fTask
	nbVertices := *fNbVertices
	mu := *fMu
	sigma := *fSigma
	s := *fS
	v := *fV
	imax := *fImax
	h := *fH
	b := *fB
	k := *fK
	nbSnapshots := *fNbSnapshots
	deltaMu := *fDeltaMu
	deltaSigma := *fDeltaSigma
	deltaS := *fDeltaS
	deltaV := *fDeltaV
	deltaImax := *fDeltaImax
	outputFormat := *fOuputFormat

	switch task {
	case "GenStatGauss":
		extendedChungLu.GenStatGauss(nbVertices, mu, sigma, outputPath)
		break
	case "GenStatZipf":
		extendedChungLu.GenStatZipf(nbVertices, s, v, imax, outputPath)
		break
	case "GenStatGaussCom":
		extendedChungLu.GenStatGaussCom(nbVertices, mu, sigma, h, b, k, outputPath)
		break
	case "GenStatZipfCom":
		extendedChungLu.GenStatZipfCom(nbVertices, s, v, imax, h, b, k, outputPath)
		break
	case "GenDynGaussCom":
		extendedChungLu.GenDynGaussCom(nbSnapshots, nbVertices, mu, sigma, h, b, deltaMu, deltaSigma, outputPath, outputFormat)
		break
	case "GenDynZipfCom":
		extendedChungLu.GenDynZipfCom(nbSnapshots, nbVertices, s, v, imax, deltaS, deltaV, deltaImax, h, b, k, outputPath, outputFormat)
		break
	}
}
