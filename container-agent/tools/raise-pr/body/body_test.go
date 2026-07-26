package body_test

import (
	"testing"

	"github.com/chrisjpalmer/shoppinglist/scripts/raise-pr/body"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		lines   []string
		want    string
		wantErr bool
	}{
		{
			name: "basic",
			lines: []string{
				"Here is one line.",
				"Here is another.",
				"",
				"Here is a double line.",
			},
			want: `Here is one line. Here is another.

Here is a double line.`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := body.Format(tt.lines)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Format() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Format() succeeded unexpectedly")
			}

			if got != tt.want {
				t.Errorf("\ngot : %q\nwant: %q", got, tt.want)
			}
		})
	}
}
