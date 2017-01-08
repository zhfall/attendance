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

func main() {
	flag.Parse()
	if *startDate == "" || *endDate == "" || *planFile == "" || *actualFile == "" {
		fmt.Printf("Usage: %s -start $StartDate -end $EndDate -plan $planFile -actual $actualFile [-output $outputFilePath] [-error $errorFilePath]\n", os.Args[0])
		os.Exit(1)
	}

	loc, _ := time.LoadLocation("Asia/Shanghai")
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
		AttendanceRecordMap:    make(map[string][]AttendanceRecord),
		UnPlannedAttendanceMap: make(map[string][]UnPlannedAttendanceRecord),
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
				var attendanceRecord AttendanceRecord
				for k := 4; k < len(row.Cells); k = k + 2 {
					attendanceRecord = AttendanceRecord{}
					attendanceName, _ := row.Cells[2].String()
					if attendanceName != "" {
						attendanceRecord.AttendanceName = attendanceName
						tmpDate, err := firstRow.Cells[k].GetTime(false)
						if err != nil {
							fmt.Printf("Colume %d of first row is not a date!\n", k)
						}
						tmpDate = time.Date(tmpDate.Year(), tmpDate.Month(), tmpDate.Day(), 0, 0, 0, 0, loc)
						attendanceRecord.AttendanceDate = tmpDate
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
						attendances.AttendanceRecordMap[attendanceName] = append(attendances.AttendanceRecordMap[attendanceName], attendanceRecord)
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
		for _, row := range sheet.Rows {
			for _, cell := range row.Cells {
				text, _ := cell.String()
				fmt.Printf("%s\n", text)
			}
		}
	}
}
