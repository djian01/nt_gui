package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/djian01/nt/pkg/ntPinger"
	"github.com/djian01/nt_gui/pkg/ntdb"
)

func NewChartWindow(a fyne.App, testObj testObject, recording *bool, p *ntPinger.Pinger, db *sql.DB, entryChan chan ntdb.DbEntry, errChan chan error, PopUpChartWindowFlag *bool) {

	// ** "p *ntPinger.Pinger" is only for testObj.Stop(p) purpose

	// Create a global cancelable context
	var testCtx, testCancelFunc = context.WithCancel(context.Background())

	// pause flag
	chartPauseFlag := false

	// Initial Vars
	testType := testObj.GetType()
	testSummary := testObj.GetSummary()
	testChartData := testObj.GetChartData()

	// Initial New Chart Window
	newChartWindow := a.NewWindow(fmt.Sprintf("%s Chart", strings.ToUpper(testType)))
	newChartWindow.Resize(fyne.NewSize(1400, 900))
	newChartWindow.CenterOnScreen()
	newChartWindow.SetOnClosed(func() {
		// set the SetPopUpChartWindowFlag to false
		*PopUpChartWindowFlag = false
		// call the cancel func to close the go routine
		testCancelFunc()
	})

	// summary Card
	testSummaryUI := SummaryUI{}
	testSummaryUI.Initial()
	testSummaryUI.CreateCard()
	testSummaryUI.UpdateUI_Initial(testSummary)

	// Chart card

	//// Chart Btn Pause
	chartBtnPause := widget.NewButtonWithIcon("Pause Chart Update", theme.MediaPauseIcon(), func() {})
	chartBtnPause.Importance = widget.WarningImportance

	if !chartPauseFlag {
		chartBtnPause.Enable()
	} else {
		chartBtnPause.Disable()
	}
	chartBtnPauseContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(220, 30)), chartBtnPause)

	//// Chart Btn Play
	chartBtnPlay := widget.NewButtonWithIcon("Resume Chart Update", theme.MediaPlayIcon(), func() {})
	chartBtnPlay.Importance = widget.WarningImportance

	if chartPauseFlag {
		chartBtnPlay.Enable()
	} else {
		chartBtnPlay.Disable()
	}
	chartBtnPlayContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(220, 30)), chartBtnPlay)

	//// Chart Btn Record
	chartBtnRecord := widget.NewButtonWithIcon("Record Test", theme.MediaRecordIcon(), func() {})
	chartBtnRecord.Importance = widget.WarningImportance
	if *recording { // if recording is enabled, disable the recording button
		// disable the Recording btn
		chartBtnRecord.Disable()
	}
	chartBtnRecordContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(220, 30)), chartBtnRecord)

	//// Chart Btn Stop
	chartBtnStop := widget.NewButtonWithIcon("Stop Test", theme.MediaStopIcon(), func() {})
	chartBtnStop.Importance = widget.DangerImportance
	chartBtnStopContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(220, 30)), chartBtnStop)

	//// Chart Btn Export CSV
	chartBtnExport := widget.NewButtonWithIcon("Export CSV", theme.DocumentSaveIcon(), func() {})
	chartBtnExport.Importance = widget.HighImportance
	chartBtnExport.Disable()
	chartBtnExportContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(220, 30)), chartBtnExport)

	//// Read DB Progress
	chartBtnReadDBProgressBar := widget.NewProgressBarInfinite()
	chartBtnReadDBProgressLabel := widget.NewLabel(" Exporting CSV: ")
	chartBtnReadDBProgressLabel.TextStyle.Bold = true
	chartBtnReadDBProgress := container.New(layout.NewBorderLayout(nil, nil, chartBtnReadDBProgressLabel, nil), chartBtnReadDBProgressLabel, chartBtnReadDBProgressBar)
	chartBtnReadDBProgress.Hidden = true

	//// Chart Btn Container & Card
	chartBtnContainerInBtn := container.New(layout.NewHBoxLayout(), chartBtnPauseContainer, chartBtnPlayContainer, chartBtnRecordContainer, chartBtnStopContainer, chartBtnExportContainer)
	chartBtnContainerIn := container.New(layout.NewVBoxLayout(), chartBtnContainerInBtn, chartBtnReadDBProgress)
	chartBtnContainerOut := container.New(layout.NewCenterLayout(), chartBtnContainerIn)
	chartBtnCard := widget.NewCard("", "", chartBtnContainerOut)

	//// chart body
	chartBody := Chart{}
	chartBody.Initial()

	// Slider Card
	chartSlider := Slider{}
	chartSlider.Initial(0, 100, 0, 100)
	chartSlider.chartData = testChartData
	chartSlider.CreateCard()
	chartSlider.sliderCard.Hidden = true
	chartSlider.rangeSlider.OnChanged = func() { chartSlider.update() }

	// close window Btn Container
	chartWindowCloseBtn := widget.NewButton("Close Window", func() {
		newChartWindow.Close()
	})
	chartWindowCloseBtn.Importance = widget.HighImportance
	ChartWindowCloseContainer := container.New(layout.NewCenterLayout(), container.New(layout.NewGridWrapLayout(fyne.NewSize(200, 40)), chartWindowCloseBtn))

	// if the test is already stopped.
	if !existingTestCheck(&testRegister, testObj.GetUUID()) {
		// disable the all buttons
		chartBtnPause.Disable()
		chartBtnPlay.Disable()
		chartBtnStop.Disable()
		chartBtnRecord.Disable()
		// if the test is already stopped & recording is enabed, enable the export button
		if *recording {
			chartBtnExport.Enable()
		}

		// chart
		(chartBody).ChartUpdate(testType, testChartData)

		// update slider
		chartSlider.sliderCard.Hidden = false
		chartSlider.rangeSlider.UpdateValues(0, float64(len(*testChartData)-1), 0, float64(len(*testChartData)-1))
		chartSlider.update()

		// slider update chart image btn
		chartSlider.chartUpdateBtn.OnTapped = func() {
			chartSlider.BuildSliderChartData()
			chartSlider.UpdateChartImage(testType, &chartBody)
		}

		// slider reset chart image
		chartSlider.chartResetBtn.OnTapped = func() {
			chartSlider.ResetChartImage(testType, &chartBody)
		}

		// update summary UI
		(testSummaryUI).UpdateStaticUI(testSummary)

		// relaunch btn
		testSummaryUI.ntCmdBtn.OnTapped = func() {
			_, iv, err := NtCmd2Iv(testSummary.ntCmd)
			if err != nil {
				logger.Println(err)
			}

			// launch new test
			switch testType {
			case "dns":
				go DnsAddPingRow(a, &ntGlobal.dnsIndex, &iv, ntGlobal.dnsTable, *recording, db, entryChan, errChan)
			case "http":
				go HttpAddPingRow(a, &ntGlobal.httpIndex, &iv, ntGlobal.httpTable, *recording, db, entryChan, errChan)
			case "tcp":
				go TcpAddPingRow(a, &ntGlobal.tcpIndex, &iv, ntGlobal.tcpTable, *recording, db, entryChan, errChan)
			case "icmp":
				go IcmpAddPingRow(a, &ntGlobal.icmpIndex, &iv, ntGlobal.icmpTable, *recording, db, entryChan, errChan)
			}
			testSummaryUI.ntCmdBtn.Disable()
		}
	} else {
		// kickoff the go routine for chart/summary update
		go NewChartUpdate(testCtx, &chartPauseFlag, testObj, &testSummaryUI, &chartBody)
	}

	// ** update btn functions **

	//// Pause btn Function
	chartBtnPause.OnTapped = func() {

		chartPauseFlag = true
		chartBtnPause.Disable()
		chartBtnPlay.Enable()

		// chartDataSnapshot
		chartDataSnapshot := CloneChartPoints(testChartData)

		// chart
		(chartBody).ChartUpdate(testType, &chartDataSnapshot)

		// update slider value
		chartSlider.chartData = &chartDataSnapshot
		chartSlider.rangeSlider.UpdateValues(0, float64(len(chartDataSnapshot)-1), 0, float64(len(chartDataSnapshot)-1))
		chartSlider.update()

		// slider visible
		chartSlider.sliderCard.Hidden = false
		chartSlider.sliderCard.Refresh()

		// slider btn
		chartSlider.chartUpdateBtn.OnTapped = func() {
			chartSlider.BuildSliderChartData()
			chartSlider.UpdateChartImage(testType, &chartBody)
		}
		chartSlider.chartResetBtn.OnTapped = func() {
			chartSlider.ResetChartImage(testType, &chartBody)
		}
	}
	//// Play Btn Function
	chartBtnPlay.OnTapped = func() {

		chartPauseFlag = false
		chartBtnPause.Enable()
		chartBtnPlay.Disable()

		// slider visible
		chartSlider.sliderCard.Hidden = true
		chartSlider.sliderCard.Refresh()
	}
	//// Record Btn Function
	chartBtnRecord.OnTapped = func() {
		// disable the Recording btn
		chartBtnRecord.Disable()

		// build recording table if "recording" is enabled
		recordingTableName := fmt.Sprintf("%s_%s", testType, testObj.GetUUID())

		err := ntdb.CreateTestResultsTable(db, testType, recordingTableName)
		if err != nil {
			errChan <- err
		}

		// update the history record with recording on
		err = ntdb.UpdateFieldValue(db, "history", "uuid", "string", testObj.GetUUID(), "recorded", "int", "1")
		if err != nil {
			errChan <- err
		}

		// update the DNS ping Row recording field
		testObj.UpdateRecording(true)

		// set the recording to true
		*recording = true
	}

	//// Stop Btn Function
	chartBtnStop.OnTapped = func() {
		// if the test has not yet stopped, stop it
		if existingTestCheck(&testRegister, testObj.GetUUID()) {
			testObj.Stop(p)
		}

		chartBtnPause.Disable()
		chartBtnPlay.Disable()
		chartBtnStop.Disable()
		chartBtnRecord.Disable()

		if *recording {
			chartBtnExport.Enable()
		}

		// chart
		(chartBody).ChartUpdate(testType, testChartData)

		// update slider value
		chartSlider.chartData = testChartData
		chartSlider.rangeSlider.UpdateValues(0, float64(len(*testChartData)-1), 0, float64(len(*testChartData)-1))
		chartSlider.update()

		// slider visible
		chartSlider.sliderCard.Hidden = false
		chartSlider.sliderCard.Refresh()

		// slider btns
		chartSlider.chartUpdateBtn.OnTapped = func() {
			chartSlider.BuildSliderChartData()
			chartSlider.UpdateChartImage(testType, &chartBody)
		}
		chartSlider.chartResetBtn.OnTapped = func() {
			chartSlider.ResetChartImage(testType, &chartBody)
		}

		// update summary
		testSummary.EndTime = time.Now()
		testSummaryUI.UpdateUI_Ended(testSummary)

		// relaunch btn
		testSummaryUI.ntCmdBtn.Enable()
		testSummaryUI.ntCmdBtn.OnTapped = func() {
			_, iv, err := NtCmd2Iv(testSummary.ntCmd)
			if err != nil {
				logger.Println(err)
			}

			// launch new test
			switch testType {
			case "dns":
				go DnsAddPingRow(a, &ntGlobal.dnsIndex, &iv, ntGlobal.dnsTable, *recording, db, entryChan, errChan)
			case "http":
				go HttpAddPingRow(a, &ntGlobal.httpIndex, &iv, ntGlobal.httpTable, *recording, db, entryChan, errChan)
			case "tcp":
				go TcpAddPingRow(a, &ntGlobal.tcpIndex, &iv, ntGlobal.tcpTable, *recording, db, entryChan, errChan)
			case "icmp":
				go IcmpAddPingRow(a, &ntGlobal.icmpIndex, &iv, ntGlobal.icmpTable, *recording, db, entryChan, errChan)
			}
			testSummaryUI.ntCmdBtn.Disable()
		}
	}

	// chartBtnExport func - in Fyne, the btn.OnTapped function won't be run till the previous tappped function is finished. So no need to monitor the function progressing
	chartBtnExport.OnTapped = func() {

		// show progress bar
		chartBtnReadDBProgress.Hidden = false

		// initial dbTestEntries
		var dbTestEntries *[]ntdb.DbTestEntry
		var err error

		// create Input Var
		_, iv, err := NtCmd2Iv(testObj.GetSummary().ntCmd)
		if err != nil {
			errChan <- err
			return
		}

		// Read DB Test Entries
		dbTestEntries, err = ntdb.ReadTestTableEntries(db, fmt.Sprintf("%s_%s", testObj.GetType(), testObj.GetUUID()))
		if err != nil {
			errChan <- err
			return
		}

		// dialog for file save
		exportFileSaveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
			if err != nil {
				errChan <- err
				chartBtnReadDBProgress.Hidden = true
				return
			}

			if writer == nil {
				errChan <- fmt.Errorf("no file saved")
				chartBtnReadDBProgress.Hidden = true
				return
			}
			defer writer.Close()

			err = SaveToCSV(writer.URI().Path(), iv, dbTestEntries)
			if err != nil {
				errChan <- err
				chartBtnReadDBProgress.Hidden = true
				return
			}

			// end dialog, hide progress bar
			chartBtnReadDBProgress.Hidden = true

		}, newChartWindow)

		now := time.Now()
		record_csv_name := fmt.Sprintf("Record_%s_%s_%s.csv", testObj.GetType(), testObj.GetSummary().DestHost, now.Format("20060102150405"))
		exportFileSaveDialog.SetFileName(record_csv_name)
		exportFileSaveDialog.Resize(fyne.NewSize(750, 550))

		// get current <current_user>/Document/<app_name> path
		exportPath, err := GetDefaultExportFolder("nt_gui")
		if err != nil {
			errChan <- err
		}

		// set exportPath as the default path for Dialog
		exportPathURI, _ := storage.ListerForURI(storage.NewFileURI(exportPath))
		exportFileSaveDialog.SetLocation(exportPathURI)

		// export dialog show()
		exportFileSaveDialog.Show()

	}

	// New Chart Window Container
	chartContainerMainIn := container.New(layout.NewVBoxLayout(), testSummaryUI.summaryCard, chartBtnCard, chartBody.chartCard, chartSlider.sliderCard, ChartWindowCloseContainer)

	chartWindowSpaceHolder := widget.NewLabel("         ")
	chartContainerMainOut := container.New(layout.NewBorderLayout(chartWindowSpaceHolder, chartWindowSpaceHolder, chartWindowSpaceHolder, chartWindowSpaceHolder), chartWindowSpaceHolder, chartContainerMainIn)

	newChartWindow.SetContent(chartContainerMainOut)
	newChartWindow.Show()
}

// Func: Chart Update Go Routine
func NewChartUpdate(testCtx context.Context, pauseFlag *bool, testObj testObject, testSummaryUI *SummaryUI, testChart *Chart) {

	testType := testObj.GetType()
	testSummary := testObj.GetSummary()
	testChartData := testObj.GetChartData()

	for {
		// if pauseFlag is true, sleep for 1 sec and skip the current loop
		if *pauseFlag {
			time.Sleep(1 * time.Second)
			continue
		}

		// if the test is ended, exit
		if !existingTestCheck(&testRegister, testObj.GetUUID()) {
			return
		}

		select {
		// if the chart window closed, exit
		case <-testCtx.Done():
			return
		// if the ntGUI app closed, exit
		case <-appCtx.Done():
			return
		default:
			// update summary
			(*testSummaryUI).UpdateUI_Running(testSummary)
			// update chart
			(*testChart).ChartUpdate(testType, testChartData)

			time.Sleep(1 * time.Second)
		}
	}
}
