package ntdb

type HistoryEntry struct {
	Id        string
	Type      string
	Date      string
	Time      string
	Command   string
	TableName string
}

func (h *HistoryEntry) GetTableName() string {
	return (*h).TableName
}

type Entry interface {
	GetTableName() string
}
