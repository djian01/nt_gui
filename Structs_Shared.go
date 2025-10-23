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
	"github.com/djian01/nt/pkg/ntPinger"
	ntchart "github.com/djian01/nt_gui/pkg/chart"
	"github.com/djian01/nt_gui/pkg/ntdb"
	"github.com/djian01/nt_gui/pkg/ntwidget"
)

// Interface testGUIRow
type testGUIRow interface {
	Initial()
	GenerateHeaderRow() *fyne.Container
	UpdateRow(p *ntPinger.Packet)
}

// Interace testObject
type testObject interface {
	Initial(*SummaryData)
	UpdateChartData(pkt *ntPinger.Packet)
	DisplayChartDataTerminal()
	Stop(p *ntPinger.Pinger)
	GetType() string
	GetSummary() *SummaryData
	GetChartData() *[]ntchart.ChartPoint
	GetUUID() string
	UpdateRecording(recording bool)
}

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

// ********* Summary Data ***************

type SummaryData struct {
	Type            string
	DestHost        string
	StartTime       time.Time
	EndTime         time.Time
	PacketSent      int
	SuccessResponse int
	FailRate        string
	MinRTT          time.Duration
	MaxRTT          time.Duration
	AvgRTT          time.Duration
	ntCmd           string
	uuid            string
	AddInfo         string
	//testEnded       bool
}

// Update Summary Data - Initial
func (sd *SummaryData) Initial(pType, destHost, ntCmd string, startTime time.Time, uuid string) {
	sd.Type = pType
	sd.ntCmd = ntCmd
	sd.StartTime = startTime
	sd.DestHost = destHost
	sd.uuid = uuid
	//sd.testEnded = false
}

// Update Summary Data - Running
func (sd *SummaryData) UpdateRunning(p *ntPinger.Packet) {
	switch sd.Type {
	case "dns":
		myPacket := (*p).(*ntPinger.PacketDNS)
		sd.PacketSent = myPacket.PacketsSent
		sd.SuccessResponse = myPacket.PacketsRecv
		sd.FailRate = fmt.Sprintf("%.2f%%", float64(myPacket.PacketLoss*100))
		sd.MinRTT = myPacket.MinRtt
		sd.MaxRTT = myPacket.MaxRtt
		sd.AvgRTT = myPacket.AvgRtt
		sd.AddInfo = myPacket.AdditionalInfo
	case "http":
		myPacket := (*p).(*ntPinger.PacketHTTP)
		sd.PacketSent = myPacket.PacketsSent
		sd.SuccessResponse = myPacket.PacketsRecv
		sd.FailRate = fmt.Sprintf("%.2f%%", float64(myPacket.PacketLoss*100))
		sd.MinRTT = myPacket.MinRtt
		sd.MaxRTT = myPacket.MaxRtt
		sd.AvgRTT = myPacket.AvgRtt
		sd.AddInfo = myPacket.AdditionalInfo
	case "tcp":
		myPacket := (*p).(*ntPinger.PacketTCP)
		sd.PacketSent = myPacket.PacketsSent
		sd.SuccessResponse = myPacket.PacketsRecv
		sd.FailRate = fmt.Sprintf("%.2f%%", float64(myPacket.PacketLoss*100))
		sd.MinRTT = myPacket.MinRtt
		sd.MaxRTT = myPacket.MaxRtt
		sd.AvgRTT = myPacket.AvgRtt
		sd.AddInfo = myPacket.AdditionalInfo
	case "icmp":
		myPacket := (*p).(*ntPinger.PacketICMP)
		sd.PacketSent = myPacket.PacketsSent
		sd.SuccessResponse = myPacket.PacketsRecv
		sd.FailRate = fmt.Sprintf("%.2f%%", float64(myPacket.PacketLoss*100))
		sd.MinRTT = myPacket.MinRtt
		sd.MaxRTT = myPacket.MaxRtt
		sd.AvgRTT = myPacket.AvgRtt
		sd.AddInfo = myPacket.AdditionalInfo
	}
}

// func: get UUID
func (sd *SummaryData) GetUUID() string {
	return sd.uuid
}

// func: DB Test Entry -> SummaryData
func DbTestEntry2SummaryData(historyEntry ntdb.HistoryEntry, firstEntry, lastEntry ntdb.DbTestEntry) (SummaryData, error) {

	// inital SummaryData
	sumData := SummaryData{}

	// create Input Var
	_, iv, err := NtCmd2Iv(historyEntry.Command)
	if err != nil {
		return sumData, err
	}

	// construct summary data
	sumData.Type = iv.Type
	sumData.DestHost = iv.DestHost
	sumData.StartTime = firstEntry.GetSendTime()
	sumData.EndTime = lastEntry.GetSendTime()
	sumData.PacketSent = lastEntry.GetPacketSent()
	sumData.SuccessResponse = lastEntry.GetSuccessResponse()
	sumData.FailRate = lastEntry.GetFailRate()
	sumData.MinRTT, _ = time.ParseDuration(lastEntry.GetMinRtt())
	sumData.MaxRTT, _ = time.ParseDuration(lastEntry.GetMaxRtt())
	sumData.AvgRTT, _ = time.ParseDuration(lastEntry.GetAvgRtt())
	sumData.ntCmd = historyEntry.Command
	sumData.uuid = historyEntry.UUID

	return sumData, err
}

