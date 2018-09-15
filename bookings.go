package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

// LOCATION is the timezone the bookings are made in, for use in creating and comparing dates & times
var LOCATION = time.UTC

/*
 * { "properties": [
 *   { "long_name" : "FooBarBaz",
 *     "short_name" : "FB",
 *     "calendar" : "calendar@mycalendar.com",
 *     "commission" : 0.1,
 *     "booking_commission" : 0.0,
 *     "house_owner_commission" : 0.1,
 *     "greeting" : 15,
 *     "laundry" : [10,10,15,15,25,25],
 *     "cleaning" : 35,
 *     "consumables" : [15,15,25,25,35,35]
 *   },
 *   { "long_name" : "WibbleWobbleWoo",
 *     "short_name" : "WW",
 *     "calendar" : "calendar2@mycalendar.com",
 *     "commission" : 0.2,
 *     "booking_commission" : 0.1,
 *     "house_owner_commission" : 0.3,
 *     "greeting" : 25,
 *     "laundry" : [15,15,20,20,35,35],
 *     "cleaning" : 35,
 *     "consumables" : [15,15,25,25,35,35] }
 * ]}
 */

// Property is used to unmarshal a JSON file of the above form
type Property struct {
	LongName             string    `json:"long_name"`
	ShortName            string    `json:"short_name"`
	Calendar             string    `json:"calendar"`
	Commission           float64   `json:"commission"`
	BookingCommission    float64   `json:"booking_commission"`
	HouseOwnerCommission float64   `json:"house_owner_commission"`
	Greeting             float64   `json:"greeting"`
	Laundry              []float64 `json:"laundry"`
	Cleaning             float64   `json:"cleaning"`
	Consumables          []float64 `json:"consumables"`
}

// Settings holds the settings for each property
type Settings struct {
	Properties []Property `json:"properties"`
}

// Source is an Enum
type Source int

// BookingCom and others are sources the booking originated from
const (
	BookingCom Source = 1 + iota
	AirBnb
	Email
	Phone
	VisitBath
	Other
	nSources = int(Other)
)

type Month int

const (
	Jan Month = 1 + iota
	Feb
	Mar
	Apr
	May
	Jun
	Jul
	Aug
	Sep
	Oct
	Nov
	Dec
	nMonths = int(Dec)
)

var abbrevMonths = map[string]Month{
	"JAN": Jan,
	"FEB": Feb,
	"MAR": Mar,
	"APR": Apr,
	"MAY": May,
	"JUN": Jun,
	"JUL": Jul,
	"AUG": Aug,
	"SEP": Sep,
	"OCT": Oct,
	"NOV": Nov,
	"DEC": Dec,
}

func (m Month) String() string {
	foo := make(map[Month]string)
	for k, v := range abbrevMonths {
		foo[v] = k
	}
	return foo[m]
}

// FormInput holds all the form Inputs from the user
type FormInput struct {
	BookingRef     string
	FirstName      string
	LastName       string
	Email          string
	Mobile         string
	Notes          string
	Source         Source
	NumberOfPeople int
	Gross          float64
	IsGreeting     bool
	IsLaundry      bool
	IsCleaning     bool
	IsConsumables  bool
	BookingDate    time.Time
}

// Booking holds a booking
type Booking struct {
	Form          FormInput
	Property      Property
	Arrival       time.Time
	Departure     time.Time
	BookingDate   time.Time
	BookingFee    float64
	HouseOwnerFee float64
	Net           float64
	TotalFees     float64
	OwnerIncome   float64
}

// SpreadsheetRow holds a spreadsheet row
type SpreadsheetRow struct {
	BookingRef       string
	PropertyLongName string
	FirstName        string
	LastName         string
	Email            string
	Mobile           string
	Notes            string
	BookingDate      time.Time
	Source           Source
	Arrival          time.Time
	Departure        time.Time
	NumberOfPeople   int
	Gross            float64
	Net              float64
	IsDiscount       bool
	Commission       float64
	DueDate          time.Time
	IsCommission     bool
	Greeting         float64
	Laundry          float64
	Cleaning         float64
	Consumables      float64
	BookingFee       float64
	HouseOwnerFee    float64
	TotalFees        float64
	OwnerIncome      float64
}

