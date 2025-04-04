package ntdb

import (
	"fmt"
	"time"
)

// DB Entry Interface
type DbEntry interface {
	GetTableName() string
	GetTestType() string
	GetRtt() string
	GetSendTime() time.Time
	GetStatus() bool
	GetPacketSent() int
	GetFailRate() string
	GetMinRtt() string
	GetMaxRtt() string
	GetAvgRtt() string
	GetSuccessResponse() int
}

// check interafce implementation
var _ DbEntry = (*HistoryEntry)(nil)
var _ DbEntry = (*RecordDNSEntry)(nil)
var _ DbEntry = (*RecordHTTPEntry)(nil)
var _ DbEntry = (*RecordTCPEntry)(nil)
var _ DbEntry = (*RecordICMPEntry)(nil)

// ** DB Entry: History Entry **
type HistoryEntry struct {
	Id        string
	TableName string
	TestType  string
	StartTime string
	Command   string
	UUID      string
	Recorded  bool
}

func (h *HistoryEntry) GetSubRecordTableName() string {
	return fmt.Sprintf("recordTable-%s-%s", (*h).TestType, (*h).UUID)
}

func (h *HistoryEntry) GetTableName() string {
	return h.TableName
}

func (h *HistoryEntry) GetTestType() string {
	return h.TestType
}

func (r *HistoryEntry) GetRtt() string {
	return ""
}

func (r *HistoryEntry) GetSendTime() time.Time {
	layout := "2006-01-02 15:04:05 MST"
	t, _ := time.Parse(layout, r.StartTime)
	return t
}
func (r *HistoryEntry) GetStatus() bool {
	return false
}

func (r *HistoryEntry) GetPacketSent() int {
	return 0
}

func (r *HistoryEntry) GetSuccessResponse() int {
	return 0
}

func (r *HistoryEntry) GetFailRate() string {
	return ""
}

func (r *HistoryEntry) GetMinRtt() string {
	return ""
}

func (r *HistoryEntry) GetMaxRtt() string {
	return ""
}
func (r *HistoryEntry) GetAvgRtt() string {
	return ""
}

// ** Record Table Entry: DNS Record Entry **
type RecordDNSEntry struct {
	Id              string
	TableName       string
	TestType        string
	Seq             int
	Status          string
	DnsResponse     string
	DnsRecord       string
	ResponseTime    string
	SendDateTime    string
	SuccessResponse int
	FailRate        string
	MinRTT          string
	MaxRTT          string
	AvgRTT          string
	AddInfo         string
}

func (r *RecordDNSEntry) GetTableName() string {
	return r.TableName
}

func (r *RecordDNSEntry) GetTestType() string {
	return r.TestType
}

func (r *RecordDNSEntry) GetRtt() string {
	return r.ResponseTime
}
func (r *RecordDNSEntry) GetSendTime() time.Time {
	layout := "2006-01-02 15:04:05 MST"
	t, _ := time.Parse(layout, r.SendDateTime)
	return t
}
func (r *RecordDNSEntry) GetStatus() bool {
	if r.Status == "true" {
		return true
	} else {
		return false
	}
}

func (r *RecordDNSEntry) GetPacketSent() int {
	return (r.Seq + 1)
}

func (r *RecordDNSEntry) GetSuccessResponse() int {
	return r.SuccessResponse
}

func (r *RecordDNSEntry) GetFailRate() string {
	return r.FailRate
}

func (r *RecordDNSEntry) GetMinRtt() string {
	return r.MinRTT
}

func (r *RecordDNSEntry) GetMaxRtt() string {
	return r.MaxRTT
}
func (r *RecordDNSEntry) GetAvgRtt() string {
	return r.AvgRTT
}

// ** Record Table Entry: HTTP Record Entry **
type RecordHTTPEntry struct {
	Id              string
	TableName       string
	TestType        string
	Seq             int
	Status          string
	ResponseCode    string
	ResponsePhase   string
	ResponseTime    string
	SendDateTime    string
	SuccessResponse int
	FailRate        string
	MinRTT          string
	MaxRTT          string
	AvgRTT          string
	AddInfo         string
}

func (r *RecordHTTPEntry) GetTableName() string {
	return r.TableName
}

