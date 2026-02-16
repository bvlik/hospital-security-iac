package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type LocalUser struct {
	Name    string `json:"Name"`
	Enabled bool   `json:"Enabled"`
}

type LocalGroup struct {
	Name string `json:"Name"`
}

type WinEvent struct {
	Id          int    `json:"Id"`
	TimeCreated string `json:"TimeCreated"`
	Message     string `json:"Message"`
	Target      string `json:"Target"`
	RecordId    int64  `json:"RecordId"`
}

var (
	app        *tview.Application
	userTable  *tview.Table
	groupTable *tview.Table
	logView    *tview.Table
	seenLogs   map[int64]bool
)

func main() {
	app = tview.NewApplication()
	seenLogs = make(map[int64]bool)

	header := tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("🛡️ SOC MONITORING - HOSPITAL 🛡️").SetBackgroundColor(tcell.ColorGreen).SetTextColor(tcell.ColorBlack)
	userTable = tview.NewTable().SetBorders(true)
	userTable.SetTitle(" 👤 USERS ").SetTitleColor(tcell.ColorAqua).SetBorder(true)
	groupTable = tview.NewTable().SetBorders(true)
	groupTable.SetTitle(" 👥 GROUPS ").SetTitleColor(tcell.ColorPurple).SetBorder(true)
	logView = tview.NewTable().SetBorders(false)
	logView.SetTitle(" 📜 AUDIT TRAIL ").SetTitleColor(tcell.ColorYellow).SetBorder(true)

	leftFlex := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(userTable, 0, 1, false).AddItem(groupTable, 0, 1, false)
	mainFlex := tview.NewFlex().AddItem(leftFlex, 0, 1, false).AddItem(logView, 0, 2, false)
	grid := tview.NewFlex().SetDirection(tview.FlexRow).AddItem(header, 1, 1, false).AddItem(mainFlex, 0, 1, false)

	go func() {
		for {
			users, groups := fetchUsers(), fetchGroups()
			app.QueueUpdateDraw(func() { updateUserTable(users); updateGroupTable(groups) })
			time.Sleep(2 * time.Second)
		}
	}()

	go func() {
		for {
			logs := fetchNewLogs()
			app.QueueUpdateDraw(func() { prependLogs(logs) })
			time.Sleep(1 * time.Second)
		}
	}()

	if err := app.SetRoot(grid, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func fetchUsers() []LocalUser {
	output, _ := runPS("Get-LocalUser | Select-Object Name, Enabled | ConvertTo-Json -Compress")
	var users []LocalUser
	parseJSONList(output, &users)
	return users
}

func fetchGroups() []LocalGroup {
	output, _ := runPS("Get-LocalGroup | Select-Object Name | ConvertTo-Json -Compress")
	var groups []LocalGroup
	parseJSONList(output, &groups)
	return groups
}

func fetchNewLogs() []WinEvent {
	// Targeted extraction of security events related to accounts and groups
	psCmd := `[Console]::OutputEncoding = [System.Text.Encoding]::UTF8; Get-WinEvent -FilterHashTable @{LogName='Security'; Id=4720,4726,4732,4733} -MaxEvents 5 -ErrorAction SilentlyContinue | Select-Object Id, @{N='TimeCreated';E={$_.TimeCreated.ToString('HH:mm:ss')}}, Message, RecordId | ConvertTo-Json -Depth 1 -Compress`
	output, _ := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-Command", psCmd).CombinedOutput()

	var events, newEvents []WinEvent
	parseJSONList(output, &events)

	for i := len(events) - 1; i >= 0; i-- {
		e := events[i]
		if e.Target == "" {
			lines := strings.Split(e.Message, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if (strings.Contains(line, "Compte cible") || strings.Contains(line, "Account Name") || strings.Contains(line, "Membre")) && !strings.Contains(line, "$") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 && len(strings.TrimSpace(parts[1])) > 1 {
						e.Target = strings.TrimSpace(parts[1])
						break
					}
				}
			}
		}
		if e.Target == "" {
			e.Target = "Detected action"
		}

		if !seenLogs[e.RecordId] {
			seenLogs[e.RecordId] = true
			newEvents = append(newEvents, e)
		}
	}
	return newEvents
}

func updateUserTable(users []LocalUser) {
	userTable.Clear()
	for i, u := range users {
		color := tcell.ColorWhite
		if strings.Contains(strings.ToLower(u.Name), "admin") {
			color = tcell.ColorRed
		}
		status, stColor := " ✔ ", tcell.ColorGreen
		if !u.Enabled {
			status, stColor, color = " X ", tcell.ColorDarkGray, tcell.ColorGray
		}
		userTable.SetCell(i, 0, tview.NewTableCell(" "+u.Name).SetTextColor(color))
		userTable.SetCell(i, 1, tview.NewTableCell(status).SetTextColor(stColor))
	}
}

func updateGroupTable(groups []LocalGroup) {
	groupTable.Clear()
	for i, g := range groups {
		color := tcell.ColorWhite
		if strings.HasPrefix(g.Name, "G_") {
			color = tcell.ColorYellow
		}
		groupTable.SetCell(i, 0, tview.NewTableCell(" "+g.Name).SetTextColor(color))
	}
}

func prependLogs(events []WinEvent) {
	for _, e := range events {
		logView.InsertRow(0)
		actionStr, actionColor := " EVENT ", tcell.ColorGray
		switch e.Id {
		case 4720:
			actionStr, actionColor = " CREATED USER ", tcell.ColorGreen
		case 4726:
			actionStr, actionColor = " DELETED USER ", tcell.ColorRed
		case 4732:
			actionStr, actionColor = " MEMBER ADDED ", tcell.ColorYellow
		case 4733:
			actionStr, actionColor = " MEMBER REMOVED ", tcell.ColorOrange
		}
		logView.SetCell(0, 0, tview.NewTableCell(fmt.Sprintf(" %s ", e.TimeCreated)).SetTextColor(tcell.ColorGray))
		logView.SetCell(0, 1, tview.NewTableCell(actionStr).SetTextColor(tcell.ColorBlack).SetBackgroundColor(actionColor))
		logView.SetCell(0, 2, tview.NewTableCell(" ➜ ").SetTextColor(tcell.ColorWhite))
		logView.SetCell(0, 3, tview.NewTableCell(fmt.Sprintf(" %s ", e.Target)).SetTextColor(tcell.ColorWhite).SetAttributes(tcell.AttrBold))
	}
	if logView.GetRowCount() > 50 {
		logView.RemoveRow(50)
	}
}

func runPS(cmdStr string) ([]byte, error) {
	return exec.Command("powershell", "-NoProfile", "-Command", fmt.Sprintf("[Console]::OutputEncoding = [System.Text.Encoding]::UTF8; %s", cmdStr)).CombinedOutput()
}

func parseJSONList(data []byte, target interface{}) {
	str := strings.TrimSpace(string(data))
	if str == "" {
		return
	}
	if !strings.HasPrefix(str, "[") {
		str = "[" + str + "]"
	}
	json.Unmarshal([]byte(str), target)
}
