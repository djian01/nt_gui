package main

import (
	"database/sql"
	"fmt"
	"image/color"
	"net"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/djian01/nt/pkg/ntPinger"
	ntchart "github.com/djian01/nt_gui/pkg/chart"
	"github.com/djian01/nt_gui/pkg/ntdb"
)

// check interafce implementation
var _ testGUIRow = (*icmpGUIRow)(nil)
var _ testObject = (*icmpObject)(nil)

// ******* struct icmpGUIRow ********

type icmpGUIRow struct {
	Index     pingCell
	Seq       pingCell
	Status    pingCell
	HostName  pingCell // fixed
	IP        pingCell // fixed
	Payload   pingCell // fixed
	DF        pingCell //fixed
	RTT       pingCell
	StartTime pingCell // sendDateTime
	Fail      pingCell
	AvgRTT    pingCell
	Recording pingCell

	ChartBtn     *widget.Button
	StopBtn      *widget.Button
	ReplayBtn    *widget.Button
	CloseBtn     *widget.Button
	Action       pingCell
	icmpTableRow *fyne.Container
}

func (d *icmpGUIRow) Initial() {

	chartIcon := theme.NewThemedResource(resourceChartSvg)
	d.ChartBtn = widget.NewButtonWithIcon("", chartIcon, func() {})

	d.StopBtn = widget.NewButtonWithIcon("", theme.MediaStopIcon(), func() {})
	d.StopBtn.Importance = widget.DangerImportance

	d.ReplayBtn = widget.NewButtonWithIcon("", theme.MediaReplayIcon(), func() {})
	d.ReplayBtn.Importance = widget.HighImportance
	d.ReplayBtn.Disable()

	d.CloseBtn = widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {})
	d.CloseBtn.Importance = widget.WarningImportance
	d.CloseBtn.Disable()

	d.Action.Label = "Action"
	d.Action.Length = 110
	d.Action.Object = container.New(layout.NewGridLayoutWithColumns(4), d.ChartBtn, d.StopBtn, d.ReplayBtn, d.CloseBtn)

	d.Index.Label = "Index"
	d.Index.Length = 50
	d.Index.Object = widget.NewLabelWithStyle("--", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	d.Seq.Label = "Seq"
	d.Seq.Length = 50
	d.Seq.Object = widget.NewLabelWithStyle("--", fyne.TextAlignCenter, fyne.TextStyle{Bold: false})

	d.Status.Label = "Status"
	d.Status.Length = 65
	d.Status.Object = canvas.NewText("--", color.Black)

	d.HostName.Label = "HostName"
	d.HostName.Length = 180
	d.HostName.Object = widget.NewLabel("--")

	d.IP.Label = "IP"
	d.IP.Length = 180
	d.IP.Object = widget.NewLabel("--")

	d.Payload.Label = "Payload"
	d.Payload.Length = 90
	d.Payload.Object = widget.NewLabel("--")

	d.DF.Label = "DF"
	d.DF.Length = 50
	d.DF.Object = widget.NewLabel("--")

	d.RTT.Label = "RTT"
	d.RTT.Length = 90
	d.RTT.Object = widget.NewLabel("--")

	d.StartTime.Label = "StartTime"
	d.StartTime.Length = 190
	d.StartTime.Object = widget.NewLabel("--")

	d.Fail.Label = "Fail"
	d.Fail.Length = 80
	d.Fail.Object = widget.NewLabel("--")

	d.AvgRTT.Label = "AvgRTT"
	d.AvgRTT.Length = 110
	d.AvgRTT.Object = widget.NewLabel("--")

	d.Recording.Label = "Recording"
	d.Recording.Length = 80
	d.Recording.Object = widget.NewLabel("OFF")

	// table row
	row := container.New(layout.NewHBoxLayout(),
		container.NewGridWrap(fyne.NewSize(float32(d.Action.Length), 30), d.Action.Object),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Index.Length), 30), container.NewCenter(d.Index.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Seq.Length), 30), container.NewCenter(d.Seq.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Status.Length), 30), container.NewCenter(d.Status.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.HostName.Length), 30), container.NewCenter(d.HostName.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.IP.Length), 30), container.NewCenter(d.IP.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Payload.Length), 30), container.NewCenter(d.Payload.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.DF.Length), 30), container.NewCenter(d.DF.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.RTT.Length), 30), container.NewCenter(d.RTT.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.StartTime.Length), 30), container.NewCenter(d.StartTime.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Fail.Length), 30), container.NewCenter(d.Fail.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.AvgRTT.Length), 30), container.NewCenter(d.AvgRTT.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Recording.Length), 30), container.NewCenter(d.Recording.Object)),
	)

	// Create a thick line using a rectangle
	thickLine := canvas.NewRectangle(color.RGBA{200, 200, 200, 255})
	thickLine.SetMinSize(fyne.NewSize(200, 2)) // Adjust width & thickness

	d.icmpTableRow = container.New(layout.NewVBoxLayout(),
		row,
		thickLine,
	)
}

