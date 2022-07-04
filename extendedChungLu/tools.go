package extendedChungLu

import (
	"bufio"
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/tealeg/xlsx"
	"github.com/xuri/excelize/v2"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func randStringGen(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

type distributionPoint struct {
	degree       int
	nbOccurences int
	timestamp    int
}

func appendPoints(distributionPoints []distributionPoint, timestamp int, distribution map[int]int) []distributionPoint {
	for degree, nbOccurences := range distribution {
		point := distributionPoint{
			degree:       degree,
			nbOccurences: nbOccurences,
			timestamp:    timestamp,
		}
		distributionPoints = append(distributionPoints, point)
	}
	return distributionPoints
}

func writeSurfaceDistributionToExcell(distributionPoints []distributionPoint, outputPath string) {
	file, err := excelize.OpenFile(outputPath)
	if err != nil {
		log.Fatalf("Could not open file: %v", err)
	}
	//index := file.NewSheet("test")
	file.NewSheet("distributionPoints")
	streamWriter, err := file.NewStreamWriter("distributionPoints")
	if err != nil {
		fmt.Printf("cannot open Feuil 1, ERROR: %v\n", err)
	}
	rowID := 1
	for _, point := range distributionPoints {
		row := make([]interface{}, 3)
		row[0] = point.timestamp
		row[1] = point.degree
		row[2] = point.nbOccurences
		cell, _ := excelize.CoordinatesToCellName(1, rowID)
		streamWriter.SetRow(cell, row)
		rowID++
	}
	streamWriter.Flush()
	file.SaveAs(outputPath)
}

/*func writeDurationToExcell(outputPath string) {
	file, err := excelize.OpenFile(outputPath)
	if err != nil {
		log.Fatalf("Could not open file: %v", err)
	}
	//index := file.NewSheet("test")
	file.NewSheet(strconv.Itoa(timestamp))
	streamWriter, err := file.NewStreamWriter(strconv.Itoa(timestamp))
	if err != nil {
		fmt.Printf("cannot open Feuil 1, ERROR: %v\n", err)
	}
	rowID := 1
	for deg, nbOcc := range degreeDistribution {
		row := make([]interface{}, 3)
		row[0] = timestamp
		row[1] = deg
		row[2] = nbOcc
		cell, _ := excelize.CoordinatesToCellName(1, rowID)
		streamWriter.SetRow(cell, row)
		rowID++
	}
	streamWriter.Flush()
	file.SaveAs(outputPath)
}*/

func writeToExcell(degreeDistribution map[int]int, timestamp int, outputPath string) {
	file, err := excelize.OpenFile(outputPath)
	if err != nil {
		log.Fatalf("Could not open file: %v", err)
	}
	//index := file.NewSheet("test")
	file.NewSheet(strconv.Itoa(timestamp))
	streamWriter, err := file.NewStreamWriter(strconv.Itoa(timestamp))
	if err != nil {
		fmt.Printf("cannot open Feuil 1, ERROR: %v\n", err)
	}
	rowID := 1
	for deg, nbOcc := range degreeDistribution {
		row := make([]interface{}, 3)
		row[0] = timestamp
		row[1] = deg
		row[2] = nbOcc
		cell, _ := excelize.CoordinatesToCellName(1, rowID)
		streamWriter.SetRow(cell, row)
		rowID++
	}
	streamWriter.Flush()
	file.SaveAs(outputPath)
}
func plotFormExcell(outputPath string, paths ...string) {
	p := plot.New()
	/////////////////////////////////////////////////
	for _, path := range paths {
		excelFileName := path
		var dist = make(map[int]float64)
		var degSlice []int
		var distSlice []float64
		xlFile, err := xlsx.OpenFile(excelFileName)
		r := uint8(rand.Intn(255))
		g := uint8(rand.Intn(255))
		b := uint8(rand.Intn(255))
		if err != nil {
			fmt.Printf("Unable to open xlsx File, ERROR: %v \n", err)
		}
		for _, sheet := range xlFile.Sheets {
			degRow := sheet.Rows[0]
			distRow := sheet.Rows[1]
			for _, cell := range degRow.Cells {
				d, err := cell.Int()
				if err != nil {
					fmt.Printf("unable to parse degree to int, ERROR: %v \n", err)
				}
				degSlice = append(degSlice, d)

			}
			for _, cell := range distRow.Cells {
				p, err := cell.Float()
				if err != nil {
					fmt.Printf("unable to parse degree to int, ERROR: %v \n", err)
				}
				distSlice = append(distSlice, p)
			}
			for i, d := range degSlice {
				dist[d] = distSlice[i]
			}
		}
		pxysRes := make(plotter.XYs, len(dist))
		i := 0
		cumY := 0.0
		fmt.Printf("///////////////////new Distribution////////////")
		for deg, nbOcc := range dist {
			pxysRes[i].X = float64(deg)
			pxysRes[i].Y = float64(nbOcc)
			fmt.Printf("deg: %v, nbOcc: %v\n", pxysRes[i].X, pxysRes[i].Y)
			i++
		}
		sort.Slice(pxysRes, func(i, j int) bool {
			return pxysRes[i].X < pxysRes[j].X
		})
		for j := 0; j < len(pxysRes); j++ {
			cumY += pxysRes[j].Y
			pxysRes[j].Y = cumY

		}

		sRes, err := plotter.NewScatter(pxysRes)
		if err != nil {
			log.Fatalf("Could not create scatter: %v", err)
		}
		//sRes.GlyphStyle.Shape = draw.CrossGlyph{}
		sRes.Color = color.RGBA{R: r, G: g, B: b, A: 255}

		lpLine, _, err := plotter.NewLinePoints(pxysRes)
		if err != nil {
			panic(err)
		}

		lpLine.Color = color.RGBA{R: r, G: g, B: b, A: 255}
		//lpPoints.Shape = draw.PyramidGlyph{}
		//lpPoints.Color = color.RGBA{R: r, G: g, B: b, A: 255}
		//p.Add(sRes, lpLine, lpPoints)
		//p.Add(sRes, lpLine)
		p.Add(lpLine)
	}

	/////////////////////////////////////////////////
	wt, err := p.WriterTo(512, 256, "png")
	if err != nil {
		log.Fatalf("Could not create writer: %v", err)
	}
	f, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("Could not create file: %v", err)
	}
	defer f.Close()

	_, err = wt.WriteTo(f)
	if err != nil {
		log.Fatalf("could not write to out.png: %v", err)
	}
}

func plotDistributions(dist1 map[int]int, dist2 map[int]int, path string) {
	p := plot.New()
	p.Title.Text = "Comparison between two degree distributions"
	p.X.Label.Text = "Degree"
	p.Y.Label.Text = "Number of vertices"
	/////////////////////////////////////////////////
	lenDist1 := 0
	lenDist2 := 0
	for _, nbOcc := range dist1 {
		if nbOcc <= 1 {
			continue
		}
		lenDist1++
	}
	for _, nbOcc := range dist2 {
		if nbOcc <= 1 {
			continue
		}
		lenDist2++
	}
	pxysRes1 := make(plotter.XYs, lenDist1)
	i := 0
	for deg, nbOcc := range dist1 {
		if nbOcc <= 1 {
			continue
		}
		pxysRes1[i].X = float64(deg)
		pxysRes1[i].Y = float64(nbOcc)
		i++
	}
	sort.Slice(pxysRes1, func(i, j int) bool {
		return pxysRes1[i].X < pxysRes1[j].X
	})
	sRes1, err := plotter.NewScatter(pxysRes1)
	if err != nil {
		log.Fatalf("Could not create scatter: %v", err)
	}
	sRes1.GlyphStyle.Shape = draw.CrossGlyph{}
	sRes1.Color = color.RGBA{R: 255, A: 255}

	lpLine1, lpPoints1, err := plotter.NewLinePoints(pxysRes1)
	if err != nil {
		panic(err)
	}
	lpLine1.Color = color.RGBA{R: 255, A: 255}
	lpPoints1.Shape = draw.PyramidGlyph{}
	lpPoints1.Color = color.RGBA{R: 255, A: 255}
	/////////////////////////////////////////////////

	pxysRes2 := make(plotter.XYs, lenDist2)
	i = 0

	for deg, nbOcc := range dist2 {
		if nbOcc <= 1 {
			continue
		}
		pxysRes2[i].X = float64(deg)
		pxysRes2[i].Y = float64(nbOcc)
		i++
	}
	sort.Slice(pxysRes2, func(i, j int) bool {
		return pxysRes2[i].X < pxysRes2[j].X
	})
	sRes2, err := plotter.NewScatter(pxysRes2)
	if err != nil {
		log.Fatalf("Could not create scatter: %v", err)
	}
	sRes2.GlyphStyle.Shape = draw.CrossGlyph{}
	sRes2.Color = color.RGBA{B: 255, A: 255}

	lpLine2, lpPoints2, err := plotter.NewLinePoints(pxysRes2)
	if err != nil {
		panic(err)
	}
	lpLine2.Color = color.RGBA{B: 255, A: 255}
	lpPoints2.Shape = draw.PyramidGlyph{}
	lpPoints2.Color = color.RGBA{B: 255, A: 255}
	////////////////////////////////////////////////

	p.Add(sRes1, lpLine1, lpPoints1)
	p.Add(sRes2, lpLine2, lpPoints2)

	p.Legend.Add("Source distribution", sRes1, lpLine1, lpPoints1)
	p.Legend.Add("Target distribution", sRes2, lpLine2, lpPoints2)

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
func plotDistribution(distribution map[int]int, path string) {
	p := plot.New()
	lenDist := 0
	for _, nbOcc := range distribution {
		if nbOcc <= 1 {
			continue
		}
		lenDist++
	}
	pxysRes := make(plotter.XYs, lenDist)
	i := 0
	for deg, nbOcc := range distribution {
		if nbOcc <= 1 {
			continue
		}
		pxysRes[i].X = float64(deg)
		pxysRes[i].Y = float64(nbOcc)
		i++
	}
	sRes, err := plotter.NewScatter(pxysRes)
	if err != nil {
		log.Fatalf("Could not create scatter: %v", err)
	}
	sRes.GlyphStyle.Shape = draw.CrossGlyph{}
	sRes.Color = color.RGBA{R: 255, A: 255}
	p.Add(sRes)

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

func graphToNetworkXFile(graph *graph, path string) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Could not open file: %v", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	for _, vertex := range graph.vertices {
		for _, neighbor := range vertex.neighbors {
			line := strconv.Itoa(vertex.id) + " " + strconv.Itoa(neighbor.id) + "\n"
			w.WriteString(line)
		}
	}
}

func verticesToVertexFile(vertices map[int]*vertex, outputPath string) {
	f, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("Could not open file: %v", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	w.WriteString("Vertex;ComID\n")
	for _, vertex := range vertices {
		line := strconv.Itoa(vertex.id) + ";" + strconv.Itoa(vertex.comID)
		w.WriteString(line)
	}
	w.Flush()
}

func graphToEdgeFile(graph *graph, path string, temporal bool) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Could not open file: %v", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	w.WriteString("Source;Target\n")
	for _, vertex := range graph.vertices {
		for _, neighbor := range vertex.neighbors {
			line := ""
			if !temporal {
				line = strconv.Itoa(vertex.id) + ";" + strconv.Itoa(neighbor.id) + "\n"
			} else {
				line = strconv.Itoa(vertex.id) + ";" + strconv.Itoa(vertex.comID) + ";" + strconv.Itoa(neighbor.id) + ";" + strconv.Itoa(neighbor.comID) + "\n"
			}
			w.WriteString(line)
		}
	}
}

func graphToEdgeFileComs(graph *graph, vertexPath, edgePath string) {
	vertexFile, err := os.Create(vertexPath)
	if err != nil {
		log.Fatalf("Could not open file: %v", err)
	}
	edgeFile, err := os.Create(edgePath)
	if err != nil {
		log.Fatalf("Could not open file: %v", err)
	}
	defer vertexFile.Close()
	defer edgeFile.Close()
	vertexWriter := bufio.NewWriter(vertexFile)
	vertexWriter.WriteString("Vertex;Community\n")
	edgeWriter := bufio.NewWriter(edgeFile)
	edgeWriter.WriteString("Source;Target\n")

	for _, vertex := range graph.vertices {
		line := strconv.Itoa(vertex.id) + ";" + strconv.Itoa(vertex.comID) + "\n"
		vertexWriter.WriteString(line)
		for _, neighbor := range vertex.neighbors {
			line = strconv.Itoa(vertex.id) + ";" + strconv.Itoa(neighbor.id) + "\n"
			edgeWriter.WriteString(line)
		}
	}
}

func stringToTimePrime(timestamp string) (time.Time, error) {
	t, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Unix(0, 0), err
	}
	return time.Unix(t, 0), err
}

