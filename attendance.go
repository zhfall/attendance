package main

import (
	"time"
)

// AttendanceRecord a record of attendence
type AttendanceRecord struct {
	AttendenceName   string
	AttendenceDate   time.Time
	PlannedStart     time.Time
	PlannedEnd       time.Time
	ActualStart      time.Time
	ActualEnd        time.Time
	NeedMiddleRecord bool
	MiddleRecords    [2]time.Time
	OriginalRecords  []time.Time
}

// UnPlannedAttendanceRecord is the unplanned attendence record of a person in a day
type UnPlannedAttendanceRecord struct {
	AttendenceName string
	AttendenceDate time.Time
	OriginalRecord []time.Time
}

// AttendenceSummary is the attendence record for a person
type AttendenceSummary struct {
	AttendenceName          string
	StartDate               time.Time
	EndDate                 time.Time
	AttendanceRecordList    []AttendanceRecord
	UnPlannedAttendanceList []UnPlannedAttendanceRecord
}
