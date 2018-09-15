package main

import (
	"bytes"
	"log"
	"reflect"
	"testing"
	"time"
)

func Test_yearsAfter(t *testing.T) {
	type args struct {
		years    int
		fromDate time.Time
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{"", args{years: 0, fromDate: Datetime(2017, time.January, 1)}, Datetime(2017, time.January, 1)},
		{"", args{years: 1, fromDate: Datetime(2017, time.January, 1)}, Datetime(2018, time.January, 1)},
		//TODO: we don't want to accept negative values
		{"", args{years: -1, fromDate: Datetime(2017, time.January, 1)}, Datetime(2016, time.January, 1)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := yearsAfter(tt.args.years, tt.args.fromDate); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("yearsAfter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSliceEnd(t *testing.T) {
	type args struct {
		bookingRef string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO raise error if first value is NaN
		{"", args{bookingRef: "AMJUN1719"}, 0},
		{"", args{bookingRef: "1AMJUN1719"}, 1},
		{"", args{bookingRef: "11AMJUN1719"}, 2},
		{"", args{bookingRef: "111AMJUN1719"}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSliceEnd(tt.args.bookingRef); got != tt.want {
				t.Errorf("getSliceEnd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getBookingYear(t *testing.T) {
	type args struct {
		sliceEnd   int
		bookingRef string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"", args{sliceEnd: 1, bookingRef: "6ASJUN1719"}, 2017},
		// TODO: raise error if slice causes a read of NaN
		{"", args{sliceEnd: 2, bookingRef: "6ASJUN1719"}, BusinessOpeningDate.Year()},
		{"", args{sliceEnd: 3, bookingRef: "123JUN1719"}, 2134},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getBookingYear(tt.args.sliceEnd, tt.args.bookingRef); got != tt.want {
				t.Errorf("getBookingYear() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getBookingArrivalDate(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	type args struct {
		sliceEnd    int
		bookingYear int
		bookingRef  string
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{"", args{sliceEnd: 1, bookingRef: "1ASJUN1719", bookingYear: 2012}, Datetime(2012, time.June, 17)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getBookingArrivalDate(tt.args.sliceEnd, tt.args.bookingYear, tt.args.bookingRef); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getBookingArrivalDate() = %v, want %v", got, tt.want)
			}
		})
	}

	got := getBookingArrivalDate(1, 2012, "1ASJUB1719")
	want := Datetime(0, 0, 0)
	wantLog := "unknown month in booking reference: JUB\n"
	gotLog := buf.String()
	ignore := len("2017/12/10 12:37:08 ")
	if gotLog[ignore:] != wantLog {
		t.Errorf("%#v, want %#v", gotLog[ignore:], wantLog)
	}
	if got != want {
		t.Errorf("%#v, want %#v", got, want)
	}
}

func Test_getBookingDepartureDate(t *testing.T) {
	type args struct {
		sliceEnd    int
		arrivalDate time.Time
		bookingRef  string
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{"", args{sliceEnd: 1, bookingRef: "1ASJUN1719",
			arrivalDate: Datetime(2012, time.June, 17)},
			Datetime(2012, time.June, 19)},
		{"", args{sliceEnd: 1, bookingRef: "1ASDEC3103",
			arrivalDate: Datetime(2012, time.December, 31)},
			Datetime(2013, time.January, 3)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getBookingDepartureDate(tt.args.sliceEnd, tt.args.arrivalDate, tt.args.bookingRef); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getBookingDepartureDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createBookingRef(t *testing.T) {
	type args struct {
		b Booking
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{b: Booking{
			Property:    Property{ShortName: "AM"},
			Arrival:     Datetime(2012, time.December, 17),
			Departure:   Datetime(2013, time.December, 03),
			BookingDate: Datetime(2012, time.May, 20)}},
			"1AMDEC1703"},
		{"", args{b: Booking{
			Property:    Property{ShortName: "AS"},
			Arrival:     Datetime(2012, time.June, 17),
			Departure:   Datetime(2012, time.June, 19),
			BookingDate: Datetime(2017, time.May, 20)}},
			"6ASJUN1719"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createBookingRef(tt.args.b); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createBookingRef() = %v, want %v", got, tt.want)
			}
		})
	}
}
