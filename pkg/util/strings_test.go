package util

import "testing"

func TestReverseString(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"normal", args{"1234"}, "4321"},
		{"empty", args{""}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReverseString(tt.args.in); got != tt.want {
				t.Errorf("ReverseString() = %v, want %v", got, tt.want)
			}
		})
	}
}