func (d *icmpGUIRow) GenerateHeaderRow() *fyne.Container {

	// table row
	header := container.New(layout.NewHBoxLayout(),

		container.NewGridWrap(fyne.NewSize(float32(d.Action.Length), 30), widget.NewLabelWithStyle(d.Action.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Index.Length), 30), widget.NewLabelWithStyle(d.Index.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Seq.Length), 30), widget.NewLabelWithStyle(d.Seq.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Status.Length), 30), widget.NewLabelWithStyle(d.Status.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.HostName.Length), 30), widget.NewLabelWithStyle(d.HostName.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.IP.Length), 30), widget.NewLabelWithStyle(d.IP.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Payload.Length), 30), widget.NewLabelWithStyle(d.Payload.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.DF.Length), 30), widget.NewLabelWithStyle(d.DF.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.RTT.Length), 30), widget.NewLabelWithStyle(d.RTT.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.StartTime.Length), 30), widget.NewLabelWithStyle(d.StartTime.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Fail.Length), 30), widget.NewLabelWithStyle(d.Fail.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.AvgRTT.Length), 30), widget.NewLabelWithStyle(d.AvgRTT.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Recording.Length), 30), widget.NewLabelWithStyle(d.Recording.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
	)

	// Create a thick line using a rectangle
	thickLine := canvas.NewRectangle(color.RGBA{200, 200, 200, 255})
	thickLine.SetMinSize(fyne.NewSize(200, 3)) // Adjust width & thickness

	headerRow := container.New(layout.NewVBoxLayout(),
		header,
		thickLine,
	)

	return headerRow
}

func (d *icmpGUIRow) UpdateRow(p *ntPinger.Packet) {

	// seq
	d.Seq.Object.(*widget.Label).Text = strconv.Itoa((*p).(*ntPinger.PacketICMP).Seq)
	d.Seq.Object.(*widget.Label).Refresh()

	// status
	d.Status.Object.(*canvas.Text).TextStyle.Bold = true
	d.Status.Object.(*canvas.Text).TextSize = 15
	if (*p).(*ntPinger.PacketICMP).Status {
		d.Status.Object.(*canvas.Text).Text = "Success"
		d.Status.Object.(*canvas.Text).Color = color.RGBA{0, 128, 0, 255}
	} else {
		d.Status.Object.(*canvas.Text).Text = "Fail"
		d.Status.Object.(*canvas.Text).Color = color.RGBA{255, 0, 0, 255}
	}
	d.Status.Object.(*canvas.Text).Refresh()

	// Response Time
	if (*p).(*ntPinger.PacketICMP).Status {
		d.RTT.Object.(*widget.Label).Text = (*p).(*ntPinger.PacketICMP).RTT.String()
	} else {
		d.RTT.Object.(*widget.Label).Text = "--"
	}
	d.RTT.Object.(*widget.Label).Refresh()

	// Fail Rate
	d.Fail.Object.(*widget.Label).Text = fmt.Sprintf("%.2f%%", (*p).(*ntPinger.PacketICMP).PacketLoss*100)
	d.Fail.Object.(*widget.Label).Refresh()

	// AvgRTT
	d.AvgRTT.Object.(*widget.Label).Text = (*p).(*ntPinger.PacketICMP).AvgRtt.String()
	d.AvgRTT.Object.(*widget.Label).Refresh()
}

// ******* struct icmpObject ********
type icmpObject struct {
	testSummary          *SummaryData
	ChartData            []ntchart.ChartPoint
	icmpGUI              icmpGUIRow
	PopUpChartWindowFlag bool
}

func (d *icmpObject) GetType() string {
	return d.testSummary.Type
}

func (d *icmpObject) GetSummary() *SummaryData {
	return d.testSummary
}

