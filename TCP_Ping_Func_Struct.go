package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// check interafce implementation
// var _ testGUIRow = (*tcpGUIRow)(nil)
// var _ testObject = (*tcpObject)(nil)

// ******* struct tcpGUIRow ********

type tcpGUIRow struct {
	Index     pingCell
	Seq       pingCell
	HostName  pingCell
	IP        pingCell
	Port      pingCell
	Payload   pingCell
	RTT       pingCell
	TimeStamp pingCell // sendDateTime
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
