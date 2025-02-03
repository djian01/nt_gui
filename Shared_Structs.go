package main

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	ntchart "github.com/djian01/nt_gui/pkg/chart"
)

// ********* Summary ***************

type Summary struct {
	Type            string
	DestHost        string
	StartTime       time.Time
	EndTime         time.Time
	PacketSent      int
	SuccessResponse int
	FailRate        string
	MinRTT          time.Duration
	MaxRTT          time.Duration
	AvgRtt          time.Duration
	ntCmd           string
}

type SummaryUI struct {
	typeLabel *widget.Label
	typeEntry *widget.Entry

	destHostLabel *widget.Label
	destHostEntry *widget.Entry

	startTimeLabel *widget.Label
	startTimeEntry *widget.Entry

	endTimeLabel *widget.Label
	endTimeEntry *widget.Entry

	packetSentLabel *widget.Label
	packetSentEntry *widget.Entry

	successRespLabel     *widget.Label
	successResponseEntry *widget.Entry

	failRateLabel *widget.Label
	failRateEntry *widget.Entry

	minRttLabel *widget.Label
	minRttEntry *widget.Entry

	maxRttLabel *widget.Label
	maxRttEntry *widget.Entry

	avgRttLabel *widget.Label
	avgRttEntry *widget.Entry

	ntCmdLabel *widget.Label
	ntCmdEntry *widget.Entry
	ntCmdBtn   *widget.Button
}

func (s *SummaryUI) Initial() {
	s.typeLabel = widget.NewLabel("Type              ")
	s.typeEntry = widget.NewEntry()

	s.destHostLabel = widget.NewLabel("Dest Host/IP")
	s.destHostEntry = widget.NewEntry()

	s.startTimeLabel = widget.NewLabel("Start Time    ")
	s.startTimeEntry = widget.NewEntry()

	s.endTimeLabel = widget.NewLabel("End Time      ")
	s.endTimeEntry = widget.NewEntry()

	s.packetSentLabel = widget.NewLabel("Packets Sent")
	s.packetSentEntry = widget.NewEntry()

	s.successRespLabel = widget.NewLabel("Success Probs")
	s.successResponseEntry = widget.NewEntry()

	s.failRateLabel = widget.NewLabel("Fail Rate    ")
	s.failRateEntry = widget.NewEntry()

	s.minRttLabel = widget.NewLabel("Min RTT        ")
	s.minRttEntry = widget.NewEntry()

	s.maxRttLabel = widget.NewLabel("Max RTT          ")
	s.maxRttEntry = widget.NewEntry()

	s.avgRttLabel = widget.NewLabel("Avg RTT     ")
	s.avgRttEntry = widget.NewEntry()

	s.ntCmdLabel = widget.NewLabel("nt CMD         ")
	s.ntCmdEntry = widget.NewEntry()
	s.ntCmdBtn = widget.NewButton("Relaunch CMD", func() {})
	s.ntCmdBtn.Importance = widget.HighImportance
}

func (s *SummaryUI) CreateCard() *widget.Card {

	// cell containers
	typeContainer := container.New(layout.NewBorderLayout(nil, nil, s.typeLabel, nil), s.typeLabel, s.typeEntry)
	destHostContainer := container.New(layout.NewBorderLayout(nil, nil, s.destHostLabel, nil), s.destHostLabel, s.destHostEntry)
	startTimeContainer := container.New(layout.NewBorderLayout(nil, nil, s.startTimeLabel, nil), s.startTimeLabel, s.startTimeEntry)
	endTimeContainer := container.New(layout.NewBorderLayout(nil, nil, s.endTimeLabel, nil), s.endTimeLabel, s.endTimeEntry)
	packetSentContainer := container.New(layout.NewBorderLayout(nil, nil, s.packetSentLabel, nil), s.packetSentLabel, s.packetSentEntry)
	successRespContainer := container.New(layout.NewBorderLayout(nil, nil, s.successRespLabel, nil), s.successRespLabel, s.successResponseEntry)
	failRateContainer := container.New(layout.NewBorderLayout(nil, nil, s.failRateLabel, nil), s.failRateLabel, s.failRateEntry)
	minRttContainer := container.New(layout.NewBorderLayout(nil, nil, s.minRttLabel, nil), s.minRttLabel, s.minRttEntry)
	maxRttContainer := container.New(layout.NewBorderLayout(nil, nil, s.maxRttLabel, nil), s.maxRttLabel, s.maxRttEntry)
	avgRttContainer := container.New(layout.NewBorderLayout(nil, nil, s.avgRttLabel, nil), s.avgRttLabel, s.avgRttEntry)
	ntCmdContainer := container.New(layout.NewBorderLayout(nil, nil, s.ntCmdLabel, nil), s.ntCmdLabel, s.ntCmdEntry)

	// rows
	summaryRow1 := container.New(layout.NewGridLayoutWithColumns(2), typeContainer, destHostContainer)
	summaryRow2 := container.New(layout.NewGridLayoutWithColumns(2), startTimeContainer, endTimeContainer)
	summaryRow3 := container.New(layout.NewGridLayoutWithColumns(3), packetSentContainer, successRespContainer, failRateContainer)
	summaryRow4 := container.New(layout.NewGridLayoutWithColumns(3), minRttContainer, maxRttContainer, avgRttContainer)
	summaryRow5 := container.New(layout.NewBorderLayout(nil, nil, nil, s.ntCmdBtn), s.ntCmdBtn, ntCmdContainer)

	// overall container and card
	summaryContainer := container.New(layout.NewGridLayoutWithRows(5), summaryRow1, summaryRow2, summaryRow3, summaryRow4, summaryRow5)
	summaryCard := widget.NewCard("", "", summaryContainer)

	return summaryCard
}