func stringToTime(timestamp string) (time.Time, error) {
	timestamp = strings.TrimSuffix(timestamp, "000")
	t, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return time.Unix(0, 0), err
	}
	return time.Unix(t, 0), err
}
func deleteFromSlice(slice []int, e int) []int {
	for i, s := range slice {
		if s == e {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return nil
}

func timeToString(timestamp time.Time) string {
	tint := timestamp.Unix()
	tint = tint * 1000
	return strconv.FormatInt(tint, 10)
}

func graphToOutput(graph *graph, coms *coms, path string, startTime time.Time) (err error) {
	timestamp := startTime
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	var communityIDs string
	for _, comID := range coms.comIDs {
		communityIDs += strconv.Itoa(comID) + ","
	}
	communityIDs = strings.Trim(communityIDs, ",")
	communityIDs += "\n"
	w.WriteString(communityIDs)
	for _, vertex := range graph.vertices {
		name := randStringGen(5)
		studyField := randStringGen(10)
		comID := strconv.Itoa(vertex.comID)
		id := strconv.Itoa(vertex.id)
		timestamp = timestamp.Add(time.Second)
		timestamps := timeToString(timestamp)
		w.WriteString("person," + comID + "," + id + "," + timestamps + "," + "start" + "," + name + "," + studyField + "\n")
	}
	for _, src := range graph.vertices {
		for _, tgt := range src.neighbors {
			bondtype := randStringGen(8)
			comIDSrc := strconv.Itoa(src.comID)
			srcID := strconv.Itoa(src.id)
			tgtID := strconv.Itoa(tgt.id)
			timestamp = timestamp.Add(time.Second)
			timestamps := timeToString(timestamp)
			w.WriteString("knows," + comIDSrc + "," + srcID + "," + timestamps + "," + tgtID + "," + "start" + "," + bondtype + "\n")

		}
	}
	fmt.Printf("last timestamp: %v \n", timestamp)
	err = w.Flush()
	if err != nil {
		return err
	}
	return
}

func randate(startTime, endTime time.Time) time.Time {
	min := startTime.Unix()
	max := endTime.Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}

func computeFrobeniusNorm(matrix mat.Dense) float64 {
	r, c := matrix.Dims()
	norm := 0.0
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			norm += math.Pow(matrix.At(i, j), 2)
		}
	}
	return math.Sqrt(norm)
}
