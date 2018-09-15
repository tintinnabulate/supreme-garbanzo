package main

import (
	"io/ioutil"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	jsonByteArray, err := ioutil.ReadFile("settings.json")
	check(err)
	settings := GetSettings(jsonByteArray)
	spreadsheet := FixCSV("bookings.csv", settings)
	WriteFixedCSV(spreadsheet)
}
