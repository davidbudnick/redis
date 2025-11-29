package types

import "github.com/charmbracelet/lipgloss"

// Theme defines the color scheme for the UI
type Theme struct {
	Name          string
	Primary       lipgloss.Color
	Secondary     lipgloss.Color
	Background    lipgloss.Color
	Foreground    lipgloss.Color
	Muted         lipgloss.Color
	Accent        lipgloss.Color
	Error         lipgloss.Color
	Warning       lipgloss.Color
	Success       lipgloss.Color
	Info          lipgloss.Color
	Border        lipgloss.Color
	Selection     lipgloss.Color
	SelectionText lipgloss.Color
	TypeString    lipgloss.Color
	TypeList      lipgloss.Color
	TypeSet       lipgloss.Color
	TypeZSet      lipgloss.Color
	TypeHash      lipgloss.Color
	TypeStream    lipgloss.Color
}

// DarkTheme is the default dark color scheme
var DarkTheme = Theme{
	Name:          "Dark",
	Primary:       lipgloss.Color("39"),
	Secondary:     lipgloss.Color("63"),
	Background:    lipgloss.Color("0"),
	Foreground:    lipgloss.Color("15"),
	Muted:         lipgloss.Color("240"),
	Accent:        lipgloss.Color("135"),
	Error:         lipgloss.Color("1"),
	Warning:       lipgloss.Color("3"),
	Success:       lipgloss.Color("2"),
	Info:          lipgloss.Color("4"),
	Border:        lipgloss.Color("240"),
	Selection:     lipgloss.Color("39"),
	SelectionText: lipgloss.Color("0"),
	TypeString:    lipgloss.Color("2"),
	TypeList:      lipgloss.Color("3"),
	TypeSet:       lipgloss.Color("4"),
	TypeZSet:      lipgloss.Color("5"),
	TypeHash:      lipgloss.Color("6"),
	TypeStream:    lipgloss.Color("13"),
}

// LightTheme is a light color scheme
var LightTheme = Theme{
	Name:          "Light",
	Primary:       lipgloss.Color("27"),
	Secondary:     lipgloss.Color("63"),
	Background:    lipgloss.Color("15"),
	Foreground:    lipgloss.Color("0"),
	Muted:         lipgloss.Color("245"),
	Accent:        lipgloss.Color("99"),
	Error:         lipgloss.Color("160"),
	Warning:       lipgloss.Color("208"),
	Success:       lipgloss.Color("34"),
	Info:          lipgloss.Color("33"),
	Border:        lipgloss.Color("250"),
	Selection:     lipgloss.Color("27"),
	SelectionText: lipgloss.Color("15"),
	TypeString:    lipgloss.Color("28"),
	TypeList:      lipgloss.Color("130"),
	TypeSet:       lipgloss.Color("25"),
	TypeZSet:      lipgloss.Color("127"),
	TypeHash:      lipgloss.Color("30"),
	TypeStream:    lipgloss.Color("128"),
}

// NordTheme is a Nord-inspired color scheme
var NordTheme = Theme{
	Name:          "Nord",
	Primary:       lipgloss.Color("#88C0D0"),
	Secondary:     lipgloss.Color("#81A1C1"),
	Background:    lipgloss.Color("#2E3440"),
	Foreground:    lipgloss.Color("#ECEFF4"),
	Muted:         lipgloss.Color("#4C566A"),
	Accent:        lipgloss.Color("#B48EAD"),
	Error:         lipgloss.Color("#BF616A"),
	Warning:       lipgloss.Color("#EBCB8B"),
	Success:       lipgloss.Color("#A3BE8C"),
	Info:          lipgloss.Color("#5E81AC"),
	Border:        lipgloss.Color("#4C566A"),
	Selection:     lipgloss.Color("#5E81AC"),
	SelectionText: lipgloss.Color("#ECEFF4"),
	TypeString:    lipgloss.Color("#A3BE8C"),
	TypeList:      lipgloss.Color("#EBCB8B"),
	TypeSet:       lipgloss.Color("#81A1C1"),
	TypeZSet:      lipgloss.Color("#B48EAD"),
	TypeHash:      lipgloss.Color("#88C0D0"),
	TypeStream:    lipgloss.Color("#D08770"),
}

// DraculaTheme is a Dracula-inspired color scheme
var DraculaTheme = Theme{
	Name:          "Dracula",
	Primary:       lipgloss.Color("#BD93F9"),
	Secondary:     lipgloss.Color("#8BE9FD"),
	Background:    lipgloss.Color("#282A36"),
	Foreground:    lipgloss.Color("#F8F8F2"),
	Muted:         lipgloss.Color("#6272A4"),
	Accent:        lipgloss.Color("#FF79C6"),
	Error:         lipgloss.Color("#FF5555"),
	Warning:       lipgloss.Color("#FFB86C"),
	Success:       lipgloss.Color("#50FA7B"),
	Info:          lipgloss.Color("#8BE9FD"),
	Border:        lipgloss.Color("#44475A"),
	Selection:     lipgloss.Color("#44475A"),
	SelectionText: lipgloss.Color("#F8F8F2"),
	TypeString:    lipgloss.Color("#50FA7B"),
	TypeList:      lipgloss.Color("#FFB86C"),
	TypeSet:       lipgloss.Color("#8BE9FD"),
	TypeZSet:      lipgloss.Color("#FF79C6"),
	TypeHash:      lipgloss.Color("#BD93F9"),
	TypeStream:    lipgloss.Color("#FF5555"),
}

// AvailableThemes lists all available themes
var AvailableThemes = []Theme{DarkTheme, LightTheme, NordTheme, DraculaTheme}

// GetThemeByName returns a theme by name
func GetThemeByName(name string) Theme {
	for _, t := range AvailableThemes {
		if t.Name == name {
			return t
		}
	}
	return DarkTheme
}

// GetTypeColor returns the color for a key type based on the current theme
func (t Theme) GetTypeColor(keyType KeyType) lipgloss.Color {
	switch keyType {
	case KeyTypeString:
		return t.TypeString
	case KeyTypeList:
		return t.TypeList
	case KeyTypeSet:
		return t.TypeSet
	case KeyTypeZSet:
		return t.TypeZSet
	case KeyTypeHash:
		return t.TypeHash
	case KeyTypeStream:
		return t.TypeStream
	default:
		return t.Foreground
	}
}
