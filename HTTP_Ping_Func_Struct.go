package main

import (
	"database/sql"
	"fmt"
	"image/color"
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
var _ testGUIRow = (*httpGUIRow)(nil)
var _ testObject = (*httpObject)(nil)

// ******* struct httpGUIRow ********

type httpGUIRow struct {
	Index         pingCell
	Seq           pingCell
	Status        pingCell
	Method        pingCell // i.e GET, Fixed
	URL           pingCell // Fixed
	Response_Code pingCell
	Response_Time pingCell
	StartTime     pingCell // fixed
	Fail          pingCell
	AvgRTT        pingCell
	Recording     pingCell

	ChartBtn     *widget.Button
	StopBtn      *widget.Button
	ReplayBtn    *widget.Button
	CloseBtn     *widget.Button
	Action       pingCell
	httpTableRow *fyne.Container
}

func (d *httpGUIRow) Initial() {

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

	d.Method.Label = "Method"
	d.Method.Length = 65
	d.Method.Object = widget.NewLabel("--")

	d.URL.Label = "URL"
	d.URL.Length = 300
	d.URL.Object = widget.NewLabel("--")

	d.Response_Code.Label = "Code"
	d.Response_Code.Length = 100
	d.Response_Code.Object = widget.NewLabel("--")

	d.Response_Time.Label = "RTT"
	d.Response_Time.Length = 90
	d.Response_Time.Object = widget.NewLabel("--")

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
		container.NewGridWrap(fyne.NewSize(float32(d.Method.Length), 30), container.NewCenter(d.Method.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.URL.Length), 30), container.NewCenter(d.URL.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Response_Code.Length), 30), container.NewCenter(d.Response_Code.Object)),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Response_Time.Length), 30), container.NewCenter(d.Response_Time.Object)),
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

	d.httpTableRow = container.New(layout.NewVBoxLayout(),
		row,
		thickLine,
	)
}

func (d *httpGUIRow) GenerateHeaderRow() *fyne.Container {

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
		container.NewGridWrap(fyne.NewSize(float32(d.Method.Length), 30), widget.NewLabelWithStyle(d.Method.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.URL.Length), 30), widget.NewLabelWithStyle(d.URL.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Response_Code.Length), 30), widget.NewLabelWithStyle(d.Response_Code.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
		GUIVerticalSeparator(),
		container.NewGridWrap(fyne.NewSize(float32(d.Response_Time.Length), 30), widget.NewLabelWithStyle(d.Response_Time.Label, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})),
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

func (d *httpGUIRow) UpdateRow(p *ntPinger.Packet) {

	// seq
	d.Seq.Object.(*widget.Label).Text = strconv.Itoa((*p).(*ntPinger.PacketHTTP).Seq)
	d.Seq.Object.(*widget.Label).Refresh()

	// status
	d.Status.Object.(*canvas.Text).TextStyle.Bold = true
	d.Status.Object.(*canvas.Text).TextSize = 15
	if (*p).(*ntPinger.PacketHTTP).Status {
		d.Status.Object.(*canvas.Text).Text = "Success"
		d.Status.Object.(*canvas.Text).Color = color.RGBA{0, 128, 0, 255}
	} else {
		d.Status.Object.(*canvas.Text).Text = "Fail"
		d.Status.Object.(*canvas.Text).Color = color.RGBA{255, 0, 0, 255}
	}
	d.Status.Object.(*canvas.Text).Refresh()

	// Response_Code
	d.Response_Code.Object.(*widget.Label).Text = strconv.Itoa((*p).(*ntPinger.PacketHTTP).Http_response_code)
	d.Response_Code.Object.(*widget.Label).Refresh()

	// Response Time
	if (*p).(*ntPinger.PacketHTTP).Status {
		d.Response_Time.Object.(*widget.Label).Text = (*p).(*ntPinger.PacketHTTP).RTT.String()
	} else {
		d.Response_Time.Object.(*widget.Label).Text = "--"
	}
	d.Response_Time.Object.(*widget.Label).Refresh()

	// Fail Rate
	d.Fail.Object.(*widget.Label).Text = fmt.Sprintf("%.2f%%", (*p).(*ntPinger.PacketHTTP).PacketLoss*100)
	d.Fail.Object.(*widget.Label).Refresh()

	// AvgRTT
	d.AvgRTT.Object.(*widget.Label).Text = (*p).(*ntPinger.PacketHTTP).AvgRtt.String()
	d.AvgRTT.Object.(*widget.Label).Refresh()
}

