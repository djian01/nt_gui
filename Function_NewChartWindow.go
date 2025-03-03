package main

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/djian01/nt/pkg/ntPinger"
)

func NewChartWindow(a fyne.App, testObj testObject, recording bool, p *ntPinger.Pinger) {

	// Initial Vars
	testType := testObj.GetType()
	testSummary := testObj.GetSummary()
	testChart := testObj.GetChartData()

	// Initial New Chart Window
	newChartWindow := a.NewWindow(fmt.Sprintf("%s Chart", strings.ToUpper(testType)))
	newChartWindow.Resize(fyne.NewSize(1400, 900))
	newChartWindow.CenterOnScreen()

	// summary Card
	testSummaryUI := SummaryUI{}
	testSummaryUI.Initial()
	testSummaryUI.CreateCard()
	testSummaryUI.UpdateUI_Initial(testSummary)

	// Chart card
	chartBtnPause := widget.NewButtonWithIcon("Pause Chart Update", theme.MediaPauseIcon(), func() {})
	chartBtnPause.Importance = widget.WarningImportance
	chartBtnPauseContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(200, 30)), chartBtnPause)

	chartBtnPlay := widget.NewButtonWithIcon("Continue Chart Update", theme.MediaPlayIcon(), func() {})
	chartBtnPlay.Importance = widget.WarningImportance
	chartBtnPlayContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(200, 30)), chartBtnPlay)

	chartBtnRecord := widget.NewButtonWithIcon("Record Test", theme.MediaRecordIcon(), func() {})
	chartBtnRecord.Importance = widget.WarningImportance
	if recording { // if recording is enabled, disable the recording button
		chartBtnRecord.Disable()
	}
	chartBtnRecordContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(200, 30)), chartBtnRecord)

	chartBtnStop := widget.NewButtonWithIcon("Stop Test", theme.MediaStopIcon(), func() {})
	chartBtnStop.Importance = widget.HighImportance
	chartBtnStopContainer := container.New(layout.NewGridWrapLayout(fyne.NewSize(200, 30)), chartBtnStop)
	chartBtnStop.OnTapped = func() { // Stop on tap function
		testObj.Stop(p)
		testSummary.testEnded = true
		chartBtnPause.Disable()
		chartBtnPlay.Disable()
		chartBtnStop.Disable()
		chartBtnRecord.Disable()
	}

	if testSummary.testEnded { // if test is already stopped, disable the all buttons
		chartBtnPause.Disable()
		chartBtnPlay.Disable()
		chartBtnStop.Disable()
		chartBtnRecord.Disable()
	}

	chartBtnContainerIn := container.New(layout.NewHBoxLayout(), chartBtnPauseContainer, chartBtnPlayContainer, chartBtnRecordContainer, chartBtnStopContainer)
	chartBtnContainerOut := container.New(layout.NewCenterLayout(), chartBtnContainerIn)
	chartBtnCard := widget.NewCard("", "", chartBtnContainerOut)

	//// chart update pause flag
	//chartUpdatePause := false

	//// chart body
	chartBody := Chart{}
	chartBody.Initial()

	// Slider Card
	chartSlider := Slider{}
	chartSlider.Initial(0, 100, 0, 100)
	chartSlider.chartData = testChart
	chartSlider.CreateCard()
	//chartSlider.sliderCard.Hidden = true

	// close window Btn Container
	chartWindowCloseBtn := widget.NewButton("Close Window", func() {
		newChartWindow.Close()
	})
	ChartWindowCloseContainer := container.New(layout.NewCenterLayout(), container.New(layout.NewGridWrapLayout(fyne.NewSize(200, 40)), chartWindowCloseBtn))

	// New Chart Window Container
	chartContainerMainIn := container.New(layout.NewVBoxLayout(), testSummaryUI.summaryCard, chartBtnCard, chartBody.chartCard, chartSlider.sliderCard, ChartWindowCloseContainer)

	chartWindowSpaceHolder := widget.NewLabel("         ")
	chartContainerMainOut := container.New(layout.NewBorderLayout(chartWindowSpaceHolder, chartWindowSpaceHolder, chartWindowSpaceHolder, chartWindowSpaceHolder), chartWindowSpaceHolder, chartContainerMainIn)

	newChartWindow.SetContent(chartContainerMainOut)
	newChartWindow.Show()
}
