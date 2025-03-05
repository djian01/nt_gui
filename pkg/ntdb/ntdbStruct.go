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
	DateTime  string
	Command   string
	UUID      string
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
	Status          string
	DnsResponse     string
	DnsRecord       string
	ResponseTime    string
	SendTime        string
	SuccessResponse string
	MinRTT          string
	MaxRTT          string
	AvgRtt          string
	AddInfo         string
}

func (r *RecordDNSEntry) GetTableName() string {
	return r.TableName
}

func (r *RecordDNSEntry) GetTestType() string {
	return r.TestType
}
