package messageformat

import (
	"testing"

	"github.com/stretchr/testify/require"
)

//nolint:lll
func Test_ConvertToMessageFormat2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Placeholder tests
		{
			name:     "python old formatting",
			input:    "%(object)s does not exist in this database.",
			expected: "{{:Placeholder format=pythonVar name=object type=string} does not exist in this database.}",
		},
		{
			name:     "python old formatting multiple placeholders",
			input:    "%(object)s does not exist in the database #%(num)d.",
			expected: "{{:Placeholder format=pythonVar name=object type=string} does not exist in the database \\#{:Placeholder format=pythonVar name=num type=int}.}",
		},
		{
			name:  "python old formatting with newlines and special chars",
			input: "*%(name)s*\n\n%(description)s\n\n<%(url)s|Explore in Superset>\n\n%(table)s\n",
			expected: `{\*{:Placeholder format=pythonVar name=name type=string}\*

{:Placeholder format=pythonVar name=description type=string}

\<{:Placeholder format=pythonVar name=url type=string}\|Explore in Superset\>

{:Placeholder format=pythonVar name=table type=string}
}`,
		},
		{
			name:     "printf style",
			input:    "%s does not exist in this database.",
			expected: "{{:Placeholder format=printf type=string} does not exist in this database.}",
		},
		{
			name:     "printf style multiple placeholders",
			input:    "%s does not exist in the database #%d.",
			expected: "{{:Placeholder format=printf type=string} does not exist in the database \\#{:Placeholder format=printf type=int}.}",
		},
		{
			name:     "printf style with exclamation mark",
			input:    "Welcome, %s!",
			expected: "{Welcome, {:Placeholder format=printf type=string}\\!}",
		},
		{
			name:     "brackets with variable",
			input:    "{object} does not exist in this database.",
			expected: "{{:Placeholder format=bracketVar name=object} does not exist in this database.}",
		},
		{
			name:     "brackets with variable multiple placeholders",
			input:    "{object} does not exist in the database #{num}.",
			expected: "{{:Placeholder format=bracketVar name=object} does not exist in the database \\#{:Placeholder format=bracketVar name=num}.}",
		},
		{
			name:     "brackets with order",
			input:    "{0} does not exist in this database.",
			expected: "{{:Placeholder format=bracketVar name=0} does not exist in this database.}",
		},
		{
			name:     "brackets with order multiple placeholders",
			input:    "{0} does not exist in the database #{1}.",
			expected: "{{:Placeholder format=bracketVar name=0} does not exist in the database \\#{:Placeholder format=bracketVar name=1}.}",
		},
		{
			name:     "brackets without variable",
			input:    "{} does not exist in this database.",
			expected: "{{:Placeholder format=emptyBracket} does not exist in this database.}",
		},
		{
			name:     "brackets without variable multiple placeholders",
			input:    "{} does not exist in the database #{}.",
			expected: "{{:Placeholder format=emptyBracket} does not exist in the database \\#{:Placeholder format=emptyBracket}.}",
		},
		// Tests for checking false positives
		{
			name:     "no placeholders",
			input:    "Happy days",
			expected: "{Happy days}",
		},
		{
			name:     "no placeholders with percent",
			input:    "Happy days, 20% left",
			expected: "{Happy days, 20\\% left}",
		},
		{
			name:     "empty message",
			input:    "",
			expected: "",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// TODO: Delete skip when we have ast to mf2 string
			t.Skip("We need ast to mf2 string")

			actual := ToMessageFormat2(tt.input)
			require.Equal(t, tt.expected, actual)
		})
	}
}
