package metrics

import "testing"

func TestClassifyIsBounded(t *testing.T) {
	known := map[string]bool{"bananabot": true, "bananatron": true}

	tests := map[string]struct{ route, bot string }{
		"/":                     {"/", ""},
		"/metrics":              {"/metrics", ""},
		"/bananatron":           {"/{bot}", "bananatron"},
		"/bananatron/":          {"/{bot}", "bananatron"},
		"/bananatron/move":      {"/{bot}/move", "bananatron"},
		"/bananabot/start":      {"/{bot}/start", "bananabot"},
		"/bananabot/end":        {"/{bot}/end", "bananabot"},
		"/bananatron/nonsense":  {"other", "bananatron"},
		"/wp-admin":             {"other", "unknown"},
		"/attacker/move":        {"other", "unknown"},
		"/bananatron/move/deep": {"other", "bananatron"},
	}

	for path, want := range tests {
		route, bot := Classify(path, known)
		if route != want.route || bot != want.bot {
			t.Errorf("Classify(%q) = (%q, %q), want (%q, %q)", path, route, bot, want.route, want.bot)
		}
	}
}
