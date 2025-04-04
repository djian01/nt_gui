package ntdb

import (
	"fmt"
	"time"
)

// DB Entry Interface
type DbEntry interface {
	GetTableName() string
	GetTestType() string
}

// check interafce implementation for DB Entry
var _ DbEntry = (*HistoryEntry)(nil)
var _ DbEntry = (*RecordDNSEntry)(nil)
var _ DbEntry = (*RecordHTTPEntry)(nil)
var _ DbEntry = (*RecordTCPEntry)(nil)
var _ DbEntry = (*RecordICMPEntry)(nil)

// History Record Interface
type HistoryRecord interface {
	GetRtt() time.Duration
	GetSendTime() time.Time
	GetStatus() bool
	GetPacketSent() int
	GetFailRate() string
	GetMinRtt() string
	GetMaxRtt() string
	GetAvgRtt() string
	GetSuccessResponse() int
}

// check interafce implementation for DB Entry
var _ HistoryRecord = (*HistoryEntry)(nil)
var _ HistoryRecord = (*RecordDNSEntry)(nil)
var _ HistoryRecord = (*RecordHTTPEntry)(nil)
var _ HistoryRecord = (*RecordTCPEntry)(nil)
var _ HistoryRecord = (*RecordICMPEntry)(nil)

// ** DB Entry: History Entry **
type HistoryEntry struct {
	Id        string
	TableName string
	TestType  string
	StartTime time.Time
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

// ** Record Table Entry: DNS Record Entry **
type RecordDNSEntry struct {
	Id              string
	TableName       string
	TestType        string
	Seq             int
	Status          string
	DnsResponse     string
	DnsRecord       string
	ResponseTime    time.Duration
	SendDateTime    time.Time
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
	ResponseTime    time.Duration
	SendDateTime    time.Time
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
	RTT            time.Duration
	SendDateTime   time.Time
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

func (r *RecordTCPEntry) GetRtt() time.Duration {
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
	RTT            time.Duration
	SendDateTime   time.Time
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