func (d *icmpObject) GetChartData() *[]ntchart.ChartPoint {
	return &d.ChartData
}

func (d *icmpObject) GetUUID() string {
	return d.testSummary.GetUUID()
}

func (d *icmpObject) UpdateRecording(recording bool) {
	if recording {
		d.icmpGUI.Recording.Object.(*widget.Label).Text = "ON"
	} else {
		d.icmpGUI.Recording.Object.(*widget.Label).Text = "OFF"
	}

	d.icmpGUI.Recording.Object.(*widget.Label).Refresh()
}

func (d *icmpObject) Initial(testSummary *SummaryData) {

	// initial the PopUpChartWindowFlag
	d.PopUpChartWindowFlag = false

	// test Summary
	d.testSummary = testSummary

	// ChartData
	d.ChartData = []ntchart.ChartPoint{}

	// icmp GUI
	d.icmpGUI = icmpGUIRow{}
	d.icmpGUI.Initial()
}

func (d *icmpObject) UpdateChartData(pkt *ntPinger.Packet) {
	d.ChartData = append(d.ChartData, ntchart.ConvertFromPacketToChartPoint(*pkt))
}

func (d *icmpObject) DisplayChartDataTerminal() {
	for _, d := range d.ChartData {
		fmt.Println(d)
	}
}

// Stop the pinger
func (d *icmpObject) Stop(p *ntPinger.Pinger) {
	p.PingerEnd = true
	time.Sleep(200 * time.Millisecond) // wait for the test to stop

	d.icmpGUI.StopBtn.Disable()
	d.icmpGUI.CloseBtn.Enable()
	d.icmpGUI.ReplayBtn.Enable()

	d.icmpGUI.Status.Object.(*canvas.Text).Text = "Stop"
	d.icmpGUI.Status.Object.(*canvas.Text).Color = color.RGBA{165, 42, 42, 255}
	d.icmpGUI.Status.Object.(*canvas.Text).Refresh()

	// Unregister test from test register
	UnregisterTest(&testRegister, d.GetUUID())
}

