package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"

	"github.com/anton2920/imgo"
	"github.com/anton2920/imgo/gr"
)

type Neuron struct {
	Weights       []float32
	X, Y          float32
	Width, Height float32
}

type SOM struct {
	Neurons      []Neuron
	StartingRate float32
	CurrentRate  float32
	Radius       float32
	Time         float32
	Count        int
	MaxCount     int
	Trained      bool
}

func (n *Neuron) AdjustWeights(inputs []float32, rate, influence float32) {
	for i := 0; i < len(n.Weights); i++ {
		n.Weights[i] += rate * influence * (inputs[i] - n.Weights[i])
	}
}

func (n *Neuron) DistanceTo(inputs []float32) float32 {
	var distance float32

	for i := 0; i < len(n.Weights); i++ {
		distance += (inputs[i] - n.Weights[i]) * (inputs[i] - n.Weights[i])
	}

	return distance
}

func (n *Neuron) Render(renderer *imgo.Renderer) {
	const offset = 0

	var components [3]float32
	weights := [3]float32{1.0, 0.5, 0.0}

	for i := 0; i < len(components); i++ {
		for _, w := range n.Weights {
			components[i] += float32(math.Abs(float64(weights[i] * w)))
			weights[i] -= 1 / float32(len(n.Weights)-1)
		}
		components[i] *= 255
	}
	color := gr.ColorRGB(byte(components[0]), byte(components[1]), byte(components[2]))

	renderer.GraphSolidRectWH(int(n.X-n.Width*0.5)+offset, int(n.Y-n.Height*0.5)+offset, int(n.Width)-offset, int(n.Height)-offset, color)
}

func (s *SOM) Init(width, height int, nrows, ncols int, ninputs int, maxCount int, rate float32) {
	s.Count = 0
	s.MaxCount = maxCount

	s.Radius = float32(max(width, height)) / 2
	s.Time = float32(float64(s.MaxCount) / math.Log(float64(s.Radius)))

	s.StartingRate = rate
	s.CurrentRate = rate

	neuronWidth := float32(width) / float32(ncols)
	neuronHeight := float32(height) / float32(nrows)

	for row := 0; row < nrows; row++ {
		for col := 0; col < ncols; col++ {
			var neuron Neuron
			neuron.Weights = make([]float32, ninputs)
			for i := 0; i < ninputs; i++ {
				neuron.Weights[i] = rand.Float32()
			}
			neuron.X = neuronWidth * (float32(col) + 0.5)
			neuron.Y = neuronHeight * (float32(row) + 0.5)
			neuron.Width = neuronWidth
			neuron.Height = neuronHeight
			s.Neurons = append(s.Neurons, neuron)
		}
	}
}

/* FindBMU returns Best Matching Node: node which is the closest to input. */
func (s *SOM) FindBMU(inputs []float32) int {
	bmuIndex := 0
	minDist := s.Neurons[0].DistanceTo(inputs)

	for i := 1; i < len(s.Neurons); i++ {
		neuron := &s.Neurons[i]
		dist := neuron.DistanceTo(inputs)

		if dist < minDist {
			minDist = dist
			bmuIndex = i
		}
	}

	return bmuIndex
}

func (s *SOM) TrainStep(trainingData [][]float32) {
	if !s.Trained {
		inputs := trainingData[rand.Int()%len(trainingData)]
		bmuIndex := s.FindBMU(inputs)
		bmu := &s.Neurons[bmuIndex]

		neighbourhoodRadius := s.Radius * float32(math.Exp(float64(-s.Count)/float64(s.Time)))

		for i := 0; i < len(s.Neurons); i++ {
			neuron := &s.Neurons[i]

			distanceSquared := (bmu.X-neuron.X)*(bmu.X-neuron.X) + (bmu.Y-neuron.Y)*(bmu.Y-neuron.Y)
			radiusSquared := neighbourhoodRadius * neighbourhoodRadius

			if distanceSquared < radiusSquared {
				influence := float32(math.Exp(float64(-distanceSquared / (2 * radiusSquared))))
				neuron.AdjustWeights(inputs, s.CurrentRate, influence)
			}
		}

		s.CurrentRate = s.StartingRate * float32(math.Exp(float64(-s.Count)/float64(s.MaxCount)))
		s.Count++
		if s.Count >= s.MaxCount {
			s.Trained = true
		}
	}
}

func (s *SOM) Render(renderer *imgo.Renderer) {
	for i := 0; i < len(s.Neurons); i++ {
		s.Neurons[i].Render(renderer)
	}
}