// *********** Summary UI **********
type SummaryUI struct {
	typeLabel *widget.Label
	typeEntry *widget.Entry

	destHostLabel *widget.Label
	destHostEntry *widget.Entry

	startTimeLabel *widget.Label
	startTimeEntry *widget.Entry

	endTimeLabel *widget.Label
	endTimeEntry *widget.Entry // if the test is still on-going, Endtime is "--"

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

	addInfoLabel *widget.Label
	addInfoEntry *widget.Entry

	ntCmdLabel *widget.Label
	ntCmdEntry *widget.Entry
	ntCmdBtn   *widget.Button

	summaryCard *widget.Card
}

func (sui *SummaryUI) Initial() {
	sui.typeLabel = widget.NewLabel("Type              ")
	sui.typeEntry = widget.NewEntry()

	sui.destHostLabel = widget.NewLabel("Dest Host/IP")
	sui.destHostEntry = widget.NewEntry()

	sui.startTimeLabel = widget.NewLabel("Start Time    ")
	sui.startTimeEntry = widget.NewEntry()

	sui.endTimeLabel = widget.NewLabel("End Time         ")
	sui.endTimeEntry = widget.NewEntry()

	sui.packetSentLabel = widget.NewLabel("Packets Sent")
	sui.packetSentEntry = widget.NewEntry()

	sui.successRespLabel = widget.NewLabel("Success Probs")
	sui.successResponseEntry = widget.NewEntry()

	sui.failRateLabel = widget.NewLabel("Fail Rate    ")
	sui.failRateEntry = widget.NewEntry()

	sui.minRttLabel = widget.NewLabel("Min RTT        ")
	sui.minRttEntry = widget.NewEntry()

	sui.maxRttLabel = widget.NewLabel("Max RTT          ")
	sui.maxRttEntry = widget.NewEntry()

	sui.avgRttLabel = widget.NewLabel("Avg RTT     ")
	sui.avgRttEntry = widget.NewEntry()

	sui.addInfoLabel = widget.NewLabel("Info            ")
	sui.addInfoEntry = widget.NewEntry()

	sui.ntCmdLabel = widget.NewLabel("nt CMD         ")
	sui.ntCmdEntry = widget.NewEntry()
	sui.ntCmdBtn = widget.NewButton("Relaunch Test", func() {})
	sui.ntCmdBtn.Disable()
	sui.ntCmdBtn.Importance = widget.HighImportance
}

func (sui *SummaryUI) CreateCard() {

	// cell containers
	typeContainer := container.New(layout.NewBorderLayout(nil, nil, sui.typeLabel, nil), sui.typeLabel, sui.typeEntry)
	destHostContainer := container.New(layout.NewBorderLayout(nil, nil, sui.destHostLabel, nil), sui.destHostLabel, sui.destHostEntry)
	startTimeContainer := container.New(layout.NewBorderLayout(nil, nil, sui.startTimeLabel, nil), sui.startTimeLabel, sui.startTimeEntry)
	endTimeContainer := container.New(layout.NewBorderLayout(nil, nil, sui.endTimeLabel, nil), sui.endTimeLabel, sui.endTimeEntry)
	packetSentContainer := container.New(layout.NewBorderLayout(nil, nil, sui.packetSentLabel, nil), sui.packetSentLabel, sui.packetSentEntry)
	successRespContainer := container.New(layout.NewBorderLayout(nil, nil, sui.successRespLabel, nil), sui.successRespLabel, sui.successResponseEntry)
	failRateContainer := container.New(layout.NewBorderLayout(nil, nil, sui.failRateLabel, nil), sui.failRateLabel, sui.failRateEntry)
	minRttContainer := container.New(layout.NewBorderLayout(nil, nil, sui.minRttLabel, nil), sui.minRttLabel, sui.minRttEntry)
	maxRttContainer := container.New(layout.NewBorderLayout(nil, nil, sui.maxRttLabel, nil), sui.maxRttLabel, sui.maxRttEntry)
	avgRttContainer := container.New(layout.NewBorderLayout(nil, nil, sui.avgRttLabel, nil), sui.avgRttLabel, sui.avgRttEntry)
	ntCmdContainer := container.New(layout.NewBorderLayout(nil, nil, sui.ntCmdLabel, nil), sui.ntCmdLabel, sui.ntCmdEntry)
	addInfoContainer := container.New(layout.NewBorderLayout(nil, nil, sui.addInfoLabel, nil), sui.addInfoLabel, sui.addInfoEntry)

	// rows
	summaryRow1 := container.New(layout.NewGridLayoutWithColumns(2), typeContainer, destHostContainer)
	summaryRow2 := container.New(layout.NewGridLayoutWithColumns(3), startTimeContainer, endTimeContainer, addInfoContainer)
	summaryRow3 := container.New(layout.NewGridLayoutWithColumns(3), packetSentContainer, successRespContainer, failRateContainer)
	summaryRow4 := container.New(layout.NewGridLayoutWithColumns(3), minRttContainer, maxRttContainer, avgRttContainer)
	summaryRow5 := container.New(layout.NewBorderLayout(nil, nil, nil, sui.ntCmdBtn), sui.ntCmdBtn, ntCmdContainer)

	// overall container and card
	summaryContainer := container.New(layout.NewGridLayoutWithRows(5), summaryRow1, summaryRow2, summaryRow3, summaryRow4, summaryRow5)
	sui.summaryCard = widget.NewCard("", "", summaryContainer)
}

