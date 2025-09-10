package scrobbler

import (
	"fmt"
	"testing"
)

func TestBaseWrapper_conversionSimplified(t *testing.T) {
	type args struct {
		target string
	}
	tests := []struct {
		name string
		m    BaseWrapper
		args args
		want string
	}{
		{
			name: "空字符串",
			m:    BaseWrapper{},
			args: args{target: ""},
			want: "",
		},
		{
			name: "只有空格",
			m:    BaseWrapper{},
			args: args{target: "   "},
			want: "   ",
		},
		{
			name: "无中文字符",
			m:    BaseWrapper{},
			args: args{target: "Hello World"},
			want: "Hello World",
		},
		{
			name: "繁体转简体",
			m:    BaseWrapper{},
			args: args{target: "我來測試一下繁轉簡"},
			want: "我来测试一下繁转简",
		},
		{
			name: "混合英文繁体转简体",
			m:    BaseWrapper{},
			args: args{target: "我來測試一下繁轉簡 feat（mj）"},
			want: "我来测试一下繁转简 feat（mj）",
		},
		{
			name: "简体保持不变",
			m:    BaseWrapper{},
			args: args{target: "我来测试一下简体"},
			want: "我来测试一下简体",
		},
		{
			name: "印地保持不变",
			m:    BaseWrapper{},
			args: args{target: "नमस्ते"},
			want: "नमस्ते",
		},
		{
			name: "日语保持不变",
			m:    BaseWrapper{},
			args: args{target: "こんにちは"},
			want: "こんにちは",
		},
		{
			name: "混合繁体日语转简体",
			m:    BaseWrapper{},
			args: args{target: "こん繁にち轉は簡 feat（mj）"},
			want: "こん繁にち转は简 feat（mj）",
		},
	}
	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := tt.m.conversionSimplified(tt.args.target); got != tt.want {
					t.Errorf("BaseWrapper.conversionSimplified() = %v, want %v", got, tt.want)
				} else {
					fmt.Println(tt.args.target)
					fmt.Println(got)
					fmt.Println("=======")
				}
			},
		)
	}
}
