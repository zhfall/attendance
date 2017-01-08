package main

import (
	"time"
)

// AttendanceRecord a record of attendence
type AttendanceRecord struct {
	AttendanceName   string
	AttendanceDate   time.Time
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
	AttendanceName string
	AttendanceDate time.Time
	OriginalRecord []time.Time
}

// AttendanceSummary is the attendence record for a person
type AttendanceSummary struct {
	StartDate              time.Time
	EndDate                time.Time
	AttendanceRecordMap    map[string][]AttendanceRecord
	UnPlannedAttendanceMap map[string][]UnPlannedAttendanceRecord
}
