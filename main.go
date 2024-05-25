package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
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

type model struct {
	l list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.l.SetWidth(msg.Width)
		m.l.SetHeight(msg.Height)
		return m, nil
	}

	var cmd tea.Cmd
	m.l, cmd = m.l.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.l.View()
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

	m := model{l: l}

	p := tea.NewProgram(m)
	if _, err = p.Run(); err != nil {
		log.Fatalln("run bubble tea program:", err)
	}
}
