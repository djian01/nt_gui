package main

import (
	"fmt"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	ntchart "github.com/djian01/nt_gui/pkg/chart"
	"github.com/djian01/nt_gui/pkg/ntwidget"
)

// ********* Chart ***************
type Chart struct {
	chartImage     *canvas.Image
	chartContainer *fyne.Container
	chartCard      *widget.Card
}

func (c *Chart) Initial() {
	// chartImage
	c.chartImage = canvas.NewImageFromResource(nil)
	c.chartImage.FillMode = canvas.ImageFillContain
	c.chartImage.SetMinSize(fyne.NewSize(400, 400))

	// chartCard
	c.chartContainer = container.New(layout.NewBorderLayout(nil, nil, nil, nil), c.chartImage)
	c.chartCard = widget.NewCard("", "", c.chartContainer)
}

func (c *Chart) ChartUpdate(RaType string, chartData *[]ntchart.ChartPoint) {
	c.chartImage.Image = ntchart.CreateChart(RaType, chartData)
	c.chartImage.FillMode = canvas.ImageFillStretch
	c.chartImage.Refresh()
	c.chartContainer.Refresh()
}

// ********* Summary ***************

type Summary struct {
	Type            string
	DestHost        string
	StartTime       time.Time
	EndTime         time.Time // if the test is still on-going, Endtime is "--"
	PacketSent      int
	SuccessResponse int
	FailRate        string
	MinRTT          time.Duration
	MaxRTT          time.Duration
	AvgRtt          time.Duration
	ntCmd           string
	UI              SummaryUI
}

func (s *Summary) UpdateUI() {

	// update summary UI
	s.UI.typeEntry.SetText((*s).Type)
	s.UI.destHostEntry.SetText((*s).DestHost)

	s.UI.startTimeEntry.SetText((*s).StartTime.Format(("2006-01-02 15:04:05 MST")))
	s.UI.endTimeEntry.SetText((*s).EndTime.Format(("2006-01-02 15:04:05 MST")))
	s.UI.packetSentEntry.SetText(strconv.Itoa((*s).PacketSent))
	s.UI.successResponseEntry.SetText(strconv.Itoa((*s).SuccessResponse))
	s.UI.failRateEntry.SetText((*s).FailRate)
	s.UI.minRttEntry.SetText(fmt.Sprintf("%d ms", (*s).MinRTT.Milliseconds()))
	s.UI.maxRttEntry.SetText(fmt.Sprintf("%d ms", (*s).MaxRTT.Milliseconds()))
	s.UI.avgRttEntry.SetText(fmt.Sprintf("%d ms", (*s).AvgRtt.Milliseconds()))
	s.UI.ntCmdEntry.SetText((*s).ntCmd)
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

	summaryCard *widget.Card
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

func (s *SummaryUI) CreateCard() {

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
	s.summaryCard = widget.NewCard("", "", summaryContainer)
}

// ********* Chart Update Slider ***************

type Slider struct {
	chartData       *[]ntchart.ChartPoint
	sliderChartData []ntchart.ChartPoint

	rangeSlider *ntwidget.RangeSlider

	startIndicate *widget.Label
	startValue    *widget.Label

	endIndicate *widget.Label
	endValue    *widget.Label

	chartUpdateBtn *widget.Button
	chartResetBtn  *widget.Button

	sliderCard *widget.Card
}

func (s *Slider) Initial(min, max, start, end float64) {
	s.rangeSlider = ntwidget.NewRangeSlider(min, max, start, end)

	s.startIndicate = widget.NewLabel("From: ")
	s.startIndicate.TextStyle = fyne.TextStyle{Bold: true}
	s.startValue = widget.NewLabel("")

	s.endIndicate = widget.NewLabel("To: ")
	s.endIndicate.TextStyle = fyne.TextStyle{Bold: true}
	s.endValue = widget.NewLabel("")

	s.chartUpdateBtn = widget.NewButton("Update Chart", func() {})
	s.chartUpdateBtn.Importance = widget.HighImportance

	s.chartResetBtn = widget.NewButton("Reset Chart", func() {})
	s.chartResetBtn.Importance = widget.WarningImportance
}

func (s *Slider) CreateCard() {
	spaceHolder := widget.NewLabel("              ")
	RangeSliderContainer := container.New(layout.NewBorderLayout(nil, nil, spaceHolder, spaceHolder), spaceHolder, s.rangeSlider)

	startContainer := formCell(s.startIndicate, 50, s.startValue, 500)
	endContainer := formCell(s.endIndicate, 50, s.endValue, 500)
	startEndContainer := container.New(layout.NewHBoxLayout(), startContainer, endContainer)
	btnContainer := container.New(layout.NewGridLayoutWithColumns(2), s.chartUpdateBtn, s.chartResetBtn)
	sliderActionItems := container.New(layout.NewBorderLayout(nil, nil, nil, btnContainer), btnContainer, startEndContainer)
	sliderContainerMain := container.New(layout.NewVBoxLayout(), RangeSliderContainer, sliderActionItems)

	s.sliderCard = widget.NewCard("", "", sliderContainerMain)
}

func (s *Slider) update() {
	s.startValue.Text = (*s.chartData)[int(s.rangeSlider.Start)].XValues.Format("2006-01-02 15:04:05 MST")
	s.startValue.Refresh()
	s.endValue.Text = (*s.chartData)[int(s.rangeSlider.End)].XValues.Format("2006-01-02 15:04:05 MST")
	s.endValue.Refresh()
	s.rangeSlider.Layout(s.rangeSlider.Size())
}

func (s *Slider) BuildSliderChartData() {
	s.sliderChartData = []ntchart.ChartPoint{}
	for i, data := range *s.chartData {
		if i >= int(s.rangeSlider.Start) {
			s.sliderChartData = append(s.sliderChartData, data)
		}
		if i > int(s.rangeSlider.End) {
			break
		}
	}
}

func (s *Slider) UpdateChartImage(RaType string, chart *Chart) {
	s.BuildSliderChartData()
	(*chart).ChartUpdate(RaType, &(s.sliderChartData))
}

func (s *Slider) ResetChartImage(RaType string, chart *Chart) {
	s.rangeSlider.Start = s.rangeSlider.Min
	s.rangeSlider.End = s.rangeSlider.Max
	s.update()
	(*chart).ChartUpdate(RaType, s.chartData)
}

// ********* Ping Row Cell ***************
type pingCell struct {
	Label  string
	Length int
	Object fyne.CanvasObject
}

// ********* nt gui ping global ***************
type ntGUIGlboal struct {
	dnsTable *fyne.Container
	dnsIndex int

	httpTable *fyne.Container
	httpIndex int

	tcpTable *fyne.Container
	tcpIndex int

	icmpTable *fyne.Container
	icmpIndex int
}