// Spreadsheet holds a whole spreadsheet
type Spreadsheet struct {
	Rows []SpreadsheetRow
}

// GetSettings is used to unmarshal a JSON settings file into a Settings struct
func GetSettings(jsonByteArray []byte) Settings {
	var settings Settings
	json.Unmarshal(jsonByteArray, &settings)
	return settings
}

// Datetime is a utility function for making dates with the same
// location, 0 hours, 0 mins, 0 secs, 0 nanosecs.
func Datetime(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, LOCATION)
}

// Now returns the time right now when it is called
func Now() time.Time {
	return time.Now().In(LOCATION)
}

// BusinessOpeningDate returns the date the business was opened
var BusinessOpeningDate = Datetime(2011, time.January, 1)

// BusinessCurrentDate is the current date
var BusinessCurrentDate = Now()

func yearsAfter(years int, fromDate time.Time) time.Time {
	return fromDate.AddDate(years, 0, 0)
}

func getSliceEnd(bookingRef string) int {
	var sliceEnd int
	for i := 0; i < len(bookingRef); i++ {
		_, err := strconv.Atoi(string(bookingRef[i]))
		if err != nil {
			break
		} else {
			sliceEnd++
		}
	}
	return sliceEnd
}

func getBookingYear(sliceEnd int, bookingRef string) int {
	number, _ := strconv.Atoi(bookingRef[:sliceEnd])
	return yearsAfter(number, BusinessOpeningDate).Year()
}

func getBookingProperty(sliceEnd int, bookingRef string, props []Property) Property {
	var property Property
	for p := range props {
		property = props[p]
		if props[p].ShortName == bookingRef[sliceEnd:sliceEnd+2] {
			break
		}
	}
	return property
}

func getBookingArrivalDate(sliceEnd int, bookingYear int, bookingRef string) time.Time {
	monthString := bookingRef[sliceEnd+2 : sliceEnd+5]
	day, _ := strconv.Atoi(bookingRef[sliceEnd+5 : sliceEnd+7])
	if month, ok := abbrevMonths[monthString]; ok {
		return Datetime(bookingYear, time.Month(month), day)
	}
	log.Println("unknown month in booking reference:", monthString)
	return Datetime(0, 0, 0)
}

func getBookingDepartureDate(sliceEnd int, arrivalDate time.Time, bookingRef string) time.Time {
	dayString := bookingRef[sliceEnd+7 : sliceEnd+9]
	day, _ := strconv.Atoi(dayString)
	month := arrivalDate.Month()
	if day <= arrivalDate.Day() {
		/* MAGIC: if `month +=1` is 13, time.Date handles this by rolling forward to next year */
		month++
	}
	return Datetime(arrivalDate.Year(), month, day)
}

func createBooking(f FormInput, settings Settings) Booking {
	sliceEnd := getSliceEnd(f.BookingRef)
	year := getBookingYear(sliceEnd, f.BookingRef)
	arrival := getBookingArrivalDate(sliceEnd, year, f.BookingRef)
	property := getBookingProperty(sliceEnd, f.BookingRef, settings.Properties)
	bookingCommission := property.BookingCommission
	if f.Source == BookingCom {
		bookingCommission = 0.15
	}
	bookingFee := bookingCommission * f.Gross
	net := f.Gross - bookingFee
	houseOwnerFee := property.HouseOwnerCommission * net
	houseOwnerFee = math.Max(35, houseOwnerFee)
	totalFees := getServicesCost(property, f) + houseOwnerFee
	return Booking{
		Form:          f,
		Property:      property,
		Arrival:       arrival,
		Departure:     getBookingDepartureDate(sliceEnd, arrival, f.BookingRef),
		BookingDate:   f.BookingDate,
		HouseOwnerFee: houseOwnerFee,
		BookingFee:    bookingFee,
		Net:           net,
		TotalFees:     totalFees,
		OwnerIncome:   net - totalFees,
	}
}

