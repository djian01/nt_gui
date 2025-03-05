package ntchart

import (
	"bytes"
	"image"
	"image/png"
	"log"
	"time"

	"github.com/djian01/nt/pkg/ntPinger"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

// ChartData holds the time series data
type ChartPoint struct {
	XValues time.Time
	YValues float64
	Status  bool
}

// func: convert Packet to ChartPoint
func ConvertFromPacketToChartPoint(pkt ntPinger.Packet) ChartPoint {

	cp := ChartPoint{}

	cp.XValues = pkt.GetSendTime()
	cp.YValues = (float64((pkt.GetRtt()).Nanoseconds())) / 1e6
	cp.Status = pkt.GetStatus()

	return cp
}

// Generate the dynamic chart or a placeholder if needed
func CreateChart(legendName string, chartData *[]ChartPoint) image.Image {

	//xValues := make([]time.Time, len(chartData))
	xValues := make([]float64, len(*chartData))
	yValues := make([]float64, len(*chartData))

	// Threshold for annotations
	annotations := []chart.Value2{}

	for i, point := range *chartData {
		//xValues[i] = point.XValues
		xValues[i] = float64(point.XValues.Unix())
		yValues[i] = point.YValues

		// Check if the status is "false"
		if !point.Status {
			yValues[i] = 0
			annotations = append(annotations, chart.Value2{
				XValue: float64(point.XValues.Unix()),
				//YValue: point.YValues,
				YValue: 0,
				Label:  "F", // Annotation text
			})
		}
	}

	graph := chart.Chart{
		Title:  "NT Test Results",
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
			Name: "Milliseconds",
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
