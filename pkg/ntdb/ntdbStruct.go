package ntdb

import "fmt"

// DB Entry Interface
type DbEntry interface {
	GetTableName() string
	GetTestType() string
}

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

// ** Record Table Entry: TCP Record Entry **
type RecordTCPEntry struct {
	Id           string
	TableName    string
	TestType     string
	Seq          int
	Status       string
	RTT          string
	SendDateTime string
	PacketRecv   int
	PacketLoss   int
	MinRTT       string
	MaxRTT       string
	AvgRTT       string
	AddInfo      string
}

func (r *RecordTCPEntry) GetTableName() string {
	return r.TableName
}

func (r *RecordTCPEntry) GetTestType() string {
	return r.TestType
}

// ** Record Table Entry: ICMP Record Entry **
type RecordICMPEntry struct {
	Id           string
	TableName    string
	TestType     string
	Seq          int
	Status       string
	RTT          string
	SendDateTime string
	PacketRecv   int
	PacketLoss   int
	MinRTT       string
	MaxRTT       string
	AvgRTT       string
	AddInfo      string
}

func (r *RecordICMPEntry) GetTableName() string {
	return r.TableName
}

func (r *RecordICMPEntry) GetTestType() string {
	return r.TestType
}
