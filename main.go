package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tealeg/xlsx"
)

var startDate = flag.String("start", "", "start date")
var endDate = flag.String("end", "", "end date")
var planFile = flag.String("plan", "", "plan file")
var actualFile = flag.String("actual", "", "acutal file")
var outputFile = flag.String("output", fmt.Sprintf("./output/output-%s-%s", *startDate, *endDate), "output file")
var errorFile = flag.String("error", fmt.Sprintf("./output/error-%s-%s", *startDate, *endDate), "error file")

func main() {
	flag.Parse()
	if *startDate == "" || *endDate == "" || *planFile == "" || *actualFile == "" {
		fmt.Printf("Usage: %s -start $StartDate -end $EndDate -plan $planFile -actual $actualFile [-output $outputFilePath] [-error $errorFilePath]\n", os.Args[0])
		os.Exit(1)
	}

	planFile, err := xlsx.OpenFile(*planFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, sheet := range planFile.Sheets {
		for _, row := range sheet.Rows {
			for _, cell := range row.Cells {
				text, _ := cell.String()
				fmt.Printf("%s\n", text)
			}
		}
	}

	actualFile, err := xlsx.OpenFile(*actualFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, sheet := range actualFile.Sheets {
		for _, row := range sheet.Rows {
			for _, cell := range row.Cells {
				text, _ := cell.String()
				fmt.Printf("%s\n", text)
			}
		}
	}
}
