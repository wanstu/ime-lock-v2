package main

import "testing"

func TestHasAutoStartArg(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{name: "empty", args: nil, want: false},
		{name: "manual", args: []string{"--debug"}, want: false},
		{name: "autostart", args: []string{"--autostart"}, want: true},
		{name: "legacy minimized", args: []string{"--minimized"}, want: true},
		{name: "case insensitive", args: []string{" --AUTOSTART "}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasAutoStartArg(tt.args); got != tt.want {
				t.Fatalf("hasAutoStartArg(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}
