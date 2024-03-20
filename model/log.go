package model

import "time"

type LogBlock struct {
	From          time.Time    `json:"from"`
	To            time.Time    `json:"to"`
	TotalLogs     int          `json:"totalLogs"`
	ReturnedLogs  int          `json:"returnedLogs"`
	Namespace     string       `json:"namespace"`
	App           string       `json:"app"`
	LastTimestamp time.Time    `json:"lastTimestamp"`
	Content       []LogContent `json:"content"`
}

type LogContent struct {
	Timestamp time.Time `json:"timestamp"`
	Content   string    `json:"content"`
}
