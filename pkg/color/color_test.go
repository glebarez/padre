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
		{"nocolor", args{"x"}, 1},
		{"colored", args{YellowBold("x")}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TrueLen(tt.args.s); got != tt.want {
				t.Errorf("TrueLen() = %v, want %v", got, tt.want)
			}
		})
	}
}