// ******* struct httpObject ********
type httpObject struct {
	testSummary          *SummaryData
	ChartData            []ntchart.ChartPoint
	httpGUI              httpGUIRow
	PopUpChartWindowFlag bool
}

func (d *httpObject) GetType() string {
	return d.testSummary.Type
}

func (d *httpObject) GetSummary() *SummaryData {
	return d.testSummary
}

func (d *httpObject) GetChartData() *[]ntchart.ChartPoint {
	return &d.ChartData
}

func (d *httpObject) GetUUID() string {
	return d.testSummary.GetUUID()
}

func (d *httpObject) UpdateRecording(recording bool) {
	if recording {
		d.httpGUI.Recording.Object.(*widget.Label).Text = "ON"
	} else {
		d.httpGUI.Recording.Object.(*widget.Label).Text = "OFF"
	}

	d.httpGUI.Recording.Object.(*widget.Label).Refresh()
}

func (d *httpObject) Initial(testSummary *SummaryData) {

	// initial the PopUpChartWindowFlag
	d.PopUpChartWindowFlag = false

	// test Summary
	d.testSummary = testSummary

	// ChartData
	d.ChartData = []ntchart.ChartPoint{}

	// http GUI
	d.httpGUI = httpGUIRow{}
	d.httpGUI.Initial()
}

func (d *httpObject) UpdateChartData(pkt *ntPinger.Packet) {
	d.ChartData = append(d.ChartData, ntchart.ConvertFromPacketToChartPoint(*pkt))
}

func (d *httpObject) DisplayChartDataTerminal() {
	for _, d := range d.ChartData {
		fmt.Println(d)
	}
}

// Stop the pinger
func (d *httpObject) Stop(p *ntPinger.Pinger) {
	p.PingerEnd = true
	time.Sleep(200 * time.Millisecond) // wait for the test to stop

	d.httpGUI.StopBtn.Disable()
	d.httpGUI.CloseBtn.Enable()
	d.httpGUI.ReplayBtn.Enable()

	d.httpGUI.Status.Object.(*canvas.Text).Text = "Stop"
	d.httpGUI.Status.Object.(*canvas.Text).Color = color.RGBA{165, 42, 42, 255}
	d.httpGUI.Status.Object.(*canvas.Text).Refresh()

	// Unregister test from test register
	UnregisterTest(&testRegister, d.GetUUID())
}

