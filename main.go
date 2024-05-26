package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type profile struct {
	name      string
	accountID string
	region    string
	roleName  string
}

const (
	envVarHomeDir = "HOME"

	pathAWSConfigFile = ".aws/config"

	displayFieldAccountID = "acc-id"
	displayFieldRegion    = "region"
	displayFieldRoleName  = "role"
	displayFieldDelim     = ": "

	displayWidthDefault  = 300
	displayHeightDefault = 300

	keypressDisplayLoginCmd = "l"
	keypressClose           = "esc"

	configKeyRegion    = "region"
	configKeyAccountID = "sso_account_id"
	configKeyRoleName  = "sso_role_name"
)

func (p *profile) Title() string {
	return p.name
}

func (p *profile) Description() string {
	var fields []string
	if len(p.accountID) > 0 {
		fields = append(fields, displayFieldAccountID+displayFieldDelim+p.accountID)
	}
	if len(p.region) > 0 {
		fields = append(fields, displayFieldRegion+displayFieldDelim+p.region)
	}
	if len(p.roleName) > 0 {
		fields = append(fields, displayFieldRoleName+displayFieldDelim+p.roleName)
	}

	if len(fields) > 0 {
		return strings.Join(fields, "; ")
	}
	return "no description."
}

func (p *profile) FilterValue() string {
	return strings.Join([]string{p.name, p.accountID, p.region, p.roleName}, " ")
}

func parseConfig(filepath string) ([]list.Item, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var profiles []list.Item

	var currProfile *profile
	for _, line := range strings.Split(string(content), "\n") {
		line := strings.TrimSpace(line)

		if strings.HasPrefix(line, "[") {
			if currProfile != nil {
				profiles = append(profiles, list.Item(currProfile))
			}

			var name string
			lineParts := strings.Split(line, " ")
			if len(lineParts) < 2 {
				name = line[1 : len(line)-1]
			} else {
				name = lineParts[1][:len(lineParts[1])-1]
			}
			currProfile = &profile{name: name}
			continue
		}

		ei := strings.Index(line, "=")
		if ei == -1 {
			continue
		}

		name, value := strings.TrimSpace(line[:ei]), strings.TrimSpace(line[ei+1:])

		switch name {
		case configKeyRegion:
			currProfile.region = value
		case configKeyAccountID:
			currProfile.accountID = value
		case configKeyRoleName:
			currProfile.roleName = value
		}
	}
	if currProfile != nil {
		profiles = append(profiles, list.Item(currProfile))
	}

	return profiles, nil
}

type state int

const (
	stateListing state = iota
	stateDisplayLoginCmd
)

type model struct {
	l       list.Model
	s       state
	selItem list.Item
	h       help.Model
}

func (m model) Init() tea.Cmd {
	m.s = stateListing
	m.h = help.New()
	return nil
}

func (m model) displayLoginCmd() (tea.Model, tea.Cmd) {
	selItem := m.l.SelectedItem()
	if m.s == stateDisplayLoginCmd || selItem == nil {
		return m, nil
	}

	m.s = stateDisplayLoginCmd
	m.selItem = selItem
	return m, nil
}

func fatalInvalidState(s state) {
	log.Fatalln("invalid app state:", s)
}

func (m model) close(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.s {
	case stateDisplayLoginCmd:
		m.s = stateListing
		m.selItem = nil
		return m, nil
	case stateListing:
		var cmd tea.Cmd
		m.l, cmd = m.l.Update(msg)
		return m, cmd
	default:
		fatalInvalidState(m.s)
		return m, nil
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.h.Width = msg.Width
		m.l.SetWidth(msg.Width)
		m.l.SetHeight(msg.Height)
		return m, nil
	case tea.KeyMsg:
		keypress := msg.String()
		switch keypress {
		case keypressDisplayLoginCmd:
			return m.displayLoginCmd()
		case keypressClose:
			return m.close(msg)
		}
	}

	if m.s != stateListing {
		return m, nil
	}

	var cmd tea.Cmd
	m.l, cmd = m.l.Update(msg)
	return m, cmd
}

type keyMap struct {
	shortHelp []key.Binding
	fullHelp  [][]key.Binding
}

func (mkm keyMap) ShortHelp() []key.Binding {
	return mkm.shortHelp
}

func (mkm keyMap) FullHelp() [][]key.Binding {
	return mkm.fullHelp
}

var (
	simpleModalWinKeyMap = keyMap{
		shortHelp: []key.Binding{
			key.NewBinding(key.WithKeys(keypressClose), key.WithHelp(keypressClose, "go back")),
		},
	}

	simpleHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"})
)

func (m model) renderLoginCmd(i list.Item) string {
	p := i.(*profile)
	cmd := fmt.Sprintf("aws sso login --profile %v", p.name)

	helpText := m.h.View(simpleModalWinKeyMap)
	h := simpleHelpStyle.Render(helpText)

	return fmt.Sprintf("%v\n\n%v", cmd, h)
}

func (m model) View() string {
	switch m.s {
	case stateDisplayLoginCmd:
		return m.renderLoginCmd(m.selItem)
	case stateListing:
		return m.l.View()
	default:
		fatalInvalidState(m.s)
		return ""
	}
}

func main() {
	homeDir, exists := os.LookupEnv(envVarHomeDir)
	if !exists {
		log.Fatalln("HOME env var is unreachable")
	}

	confFilepath := path.Join(homeDir, pathAWSConfigFile)
	profiles, err := parseConfig(confFilepath)
	if err != nil {
		log.Fatalln("load config:", err)
	}

	l := list.New(profiles, list.NewDefaultDelegate(), displayWidthDefault, displayHeightDefault)
	l.Title = "Your AWS Profiles"
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{key.NewBinding(
			key.WithKeys(keypressDisplayLoginCmd),
			key.WithHelp(keypressDisplayLoginCmd, "view login cmd"),
		)}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{key.NewBinding(
			key.WithKeys(keypressDisplayLoginCmd),
			key.WithHelp(keypressDisplayLoginCmd, "render login cmd with the selected profile"),
		)}
	}

	m := model{l: l}

	p := tea.NewProgram(m)
	if _, err = p.Run(); err != nil {
		log.Fatalln("run bubble tea program:", err)
	}
}
