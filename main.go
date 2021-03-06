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
var outputFile = flag.String("output", "", "output file")
var errorFile = flag.String("error", "", "error file")

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

	days := int(end.Add(1*time.Second).Sub(start).Hours() / 24)

	if *outputFile == "" {
		*outputFile = fmt.Sprintf("./output-%s.xlsx", start.Format("2006-01"))
	}

	if *errorFile == "" {
		*errorFile = fmt.Sprintf("./error-%s.xlsx", start.Format("2006-01"))
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
				for k := 4; k < 4+days*2; k = k + 2 {
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
				attendanceName, _ := row.Cells[10].String()
				if attendanceName == "" {
					continue
				}
				textDate, err := row.Cells[5].String()
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

	blueFill := *xlsx.NewFill("solid", "00FF00FF", "FF000000")
	redFill := *xlsx.NewFill("solid", "00FF0000", "0000FF00")
	// greenFill := *xlsx.NewFill("solid", "00000000", "00FF0000")

	var redStyle = xlsx.Style{
		Fill:      redFill,
		ApplyFill: true,
	}

	var blueStyle = xlsx.Style{
		Fill:      blueFill,
		ApplyFill: true,
	}

	// var greenStyle = xlsx.Style{
	// 	Fill:      greenFill,
	// 	ApplyFill: true,
	// }

	outputExcel := xlsx.NewFile()
	outSheet, err := outputExcel.AddSheet("Sheet1")
	if err != nil {
		fmt.Printf("Cannot create out excel : %v\n", err)
		os.Exit(1)
	}

	row := outSheet.AddRow()
	cell := row.AddCell()
	cell.Value = "姓名"
	for tmpDate := start; end.Sub(tmpDate) > 0; tmpDate = tmpDate.Add(24 * time.Hour) {
		cell = row.AddCell()
		cell.Value = tmpDate.Format("2006-01-02")
		cell = row.AddCell()
		cell.Value = tmpDate.Format("2006-01-02")
	}
	cell = row.AddCell()
	cell.Value = "迟到早退"
	cell = row.AddCell()
	cell.Value = "未打卡"

	for i, sheet := range planFile.Sheets {
		if i > 0 {
			break
		}
		var firstRow *xlsx.Row
		lastName := ""
		for j, row := range sheet.Rows {
			if j == 0 {
				firstRow = row
			}
			if j > 2 {
				attendanceName, _ := row.Cells[2].String()
				if attendanceName != "" {
					lateCount := 0
					missCount := 0
					outRow := outSheet.AddRow()
					outCell := outRow.AddCell()
					outCell.Value = attendanceName
					for k := 4; k < 4+days*2; k = k + 2 {
						tmpDate, err := firstRow.Cells[k].GetTime(false)
						if err != nil {
							fmt.Printf("Colume %d of first row is not a date!\n", k)
						}
						tmpDate = time.Date(tmpDate.Year(), tmpDate.Month(), tmpDate.Day(), 0, 0, 0, 0, loc)
						attendanceKey := AttendanceKey{
							AttendanceName: attendanceName,
							AttendanceDate: tmpDate,
						}

						_, err1 := row.Cells[k].GetTime(false)
						_, err2 := row.Cells[k+1].GetTime(false)

						if err1 != nil || err2 != nil {
							outCell = outRow.AddCell()
							outCell = outRow.AddCell()
							continue
						} else {
							attendanceRecord, err := attendances.Lookup(attendanceKey, attendanceName != lastName)
							if err != nil {
								outCell = outRow.AddCell()
								outCell.Value = "未打卡"
								outCell.SetStyle(&redStyle)
								missCount++
								outCell = outRow.AddCell()
								outCell.Value = "未打卡"
								outCell.SetStyle(&redStyle)
								missCount++
							} else {
								outCell = outRow.AddCell()
								if attendanceRecord.ActualStart.Year() < 1910 {
									outCell.Value = "未打卡"
									outCell.SetStyle(&redStyle)
									missCount++
								} else {
									outCell.Value = attendanceRecord.ActualStart.Format("15:04")
									if attendanceRecord.ActualStart.Sub(attendanceRecord.PlannedStart) > 0 {
										lateCount++
										outCell.SetStyle(&blueStyle)
									}
								}
								outCell = outRow.AddCell()
								if attendanceRecord.ActualEnd.Year() < 1910 {
									outCell.Value = "未打卡"
									outCell.SetStyle(&redStyle)
									missCount++
								} else {
									outCell.Value = attendanceRecord.ActualEnd.Format("15:04")
									if attendanceRecord.ActualEnd.Sub(attendanceRecord.PlannedEnd) < 0 {
										lateCount++
										outCell.SetStyle(&blueStyle)
									}
								}
							}
						}
					}
					lastName = attendanceName
					outCell = outRow.AddCell()
					if lateCount > 0 {
						outCell.SetInt(lateCount)
						outCell.SetStyle(&blueStyle)
					} else {
						outCell.Value = ""
					}
					outCell = outRow.AddCell()
					if missCount > 0 {
						outCell.SetInt(missCount)
						outCell.SetStyle(&redStyle)
					} else {
						outCell.Value = ""
					}
				} else {
					continue
				}
			}
		}
	}

	err = outputExcel.Save(*outputFile)
	if err != nil {
		fmt.Printf("Error - Cannot craete output Excel: %v", err)
	}

	lenUnPlanned := len(attendances.UnPlannedAttendanceMap)
	if lenUnPlanned > 0 {
		errorExcel := xlsx.NewFile()
		errorSheet, err := errorExcel.AddSheet("Sheet1")
		if err != nil {
			fmt.Printf("Cannot create out excel : %v\n", err)
			os.Exit(1)
		}

		row := errorSheet.AddRow()
		cell := row.AddCell()
		cell.Value = "姓名"
		cell = row.AddCell()
		cell.Value = "日期"

		for k := range attendances.UnPlannedAttendanceMap {
			row = errorSheet.AddRow()
			cell = row.AddCell()
			cell.Value = k.AttendanceName
			cell = row.AddCell()
			cell.Value = k.AttendanceDate.Format("2006-01-02")
		}

		err = errorExcel.Save(*errorFile)
		if err != nil {
			fmt.Printf("Error - Cannot craete output Excel: %v", err)
		}
	}

	attendances.Println(false)
}