func (r *RecordHTTPEntry) GetTestType() string {
	return r.TestType
}

func (r *RecordHTTPEntry) GetRtt() string {
	return r.ResponseTime
}

func (r *RecordHTTPEntry) GetSendTime() time.Time {
	layout := "2006-01-02 15:04:05 MST"
	t, _ := time.Parse(layout, r.SendDateTime)
	return t
}
func (r *RecordHTTPEntry) GetStatus() bool {
	if r.Status == "true" {
		return true
	} else {
		return false
	}
}

func (r *RecordHTTPEntry) GetPacketSent() int {
	return (r.Seq + 1)
}

func (r *RecordHTTPEntry) GetSuccessResponse() int {
	return r.SuccessResponse
}

func (r *RecordHTTPEntry) GetFailRate() string {
	return r.FailRate
}

func (r *RecordHTTPEntry) GetMinRtt() string {
	return r.MinRTT
}

func (r *RecordHTTPEntry) GetMaxRtt() string {
	return r.MaxRTT
}
func (r *RecordHTTPEntry) GetAvgRtt() string {
	return r.AvgRTT
}

// ** Record Table Entry: TCP Record Entry **
type RecordTCPEntry struct {
	Id             string
	TableName      string
	TestType       string
	Seq            int
	Status         string
	RTT            string
	SendDateTime   string
	PacketRecv     int
	PacketLossRate string
	MinRTT         string
	MaxRTT         string
	AvgRTT         string
	AddInfo        string
}

func (r *RecordTCPEntry) GetTableName() string {
	return r.TableName
}

func (r *RecordTCPEntry) GetTestType() string {
	return r.TestType
}

func (r *RecordTCPEntry) GetRtt() string {
	return r.RTT
}

func (r *RecordTCPEntry) GetSendTime() time.Time {
	layout := "2006-01-02 15:04:05 MST"
	t, _ := time.Parse(layout, r.SendDateTime)
	return t
}
func (r *RecordTCPEntry) GetStatus() bool {
	if r.Status == "true" {
		return true
	} else {
		return false
	}
}

func (r *RecordTCPEntry) GetPacketSent() int {
	return (r.Seq + 1)
}

func (r *RecordTCPEntry) GetSuccessResponse() int {
	return r.PacketRecv
}

func (r *RecordTCPEntry) GetFailRate() string {
	return r.PacketLossRate
}

func (r *RecordTCPEntry) GetMinRtt() string {
	return r.MinRTT
}

func (r *RecordTCPEntry) GetMaxRtt() string {
	return r.MaxRTT
}
func (r *RecordTCPEntry) GetAvgRtt() string {
	return r.AvgRTT
}

// ** Record Table Entry: ICMP Record Entry **
type RecordICMPEntry struct {
	Id             string
	TableName      string
	TestType       string
	Seq            int
	Status         string
	RTT            string
	SendDateTime   string
	PacketRecv     int
	PacketLossRate string
	MinRTT         string
	MaxRTT         string
	AvgRTT         string
	AddInfo        string
}

func (r *RecordICMPEntry) GetTableName() string {
	return r.TableName
}

func (r *RecordICMPEntry) GetTestType() string {
	return r.TestType
}

func (r *RecordICMPEntry) GetRtt() string {
	return r.RTT
}

func (r *RecordICMPEntry) GetSendTime() time.Time {
	layout := "2006-01-02 15:04:05 MST"
	t, _ := time.Parse(layout, r.SendDateTime)
	return t
}
func (r *RecordICMPEntry) GetStatus() bool {
	if r.Status == "true" {
		return true
	} else {
		return false
	}
}

func (r *RecordICMPEntry) GetPacketSent() int {
	return (r.Seq + 1)
}

func (r *RecordICMPEntry) GetSuccessResponse() int {
	return r.PacketRecv
}

func (r *RecordICMPEntry) GetFailRate() string {
	return r.PacketLossRate
}

func (r *RecordICMPEntry) GetMinRtt() string {
	return r.MinRTT
}

func (r *RecordICMPEntry) GetMaxRtt() string {
	return r.MaxRTT
}
func (r *RecordICMPEntry) GetAvgRtt() string {
	return r.AvgRTT
}
