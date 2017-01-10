package main

import (
	"errors"
	"fmt"
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

// NewAttendanceRecord new a AttendanceRecord
func NewAttendanceRecord() AttendanceRecord {
	return AttendanceRecord{}
}

// UnPlannedAttendanceRecord is the unplanned attendence record of a person in a day
type UnPlannedAttendanceRecord struct {
	AttendanceName  string
	AttendanceDate  time.Time
	OriginalRecords []time.Time
}

// AttendanceKey is the key of Attendance Record Map
type AttendanceKey struct {
	AttendanceName string
	AttendanceDate time.Time
}

// AttendanceSummary is the attendence record for a person
type AttendanceSummary struct {
	StartDate              time.Time
	EndDate                time.Time
	AttendanceRecordMap    map[AttendanceKey][]*AttendanceRecord
	UnPlannedAttendanceMap map[AttendanceKey]*UnPlannedAttendanceRecord
}

// Loukup look up an attenance record
func (attendances *AttendanceSummary) Lookup(key AttendanceKey, first bool) (*AttendanceRecord, error) {
	attendanceRecordList, ok := attendances.AttendanceRecordMap[key]
	if ok {
		if first || len(attendanceRecordList) == 1 {
			return attendanceRecordList[0], nil
		}
		return attendanceRecordList[1], nil
	} else {
		return &AttendanceRecord{}, errors.New("Not Found")
	}
}

// AddAttendanceRecord add Attendance Record
func (attendances *AttendanceSummary) AddAttendanceRecord(attendanceName string, checkTime time.Time) {
	checkDate := time.Date(checkTime.Year(), checkTime.Month(), checkTime.Day(), 0, 0, 0, 0, checkTime.Location())
	// fmt.Println(checkTime, checkDate)
	attendanceKey := AttendanceKey{
		AttendanceName: attendanceName,
		AttendanceDate: checkDate,
	}
	attendanceRecordList, ok := attendances.AttendanceRecordMap[attendanceKey]
	if ok {
		if len(attendanceRecordList) == 1 {
			attendanceRecord := attendanceRecordList[0]
			attendanceRecord.OriginalRecords = append(attendanceRecord.OriginalRecords, checkTime)
			if checkTime.Sub(attendanceRecord.PlannedStart) < attendanceRecord.PlannedEnd.Sub(attendanceRecord.PlannedStart)/2 {
				if attendanceRecord.ActualStart.Year() < 1910 {
					attendanceRecord.ActualStart = checkTime
				} else {
					if checkTime.Sub(attendanceRecord.ActualStart) < 0 {
						attendanceRecord.ActualStart = checkTime
					}
				}
			} else {
				if attendanceRecord.ActualEnd.Year() < 1910 {
					attendanceRecord.ActualEnd = checkTime
				} else {
					if checkTime.Sub(attendanceRecord.ActualEnd) > 0 {
						attendanceRecord.ActualEnd = checkTime
					}
				}
			}
		} else {
			if len(attendanceRecordList) == 2 {
				var firstAttendance, secondAttendance, attendanceRecord *AttendanceRecord
				if attendanceRecordList[0].PlannedStart.Sub(attendanceRecordList[1].PlannedStart) < 0 {
					firstAttendance = attendanceRecordList[0]
					secondAttendance = attendanceRecordList[1]
				} else {
					firstAttendance = attendanceRecordList[1]
					secondAttendance = attendanceRecordList[0]
				}
				if checkTime.Sub(firstAttendance.PlannedEnd) < secondAttendance.PlannedStart.Sub(firstAttendance.PlannedEnd)/2 {
					attendanceRecord = firstAttendance
				} else {
					attendanceRecord = secondAttendance
				}
				attendanceRecord.OriginalRecords = append(attendanceRecord.OriginalRecords, checkTime)
				if checkTime.Sub(attendanceRecord.PlannedStart) < attendanceRecord.PlannedEnd.Sub(attendanceRecord.PlannedStart)/2 {
					if attendanceRecord.ActualStart.Year() < 1910 {
						attendanceRecord.ActualStart = checkTime
					} else {
						if checkTime.Sub(attendanceRecord.ActualStart) < 0 {
							attendanceRecord.ActualStart = checkTime
						}
					}
				} else {
					if attendanceRecord.ActualEnd.Year() < 1910 {
						attendanceRecord.ActualEnd = checkTime
					} else {
						if checkTime.Sub(attendanceRecord.ActualEnd) > 0 {
							attendanceRecord.ActualEnd = checkTime
						}
					}
				}
			}
		}
	} else {
		var unPlannedAttendanceRecord *UnPlannedAttendanceRecord
		unPlannedAttendanceRecord, ok := attendances.UnPlannedAttendanceMap[attendanceKey]
		if !ok {
			unPlannedAttendanceRecord = &UnPlannedAttendanceRecord{
				AttendanceName: attendanceName,
				AttendanceDate: checkDate,
			}
			attendances.UnPlannedAttendanceMap[attendanceKey] = unPlannedAttendanceRecord
		}
		unPlannedAttendanceRecord.OriginalRecords = append(unPlannedAttendanceRecord.OriginalRecords, checkTime)
	}
}

// Println print the result
func (attendances *AttendanceSummary) Println(verbose bool) {
	count := 0
	for key := range attendances.AttendanceRecordMap {
		count++
		// fmt.Println(key)
		attendanceRecord, ok := attendances.AttendanceRecordMap[key]
		if ok {
			for _, attendance := range attendanceRecord {
				if verbose {
					fmt.Println(key, attendance)
				}
			}
		}
	}
	fmt.Println("Total Attendances Records", count)

	lenUnPlanned := len(attendances.UnPlannedAttendanceMap)
	if lenUnPlanned > 0 {
		fmt.Printf("Warning: Some record(%d) are not found in plans!\n", lenUnPlanned)
		for k, v := range attendances.UnPlannedAttendanceMap {
			if verbose {
				fmt.Println(k, v)
			}
		}
	}
}
