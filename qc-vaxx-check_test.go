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
