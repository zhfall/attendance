package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/tealeg/xlsx"
)

var startDate = flag.String("start", "", "start date")
var endDate = flag.String("end", "", "end date")
var planFile = flag.String("plan", "", "plan file")
var actualFile = flag.String("actual", "", "acutal file")
var outputFile = flag.String("output", fmt.Sprintf("./output/output-%s-%s", *startDate, *endDate), "output file")
var errorFile = flag.String("error", fmt.Sprintf("./output/error-%s-%s", *startDate, *endDate), "error file")

var attendances AttendanceSummary
var loc *time.Location

func main() {
	flag.Parse()
	if *startDate == "" || *endDate == "" || *planFile == "" || *actualFile == "" {
		fmt.Printf("Usage: %s -start $StartDate -end $EndDate -plan $planFile -actual $actualFile [-output $outputFilePath] [-error $errorFilePath]\n", os.Args[0])
		os.Exit(1)
	}

	loc, _ = time.LoadLocation("Asia/Shanghai")

	start, err := time.ParseInLocation("2006-01-02", *startDate, loc)
	if err != nil {
		fmt.Println("state date is invalid!")
		os.Exit(1)
	}

	end, err := time.ParseInLocation("2006-01-02", *endDate, loc)
	if err != nil {
		fmt.Println("end date is invalid!")
		os.Exit(1)
	} else {
		end = end.Add((24*60*60 - 1) * time.Second)
	}

	if end.Sub(start) >= 31*24*time.Hour {
		fmt.Println("We don't support duration more than 1 month!")
		os.Exit(1)
	}

	planFile, err := xlsx.OpenFile(*planFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	actualFile, err := xlsx.OpenFile(*actualFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	attendances = AttendanceSummary{
		StartDate:              start,
		EndDate:                end,
		AttendanceRecordMap:    make(map[AttendanceKey][]*AttendanceRecord),
		UnPlannedAttendanceMap: make(map[AttendanceKey]*UnPlannedAttendanceRecord),
	}

	for i, sheet := range planFile.Sheets {
		if i > 0 {
			break
		}
		var firstRow *xlsx.Row
		for j, row := range sheet.Rows {
			if j == 0 {
				firstRow = row
			}
			if j > 2 {
				for k := 4; k < len(row.Cells); k = k + 2 {
					attendanceRecord := NewAttendanceRecord()
					attendanceName, _ := row.Cells[2].String()
					if attendanceName != "" {
						attendanceRecord.AttendanceName = attendanceName
						tmpDate, err := firstRow.Cells[k].GetTime(false)
						if err != nil {
							fmt.Printf("Colume %d of first row is not a date!\n", k)
						}
						tmpDate = time.Date(tmpDate.Year(), tmpDate.Month(), tmpDate.Day(), 0, 0, 0, 0, loc)
						attendanceRecord.AttendanceDate = tmpDate
						attendanceKey := AttendanceKey{
							AttendanceName: attendanceName,
							AttendanceDate: tmpDate,
						}

						timeOn, err := row.Cells[k].GetTime(false)
						if err != nil {
							continue
						}
						timeOff, err := row.Cells[k+1].GetTime(false)
						if err != nil {
							continue
						}
						// fmt.Printf("%v-%s", timeOn, timeOff)
						attendanceRecord.PlannedStart = time.Date(tmpDate.Year(), tmpDate.Month(), tmpDate.Day(), timeOn.Hour(), timeOn.Minute(), 0, 0, loc)
						attendanceRecord.PlannedEnd = time.Date(tmpDate.Year(), tmpDate.Month(), tmpDate.Day(), timeOff.Hour(), timeOff.Minute(), 0, 0, loc)
						// fmt.Println(attendanceRecord)
						if timeOff.Sub(timeOn) > 8*time.Hour {
							attendanceRecord.NeedMiddleRecord = true
						}

						attendances.AttendanceRecordMap[attendanceKey] = append(attendances.AttendanceRecordMap[attendanceKey], &attendanceRecord)
					} else {
						continue
					}

				}
			}
		}
	}

	for i, sheet := range actualFile.Sheets {
		if i > 0 {
			break
		}
		for j, row := range sheet.Rows {
			if j > 0 {
				attendanceName, _ := row.Cells[2].String()
				if attendanceName == "" {
					continue
				}
				textDate, err := row.Cells[7].String()
				if err != nil {
					fmt.Printf("Error in actual date, row number: %d, err: %v\n", j, err)
					fmt.Println("Row: ", row.Cells)
				}
				tmpDate, err := time.ParseInLocation("2006-01-02 15:04:05", textDate, loc)
				if err != nil {
					fmt.Printf("Error in actual date, row number: %d, err: %v\n", j, err)
					fmt.Println("Row: ", row.Cells)
				}
				attendances.AddAttendanceRecord(attendanceName, tmpDate)
			}
		}
	}

	count := 0
	for key := range attendances.AttendanceRecordMap {
		count++
		// fmt.Println(key)
		attendanceRecord, ok := attendances.AttendanceRecordMap[key]
		if ok {
			for _, attendance := range attendanceRecord {
				fmt.Println(key, attendance)
			}
		}
	}
	println(count)

	lenUnPlanned := len(attendances.UnPlannedAttendanceMap)
	if lenUnPlanned > 0 {
		fmt.Printf("Warning: Some record(%d) are not found in plans!\n", lenUnPlanned)
		for k, v := range attendances.UnPlannedAttendanceMap {
			fmt.Println(k, v)
			// _ = fmt.Sprintln(k)
		}
	}
}
