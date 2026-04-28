package cmdutil_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/brpaz/gh-secrets-sync/internal/cmdutil"
)

func TestSplitRepos(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "comma-separated single string",
			input: []string{"owner/repo1,owner/repo2"},
			want:  []string{"owner/repo1", "owner/repo2"},
		},
		{
			name:  "repeated flag values",
			input: []string{"owner/repo1", "owner/repo2"},
			want:  []string{"owner/repo1", "owner/repo2"},
		},
		{
			name:  "mixed comma-separated and repeated",
			input: []string{"owner/repo1,owner/repo2", "owner/repo3"},
			want:  []string{"owner/repo1", "owner/repo2", "owner/repo3"},
		},
		{
			name:  "trims surrounding whitespace",
			input: []string{" owner/repo1 , owner/repo2 "},
			want:  []string{"owner/repo1", "owner/repo2"},
		},
		{
			name:  "empty input",
			input: []string{},
			want:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, cmdutil.SplitRepos(tc.input))
		})
	}
}