func createBookingRef(b Booking) string {
	foo := make(map[time.Month]Month)
	for _, v := range abbrevMonths {
		foo[time.Month(v)] = Month(v)
	}
	var returnString string
	returnString += fmt.Sprintf("%d", b.BookingDate.Year()-BusinessOpeningDate.Year())
	returnString += b.Property.ShortName
	returnString += fmt.Sprintf("%s", foo[b.Arrival.Month()])
	returnString += fmt.Sprintf("%.2d", b.Arrival.Day())
	returnString += fmt.Sprintf("%.2d", b.Departure.Day())
	return returnString
}

func getServicesCost(property Property, f FormInput) float64 {
	servicesCost := 0.0
	ppl := int(math.Min(6, float64(f.NumberOfPeople)))
	if f.IsConsumables {
		servicesCost += property.Consumables[ppl-1]
	}
	if f.IsLaundry {
		servicesCost += property.Laundry[ppl-1]
	}
	if f.IsGreeting {
		servicesCost += property.Greeting
	}
	if f.IsCleaning {
		servicesCost += property.Cleaning
	}
	return servicesCost
}

func getBookingSpreadsheetRow(f FormInput, settings Settings) SpreadsheetRow {
	b := createBooking(f, settings)
	ppl := int(math.Min(6, float64(f.NumberOfPeople)))
	return SpreadsheetRow{
		BookingRef:       f.BookingRef,
		PropertyLongName: b.Property.LongName,
		FirstName:        f.FirstName,
		LastName:         f.LastName,
		Email:            f.Email,
		Mobile:           f.Mobile,
		Notes:            f.Notes,
		BookingDate:      f.BookingDate,
		Source:           f.Source,
		Arrival:          b.Arrival,
		Departure:        b.Departure,
		NumberOfPeople:   f.NumberOfPeople,
		Gross:            f.Gross,
		Net:              b.Net,
		IsDiscount:       false,
		Commission:       b.Property.Commission,
		DueDate:          f.BookingDate,
		IsCommission:     true,
		Greeting:         b.Property.Greeting,
		Laundry:          b.Property.Laundry[ppl-1],
		Cleaning:         b.Property.Cleaning,
		Consumables:      b.Property.Consumables[ppl-1],
		BookingFee:       b.BookingFee,
		HouseOwnerFee:    b.HouseOwnerFee,
		TotalFees:        b.TotalFees,
		OwnerIncome:      b.OwnerIncome,
	}
}

// FixSpreadsheetRow feeds a bad spreadsheet row back into the calculation
// to derive correct values based on settings
func FixSpreadsheetRow(bad SpreadsheetRow, settings Settings) SpreadsheetRow {
	f := FormInput{
		BookingRef:     bad.BookingRef,
		FirstName:      bad.FirstName,
		LastName:       bad.LastName,
		Email:          bad.Email,
		Mobile:         bad.Mobile,
		Notes:          bad.Notes,
		Source:         bad.Source,
		NumberOfPeople: bad.NumberOfPeople,
		BookingDate:    bad.BookingDate,
		Gross:          bad.Gross,
		IsGreeting:     true,
		IsCleaning:     true,
		IsLaundry:      true,
		IsConsumables:  true,
	}
	return getBookingSpreadsheetRow(f, settings)
}

