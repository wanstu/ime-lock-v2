//go:build windows

package main

import "testing"

func TestInputIndicatorMode(t *testing.T) {
	tests := []struct {
		name string
		acc  string
		want string
	}{
		{name: "english", acc: `任务栏输入指示 英语模式`, want: "english"},
		{name: "chinese", acc: `任务栏输入指示 中文模式`, want: "chinese"},
		{name: "unrelated accessibility event", acc: `IME Lock v2`, want: ""},
		{name: "indicator without mode", acc: `任务栏输入指示`, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inputIndicatorMode(tt.acc); got != tt.want {
				t.Fatalf("inputIndicatorMode(%q) = %q, want %q", tt.acc, got, tt.want)
			}
		})
	}
}
