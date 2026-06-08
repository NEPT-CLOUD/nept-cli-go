package utls

import (
	"testing"
)

func TestIgnoreRules(t *testing.T) {
	rules := []IgnoreRule{
		CompileIgnore("node_modules"),
		CompileIgnore("*.log"),
		CompileIgnore("/dist"),
	}

	tests := []struct {
		path   string
		expect bool
	}{
		{"node_modules/express/index.js", true},
		{"src/app.log", true},
		{"dist/main.js", true},
		{"src/dist/sub.js", false}, // Dist is anchored to root due to "/"
		{"src/index.js", false},
	}

	for _, tc := range tests {
		got := IsIgnored(tc.path, rules)
		if got != tc.expect {
			t.Errorf("IsIgnored(%q) = %v, expect %v", tc.path, got, tc.expect)
		}
	}
}
