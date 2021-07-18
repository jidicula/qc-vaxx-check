package main

import (
	"reflect"
	"testing"
)

func TestDecode(t *testing.T) {
	tests := map[string]struct {
		byteSlice []byte
		want      []byte
	}{
		"multiple of 4": {
			byteSlice: []byte("YWJj"),
			want:      []byte("abc")},
		"not multiple of 4": {
			byteSlice: []byte("YWJjZA"),
			want:      []byte("abcd")},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := decode(tt.byteSlice)
			if err != nil {
				t.Fatalf("%s: %s", name, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s: got %v, want %v", name, got, tt.want)
			}
		})
	}

}

func TestPatientFamilyNameWorkaround(t *testing.T) {
	tests := map[string]struct {
		byteSlice []byte
		want      []byte
	}{
		"0 family names": {
			byteSlice: []byte(`"family":[],`),
			want:      []byte(`"family":"",`),
		},
		"1 family name": {
			byteSlice: []byte(`"family":["Doe"],`),
			want:      []byte(`"family":"Doe",`),
		},
		"2 family names": {
			byteSlice: []byte(`"family":["Doe","Deere"],`),
			want:      []byte(`"family":"Doe Deere",`),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := PatientFamilyNameWorkaround(tt.byteSlice)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s: got `%s`, want `%s`", name, got, tt.want)
			}
		})
	}
}