// func: Add Ping Row
func HttpAddPingRow(a fyne.App, indexPing *int, inputVars *ntPinger.InputVars, httpTableBody *fyne.Container, recording bool, db *sql.DB, entryChan chan ntdb.DbEntry, errChan chan error) {

	// test uuid
	testUUID := GenerateShortUUID()

	// Register Test
	testRegister = append(testRegister, testUUID)

	// Create Summary Data
	mySumData := SummaryData{}
	mySumData.Initial("http", inputVars.DestHost, Iv2NtCmd(recording, *inputVars), time.Now(), testUUID)

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
		err := ntdb.CreateTestResultsTable(db, "http", recordingTableName)
		if err != nil {
			logger.Println(err)
			return
		}
	}

	// ResultGeneratehttp()
	myhttpPing := httpObject{}
	myhttpPing.Initial(&mySumData)

	// update index
	myPingIndex := strconv.Itoa(*indexPing)

	myhttpPing.httpGUI.Index.Object.(*widget.Label).Text = myPingIndex
	myhttpPing.httpGUI.Index.Object.(*widget.Label).Refresh()
	*indexPing++

	// Update Method
	myhttpPing.httpGUI.Method.Object.(*widget.Label).Text = inputVars.Http_method
	myhttpPing.httpGUI.Method.Object.(*widget.Label).Refresh()

	// Update URL
	testURL := ConstructURL(inputVars.Http_scheme, inputVars.DestHost, inputVars.Http_path, inputVars.DestPort)
	myhttpPing.httpGUI.URL.Object.(*widget.Label).Text = TruncateString(testURL, 42)
	myhttpPing.httpGUI.URL.Object.(*widget.Label).Refresh()

	// Update StartTime
	myhttpPing.httpGUI.StartTime.Object.(*widget.Label).Text = mySumData.StartTime.Format("2006-01-02 15:04:05 MST")
	myhttpPing.httpGUI.StartTime.Object.(*widget.Label).Refresh()

	// update recording
	if recording {
		myhttpPing.httpGUI.Recording.Object.(*widget.Label).Text = "ON"
	} else {
		myhttpPing.httpGUI.Recording.Object.(*widget.Label).Text = "OFF"
	}
	myhttpPing.httpGUI.Recording.Object.(*widget.Label).Refresh()

	// update table body
	httpTableBody.Add(myhttpPing.httpGUI.httpTableRow)
	httpTableBody.Refresh()

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
	myhttpPing.httpGUI.ChartBtn.OnTapped = func() {
		// only open the new chart window when there are more than 2 test records && No pop up window (PopUpChartWindowFlag = false)
		if len(myhttpPing.ChartData) > 2 && !PopUpChartWindowFlag {
			// set pop up window flag to true
			PopUpChartWindowFlag = true
			// pop up char window
			NewChartWindow(a, &myhttpPing, &recording, p, db, entryChan, errChan, &PopUpChartWindowFlag)
		}
	}

	// OnTapped Func - Stop btn
	myhttpPing.httpGUI.StopBtn.OnTapped = func() {
		myhttpPing.testSummary.EndTime = time.Now()
		myhttpPing.Stop(p)
	}

	// OnTapped Func - Replay btn
	myhttpPing.httpGUI.ReplayBtn.OnTapped = func() {
		// re-launch a new go routine for httpAddPingRow with the same InputVar
		go HttpAddPingRow(a, indexPing, inputVars, httpTableBody, recording, db, entryChan, errChan)
	}

	// OnTapped Func - close btn
	myhttpPing.httpGUI.CloseBtn.OnTapped = func() {
		httpTableBody.Remove(myhttpPing.httpGUI.httpTableRow)
		httpTableBody.Refresh()
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
			} else if !existingTestCheck(&testRegister, myhttpPing.GetUUID()) {
				loopClose = true
				break // break select, bypass following code in the same case
			}
			myhttpPing.httpGUI.UpdateRow(&pkt)         // update row display
			myhttpPing.UpdateChartData(&pkt)           // update chart Data slide
			myhttpPing.testSummary.UpdateRunning(&pkt) // update summary Data

			// if recording is true, add the &pkt to DB table
			if recording {
				httpEntry := ntdb.ConvertPkt2DbEntry(pkt, recordingTableName)
				entryChan <- httpEntry
			}

		// harvest the errChan input
		case err := <-errChan:
			logger.Println(err)
			return
		}
	}

}

// Construct URL
func ConstructURL(Http_scheme, DestHost, Http_path string, DestPort int) string {

	url := ""

	if Http_path == "" {
		if Http_scheme == "http" && DestPort == 80 {
			url = fmt.Sprintf("%s://%s", Http_scheme, DestHost)
		} else if Http_scheme == "https" && DestPort == 443 {
			url = fmt.Sprintf("%s://%s", Http_scheme, DestHost)
		} else {
			url = fmt.Sprintf("%s://%s:%d", Http_scheme, DestHost, DestPort)
		}

	} else {
		if Http_scheme == "http" && DestPort == 80 {
			url = fmt.Sprintf("%s://%s/%s", Http_scheme, DestHost, Http_path)
		} else if Http_scheme == "https" && DestPort == 443 {
			url = fmt.Sprintf("%s://%s/%s", Http_scheme, DestHost, Http_path)
		} else {
			url = fmt.Sprintf("%s://%s:%d/%s", Http_scheme, DestHost, DestPort, Http_path)
		}
	}
	return url
}
