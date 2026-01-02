package redis

import (
	"testing"
)

func TestFuzzyScore(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		pattern  string
		wantMin  int // Minimum expected score
		wantZero bool
	}{
		{
			name:    "exact match returns high score",
			str:     "user:123",
			pattern: "user:123",
			wantMin: 100,
		},
		{
			name:    "substring match returns high score",
			str:     "user:123:profile",
			pattern: "user:123",
			wantMin: 100,
		},
		{
			name:    "prefix match with separator bonus",
			str:     "user:profile:settings",
			pattern: "ups",
			wantMin: 30, // u + p (with separator bonus) + s (with separator bonus)
		},
		{
			name:    "sequential character match",
			str:     "configuration",
			pattern: "cfg",
			wantMin: 20,
		},
		{
			name:     "no match returns zero",
			str:      "user:123",
			pattern:  "xyz",
			wantZero: true,
		},
		{
			name:     "partial pattern match returns zero",
			str:      "ab",
			pattern:  "abc",
			wantZero: true,
		},
		{
			name:    "underscore separator bonus",
			str:     "user_profile_data",
			pattern: "upd",
			wantMin: 30, // Each char after separator gets bonus
		},
		{
			name:    "hyphen separator bonus",
			str:     "user-profile-data",
			pattern: "upd",
			wantMin: 30,
		},
		{
			name:    "empty pattern matches everything",
			str:     "anything",
			pattern: "",
			wantMin: 0,
		},
		{
			name:     "empty string with pattern returns zero",
			str:      "",
			pattern:  "test",
			wantZero: true,
		},
		{
			name:    "single character match",
			str:     "test",
			pattern: "t",
			wantMin: 10,
		},
		{
			name:    "case sensitive matching",
			str:     "UserProfile",
			pattern: "UP",
			wantMin: 20,
		},
		{
			name:     "case mismatch returns zero",
			str:      "userprofile",
			pattern:  "UP",
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fuzzyScore(tt.str, tt.pattern)

			if tt.wantZero {
				if got != 0 {
					t.Errorf("fuzzyScore(%q, %q) = %d, want 0", tt.str, tt.pattern, got)
				}
				return
			}

			if got < tt.wantMin {
				t.Errorf("fuzzyScore(%q, %q) = %d, want >= %d", tt.str, tt.pattern, got, tt.wantMin)
			}
		})
	}
}

func TestFuzzyScore_ContainsVsSequential(t *testing.T) {
	// When pattern is contained in string, score should be higher than sequential match
	containsStr := "session:user:123"
	containsPattern := "user"
	containsScore := fuzzyScore(containsStr, containsPattern)

	sequentialStr := "u_s_e_r_data"
	sequentialPattern := "user"
	sequentialScore := fuzzyScore(sequentialStr, sequentialPattern)

	if containsScore <= sequentialScore {
		t.Errorf("Contains match score (%d) should be higher than sequential match score (%d)",
			containsScore, sequentialScore)
	}
}

func TestFuzzyScore_SeparatorBonus(t *testing.T) {
	// Characters after separators should get bonus points
	withSeparator := "user:data"
	withoutSeparator := "userdatax"
	pattern := "ud"

	withSepScore := fuzzyScore(withSeparator, pattern)
	withoutSepScore := fuzzyScore(withoutSeparator, pattern)

	if withSepScore <= withoutSepScore {
		t.Errorf("Separator bonus score (%d) should be higher than without separator (%d)",
			withSepScore, withoutSepScore)
	}
}

func BenchmarkFuzzyScore(b *testing.B) {
	str := "user:profile:settings:preferences:notifications"
	pattern := "upsn"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fuzzyScore(str, pattern)
	}
}

func BenchmarkFuzzyScore_LongString(b *testing.B) {
	str := "very:long:redis:key:with:many:segments:for:testing:performance"
	pattern := "vlrk"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fuzzyScore(str, pattern)
	}
}