// ********* Chart Update Slider ***************

type Slider struct {
	chartData       *[]ntchart.ChartPoint
	sliderChartData []ntchart.ChartPoint

	sliderLeft         *widget.Slider
	sliderLeftIndicate *widget.Label
	sliderLeftValue    *widget.Label

	sliderRight         *widget.Slider
	sliderRightIndicate *widget.Label
	sliderRightValue    *widget.Label

	chartUpdateBtn *widget.Button
	chartResetBtn  *widget.Button

	ErrLabel *canvas.Text

	sliderCard *widget.Card
}

func (s *Slider) Initial() {
	s.sliderLeft = widget.NewSlider(0, 100)
	s.sliderLeftIndicate = widget.NewLabel("From: ")
	s.sliderLeftValue = widget.NewLabel("")

	s.sliderRight = widget.NewSlider(0, 100)
	s.sliderRightIndicate = widget.NewLabel("To: ")
	s.sliderRightValue = widget.NewLabel("")

	s.chartUpdateBtn = widget.NewButton("Update Chart", func() {})
	s.chartUpdateBtn.Importance = widget.HighImportance

	s.chartResetBtn = widget.NewButton("Reset Chart", func() {})
	s.chartResetBtn.Importance = widget.WarningImportance

	s.ErrLabel = canvas.NewText("No Error:", color.RGBA{255, 0, 0, 255})
	s.ErrLabel.Hidden = true
}

func (s *Slider) initialSetMax(Max float64) {
	s.sliderLeft.Min = 0
	s.sliderLeft.Max = Max - 2
	s.sliderRight.Min = 0 + 2
	s.sliderRight.Max = Max
	s.sliderLeft.SetValue(0)
	s.sliderRight.SetValue(Max)
	s.sliderUpdate()

}

func (s *Slider) CreateCard() {
	sliderLeftContainerIn := container.New(layout.NewGridLayoutWithColumns(2), s.sliderLeft, s.sliderLeftValue)
	sliderLeftContainerOut := container.New(layout.NewBorderLayout(nil, nil, s.sliderLeftIndicate, nil), s.sliderLeftIndicate, sliderLeftContainerIn)

	sliderRightContainerIn := container.New(layout.NewGridLayoutWithColumns(2), s.sliderRight, s.sliderRightValue)
	sliderRightContainerOut := container.New(layout.NewBorderLayout(nil, nil, s.sliderRightIndicate, nil), s.sliderRightIndicate, sliderRightContainerIn)

	sliderContainerIn := container.New(layout.NewGridLayoutWithColumns(2), sliderLeftContainerOut, sliderRightContainerOut)

	btnContainer := container.New(layout.NewGridLayoutWithColumns(2), s.chartUpdateBtn, s.chartResetBtn)

	sliderContainerOut := container.New(layout.NewBorderLayout(nil, nil, nil, btnContainer), btnContainer, sliderContainerIn)

	sliderContainerMain := container.New(layout.NewBorderLayout(nil, s.ErrLabel, nil, nil), sliderContainerOut, s.ErrLabel)

	s.sliderCard = widget.NewCard("", "", sliderContainerMain)
}

func (s *Slider) sliderUpdate() {
	s.sliderLeftValue.Text = (*s.chartData)[int(s.sliderLeft.Value)].XValues.Format("2006-01-02 15:04:05 MST")
	s.sliderRightValue.Text = (*s.chartData)[int(s.sliderRight.Value)].XValues.Format("2006-01-02 15:04:05 MST")
	s.sliderRight.Min = s.sliderLeft.Value + 2
	s.sliderLeft.Max = s.sliderRight.Value - 2
}

func (s *Slider) BuildSliderChartData() {
	s.sliderChartData = []ntchart.ChartPoint{}
	for i, data := range *s.chartData {
		if i >= int(s.sliderLeft.Value) {
			s.sliderChartData = append(s.sliderChartData, data)
		}
		if i > int(s.sliderRight.Value) {
			break
		}
	}
}

func (s *Slider) UpdateChartImage(RaType string, chartImage *canvas.Image) {
	s.BuildSliderChartData()
	image := ntchart.CreateChart(RaType, &(s.sliderChartData))
	chartImage.Image = image
}

func (s *Slider) ResetChartImage(RaType string, chartImage *canvas.Image) {
	s.initialSetMax(float64(len(*(s.chartData)) - 1))
	image := ntchart.CreateChart(RaType, s.chartData)
	chartImage.Image = image
}
