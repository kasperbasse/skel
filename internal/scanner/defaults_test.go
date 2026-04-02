package scanner

import (
	"testing"
)

func TestCuratedDefaultsNoDuplicates(t *testing.T) {
	seen := make(map[string]bool)
	for _, d := range curatedDefaults {
		key := d.Domain + "/" + d.Key
		if seen[key] {
			t.Errorf("duplicate curated default: %s %s", d.Domain, d.Key)
		}
		seen[key] = true
	}
}

func TestCuratedDefaultsValidTypes(t *testing.T) {
	validTypes := map[string]bool{
		"string": true,
		"int":    true,
		"float":  true,
		"bool":   true,
	}
	for _, d := range curatedDefaults {
		if !validTypes[d.Type] {
			t.Errorf("invalid type %q for %s %s", d.Type, d.Domain, d.Key)
		}
	}
}

func TestCuratedDefaultsNotEmpty(t *testing.T) {
	for _, d := range curatedDefaults {
		if d.Domain == "" {
			t.Error("empty domain in curated defaults")
		}
		if d.Key == "" {
			t.Error("empty key in curated defaults")
		}
	}
}

func TestReadDefaultNonexistent(t *testing.T) {
	// Reading a key that almost certainly doesn't exist should return false.
	_, ok := readDefault("com.skel.nonexistent.test", "this_key_does_not_exist_12345")
	if ok {
		t.Error("expected false for nonexistent key")
	}
}

func TestScanDefaultsReturnsValidProfile(t *testing.T) {
	warned := false
	warn := func(msg string) { warned = true; _ = msg }

	result := scanDefaults(warn)

	// We can't predict which settings exist on CI vs local,
	// but the function should not panic and should return a valid struct.
	if result.Settings == nil {
		// nil is fine - means no custom settings found
		return
	}

	for _, s := range result.Settings {
		if s.Domain == "" || s.Key == "" || s.Type == "" {
			t.Errorf("incomplete setting: %+v", s)
		}
	}

	// warned should be false - scanDefaults doesn't warn currently
	if warned {
		t.Error("unexpected warning from scanDefaults")
	}
}
