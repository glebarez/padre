package color

import "testing"

func TestTrueLen(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"nocolor", args{""}, 0},
		{"nocolor", args{"x"}, 1},
		{"nocolor", args{"xxx"}, 3},
		{"colored", args{YellowBold("")}, 0},
		{"colored", args{YellowBold("x")}, 1},
		{"colored", args{YellowBold("xxx")}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrueLen(tt.args.s); got != tt.want {
				t.Errorf("TrueLen() = %v, want %v", got, tt.want)
			}
		})
	}
}