// Update Summary UI with all fields
func (sui *SummaryUI) UpdateStaticUI(sd *SummaryData) {

	// update summary UI
	sui.typeEntry.SetText((*sd).Type)
	sui.destHostEntry.SetText((*sd).DestHost)
	sui.startTimeEntry.SetText((*sd).StartTime.Format(("2006-01-02 15:04:05 MST")))
	sui.endTimeEntry.SetText((*sd).EndTime.Format(("2006-01-02 15:04:05 MST")))
	sui.packetSentEntry.SetText(strconv.Itoa((*sd).PacketSent))
	sui.successResponseEntry.SetText(strconv.Itoa((*sd).SuccessResponse))
	sui.failRateEntry.SetText((*sd).FailRate)
	sui.minRttEntry.SetText(fmt.Sprintf("%d ms", (*sd).MinRTT.Milliseconds()))
	sui.maxRttEntry.SetText(fmt.Sprintf("%d ms", (*sd).MaxRTT.Milliseconds()))
	sui.avgRttEntry.SetText(fmt.Sprintf("%d ms", (*sd).AvgRTT.Milliseconds()))
	sui.ntCmdEntry.SetText((*sd).ntCmd)

	// ntCmdBtn
	if !existingTestCheck(&testRegister, sd.GetUUID()) {
		sui.ntCmdBtn.Enable()
	} else {
		sui.ntCmdBtn.Disable()
	}
}

// Update Summary UI Initial - when the test is running.
func (sui *SummaryUI) UpdateUI_Initial(sd *SummaryData) {
	// update summary UI
	sui.typeEntry.SetText((*sd).Type)
	sui.destHostEntry.SetText((*sd).DestHost)
	sui.startTimeEntry.SetText((*sd).StartTime.Format(("2006-01-02 15:04:05 MST")))
	// s.UI.endTimeEntry.SetText("--")
	// s.UI.packetSentEntry.SetText("--")
	// s.UI.successResponseEntry.SetText("--")
	// s.UI.failRateEntry.SetText("--")
	// s.UI.minRttEntry.SetText("--")
	// s.UI.maxRttEntry.SetText("--")
	// s.UI.avgRttEntry.SetText("--")
	sui.ntCmdEntry.SetText((*sd).ntCmd)
}

// Update Summary UI Initial - when the test is running.
func (sui *SummaryUI) UpdateUI_Running(sd *SummaryData) {
	sui.packetSentEntry.SetText(strconv.Itoa((*sd).PacketSent))
	sui.successResponseEntry.SetText(strconv.Itoa((*sd).SuccessResponse))
	sui.failRateEntry.SetText((*sd).FailRate)
	sui.minRttEntry.SetText(fmt.Sprintf("%d ms", (*sd).MinRTT.Milliseconds()))
	sui.maxRttEntry.SetText(fmt.Sprintf("%d ms", (*sd).MaxRTT.Milliseconds()))
	sui.avgRttEntry.SetText(fmt.Sprintf("%d ms", (*sd).AvgRTT.Milliseconds()))
	sui.addInfoEntry.SetText((*sd).AddInfo)
}

// Update Summary UI Initial - when the test is ended.
func (sui *SummaryUI) UpdateUI_Ended(sd *SummaryData) {
	sui.endTimeEntry.SetText((*sd).EndTime.Format(("2006-01-02 15:04:05 MST")))
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

	historyTable *fyne.Container
}

// ************ Http Vars **************
type HttpVars struct {
	Scheme   string
	Hostname string
	Port     int
	Path     string
}
