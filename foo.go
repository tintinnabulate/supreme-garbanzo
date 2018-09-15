package main

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing/quick"
	"time"
)

// Date creates a random date useful for randomness
type Date int64

// Generate allows Date to be used within quickcheck scenarios.
func (Date) Generate(r *rand.Rand, size int) reflect.Value {
	var (
		min = time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
		max = time.Date(2070, 1, 0, 0, 0, 0, 0, time.UTC).Unix()

		sec = r.Int63n(max-min) + min
	)
	return reflect.ValueOf(Date(sec))
}

func (d Date) Date() time.Time {
	return time.Unix(int64(d), 0)
}

func (d Date) String() string {
	return d.Date().Format(time.RFC3339)
}

// Property is used to unmarshal a JSON file of the above form
// type Property struct {
// 	LongName             string
// 	ShortName            string
// 	Calendar             string
// 	Commission           float64
// 	BookingCommission    float64
// 	HouseOwnerCommission float64
// 	Greeting             float64
// 	Laundry              []float64
// 	Cleaning             float64
// 	Consumables          []float64
// }

// Settings holds the settings for each property
// type Settings struct {
// 	Properties []Property `json:"properties"`
// }

// Source is an Enum
// type Source int

// BookingCom and others are sources the booking originated from
// const (
// 	BookingCom Source = 1 + iota
// 	AirBnb
// 	Email
// 	Phone
// 	VisitBath
// 	Other
// )

func (Source) Generate(rand *rand.Rand, size int) reflect.Value {
	return reflect.ValueOf(Source(rand.Intn(nSources)))
}

// var months = map[string]time.Month{
// 	"JAN": time.January,
// 	"FEB": time.February,
// 	"MAR": time.March,
// 	"APR": time.April,
// 	"MAY": time.May,
// 	"JUN": time.June,
// 	"JUL": time.July,
// 	"AUG": time.August,
// 	"SEP": time.September,
// 	"OCT": time.October,
// 	"NOV": time.November,
// 	"DEC": time.December,
// }

func (m Month) Generate(rand *rand.Rand, size int) reflect.Value {
	return reflect.ValueOf(Month(rand.Intn(nMonths)))
}

// FormInput holds all the form Inputs from the user
// type FormInput struct {
// 	BookingRef     string
// 	FirstName      string
// 	LastName       string
// 	Email          string
// 	Mobile         string
// 	Notes          string
// 	Source         Source
// 	NumberOfPeople int
// 	Gross          float64
// 	IsGreeting     bool
// 	IsLaundry      bool
// 	IsCleaning     bool
// 	IsConsumables  bool
// 	BookingDate    time.Time
// }

// Booking holds a booking
// type Booking struct {
// 	Form          FormInput
// 	Property      Property
// 	Arrival       time.Time
// 	Departure     time.Time
// 	BookingDate   time.Time
// 	BookingFee    float64
// 	HouseOwnerFee float64
// 	Net           float64
// 	TotalFees     float64
// 	OwnerIncome   float64
// }

func main() {
	rnd := rand.New(rand.NewSource(42))
	t := reflect.TypeOf(Date(0))
	v, _ := quick.Value(t, rnd)
	fmt.Println("here's a time:", v.Interface())
	a := reflect.TypeOf(Month(0))
	b, _ := quick.Value(a, rnd)
	fmt.Println("here's a month:", b.Interface())
}
