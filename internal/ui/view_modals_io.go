package ui

import (
	"strings"
)

func (m Model) viewSearchValues() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Search by Value"))
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("Find keys containing a specific value"))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Search:"))
	b.WriteString("\n")
	b.WriteString(m.SearchValueInput.View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("enter:search  esc:cancel"))

	return m.renderModal(b.String())
}

func (m Model) viewExport() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Export Keys"))
	b.WriteString("\n\n")

	pattern := m.KeyPattern
	if pattern == "" {
		pattern = "*"
	}
	b.WriteString(keyStyle.Render("Pattern: "))
	b.WriteString(normalStyle.Render(pattern))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Filename:"))
	b.WriteString("\n")
	b.WriteString(m.ExportInput.View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("enter:export  esc:cancel"))

	return m.renderModal(b.String())
}

func (m Model) viewImport() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Import Keys"))
	b.WriteString("\n\n")

	b.WriteString(dimStyle.Render("Import keys from a JSON file"))
	b.WriteString("\n\n")

	b.WriteString(keyStyle.Render("Filename:"))
	b.WriteString("\n")
	b.WriteString(m.ImportInput.View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("enter:import  esc:cancel"))

	return m.renderModal(b.String())
}
