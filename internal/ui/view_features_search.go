package ui

import (
	"strings"
)

func (m Model) viewRegexSearch() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Regex Search"))
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("Search keys using regular expressions"))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Pattern:"))
	b.WriteString("\n")
	b.WriteString(m.RegexSearchInput.View())
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("Examples: user:\\d+  session:[a-f0-9]+"))
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("enter:search  esc:cancel"))

	return m.renderModal(b.String())
}

func (m Model) viewFuzzySearch() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Fuzzy Search"))
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("Search keys with fuzzy matching"))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Search:"))
	b.WriteString("\n")
	b.WriteString(m.FuzzySearchInput.View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("enter:search  esc:cancel"))

	return m.renderModal(b.String())
}

func (m Model) viewCompareKeys() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Compare Keys"))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Key 1:"))
	b.WriteString("\n")
	b.WriteString(m.CompareKey1Input.View())
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Key 2:"))
	b.WriteString("\n")
	b.WriteString(m.CompareKey2Input.View())
	b.WriteString("\n\n")

	if m.CompareResult != nil {
		if m.CompareResult.Equal {
			b.WriteString(successStyle.Render("Keys are identical"))
		} else {
			b.WriteString(errorStyle.Render("x Keys differ"))
			b.WriteString("\n\n")
			if len(m.CompareResult.Differences) > 0 {
				b.WriteString(keyStyle.Render("Differences:"))
				b.WriteString("\n")
				for _, diff := range m.CompareResult.Differences {
					b.WriteString("  * " + diff + "\n")
				}
			}
		}
		b.WriteString("\n")
	}

	b.WriteString(helpStyle.Render("tab:switch  enter:compare  esc:back"))

	return m.renderModal(b.String())
}

func (m Model) viewJSONPath() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("JSON Path Query"))
	b.WriteString("\n\n")

	if m.CurrentKey != nil {
		b.WriteString(keyStyle.Render("Key: "))
		b.WriteString(normalStyle.Render(m.CurrentKey.Key))
		b.WriteString("\n\n")
	}

	b.WriteString(keyStyle.Render("JSON Path:"))
	b.WriteString("\n")
	b.WriteString(m.JSONPathInput.View())
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("Examples: $.name  $.users[0]  $.items[*].id"))
	b.WriteString("\n\n")

	if m.JSONPathResult != "" {
		b.WriteString(keyStyle.Render("Result:"))
		b.WriteString("\n")
		b.WriteString(normalStyle.Render(m.JSONPathResult))
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("enter:query  esc:back"))

	return m.renderModal(b.String())
}
