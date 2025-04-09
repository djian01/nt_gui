package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// check interafce implementation
// var _ testGUIRow = (*icmpGUIRow)(nil)
// var _ testObject = (*icmpObject)(nil)

// ******* struct icmpGUIRow ********

type icmpGUIRow struct {
	Index     pingCell
	Seq       pingCell
	Status    pingCell
	HostName  pingCell
	IP        pingCell
	Payload   pingCell
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
