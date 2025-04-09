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
var _ testGUIRow = (*tcpGUIRow)(nil)
var _ testObject = (*tcpObject)(nil)

// ******* struct tcpGUIRow ********

type tcpGUIRow struct {
	Index     pingCell
	Seq       pingCell
	Status    pingCell
	HostName  pingCell // fixed
	IP        pingCell // fixed
	Port      pingCell // fixed
	Payload   pingCell // fixed
	RTT       pingCell
	StartTime pingCell // sendDateTime
	Fail      pingCell
	AvgRTT    pingCell
	Recording pingCell

	ChartBtn    *widget.Button
	StopBtn     *widget.Button
	ReplayBtn   *widget.Button
	CloseBtn    *widget.Button
	Action      pingCell
	tcpTableRow *fyne.Container
}

func (d *tcpGUIRow) Initial() {

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

	d.Port.Label = "Port"
	d.Port.Length = 100
	d.Port.Object = widget.NewLabel("--")

	d.Payload.Label = "Payload"
	d.Payload.Length = 90
	d.Payload.Object = widget.NewLabel("--")

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
		container.NewGridWrap(fyne.NewSize(float32(d.Port.Length), 30), container.NewCenter(d.Port.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Payload.Length), 30), container.NewCenter(d.Payload.Object)),
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

	d.tcpTableRow = container.New(layout.NewVBoxLayout(),
		row,
		thickLine,
	)
}