func GenerateTrainingData(basis [][]float32, maxOffset float32, count int) [][]float32 {
	var result [][]float32

	for k := 0; k < count; k++ {
		i := rand.Int() % len(basis)

		row := make([]float32, len(basis[0]))
		for j := 0; j < len(row); j++ {
			row[j] = basis[i][j] + maxOffset*rand.Float32()
		}
		result = append(result, row)
	}

	return result
}

func NormalizeTrainingData(trainingData [][]float32) {
	minVector := make([]float32, len(trainingData[0]))
	maxVector := make([]float32, len(trainingData[0]))

	for j := 0; j < len(trainingData[0]); j++ {
		minVector[j] = trainingData[0][j]
		maxVector[j] = trainingData[0][j]
	}

	for i := 0; i < len(trainingData); i++ {
		for j := 0; j < len(trainingData[0]); j++ {
			minVector[j] = min(minVector[j], float32(math.Abs(float64(trainingData[i][j]))))
			maxVector[j] = max(maxVector[j], float32(math.Abs(float64(trainingData[i][j]))))
		}
	}

	for i := 0; i < len(trainingData); i++ {
		for j := 0; j < len(trainingData[0]); j++ {
			trainingData[i][j] = (trainingData[i][j] - minVector[j]) / (maxVector[j] - minVector[j])
		}
	}
}

func Fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func main() {
	const multiplier = 1
	const width = 400 * multiplier
	const height = 400 * multiplier
	const nrows = 20 * multiplier * 2
	const ncols = 20 * multiplier * 2

	var som SOM

	window, err := imgo.NewWindow("Self-organizing map", 0, 0, width, height, 0)
	if err != nil {
		Fatalf("Failed to create new window: %s\n", err.Error())
	}
	defer window.Close()

	ui := &window.UI

	trainingData := GenerateTrainingData([][]float32{
		{53.2521, 34.3717}, /* Bryansk. */
		{52.9651, 36.0785}, /* Orel. */
		{54.7818, 32.0401}, /* Smolensk. */
		{54.5293, 36.2754}, /* Kaluga. */
		{54.1961, 37.6182}, /* Tula. */
	}, 0.15, 50)
	NormalizeTrainingData(trainingData)

	som.Init(width, height, nrows, ncols, len(trainingData[0]), 5000, 0.1)

	started := false
	running := false
	drawCount := true
	drawPoints := false

	quit := false
	for !quit {
		for window.HasEvents() {
			event := window.GetEvent()
			switch event := event.(type) {
			case imgo.DestroyEvent:
				quit = true
			case imgo.MouseButtonDownEvent:
				if started {
					switch event.Button {
					case imgo.Button1:
						running = !running
					case imgo.Button2:
						if !running {
							som.TrainStep(trainingData)
						} else {
							drawCount = !drawCount
						}
					case imgo.Button3:
						drawPoints = !drawPoints
					}
				}
				started = true
				window.MouseButtonDownEvent(event.X, event.Y, event.Button)
			default:
				window.HandleEvent(event)
			}
		}

		ui.Begin()
		ui.Renderer.GraphSolidRectWH(0, 0, window.Width(), window.Height(), gr.ColorBlack)
		if (running) && (!som.Trained) {
			som.TrainStep(trainingData)
		}
		som.Render(&ui.Renderer)
		if drawPoints {
			ui.Renderer.GraphSolidRectWH(0, 0, window.Width(), window.Height(), gr.ColorRGBA(0, 0, 0, 50))
			for _, point := range trainingData {
				x := int(point[0] * width)
				y := int(point[1] * height)
				ui.Renderer.GraphPoint(x, y, 2, gr.ColorRGB(255, 0, 0))
			}
		}
		if drawCount {
			countString := strconv.Itoa(som.Count)
			ui.Renderer.GraphSolidRectWH(0, 0, ui.Font.TextWidth(countString), ui.Font.CharHeight('0'), gr.ColorRGBA(0, 0, 0, 150))
			ui.Renderer.GraphText(countString, ui.Font, 0, 0, gr.ColorWhite)
		}
		if !started {
			text := "Press any mouse button to start."
			ui.Renderer.GraphSolidRectWH(0, 0, window.Width(), window.Height(), gr.ColorRGBA(0, 0, 0, 200))
			ui.Renderer.GraphText(text, ui.Font, window.Width()/2-ui.Font.TextWidth(text)/2, window.Height()/2-ui.Font.CharHeight('g')/2, gr.ColorWhite)
		}
		ui.End()

		const FPS = 60
		window.PaintEventCapped(FPS)
	}
}