// ParseCSV reads csv into FormInput manually.
func ParseCSV(file string) ([]FormInput, error) {
	var forms []FormInput
	sources := [6]string{"booking.com", "airbnb", "email", "phone", "visit", "other"}
	f, err := os.Open(file)
	if err != nil {
		return forms, err
	}
	defer f.Close()
	csvr := csv.NewReader(f)
	for {
		row, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return forms, err
		}
		var f FormInput
		f.BookingRef = row[0]
		f.FirstName = row[2]
		f.LastName = row[3]
		f.Email = row[4]
		f.Mobile = row[5]
		f.Notes = row[6]
		var sc = row[8]
		var source = BookingCom
		for x := BookingCom; x <= Other; x++ {
			source = x
			if sources[x-1] == sc {
				break
			}
		}
		f.Source = source
		f.NumberOfPeople, _ = strconv.Atoi(row[11])
		date := strings.Split(row[7], "-")
		year, _ := strconv.Atoi(date[0])
		month, _ := strconv.Atoi(date[1])
		day, _ := strconv.Atoi(date[2])
		f.BookingDate = Datetime(year, time.Month(month), day)
		f.Gross, _ = strconv.ParseFloat(row[12], 64)
		f.IsGreeting = true
		f.IsLaundry = true
		f.IsCleaning = true
		f.IsConsumables = true
		forms = append(forms, f)
	}
}

// FixCSV fixes a CSV!
func FixCSV(file string, settings Settings) Spreadsheet {
	var rows []SpreadsheetRow
	lines, _ := ParseCSV(file)
	for i := 0; i < len(lines); i++ {
		rows = append(rows,
			FixSpreadsheetRow(
				// first we derive the data using the correct calculations,
				getBookingSpreadsheetRow(lines[i], settings),
				// then we do any fixing in FixSpreadsheetRow as necessary,
				// e.g. using different settings.
				// this currently just uses the same settings
				settings))
	}
	return Spreadsheet{Rows: rows}
}

// WriteFixedCSV writes fixed CSV to a new CSV
func WriteFixedCSV(spreadsheet Spreadsheet) {
	var sources = [6]string{"booking.com", "airbnb", "email", "phone", "visit", "other"}
	file, _ := os.Create("out.csv")
	defer file.Close()
	w := csv.NewWriter(file)
	defer w.Flush()
	var header = []string{"booking_ref", "property", "first_name", "last_name", "email",
		"mobile", "notes", "booking_date", "source", "arrival_date", "departure_date",
		"number_of_people", "gross", "net", "is_discount", "commission", "due_date",
		"is_commission", "greeting", "laundry", "cleaning", "consumables", "booking_fee",
		"house_owner_fee", "total_fees", "owner_income"}
	w.Write(header)
	s := spreadsheet.Rows
	for i := 0; i < len(s); i++ {
		var row []string
		row = append(row, s[i].BookingRef)
		row = append(row, s[i].PropertyLongName)
		row = append(row, s[i].FirstName)
		row = append(row, s[i].LastName)
		row = append(row, s[i].Email)
		row = append(row, s[i].Mobile)
		row = append(row, s[i].Notes)
		row = append(row, s[i].BookingDate.Format("2006-01-02"))
		for x := BookingCom; x <= Other; x++ {
			if s[i].Source == x {
				row = append(row, sources[x-1])
			}
		}
		row = append(row, s[i].Arrival.Format("2006-01-02"))
		row = append(row, s[i].Departure.Format("2006-01-02"))
		row = append(row, fmt.Sprintf("%d", s[i].NumberOfPeople))
		row = append(row, fmt.Sprintf("%.2f", s[i].Gross))
		row = append(row, fmt.Sprintf("%.2f", s[i].Net))
		row = append(row, "FALSE")
		row = append(row, fmt.Sprintf("%.3f", s[i].Commission))
		row = append(row, s[i].BookingDate.Format("2006-01-02"))
		row = append(row, "TRUE")
		row = append(row, fmt.Sprintf("%.2f", s[i].Greeting))
		row = append(row, fmt.Sprintf("%.2f", s[i].Laundry))
		row = append(row, fmt.Sprintf("%.2f", s[i].Cleaning))
		row = append(row, fmt.Sprintf("%.2f", s[i].Consumables))
		row = append(row, fmt.Sprintf("%.2f", s[i].BookingFee))
		row = append(row, fmt.Sprintf("%.2f", s[i].HouseOwnerFee))
		row = append(row, fmt.Sprintf("%.2f", s[i].TotalFees))
		row = append(row, fmt.Sprintf("%.2f", s[i].OwnerIncome))
		w.Write(row)
	}
}