// func: Add Ping Row
func IcmpAddPingRow(a fyne.App, indexPing *int, inputVars *ntPinger.InputVars, icmpTableBody *fyne.Container, recording bool, db *sql.DB, entryChan chan ntdb.DbEntry, errChan chan error) {

	// test uuid
	testUUID := GenerateShortUUID()

	// Register Test
	testRegister = append(testRegister, testUUID)

	// Create Summary Data
	mySumData := SummaryData{}
	mySumData.Initial("icmp", inputVars.DestHost, Iv2NtCmd(recording, *inputVars), time.Now(), testUUID)

	// Add History DB record
	historyRecord := ntdb.HistoryEntry{}
	historyRecord.TableName = "history"
	historyRecord.StartTime = time.Now()
	historyRecord.TestType = mySumData.Type
	historyRecord.Command = mySumData.ntCmd
	historyRecord.UUID = testUUID
	historyRecord.Recorded = recording

	entryChan <- &historyRecord

	// build recording table if "recording" is true
	recordingTableName := fmt.Sprintf("%s_%s", historyRecord.TestType, historyRecord.UUID)
	if recording {
		err := ntdb.CreateTestResultsTable(db, "icmp", recordingTableName)
		if err != nil {
			logger.Println(err)
			return
		}
	}

	// ResultGenerateicmp()
	myicmpPing := icmpObject{}
	myicmpPing.Initial(&mySumData)

	// update index
	myPingIndex := strconv.Itoa(*indexPing)

	myicmpPing.icmpGUI.Index.Object.(*widget.Label).Text = myPingIndex
	myicmpPing.icmpGUI.Index.Object.(*widget.Label).Refresh()
	*indexPing++

	// Update HostName
	testDestHost := inputVars.DestHost
	myicmpPing.icmpGUI.HostName.Object.(*widget.Label).Text = testDestHost
	myicmpPing.icmpGUI.HostName.Object.(*widget.Label).Refresh()

	// Update IP
	DestHostIPSlide, _ := net.LookupHost(testDestHost) // ignore err check as resolvable is checked in the NewTest validation
	myicmpPing.icmpGUI.IP.Object.(*widget.Label).Text = DestHostIPSlide[0]
	(*inputVars).DestHost = DestHostIPSlide[0] // update the input Var DestHost to be IP Address
	myicmpPing.icmpGUI.IP.Object.(*widget.Label).Refresh()

	// Update Payload
	myicmpPing.icmpGUI.Payload.Object.(*widget.Label).Text = strconv.Itoa(inputVars.PayLoadSize)
	myicmpPing.icmpGUI.Payload.Object.(*widget.Label).Refresh()

	// Update DF
	DfBit := "OFF"
	if inputVars.Icmp_DF {
		DfBit = "ON"
	} else {
		DfBit = "OFF"
	}
	myicmpPing.icmpGUI.DF.Object.(*widget.Label).Text = DfBit
	myicmpPing.icmpGUI.DF.Object.(*widget.Label).Refresh()

	// Update StartTime
	myicmpPing.icmpGUI.StartTime.Object.(*widget.Label).Text = mySumData.StartTime.Format("2006-01-02 15:04:05 MST")
	myicmpPing.icmpGUI.StartTime.Object.(*widget.Label).Refresh()

	// update recording
	if recording {
		myicmpPing.icmpGUI.Recording.Object.(*widget.Label).Text = "ON"
	} else {
		myicmpPing.icmpGUI.Recording.Object.(*widget.Label).Text = "OFF"
	}
	myicmpPing.icmpGUI.Recording.Object.(*widget.Label).Refresh()

	// update table body
	icmpTableBody.Add(myicmpPing.icmpGUI.icmpTableRow)
	icmpTableBody.Refresh()

	// ** start ntPinger Probe **

	// Start Ping Main Command
	p, err := ntPinger.NewPinger(*inputVars)

	if err != nil {
		fmt.Println(err)
		logger.Println(err)
		return
	}

	// OnTapped Func - Chart btn
	PopUpChartWindowFlag := false
	myicmpPing.icmpGUI.ChartBtn.OnTapped = func() {
		// only open the new chart window when there are more than 2 test records && No pop up window (PopUpChartWindowFlag = false)
		if len(myicmpPing.ChartData) > 2 && !PopUpChartWindowFlag {
			// set pop up window flag to true
			PopUpChartWindowFlag = true
			// pop up char window
			NewChartWindow(a, &myicmpPing, &recording, p, db, entryChan, errChan, &PopUpChartWindowFlag)
		}
	}

	// OnTapped Func - Stop btn
	myicmpPing.icmpGUI.StopBtn.OnTapped = func() {
		myicmpPing.testSummary.EndTime = time.Now()
		myicmpPing.Stop(p)
	}

	// OnTapped Func - Replay btn
	myicmpPing.icmpGUI.ReplayBtn.OnTapped = func() {
		// re-launch a new go routine for icmpAddPingRow with the same InputVar
		go IcmpAddPingRow(a, indexPing, inputVars, icmpTableBody, recording, db, entryChan, errChan)
	}

	// OnTapped Func - close btn
	myicmpPing.icmpGUI.CloseBtn.OnTapped = func() {
		icmpTableBody.Remove(myicmpPing.icmpGUI.icmpTableRow)
		icmpTableBody.Refresh()
	}

	// start ping go routing
	go p.Run(errChan)

	// harvest the result
	loopClose := false

	for {
		// check loopClose Flag
		if loopClose || p.PingerEnd {
			break
		}

		// select option
		select {

		// ends this test when app is closing
		case <-appCtx.Done():
			p.PingerEnd = true
			loopClose = true
			//fmt.Printf("Closing Testing: %s\n", myPingIndex)

		// harvest the Probe results
		case pkt, ok := <-p.ProbeChan:

			// if p.ProbeChan is closed, exit
			if !ok {
				loopClose = true
				break // break select, bypass following code in the same case
				// if the current test is NOT in the test register, exit
			} else if !existingTestCheck(&testRegister, myicmpPing.GetUUID()) {
				loopClose = true
				break // break select, bypass following code in the same case
			}
			myicmpPing.icmpGUI.UpdateRow(&pkt)         // update row display
			myicmpPing.UpdateChartData(&pkt)           // update chart Data slide
			myicmpPing.testSummary.UpdateRunning(&pkt) // update summary Data

			// if recording is true, add the &pkt to DB table
			if recording {
				icmpEntry := ntdb.ConvertPkt2DbEntry(pkt, recordingTableName)
				entryChan <- icmpEntry
			}

		// harvest the errChan input
		case err := <-errChan:
			logger.Println(err)
			return
		}
	}
}
