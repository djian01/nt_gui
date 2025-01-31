package main

import (
	"time"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

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
	summaryRow5 := container.New(layout.NewGridLayoutWithColumns(1), ntCmdContainer)

	// overall container and card
	summaryContainer := container.New(layout.NewGridLayoutWithRows(5), summaryRow1, summaryRow2, summaryRow3, summaryRow4, summaryRow5)
	summaryCard := widget.NewCard("", "", summaryContainer)

	return summaryCard
}
