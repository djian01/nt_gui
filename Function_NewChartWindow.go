package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/djian01/nt/pkg/ntPinger"
	"github.com/djian01/nt_gui/pkg/ntdb"
)

func NewChartWindow(a fyne.App, testObj testObject, recording bool, p *ntPinger.Pinger, db *sql.DB, entryChan chan ntdb.DbEntry) {

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
		// call the cancel func to close the go routine
		testCancelFunc()
	})

	// summary Card
	testSummaryUI := SummaryUI{}
	testSummaryUI.Initial()
	testSummaryUI.CreateCard()
	testSummaryUI.UpdateUI_Initial(testSummary)

	// Chart card
	chartBtnPause := widget.NewButtonWithIcon("Pause Chart Update", theme.MediaPauseIcon(), func() {})
	chartBtnPause.Importance = widget.WarningImportance

	if !chartPauseFlag {
		chartBtnPause.Enable()
	} else {
		chartBtnPause.Disable()
	}

	chartBtnPauseContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(200, 30)), chartBtnPause)

	chartBtnPlay := widget.NewButtonWithIcon("Continue Chart Update", theme.MediaPlayIcon(), func() {})
	chartBtnPlay.Importance = widget.WarningImportance

	if chartPauseFlag {
		chartBtnPlay.Enable()
	} else {
		chartBtnPlay.Disable()
	}
	chartBtnPlayContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(200, 30)), chartBtnPlay)

	chartBtnRecord := widget.NewButtonWithIcon("Record Test", theme.MediaRecordIcon(), func() {})
	chartBtnRecord.Importance = widget.WarningImportance
	if recording { // if recording is enabled, disable the recording button
		chartBtnRecord.Disable()
	}
	chartBtnRecordContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(200, 30)), chartBtnRecord)

	chartBtnStop := widget.NewButtonWithIcon("Stop Test", theme.MediaStopIcon(), func() {})
	chartBtnStop.Importance = widget.DangerImportance
	chartBtnStopContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(200, 30)), chartBtnStop)

	chartBtnContainerIn := container.New(layout.NewHBoxLayout(), chartBtnPauseContainer, chartBtnPlayContainer, chartBtnRecordContainer, chartBtnStopContainer)
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
				go DnsAddPingRow(a, &ntGlobal.dnsIndex, &iv, ntGlobal.dnsTable, recording, db, entryChan)
			case "http":
			case "tcp":
			case "icmp":
			}
			testSummaryUI.ntCmdBtn.Disable()
		}
	} else {
		// kickoff the go routine for chart/summary update
		go NewChartUpdate(testCtx, &chartPauseFlag, &testObj, &testSummaryUI, &chartBody)
	}

	// update btn functions
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

	}

	//// Stop Btn Function
	chartBtnStop.OnTapped = func() { // Stop on tap function
		// if the test has not yet stopped, stop it
		if existingTestCheck(&testRegister, testObj.GetUUID()) {
			testObj.Stop(p)
		}

		chartBtnPause.Disable()
		chartBtnPlay.Disable()
		chartBtnStop.Disable()
		chartBtnRecord.Disable()

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
				go DnsAddPingRow(a, &ntGlobal.dnsIndex, &iv, ntGlobal.dnsTable, recording, db, entryChan)
			case "http":
			case "tcp":
			case "icmp":
			}
			testSummaryUI.ntCmdBtn.Disable()
		}
	}

	// New Chart Window Container
	chartContainerMainIn := container.New(layout.NewVBoxLayout(), testSummaryUI.summaryCard, chartBtnCard, chartBody.chartCard, chartSlider.sliderCard, ChartWindowCloseContainer)

	chartWindowSpaceHolder := widget.NewLabel("         ")
	chartContainerMainOut := container.New(layout.NewBorderLayout(chartWindowSpaceHolder, chartWindowSpaceHolder, chartWindowSpaceHolder, chartWindowSpaceHolder), chartWindowSpaceHolder, chartContainerMainIn)

	newChartWindow.SetContent(chartContainerMainOut)
	newChartWindow.Show()
}

// Func: Chart Update Go Routine
func NewChartUpdate(testCtx context.Context, pauseFlag *bool, testObj *testObject, testSummaryUI *SummaryUI, testChart *Chart) {

	testType := (*testObj).GetType()
	testSummary := (*testObj).GetSummary()
	testChartData := (*testObj).GetChartData()

	for {
		// if pauseFlag is true, sleep for 1 sec and skip the current loop
		if *pauseFlag {
			time.Sleep(1 * time.Second)
			continue
		}

		// if the test is ended, exit
		if !existingTestCheck(&testRegister, (*testObj).GetUUID()) {
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
