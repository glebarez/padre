package util

import (
	"net/http"
	"reflect"
	"testing"
)

func TestParseCookies(t *testing.T) {
	type args struct {
		cookies string
	}
	tests := []struct {
		name          string
		args          args
		wantCookSlice []*http.Cookie
		wantErr       bool
	}{
		{"empty", args{""}, []*http.Cookie{}, false},
		{"normal", args{"key=val"}, []*http.Cookie{{Name: "key", Value: "val"}}, false},
		{"errornous", args{"key=val=1"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCookSlice, err := ParseCookies(tt.args.cookies)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCookies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCookSlice, tt.wantCookSlice) {
				t.Errorf("ParseCookies() = %v, want %v", gotCookSlice, tt.wantCookSlice)
			}
		})
	}
}

func TestDetectContentType(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"json-object", args{"{'a':1}"}, "application/json"},
		{"json-array", args{"[{'a':1}]"}, "application/json"},
		{"form", args{"a=1&b=2"}, "application/x-www-form-urlencoded"},
		{"text", args{"text"}, http.DetectContentType([]byte("text"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DetectContentType(tt.args.data); got != tt.want {
				t.Errorf("DetectContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}
