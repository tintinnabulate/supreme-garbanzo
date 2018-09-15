package main

import (
	"io/ioutil"

	"github.com/tintinnabulate/supreme-garbanzo/generators"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	generators.Run()
	jsonByteArray, err := ioutil.ReadFile("settings.json")
	check(err)
	settings := GetSettings(jsonByteArray)
	spreadsheet := FixCSV("bookings.csv", settings)
	WriteFixedCSV(spreadsheet)
}
