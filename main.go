package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var selectClause, whereClause string

func init() {
	const (
		selectDefault = "*"
		selectUsage   = "SELECT clause"
		whereDefault  = ""
		whereUsage    = "WHERE clause"
	)
	flag.StringVar(&selectClause, "S", selectDefault, selectUsage)
	flag.StringVar(&selectClause, "select", selectDefault, selectUsage+" (shorthand)")
	flag.StringVar(&whereClause, "W", whereDefault, whereUsage)
	flag.StringVar(&whereClause, "where", whereDefault, whereUsage+" (shorthand)")
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table table.Model
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Batch(
				tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n  " + m.table.HelpView() + "\n"
}

var selectedColumnIndices []int

func parseSelectClause(dataHeader []string) []int {
	// fmt.Print("parseSelectClause \n")
	// fmt.Printf("%+v", selectClause)
	if len(selectClause) == 0 || selectClause == "*" {
		// fmt.Print("all columns \n")
		// Get indices of selected columns
		for i := range dataHeader {
			selectedColumnIndices = append(selectedColumnIndices, i)
		}
	} else {
		fmt.Print("selected columns \n")
		// If * is used or no select clause is added we assume all columns are returned
		selectedColumns := strings.Split(selectClause, ",")
		for i, field := range dataHeader {
			if containsStr(selectedColumns, string(field)) {
				selectedColumnIndices = append(selectedColumnIndices, i)
			}
		}
	}
	fmt.Printf("%+v", selectedColumnIndices)

	return selectedColumnIndices
}

func parseCsvData(data [][]string, selectedColumnIndices []int) ([]table.Row, int) {
	// fmt.Print("parseCsvData \n")
	if len(data) == 0 {
		return nil, 0
	}

	maxWidth := 0
	var rows []table.Row

	for _, line := range data[1:] {
		var row []string
		for i, field := range line {
			if containsInt(selectedColumnIndices, i) {
				row = append(row, field)
				if len(field) > maxWidth {
					maxWidth = len(field)
				}
			}
		}
		rows = append(rows, table.Row(row))
	}

	return rows, maxWidth
}

func buildHeaderRow(dataHeader []string, selectedColumnIndices []int, maxWidth int) []table.Column {
	fmt.Print("buildHeaderRow \n")
	// Now that we have the indices of the selected (or all) columns we can use those to build the column header list
	var columns []table.Column
	for _, columnIdx := range selectedColumnIndices {
		columns = append(columns, table.Column{Title: dataHeader[columnIdx], Width: maxWidth + 2})
	}
	return columns
}

func containsStr(data []string, value string) bool {
	for _, v := range data {
		if v == value {
			return true
		}
	}
	return false
}

func containsInt(data []int, value int) bool {
	for _, v := range data {
		if v == value {
			return true
		}
	}
	return false
}

func main() {
	flag.Parse()

	f, err := os.Open("test.csv")
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	rows, maxWidth := parseCsvData(data, parseSelectClause(data[0]))
	columns := buildHeaderRow(data[0], selectedColumnIndices, maxWidth)

	if len(columns) == 0 || len(rows) == 0 {
		log.Fatal("CSV file is empty or improperly formatted")
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	t.SetStyles(s)

	m := model{t}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
