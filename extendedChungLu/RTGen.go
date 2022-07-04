package extendedChungLu

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"time"
)

func GenStatGauss(nbVertices int, mu, sigma float64, outputPath string) {
	fmt.Println("Generation started...")
	tStart := time.Now()
	pools, graph, _ := generateNormalDegDist(mu, sigma, nbVertices)
	connectVertices(pools)
	tEnd := time.Now()
	graphToEdgeFile(graph, outputPath+"edgeFile.csv", false)
	fmt.Println("Generation completed")
	fmt.Printf("Execution time: %v \n", tEnd.Sub(tStart))
}

func GenStatZipf(nbVertices int, s, v float64, imax int, outputPath string) {
	fmt.Println("Generating graph...")
	tStart := time.Now()
	pools, graph, _ := generateZipfianDegDist(nbVertices, 23, s, v, imax)
	connectVertices(pools)
	tEnd := time.Now()
	fmt.Printf("Generation completed, Execution time: %v \n", tEnd.Sub(tStart))
	fmt.Println("Writing graph in csv file...")
	graphToEdgeFile(graph, outputPath+"edgeFile.csv", false)
}

func GenStatGaussCom(nbVertices int, mu, sigma float64, h, b int, k float64, outputPath string) {
	fmt.Println("Generating graph...")
	pools, graph, _ := generateNormalDegDist(mu, sigma, nbVertices)
	coms := initComs()
	createHTreeComsT(h, b, 4, k, coms)
	coms.distributePools(pools)
	coms.computeCDF()
	coms.computeComNbEgdes()
	coms.distributeVertices(pools)
	coms.computeCDFPools()
	tStart := time.Now()
	connectVerticesComs(pools.nbEdges/2, coms)
	tEnd := time.Now()
	fmt.Printf("Generation completed, Execution time: %v \n", tEnd.Sub(tStart))
	fmt.Println("Writing graph in csv files...")
	graphToEdgeFileComs(graph, outputPath+"vertexFile.csv", outputPath+"edgeFile.csv")
}

func GenStatZipfCom(nbVertices int, s, v float64, imax int, h, b int, k float64, outputPath string) {
	fmt.Println("Generating graph...")
	pools, graph, _ := generateZipfianDegDist(nbVertices, 23, s, v, imax)
	coms := initComs()
	createHTreeComsT(h, b, 4, k, coms)
	coms.distributePools(pools)
	coms.computeCDF()
	coms.computeComNbEgdes()
	coms.distributeVertices(pools)
	coms.computeCDFPools()
	tStart := time.Now()
	connectVerticesComs(pools.nbEdges/2, coms)
	tEnd := time.Now()
	fmt.Printf("Generation completed, Execution time: %v \n", tEnd.Sub(tStart))
	fmt.Println("Writing graph in csv files...")
	graphToEdgeFileComs(graph, outputPath+"vertexFile.csv", outputPath+"edgeFile.csv")
}

func copyOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

func GenDynGaussCom(nbIterations, nbVertices int, mu, sigma float64, h, b int, deltaMu, deltaSigma float64, outputPath, outputFormat string) {
	nbEdges := 0
	nbIterations--
	pools, graph, _ := generateNormalDegDist(mu, sigma, nbVertices)
	var executionTime time.Duration
	pools.computeCDF()
	graph.getGivenDistribution()
	start := time.Unix(1273764761, 0)
	coms := initComsT(start, time.Minute, outputPath+"vertexFile.csv", outputPath+"edgeFile.csv")
	createHTreeComsT(h, b, 4, 20, coms)
	path := "extendedChungLu/"
	coms.distributePools(pools)
	coms.computeCDF()
	coms.computeComNbEgdes()

	coms.distributeVertices(pools)
	coms.computeCDFPools()
	nbEdges += pools.nbEdges
	tStart := time.Now()
	connectVerticesT(pools.nbEdges/2, coms)
	tEnd := time.Now()
	executionTime += tEnd.Sub(tStart)
	graph.getDistribution()
	if outputFormat == "snapshots" {
		graphToEdgeFile(graph, outputPath+"edgeFile_0.csv", false)
	}
	for i := 0; i < nbIterations; i++ {
		tStart = time.Now()

		_, graph2, _ := generateNormalDegDist(mu+float64(i)*deltaMu, sigma+float64(i)*deltaSigma, nbVertices)
		graph2.getGivenDistribution()

		degreeSource := extractDegreeDist(graph.degreeDistribution, path+"sourceDistribution.txt")
		degreeTarget := extractDegreeDist(graph2.givenDistribution, path+"targetDistribution.txt")

		cmd := exec.Command("POT/.venv/python.exe",
			"POT/.venv/OTSolver.py",
			"POT/.venv/OTSolver.py")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			panic(err)
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			panic(err)
		}
		err = cmd.Start()
		if err != nil {
			panic(err)
		}
		go copyOutput(stdout)
		go copyOutput(stderr)
		cmd.Wait()
		ts, trans := getTransformations(degreeSource, degreeTarget, path+"OTMATRIX.txt")

		coms.computeComNbEdges(ts)
		coms.generateDegreeGraphs()
		getEquations(*trans, degreeSource, degreeTarget, graph.degreeDistribution, graph2.givenDistribution, coms)
		cmd = exec.Command("POT/.venv/python.exe",
			"POT/.venv/EquationSolver.py",
			"POT/.venv/EquationSolver.py")
		stdout, err = cmd.StdoutPipe()
		if err != nil {
			panic(err)
		}
		stderr, err = cmd.StderrPipe()
		if err != nil {
			panic(err)
		}
		err = cmd.Start()
		if err != nil {
			panic(err)
		}
		go copyOutput(stdout)
		go copyOutput(stderr)
		cmd.Wait()
		createSystemPools(coms, degreeSource, degreeTarget, *trans)
		coms.computeCDFPools()
		connectVerticesComsT(coms)
		tEnd = time.Now()
		executionTime += tEnd.Sub(tStart)
		coms.updatePc()
		if outputFormat == "snapshots" {
			graphToEdgeFile(graph, outputPath+"edgeFile_"+strconv.Itoa(i+1)+".csv", false)
		}
		graph.getDistribution()

	}
	fmt.Printf("Generation completed, Execution time: %v \n", executionTime.Seconds())
	fmt.Printf("Total Number of edges: %v \n", nbEdges)
	fmt.Println("Writing graph in csv files...")
	if outputFormat == "events" {
		err := coms.vertexWriter.Flush()
		if err != nil {
			panic(err)
		}

		err = coms.edgeWriter.Flush()
		if err != nil {
			panic(err)
		}
	} else {
		verticesToVertexFile(graph.vertices, outputPath)
	}
}