func (d *tcpGUIRow) GenerateHeaderRow() *fyne.Container {

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
		container.NewGridWrap(fyne.NewSize(float32(d.Port.Length), 30), widget.NewLabelWithStyle(d.Port.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Payload.Length), 30), widget.NewLabelWithStyle(d.Payload.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
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

func (d *tcpGUIRow) UpdateRow(p *ntPinger.Packet) {

	// seq
	d.Seq.Object.(*widget.Label).Text = strconv.Itoa((*p).(*ntPinger.PacketTCP).Seq)
	d.Seq.Object.(*widget.Label).Refresh()

	// status
	d.Status.Object.(*canvas.Text).TextStyle.Bold = true
	d.Status.Object.(*canvas.Text).TextSize = 15
	if (*p).(*ntPinger.PacketTCP).Status {
		d.Status.Object.(*canvas.Text).Text = "Success"
		d.Status.Object.(*canvas.Text).Color = color.RGBA{0, 128, 0, 255}
	} else {
		d.Status.Object.(*canvas.Text).Text = "Fail"
		d.Status.Object.(*canvas.Text).Color = color.RGBA{255, 0, 0, 255}
	}
	d.Status.Object.(*canvas.Text).Refresh()

	// Response Time
	if (*p).(*ntPinger.PacketTCP).Status {
		d.RTT.Object.(*widget.Label).Text = (*p).(*ntPinger.PacketTCP).RTT.String()
	} else {
		d.RTT.Object.(*widget.Label).Text = "--"
	}
	d.RTT.Object.(*widget.Label).Refresh()

	// Fail Rate
	d.Fail.Object.(*widget.Label).Text = fmt.Sprintf("%.2f%%", (*p).(*ntPinger.PacketTCP).PacketLoss*100)
	d.Fail.Object.(*widget.Label).Refresh()

	// AvgRTT
	d.AvgRTT.Object.(*widget.Label).Text = (*p).(*ntPinger.PacketTCP).AvgRtt.String()
	d.AvgRTT.Object.(*widget.Label).Refresh()
}

// ******* struct tcpObject ********
type tcpObject struct {
	testSummary          *SummaryData
	ChartData            []ntchart.ChartPoint
	tcpGUI               tcpGUIRow
	PopUpChartWindowFlag bool
}

func (d *tcpObject) GetType() string {
	return d.testSummary.Type
}

func (d *tcpObject) GetSummary() *SummaryData {
	return d.testSummary
}

func (d *tcpObject) GetChartData() *[]ntchart.ChartPoint {
	return &d.ChartData
}

func (d *tcpObject) GetUUID() string {
	return d.testSummary.GetUUID()
}

func (d *tcpObject) UpdateRecording(recording bool) {
	if recording {
		d.tcpGUI.Recording.Object.(*widget.Label).Text = "ON"
	} else {
		d.tcpGUI.Recording.Object.(*widget.Label).Text = "OFF"
	}

	d.tcpGUI.Recording.Object.(*widget.Label).Refresh()
}

func (d *tcpObject) Initial(testSummary *SummaryData) {

	// initial the PopUpChartWindowFlag
	d.PopUpChartWindowFlag = false

	// test Summary
	d.testSummary = testSummary

	// ChartData
	d.ChartData = []ntchart.ChartPoint{}

	// tcp GUI
	d.tcpGUI = tcpGUIRow{}
	d.tcpGUI.Initial()
}

func (d *tcpObject) UpdateChartData(pkt *ntPinger.Packet) {
	d.ChartData = append(d.ChartData, ntchart.ConvertFromPacketToChartPoint(*pkt))
}

func (d *tcpObject) DisplayChartDataTerminal() {
	for _, d := range d.ChartData {
		fmt.Println(d)
	}
}

// Stop the pinger
func (d *tcpObject) Stop(p *ntPinger.Pinger) {
	p.PingerEnd = true
	time.Sleep(200 * time.Millisecond) // wait for the test to stop

	d.tcpGUI.StopBtn.Disable()
	d.tcpGUI.CloseBtn.Enable()
	d.tcpGUI.ReplayBtn.Enable()

	d.tcpGUI.Status.Object.(*canvas.Text).Text = "Stop"
	d.tcpGUI.Status.Object.(*canvas.Text).Color = color.RGBA{165, 42, 42, 255}
	d.tcpGUI.Status.Object.(*canvas.Text).Refresh()

	// Unregister test from test register
	UnregisterTest(&testRegister, d.GetUUID())
}

// func: Add Ping Row
func TcpAddPingRow(a fyne.App, indexPing *int, inputVars *ntPinger.InputVars, tcpTableBody *fyne.Container, recording bool, db *sql.DB, entryChan chan ntdb.DbEntry, errChan chan error) {

	// test uuid
	testUUID := GenerateShortUUID()

	// Register Test
	testRegister = append(testRegister, testUUID)

	// Create Summary Data
	mySumData := SummaryData{}
	mySumData.Initial("tcp", inputVars.DestHost, Iv2NtCmd(recording, *inputVars), time.Now(), testUUID)

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
		err := ntdb.CreateTestResultsTable(db, "tcp", recordingTableName)
		if err != nil {
			logger.Println(err)
			return
		}
	}

	// ResultGeneratetcp()
	mytcpPing := tcpObject{}
	mytcpPing.Initial(&mySumData)

	// update index
	myPingIndex := strconv.Itoa(*indexPing)

	mytcpPing.tcpGUI.Index.Object.(*widget.Label).Text = myPingIndex
	mytcpPing.tcpGUI.Index.Object.(*widget.Label).Refresh()
	*indexPing++

	// Update HostName
	testDestHost := inputVars.DestHost
	mytcpPing.tcpGUI.HostName.Object.(*widget.Label).Text = testDestHost
	mytcpPing.tcpGUI.HostName.Object.(*widget.Label).Refresh()

	// Update IP
	DestHostIPSlide, _ := net.LookupHost(testDestHost) // ignore err check as resolvable is checked in the NewTest validation
	mytcpPing.tcpGUI.IP.Object.(*widget.Label).Text = DestHostIPSlide[0]
	(*inputVars).DestHost = DestHostIPSlide[0] // update the input Var DestHost to be IP Address
	mytcpPing.tcpGUI.IP.Object.(*widget.Label).Refresh()

	// Update Port
	mytcpPing.tcpGUI.Port.Object.(*widget.Label).Text = strconv.Itoa(inputVars.DestPort)
	mytcpPing.tcpGUI.Port.Object.(*widget.Label).Refresh()

	// Update Payload
	mytcpPing.tcpGUI.Payload.Object.(*widget.Label).Text = strconv.Itoa(inputVars.PayLoadSize)
	mytcpPing.tcpGUI.Payload.Object.(*widget.Label).Refresh()

	// Update StartTime
	mytcpPing.tcpGUI.StartTime.Object.(*widget.Label).Text = mySumData.StartTime.Format("2006-01-02 15:04:05 MST")
	mytcpPing.tcpGUI.StartTime.Object.(*widget.Label).Refresh()

	// update recording
	if recording {
		mytcpPing.tcpGUI.Recording.Object.(*widget.Label).Text = "ON"
	} else {
		mytcpPing.tcpGUI.Recording.Object.(*widget.Label).Text = "OFF"
	}
	mytcpPing.tcpGUI.Recording.Object.(*widget.Label).Refresh()

	// update table body
	tcpTableBody.Add(mytcpPing.tcpGUI.tcpTableRow)
	tcpTableBody.Refresh()

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
	mytcpPing.tcpGUI.ChartBtn.OnTapped = func() {
		// only open the new chart window when there are more than 2 test records && No pop up window (PopUpChartWindowFlag = false)
		if len(mytcpPing.ChartData) > 2 && !PopUpChartWindowFlag {
			// set pop up window flag to true
			PopUpChartWindowFlag = true
			// pop up char window
			NewChartWindow(a, &mytcpPing, &recording, p, db, entryChan, errChan, &PopUpChartWindowFlag)
		}
	}

	// OnTapped Func - Stop btn
	mytcpPing.tcpGUI.StopBtn.OnTapped = func() {
		mytcpPing.testSummary.EndTime = time.Now()
		mytcpPing.Stop(p)
	}

	// OnTapped Func - Replay btn
	mytcpPing.tcpGUI.ReplayBtn.OnTapped = func() {
		// re-launch a new go routine for TcpAddPingRow with the same InputVar
		go TcpAddPingRow(a, indexPing, inputVars, tcpTableBody, recording, db, entryChan, errChan)
	}

	// OnTapped Func - close btn
	mytcpPing.tcpGUI.CloseBtn.OnTapped = func() {
		tcpTableBody.Remove(mytcpPing.tcpGUI.tcpTableRow)
		tcpTableBody.Refresh()
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
			} else if !existingTestCheck(&testRegister, mytcpPing.GetUUID()) {
				loopClose = true
				break // break select, bypass following code in the same case
			}
			mytcpPing.tcpGUI.UpdateRow(&pkt)          // update row display
			mytcpPing.UpdateChartData(&pkt)           // update chart Data slide
			mytcpPing.testSummary.UpdateRunning(&pkt) // update summary Data

			// if recording is true, add the &pkt to DB table
			if recording {
				tcpEntry := ntdb.ConvertPkt2DbEntry(pkt, recordingTableName)
				entryChan <- tcpEntry
			}

		// harvest the errChan input
		case err := <-errChan:
			logger.Println(err)
			return
		}
	}
}
