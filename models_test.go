package main

import (
	"testing"
)

func TestSanitizeSensitive(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"您的验证码是 123456，请尽快使用", "您的验证码是 ****，请尽快使用"},
		{"密码是 ABCDEF", "密码是 ****"},
		{"token: abcdefg", "****: ****g"},
		{"纯大写 TESTING 和数字 8888", "纯大写 ****G 和数字 ****"},
		{"", ""},
		{"12345", "****"},
		{"1234567", "****7"},
		{"混合 TestIng", "混合 TestIng"},
	}

	for _, tt := range tests {
		got := sanitizeSensitive(tt.input)
		if got != tt.expected {
			t.Errorf("sanitizeSensitive(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}