func GenDynZipfCom(nbIterations, nbVertices int, s, v float64, imax int, deltaS, deltaV float64, deltaImax int, h, b int, k float64, outputPath, outputFormat string) {
	nbEdges := 0
	nbIterations--
	pools, graph, _ := generateZipfianDegDist(nbVertices, 23, s, v, imax)
	var executionTime time.Duration
	pools.computeCDF()
	graph.getGivenDistribution()
	start := time.Unix(1273764761, 0)
	coms := initComsT(start, time.Minute, outputPath+"vertexFile.csv", outputPath+"edgeFile.csv")
	createHTreeComsT(h, b, 4, 20, coms)
	path := "extendedChungLu/"
	coms.distributePools(pools)
	coms.computeCDF()
	coms.computeComNbEgdes()

	coms.distributeVertices(pools)
	coms.computeCDFPools()
	nbEdges += pools.nbEdges
	tStart := time.Now()
	connectVerticesT(pools.nbEdges/2, coms)
	tEnd := time.Now()
	executionTime += tEnd.Sub(tStart)
	graph.getDistribution()
	if outputFormat == "snapshots" {
		graphToEdgeFile(graph, outputPath+"edgeFile_0.csv", false)
	}
	for i := 0; i < nbIterations; i++ {
		tStart = time.Now()

		_, graph2, _ := generateZipfianDegDist(nbVertices, 23, s+float64(i)*deltaS, v+float64(i)*deltaV, imax+i*deltaImax)
		graph2.getGivenDistribution()

		degreeSource := extractDegreeDist(graph.degreeDistribution, path+"sourceDistribution.txt")
		degreeTarget := extractDegreeDist(graph2.givenDistribution, path+"targetDistribution.txt")

		cmd := exec.Command("POT/.venv/python.exe",
			"POT/.venv/OTSolver.py",
			"POT/.venv/OTSolver.py")
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			panic(err)
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			panic(err)
		}
		err = cmd.Start()
		if err != nil {
			panic(err)
		}
		go copyOutput(stdout)
		go copyOutput(stderr)
		cmd.Wait()
		fmt.Println("done")
		ts, trans := getTransformations(degreeSource, degreeTarget, path+"OTMATRIX.txt")

		coms.computeComNbEdges(ts)
		coms.generateDegreeGraphs()
		getEquations(*trans, degreeSource, degreeTarget, graph.degreeDistribution, graph2.givenDistribution, coms)
		cmd = exec.Command("POT/.venv/python.exe",
			"POT/.venv/EquationSolver.py",
			"POT/.venv/EquationSolver.py")
		stdout, err = cmd.StdoutPipe()
		if err != nil {
			panic(err)
		}
		stderr, err = cmd.StderrPipe()
		if err != nil {
			panic(err)
		}
		err = cmd.Start()
		if err != nil {
			panic(err)
		}
		go copyOutput(stdout)
		go copyOutput(stderr)
		cmd.Wait()
		createSystemPools(coms, degreeSource, degreeTarget, *trans)
		coms.computeCDFPools()
		connectVerticesComsT(coms)
		tEnd = time.Now()
		executionTime += tEnd.Sub(tStart)
		coms.updatePc()
		graph.getDistribution()
		if outputFormat == "snapshots" {
			graphToEdgeFile(graph, outputPath+"edgeFile_"+strconv.Itoa(i+1)+".csv", false)
		}
	}

	fmt.Printf("Generation completed, Execution time: %v \n", executionTime.Seconds())
	fmt.Printf("Total Number of edges: %v \n", nbEdges)
	fmt.Println("Writing graph in csv files...")
	if outputFormat == "events" {
		err := coms.vertexWriter.Flush()
		if err != nil {
			panic(err)
		}
		err = coms.edgeWriter.Flush()
		if err != nil {
			panic(err)
		}
	} else {
		verticesToVertexFile(graph.vertices, outputPath+"vertexFile.csv")
	}
}