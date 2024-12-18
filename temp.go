//go:build exclude

package main

import (
	"bytes"
	"image"
	"image/png"
	"log"
	"math/rand"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	chart "github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

// ChartData holds the time series data
type ChartPoint struct {
	XValues time.Time
	YValues float64
	Status  bool
}

var chartData []ChartPoint

func initializeChartData() {
	for i := 0; i < 20; i++ {
		chartData = append(chartData, ChartPoint{
			XValues: time.Now().Add(time.Duration(-20+i) * time.Second),
			YValues: rand.Float64() * 10,
		})
	}
}

// Generate the dynamic chart or a placeholder if needed
func createDynamicChart(legendName string) image.Image {

	//xValues := make([]time.Time, len(chartData))
	xValues := make([]float64, len(chartData))
	yValues := make([]float64, len(chartData))

	// Threshold for annotations
	threshold := 7.0
	annotations := []chart.Value2{}

	for i, point := range chartData {
		//xValues[i] = point.XValues
		xValues[i] = float64(point.XValues.Unix())
		yValues[i] = point.YValues

		// Check if the value exceeds the threshold
		if point.YValues >= threshold {
			annotations = append(annotations, chart.Value2{
				XValue: float64(point.XValues.Unix()),
				YValue: point.YValues,
				Label:  "Alert", // Annotation text
			})
		}
	}

	graph := chart.Chart{
		Title:  "Dynamic Bar Chart",
		Height: 512,
		Background: chart.Style{
			Padding: chart.Box{Top: 100, Left: 20, Right: 50, Bottom: 20},
		},
		TitleStyle: chart.Style{
			FontSize:  18,
			Padding:   chart.Box{Top: 5, Bottom: 20}, // Space below the title
			FontColor: drawing.ColorBlack,
			// Font:                chart.DefaultFont,
			TextHorizontalAlign: chart.TextHorizontalAlignCenter,
			TextVerticalAlign:   chart.TextVerticalAlignTop,
		},
		XAxis: chart.XAxis{
			Name: "Time (HH:MM:SS)",
			ValueFormatter: func(v interface{}) string {
				if typed, ok := v.(float64); ok {
					t := time.Unix(int64(typed), 0) // Convert float64 back to time.Time
					return t.Format("15:04:05")     // Format as HH:MM:SS
				}
				return ""
			},
		},
		YAxis: chart.YAxis{
			Name: "Value",
			GridMinorStyle: chart.Style{
				Hidden:      false,
				StrokeColor: drawing.Color{R: 0, G: 0, B: 0, A: 100},
				StrokeWidth: 1.0,
			},
		},
		Series: []chart.Series{
			chart.ContinuousSeries{
				Name:    legendName, // Use input variable as legend name
				XValues: xValues,
				YValues: yValues,
				Style: chart.Style{
					Hidden:      false,
					StrokeColor: chart.ColorBlue,
					StrokeWidth: 2.0,
				},
			},
			chart.AnnotationSeries{
				Annotations: annotations,
				Style: chart.Style{
					Hidden:              false,
					FontColor:           chart.ColorRed,
					FontSize:            12,
					TextHorizontalAlign: chart.TextHorizontalAlignLeft,
					TextVerticalAlign:   chart.TextVerticalAlignTop,
				},
			},
		},
	}

	// Add the legend on the left side
	graph.Elements = []chart.Renderable{
		chart.LegendLeft(&graph),
	}

	// Render the chart
	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)
	if err != nil {
		log.Fatalf("Failed to render chart: %v", err)
	}

	img, err := png.Decode(buffer)
	if err != nil {
		log.Fatalf("Failed to decode chart image: %v", err)
	}
	return img
}

// Simulate data updates periodically
func simulateDataUpdates(chartDataChan chan<- []ChartPoint) {
	for {
		time.Sleep(1 * time.Second) // Simulate data updates every second
		chartData = append(chartData, ChartPoint{
			XValues: time.Now(),
			YValues: rand.Float64() * 10,
		})
		if len(chartData) > 20 {
			chartData = chartData[1:]
		}
	}
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Dynamic Chart Example")

	// Initialize chart data
	initializeChartData()

	// Initial chartDataChan
	chartDataChan := make(chan []ChartPoint)
	defer close(chartDataChan)

	// Create an initial in-memory chart
	chartImage := canvas.NewImageFromImage(createDynamicChart("ping 1.2.3.4"))
	chartImage.FillMode = canvas.ImageFillOriginal

	// Set up the UI layout
	content := container.NewCenter(chartImage)
	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(800, 600))

	// Start Goroutines for updates
	go func() {
		for {
			time.Sleep(1 * time.Second) // Update chart every second

			// Re-generate the chart and refresh the UI
			chartImage.Image = createDynamicChart("ping 1.2.3.4")
			canvas.Refresh(chartImage)
		}
	}()

	// Simulate data updates in a separate Goroutine
	go simulateDataUpdates()

	// Launch the application
	myWindow.ShowAndRun()
}
