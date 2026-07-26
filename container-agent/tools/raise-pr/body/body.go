package body

import (
	"fmt"
	"strings"
)

func Format(lines []string) (string, error) {

	trimLines(lines)

	body, err := joinLines(lines)
	if err != nil {
		return "", fmt.Errorf("error joining lines: %w", err)
	}

	return body, nil
}

func trimLines(lines []string) {
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
}

func joinLines(lines []string) (string, error) {
	var bld strings.Builder

	wasEmpty := false

	for i, l := range lines {
		isEmpty := len(l) == 0

		if wasEmpty {
			if _, err := bld.WriteString("\n\n"); err != nil {
				return "", fmt.Errorf("error adding new line: %w", err)
			}
		} else if i != 0 && !isEmpty {
			if _, err := bld.WriteString(" "); err != nil {
				return "", fmt.Errorf("error appending string: %w", err)
			}
		}

		if _, err := bld.WriteString(l); err != nil {
			return "", fmt.Errorf("error appending string: %w", err)
		}

		wasEmpty = isEmpty
	}

	return bld.String(), nil
}
