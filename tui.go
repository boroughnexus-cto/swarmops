package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ─── Styles ──────────────────────────────────────────────────────────────────

var (
	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#15a8a8"))

	selectedLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Underline(true).
				Foreground(lipgloss.Color("#15a8a8"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	statusRunning = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")).Render("●")
	statusStopped = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("○")
	statusAPI     = lipgloss.NewStyle().Foreground(lipgloss.Color("#fe5d26")).Render("◆")

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#023d60")).
			Padding(0, 1)

	topBarStyle = lipgloss.NewStyle().
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("241")).
			Padding(0, 1)

	topBarTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#15a8a8"))
)

const headerHeight = 0 // top bar integrated into sidebar

// sidebarWidth is the default (and adjustable) sidebar outer width in cells.
// contentWidth = terminalWidth - (sidebarWidth + 2) where 2 = separator/border.
const defaultSidebarWidth = 24

// ─── Sidebar item: unified type for sessions + pool slots ────────────────────

type sidebarItemKind int

const (
	itemSession sidebarItemKind = iota
	itemPoolSlot
)

type sidebarItem struct {
	kind      sidebarItemKind
	label     string
	indicator string
	// Session fields
	sessionID       string
	tmuxSession     string
	status          string
	activity        string // "stopped", "working", "awaiting_input", "idle"
	mission         string // optional mission statement
	directory       string // working directory for session restart
	claudeSessionID string // Claude session ID for resume
	profile         string // happier backend profile (empty = anthropic default)
	// Pool slot fields
	slotID   string
	model    string // claude model override (session) or pool model name (pool slot)
	state    string // idle, busy, starting, dead
	requests int64
	costUSD  float64
	alive    bool
}

// ─── Messages ────────────────────────────────────────────────────────────────

type tickMsg time.Time         // fast animation tick (150ms)
type dataTickMsg time.Time     // slow data refresh tick (2s)
type activityTickMsg time.Time // activity detection tick (1s)
type flashClearMsg struct{}    // auto-clear flash message
type quotaMsg struct{ data *QuotaData }
type sessionsMsg []Session
type terminalMsg string
type itemsMsg struct {
	items    []sidebarItem
	captures []sessionCapture
}
type activityCaptureMsg struct {
	captures []sessionCapture
}
type profileRestartDoneMsg struct {
	sessionID    string
	sessionLabel string
	profileLabel string
	happierFound bool // true if new happier session ID was discovered
}

// ─── Model ───────────────────────────────────────────────────────────────────

type tuiMode int

const (
	modePassthrough tuiMode = iota
	modeNewName
	modeNewDir
	modeNewMission
	modeNewModel
	modePlaneIssues
	modeIcingaAlerts
	modePopupAction
	modeRename
	modeEditMission
	modeFeedbackType
	modeFeedbackText
	modeAuditLog
	modeEditProfile // change profile and restart session
	modeEditDir     // change working directory for a session
)

// Spawner abstracts session creation for testability.
type Spawner interface {
	Spawn(ctx context.Context, name, dir string, mission *string, model, profile string, envOverrides map[string]string) (*Session, error)
}

// defaultSpawner calls the real spawnSession function.
type defaultSpawner struct{}

func (defaultSpawner) Spawn(ctx context.Context, name, dir string, mission *string, model, profile string, envOverrides map[string]string) (*Session, error) {
	return spawnSession(ctx, name, dir, mission, model, profile, envOverrides)
}

type tuiModel struct {
	items  []sidebarItem
	cursor int
	mode   tuiMode

	// Right pane
	vp      viewport.Model
	vpReady bool

	// New session wizard
	newNameInput    textinput.Model
	newDirInput     textinput.Model
	newMissionInput textinput.Model
	newModel        int    // 0=default, 1=haiku, 2=sonnet, 3=opus, 4=deepseek, 5=openai
	// Edit profile / edit directory inputs
	editProfileIdx int           // index in profileOptions slice
	editDirInput   textinput.Model

	// Pool section display
	poolExpanded bool // expanded in sidebar; default false (collapsed, SWM-49)

	// Per-session activity state for diff detection and 1-tick damper
	activityStates    map[string]*activityState
	activityInflight  bool // true while a captureActivityCmd is running

	// Sessions currently undergoing a profile-switch restart (guards against concurrent restarts)
	restartingSessionIDs map[string]bool

	// Terminal content cache
	termContent string

	// Pre-rendered content for the right pane (set in Update, read in View)
	contentCache string

	// Terminal size
	w, h int

	// Last resize dimensions — guard against redundant tmux resize-window calls
	resizeW, resizeH int
	// resizeDone is closed by resizeTmuxSessions when the goroutine finishes.
	// Tests wait on this to avoid data races with the concurrent goroutine.
	resizeDone chan struct{}

	// Adjustable sidebar width (default 24, range 18–40)
	sidebarWidth int

	// Status message
	flash string

	// Popup data
	planeIssues    []planeIssue
	icingaProblems []icingaProblem
	auditEvents    []ManagedSessionEvent
	auditScrollback string // scrollback for selected audit event's session
	popupErr       string
	popupCursor    int
	planeReqID     uint64 // incremented on each fetch; stale responses ignored
	icingaReqID    uint64

	// Popup filter & sort
	popupFilter       textinput.Model
	popupFilterActive bool
	popupSortMode     int  // 0=default, 1, 2 — meaning depends on popup type
	popupTriageMode   int  // Plane triage preset: 0=all, 1=started, 2=high+urgent, 3=backlog
	icingaGroupByHost bool // Icinga: group problems by host
	planeStates       map[string]string // state group → state ID for transitions

	// Action picker (modePopupAction)
	actionTarget    string        // display label for the selected item
	actionPrompt    string        // text to inject into the session
	actionPrevMode  tuiMode       // mode to return to on Esc
	actionSessions  []sidebarItem // running sessions to choose from
	actionCursor    int           // cursor in action picker (sessions + "new" option)
	actionChosenIdx int           // which session was chosen (-1 = not yet, len(actionSessions) = new)
	actionCtxCursor int           // cursor in context picker during dispatch

	// Scroll state
	userScrolled bool

	// Animation frame (cycles on tick)
	animFrame int

	// Quota/usage data from quota-proxy
	quota *QuotaData

	// Rename session
	renameInput textinput.Model

	// Feedback submission
	feedbackInput    textinput.Model
	feedbackType     int // 0=bug, 1=feature
	feedbackSnapshot string  // TUI state captured at Alt+F press
	feedbackPrevMode tuiMode // mode to return to after feedback cancel

	// Dependency injection for testing
	spawner Spawner

	// HTTP client for backend API (client mode)
	api swarmClient
}

func initialModel(api swarmClient) tuiModel {
	ni := textinput.New()
	ni.Placeholder = "Session name"
	ni.CharLimit = 64

	di := textinput.New()
	di.Placeholder = "Working directory"
	di.CharLimit = 256
	di.SetValue(os.Getenv("HOME"))

	mi := textinput.New()
	mi.Placeholder = "Mission (optional, enter to skip)"
	mi.CharLimit = 256

	fi := textinput.New()
	fi.Placeholder = "filter..."
	fi.CharLimit = 128

	ri := textinput.New()
	ri.Placeholder = "New name"
	ri.CharLimit = 64

	fi2 := textinput.New()
	fi2.Placeholder = "Describe the bug or feature..."
	fi2.CharLimit = 256

	edi := textinput.New()
	edi.Placeholder = "Working directory"
	edi.CharLimit = 256

	var spawner Spawner
	if api != nil {
		spawner = api
	} else {
		spawner = defaultSpawner{}
	}

	return tuiModel{
		mode:            modePassthrough,
		newNameInput:    ni,
		newDirInput:     di,
		newMissionInput: mi,
		popupFilter:     fi,
		renameInput:     ri,
		feedbackInput:   fi2,
		editDirInput:    edi,
		activityStates:       make(map[string]*activityState),
		restartingSessionIDs: make(map[string]bool),
		spawner:              spawner,
		api:                  api,
		sidebarWidth:         loadSidebarWidth(),
	}
}

func (m tuiModel) Init() tea.Cmd {
	return tea.Batch(tickCmd(), dataTickCmd(), activityTickCmd(), loadItemsCmd(m.api))
}

func tickCmd() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func dataTickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return dataTickMsg(t)
	})
}

func activityTickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return activityTickMsg(t)
	})
}

// captureActivityCmd captures tmux panes for all running sessions without reloading from DB.
func captureActivityCmd(items []sidebarItem) tea.Cmd {
	return func() tea.Msg {
		var captures []sessionCapture
		for _, item := range items {
			if item.kind == itemSession && item.status == "running" {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				out, err := exec.CommandContext(ctx, "tmux", "capture-pane", "-p", "-S", "-100", "-t", item.tmuxSession).Output()
				cancel()
				cap := sessionCapture{tmuxSession: item.tmuxSession, alive: err == nil}
				if err == nil {
					cap.capture = string(out)
				}
				captures = append(captures, cap)
			}
		}
		return activityCaptureMsg{captures: captures}
	}
}

func flashClearCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return flashClearMsg{}
	})
}

// sessionCapture holds raw tmux capture for a session, captured in the command goroutine.
// Classification happens in the Update loop where activityStates is safely accessed.
type sessionCapture struct {
	tmuxSession string
	capture     string // raw tmux capture-pane output (empty if stopped/failed)
	alive       bool   // whether tmux capture succeeded
}

// loadItemsCmd returns a tea.Cmd that builds the unified sidebar list.
// Captures raw tmux pane content but does NOT classify activity (that happens in Update
// via applyActivityClassification to avoid sharing the activityStates map with goroutines).
func loadQuotaCmd(api swarmClient) tea.Cmd {
	return func() tea.Msg {
		if api == nil {
			return quotaMsg{}
		}
		data, _ := api.quota()
		return quotaMsg{data: data}
	}
}

func loadItemsCmd(api swarmClient) tea.Cmd {
	return func() tea.Msg {
		var sessions []Session
		if api != nil {
			sessions, _ = api.listSessions()
		} else {
			ctx := context.Background()
			refreshSessionStatuses(ctx)
			sessions, _ = listSessions(ctx)
		}

		var items []sidebarItem
		var captures []sessionCapture

		for _, s := range sessions {
			activity := "stopped"
			indicator := statusStopped
			if s.Status == "running" {
				indicator = statusRunning
				// Capture tmux content in the goroutine (safe — no shared state)
				out, err := exec.Command("tmux", "capture-pane", "-p", "-S", "-100", "-t", s.TmuxSession).Output()
				cap := sessionCapture{tmuxSession: s.TmuxSession, alive: err == nil}
				if err == nil {
					cap.capture = string(out)
				}
				captures = append(captures, cap)
				activity = "pending" // placeholder — classified in Update
			}
			mission := ""
			if s.Mission != nil {
				mission = *s.Mission
			}
			items = append(items, sidebarItem{
				kind:            itemSession,
				label:           s.Name,
				indicator:       indicator,
				sessionID:       s.ID,
				tmuxSession:     s.TmuxSession,
				status:          s.Status,
				activity:        activity,
				mission:         mission,
				model:           s.Model,
				profile:         s.Profile,
				directory:       s.Directory,
				claudeSessionID: func() string { if s.ClaudeSessionID != nil { return *s.ClaudeSessionID }; return "" }(),
			})
		}

		var poolData map[string]interface{}
		if api != nil {
			poolData, _ = api.poolStatus()
		} else if globalPool != nil {
			poolData = globalPool.Status()
		}

		if poolData != nil {
			if models, ok := poolData["models"].(map[string]interface{}); ok {
				// Sort model names for stable sidebar order
				modelNames := make([]string, 0, len(models))
				for name := range models {
					modelNames = append(modelNames, name)
				}
				sort.Strings(modelNames)
				for _, model := range modelNames {
					info := models[model]
					if minfo, ok := info.(map[string]interface{}); ok {
						// Handle slots as []interface{} (JSON unmarshal) or []map[string]interface{} (in-process)
						var slotMaps []map[string]interface{}
						if typed, ok := minfo["slots"].([]map[string]interface{}); ok {
							slotMaps = typed
						} else if raw, ok := minfo["slots"].([]interface{}); ok {
							for _, r := range raw {
								if m, ok := r.(map[string]interface{}); ok {
									slotMaps = append(slotMaps, m)
								}
							}
						}
						// Sort slots by ID for stable order
						sort.Slice(slotMaps, func(i, j int) bool {
							a, _ := slotMaps[i]["id"].(string)
							b, _ := slotMaps[j]["id"].(string)
							return a < b
						})
						for _, slot := range slotMaps {
							sid, _ := slot["id"].(string)
							state, _ := slot["state"].(string)
							// JSON numbers are float64; in-process are int64
							reqs := toInt64(slot["requests"])
							cost, _ := slot["cost_usd"].(float64)
							alive, _ := slot["alive"].(bool)

							ind := statusAPI
							if state == "starting" {
								ind = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffd700")).Render("↺")
							} else if !alive || state == "dead" {
								ind = statusStopped
							}

							short := modelShortName(model)
							items = append(items, sidebarItem{
								kind:      itemPoolSlot,
								label:     fmt.Sprintf("[api] %s", short),
								indicator: ind,
								slotID:   sid,
								model:    model,
								state:    state,
								requests: reqs,
								costUSD:  cost,
								alive:    alive,
							})
						}
					}
				}
			}
		}

		return itemsMsg{items: items, captures: captures}
	}
}

// toInt64 converts a value that may be int64 (in-process) or float64 (JSON) to int64.
func toInt64(v interface{}) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case float64:
		return int64(n)
	case json.Number:
		i, _ := n.Int64()
		return i
	default:
		return 0
	}
}

// restartSessionWithProfileCmd runs the kill-sleep-restart sequence in a
// goroutine (tea.Cmd) so the Bubbletea event loop is never blocked.
func restartSessionWithProfileCmd(tmuxSess, sessionID, name string, happierParts []string, profileLabel string) tea.Cmd {
	return func() tea.Msg {
		pre := listHappierSessionIDs()
		exec.Command("tmux", "send-keys", "-t", tmuxSess, "C-c").Run()
		time.Sleep(200 * time.Millisecond)
		exec.Command("tmux", "send-keys", "-t", tmuxSess, "C-c").Run()
		time.Sleep(200 * time.Millisecond)
		exec.Command("tmux", "send-keys", "-t", tmuxSess, "exit", "Enter").Run()
		time.Sleep(500 * time.Millisecond)
		exec.Command("tmux", "send-keys", "-t", tmuxSess, strings.Join(happierParts, " "), "Enter").Run()
		found := false
		if newID := discoverNewHappierSession(pre, 15*time.Second); newID != "" {
			found = true
			updateClaudeSessionID(context.Background(), sessionID, newID)
			setHappierTitle(newID, name)
		}
		return profileRestartDoneMsg{sessionID: sessionID, sessionLabel: name, profileLabel: profileLabel, happierFound: found}
	}
}

func loadTerminal(tmuxName string) tea.Cmd {
	return func() tea.Msg {
		content, err := captureTerminal(tmuxName)
		if err != nil {
			return terminalMsg("(error: " + err.Error() + ")")
		}
		return terminalMsg(content)
	}
}

// Context fetching and MCP client helpers are in mcp_client.go

// ─── Update ──────────────────────────────────────────────────────────────────

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Non-blocking poll for externally queued keys (from swop_tui_send_key / POST /api/tui/key).
	// Injects the key as a synthetic tea.KeyMsg into the current Update cycle.
	// tea.KeyMsg.String() produces "alt+shift+left" from {Type: KeyShiftLeft, Alt: true}.
	if m.api != nil {
		if key, ok := m.api.pollTUIKey(); ok {
			km := syntheticKeyMsg(key)
			if km.Type != 0 {
				updated, _ := m.handleKey(km)
				if u, ok := updated.(tuiModel); ok {
					m = u
				}
			}
		}
	}

	// Push rendered state after every Update cycle (via defer so all return paths are covered).
	// Capture api pointer at top; View() is called via closure at return time so it uses the final state.
	api := m.api
	defer func() {
		if api != nil {
			api.pushTUIState(m.View())
		}
	}()

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.w = msg.Width
		m.h = msg.Height
		contentWidth := m.w - (m.sidebarWidth + 2)
		if contentWidth < 20 {
			contentWidth = 20
		}
		// DEBUG: log EVERY resize event to trace width-flip and startup dims
		log.Printf("[DEBUG] WindowSizeMsg: w=%d h=%d contentWidth=%d sidebarWidth=%d", m.w, m.h, contentWidth, m.sidebarWidth)
		// Match sidebar: status line (2) + sidebar padding (2) + content header (2)
		contentHeight := m.h - headerHeight - 2 - 2 - 2
		if contentHeight < 5 {
			contentHeight = 5
		}
		m.vp = viewport.New(contentWidth, contentHeight)
		// Resize tmux sessions to viewport size (not including content header).
		// Called synchronously: tmux resize-window is fast (~10ms), and making it a
		// goroutine caused a data race with the test's struct copy due to Go's stack-
		// copying behaviour in race-detector builds.
		m.resizeTmuxSessions(contentWidth, contentHeight)
		m.vp.MouseWheelEnabled = true
		m.vpReady = true
		m.updateContentCache()
		return m, nil

	case tickMsg:
		// Fast tick (150ms): animation frames + terminal refresh only
		m.animFrame++
		var cmds []tea.Cmd
		cmds = append(cmds, tickCmd())
		if m.cursor < len(m.items) {
			item := m.items[m.cursor]
			if item.kind == itemSession && item.status == "running" {
				cmds = append(cmds, loadTerminal(item.tmuxSession))
			}
		}
		return m, tea.Batch(cmds...)

	case activityTickMsg:
		// 1s tick: capture tmux panes for activity classification only (no DB reload)
		// Skip if a previous capture is still in flight to prevent overlap/stale results
		if m.activityInflight {
			return m, activityTickCmd()
		}
		m.activityInflight = true
		return m, tea.Batch(activityTickCmd(), captureActivityCmd(m.items))

	case activityCaptureMsg:
		m.activityInflight = false
		// Classify activity from the 1s capture tick
		for i := range m.items {
			item := &m.items[i]
			if item.kind != itemSession || item.status != "running" {
				continue
			}
			for _, cap := range msg.captures {
				if cap.tmuxSession == item.tmuxSession {
					if !cap.alive {
						item.activity = "stopped"
					} else {
						st, ok := m.activityStates[cap.tmuxSession]
						if !ok {
							st = &activityState{}
							m.activityStates[cap.tmuxSession] = st
						}
						item.activity = classifyActivity(cap.capture, st)
					}
					break
				}
			}
		}
		return m, nil

	case dataTickMsg:
		// Slow tick (2s): HTTP data refresh (sessions, pool status, quota)
		return m, tea.Batch(dataTickCmd(), loadItemsCmd(m.api), loadQuotaCmd(m.api))

	case quotaMsg:
		m.quota = msg.data
		return m, nil

	case flashClearMsg:
		m.flash = ""
		return m, nil

	case profileRestartDoneMsg:
		delete(m.restartingSessionIDs, msg.sessionID)
		if msg.happierFound {
			m.flash = fmt.Sprintf("Restarted %s with profile: %s", msg.sessionLabel, msg.profileLabel)
		} else {
			m.flash = fmt.Sprintf("Restarted %s (profile: %s, session ID not synced)", msg.sessionLabel, msg.profileLabel)
		}
		return m, flashClearCmd()

	case itemsMsg:
		// Classify activity in the Update loop (single-threaded) using captures from the command
		for i := range msg.items {
			item := &msg.items[i]
			if item.kind == itemSession && item.activity == "pending" {
				// Find matching capture
				for _, cap := range msg.captures {
					if cap.tmuxSession == item.tmuxSession {
						if !cap.alive {
							item.activity = "stopped"
						} else {
							st, ok := m.activityStates[cap.tmuxSession]
							if !ok {
								st = &activityState{}
								m.activityStates[cap.tmuxSession] = st
							}
							item.activity = classifyActivity(cap.capture, st)
						}
						break
					}
				}
			}
		}
		m.items = msg.items
		if m.cursor >= len(m.items) && len(m.items) > 0 {
			m.cursor = len(m.items) - 1
		}
		m.updateContentCache()
		// Resize tmux sessions to match content pane on data refresh
		if m.w > 0 {
			contentWidth := m.w - (m.sidebarWidth + 2)
			if contentWidth < 20 {
				contentWidth = 20
			}
			contentHeight := m.h - headerHeight - 2 - 2 - 2
			if contentHeight < 5 {
				contentHeight = 5
			}
			// Called synchronously: goroutine caused data races in tests (see line 571)
			m.resizeTmuxSessions(contentWidth, contentHeight)
		}
		return m, nil

	case terminalMsg:
		m.termContent = string(msg)
		m.contentCache = m.termContent
		if m.vpReady {
			m.vp.SetContent(m.contentCache)
			if !m.userScrolled {
				m.vp.GotoBottom()
			}
		}
		return m, nil

	case planeIssuesMsg:
		if m.mode == modePlaneIssues && msg.reqID == m.planeReqID {
			m.planeIssues = msg.issues
			m.popupErr = ""
			if m.popupCursor >= len(m.planeIssues) {
				m.popupCursor = max(0, len(m.planeIssues)-1)
			}
		}
		return m, nil

	case icingaProblemsMsg:
		if m.mode == modeIcingaAlerts && msg.reqID == m.icingaReqID {
			m.icingaProblems = msg.problems
			m.popupErr = ""
			if m.popupCursor >= len(m.icingaProblems) {
				m.popupCursor = max(0, len(m.icingaProblems)-1)
			}
		}
		return m, nil

	case auditEventsMsg:
		if m.mode == modeAuditLog {
			m.auditEvents = msg.events
			m.popupErr = ""
			m.popupCursor = 0
			// Fetch scrollback for the first event's session if any
			if len(msg.events) > 0 {
				return m, fetchAuditScrollback(msg.events[0].SessionID)
			}
		}
		return m, nil

	case auditScrollbackMsg:
		if m.mode == modeAuditLog {
			m.auditScrollback = msg.content
		}
		return m, nil

	case popupErrMsg:
		if (msg.source == "plane" && m.mode == modePlaneIssues && msg.reqID == m.planeReqID) ||
			(msg.source == "icinga" && m.mode == modeIcingaAlerts && msg.reqID == m.icingaReqID) {
			m.popupErr = msg.text
		}

	case planeStatesMsg:
		m.planeStates = msg.states
		return m, nil

	case popupActionDoneMsg:
		m.flash = msg.flash
		// Refresh data after write action
		if m.mode == modePlaneIssues {
			m.planeIssues = nil
			m.planeReqID++
			return m, tea.Batch(flashClearCmd(), fetchPlaneIssues(m.planeReqID, m.api))
		}
		if m.mode == modeIcingaAlerts {
			m.icingaProblems = nil
			m.icingaReqID++
			return m, tea.Batch(flashClearCmd(), fetchIcingaProblems(m.icingaReqID, m.api))
		}
		return m, flashClearCmd()

	case tea.MouseMsg:
		if m.mode == modePassthrough && m.vpReady {
			oldOffset := m.vp.YOffset
			var cmd tea.Cmd
			m.vp, cmd = m.vp.Update(msg)
			if m.vp.YOffset != oldOffset {
				m.userScrolled = true
			}
			return m, cmd
		}

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m tuiModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch m.mode {
	case modePassthrough:
		switch key {
		case "alt+a":
			if m.cursor > 0 {
				m.cursor--
				m.flash = ""
				m.userScrolled = false
				m.updateContentCache()
			}
			return m, nil
		case "alt+z":
			if m.cursor < len(m.items)-1 {
				m.cursor++
				m.flash = ""
				m.userScrolled = false
				m.updateContentCache()
			}
			return m, nil
		case "alt+n":
			m.mode = modeNewName
			m.newNameInput.SetValue("")
			m.newNameInput.Focus()
			m.flash = "New session — enter name (esc to cancel)"
			return m, textinput.Blink
		case "alt+d":
			if m.cursor < len(m.items) && m.items[m.cursor].kind == itemSession {
				item := m.items[m.cursor]
				if m.api != nil {
					m.api.deleteSession(item.sessionID)
				} else {
					deleteSession(context.Background(), item.sessionID)
				}
				m.flash = fmt.Sprintf("✓ Deleted %s", item.label)
				return m, tea.Batch(loadItemsCmd(m.api), flashClearCmd())
			}
			return m, nil
		case "alt+s":
			if m.cursor < len(m.items) && m.items[m.cursor].kind == itemSession {
				item := m.items[m.cursor]
				if item.tmuxSession != "" {
					if item.status == "running" {
						// Running: send Ctrl+C to interrupt
						exec.Command("tmux", "send-keys", "-t", item.tmuxSession, "C-c").Run()
						m.flash = fmt.Sprintf("Sent interrupt to %s (Alt+Shift+S to kill & restart)", item.label)
					} else {
						// Stopped: restart claude — recreate tmux session if needed
						if !isTmuxAlive(item.tmuxSession) {
							dir := item.directory
							if dir == "" {
								if h, err := os.UserHomeDir(); err == nil {
									dir = h
								}
							}
							cArgs := resumeClaudeCmd(item.claudeSessionID, item.model)
							args := append([]string{"new-session", "-d", "-s", item.tmuxSession, "-c", dir, "-x", "200", "-y", "50", "--"}, cArgs...)
							if out, err := exec.Command("tmux", args...).CombinedOutput(); err != nil {
								m.flash = fmt.Sprintf("Failed to recreate tmux session: %s", strings.TrimSpace(string(out)))
								return m, flashClearCmd()
							}
						}
						m.flash = fmt.Sprintf("Resuming Claude in %s", item.label)
					}
				}
				return m, flashClearCmd()
			}
			return m, nil
		case "alt+R": // Resume-reconnect — kill and resume with full history (reconnects MCPs, declines compact)
			if m.cursor < len(m.items) && m.items[m.cursor].kind == itemSession {
				item := m.items[m.cursor]
				if item.tmuxSession != "" {
					exec.Command("tmux", "send-keys", "-t", item.tmuxSession, "C-c").Run()
					time.Sleep(200 * time.Millisecond)
					exec.Command("tmux", "send-keys", "-t", item.tmuxSession, "C-c").Run()
					time.Sleep(200 * time.Millisecond)
					exec.Command("tmux", "send-keys", "-t", item.tmuxSession, "exit", "Enter").Run()
					time.Sleep(500 * time.Millisecond)
					cArgs := resumeClaudeCmd(item.claudeSessionID, item.model)
					exec.Command("tmux", "send-keys", "-t", item.tmuxSession, strings.Join(cArgs, " "), "Enter").Run()
					go compactWatcher(item.tmuxSession, 90*time.Second)
					m.flash = fmt.Sprintf("Reconnecting %s (MCPs reloading, history restored)", item.label)
				}
				return m, nil
			}
			return m, nil
		case "alt+S": // Shift variant — kill and restart claude in the session (fresh, no history)
			if m.cursor < len(m.items) && m.items[m.cursor].kind == itemSession {
				item := m.items[m.cursor]
				if item.tmuxSession != "" {
					// Kill all processes in the tmux pane, then restart claude
					exec.Command("tmux", "send-keys", "-t", item.tmuxSession, "C-c").Run()
					time.Sleep(200 * time.Millisecond)
					exec.Command("tmux", "send-keys", "-t", item.tmuxSession, "C-c").Run()
					time.Sleep(200 * time.Millisecond)
					exec.Command("tmux", "send-keys", "-t", item.tmuxSession, "exit", "Enter").Run()
					time.Sleep(500 * time.Millisecond)
					parts := []string{"happier", "--yolo"}
					parts = append(parts, profileToHappierArgs(item.profile)...)
					parts = append(parts, "--model", effectiveModel(item.model))
					preIDs := listHappierSessionIDs()
					exec.Command("tmux", "send-keys", "-t", item.tmuxSession, strings.Join(parts, " "), "Enter").Run()
					// Discover new happier ID and set title (fresh start creates a new session)
					go func(sessionID, name string, pre map[string]bool) {
						if newID := discoverNewHappierSession(pre, 15*time.Second); newID != "" {
							updateClaudeSessionID(context.Background(), sessionID, newID)
							setHappierTitle(newID, name)
						}
					}(item.sessionID, item.label, preIDs)
					m.flash = fmt.Sprintf("Restarted Claude in %s", item.label)
				}
				return m, nil
			}
			return m, nil
		case "alt+r":
			if m.cursor < len(m.items) && m.items[m.cursor].kind == itemSession {
				m.mode = modeRename
				m.renameInput.SetValue(m.items[m.cursor].label)
				m.renameInput.Focus()
				m.flash = "Rename session (esc to cancel)"
				return m, textinput.Blink
			}
			return m, nil
		case "alt+m":
			if m.cursor < len(m.items) && m.items[m.cursor].kind == itemSession {
				m.mode = modeEditMission
				m.newMissionInput.SetValue(m.items[m.cursor].mission)
				m.newMissionInput.Focus()
				m.flash = "Edit mission (esc to cancel, enter to save)"
				return m, textinput.Blink
			}
			return m, nil
		case "alt+k":
			if m.cursor < len(m.items) && m.items[m.cursor].kind == itemSession {
				item := m.items[m.cursor]
				// Set picker to current profile
				m.editProfileIdx = profileIndexFromString(item.profile)
				m.mode = modeEditProfile
				m.flash = profilePickerFlash(m.editProfileIdx)
			}
			return m, nil
		case "alt+g":
			if m.cursor < len(m.items) && m.items[m.cursor].kind == itemSession {
				item := m.items[m.cursor]
				m.editDirInput.SetValue(item.directory)
				m.editDirInput.SetCursor(len(item.directory))
				m.editDirInput.Focus()
				m.mode = modeEditDir
				m.flash = "Edit directory (tab to complete, enter to save, esc to cancel)"
				return m, textinput.Blink
			}
			return m, nil
		case "alt+f":
			// Capture TUI state before switching to feedback mode
			m.feedbackSnapshot = m.View()
			m.feedbackPrevMode = modePassthrough
			m.mode = modeFeedbackType
			m.feedbackType = 0
			m.flash = "Feedback: ←/→ Bug or Feature, Enter to continue, Esc to cancel"
			return m, nil
		case "alt+o":
			m.poolExpanded = !m.poolExpanded
			return m, nil
		case "alt+p":
			m.mode = modePlaneIssues
			m.planeIssues = nil
			m.popupErr = ""
			m.popupCursor = 0
			m.flash = ""
			m.planeReqID++
			cmds := []tea.Cmd{fetchPlaneIssues(m.planeReqID, m.api)}
			if m.planeStates == nil {
				cmds = append(cmds, fetchPlaneStates(m.api))
			}
			return m, tea.Batch(cmds...)
		case "alt+i":
			m.mode = modeIcingaAlerts
			m.icingaProblems = nil
			m.popupErr = ""
			m.popupCursor = 0
			m.flash = ""
			m.icingaReqID++
			return m, fetchIcingaProblems(m.icingaReqID, m.api)
		case "alt+q":
			return m, tea.Quit
		case "alt+l":
			m.mode = modeAuditLog
			m.auditEvents = nil
			m.auditScrollback = ""
			m.popupCursor = 0
			m.popupErr = ""
			return m, fetchAuditEvents(m.api)
		case "alt+w":
			// Close the Plane issue referenced in the current session's name (SWM-26)
			if m.cursor < len(m.items) && m.items[m.cursor].kind == itemSession {
				label := m.items[m.cursor].label
				m.flash = "Closing Plane issue for session..."
				return m, planeCloseSessionIssue(label, m.api)
			}
			return m, nil
		case "alt+e":
			// Jump to next awaiting_input session (SWM-11)
			n := len(m.items)
			for offset := 1; offset <= n; offset++ {
				idx := (m.cursor + offset) % n
				item := m.items[idx]
				if item.kind == itemSession && item.activity == "awaiting_input" {
					m.cursor = idx
					m.userScrolled = false
					m.updateContentCache()
					break
				}
			}
			return m, nil
		case "alt+b":
			// Snap viewport to bottom and resume auto-scroll
			if m.vpReady {
				m.userScrolled = false
				m.vp.GotoBottom()
			}
			return m, nil
		case "shift+alt+left", "alt+shift+left":
			if m.sidebarWidth > 18 {
				m.sidebarWidth--
				saveSidebarWidth(m.sidebarWidth)
				m.resizeTmuxSessions(m.w-(m.sidebarWidth+2), m.h)
				m.flash = fmt.Sprintf("Sidebar: %d", m.sidebarWidth)
			}
			return m, nil
		case "shift+alt+right", "alt+shift+right":
			if m.sidebarWidth < 40 {
				m.sidebarWidth++
				saveSidebarWidth(m.sidebarWidth)
				m.resizeTmuxSessions(m.w-(m.sidebarWidth+2), m.h)
				m.flash = fmt.Sprintf("Sidebar: %d", m.sidebarWidth)
			}
			return m, nil
		case "ctrl+[":
			// DEBUG: dump sidebar render output to /tmp/sidebar-debug.txt
			sidebarOut := m.renderSidebar()
			lines := strings.Split(sidebarOut, "\n")
			content := fmt.Sprintf("sidebarWidth=%d w=%d h=%d innerW=%d\nsidebar lines (%d):\n", m.sidebarWidth, m.w, m.h, m.sidebarInnerWidth(), len(lines))
			for i, l := range lines {
				content += fmt.Sprintf("  [%2d] (%3d chars) %q\n", i, len(l), l)
			}
			os.WriteFile("/tmp/sidebar-debug.txt", []byte(content), 0644)
			m.flash = "Debug written to /tmp/sidebar-debug.txt"
			return m, nil
		default:
			// DEBUG: log unknown keys to see what shift+alt+arrow produces
			log.Printf("[DEBUG] handleKey unhandled: key=%s alt=%v", key, msg.Alt)
			// Only pass keys to session items (not pool slots)
			if m.cursor < len(m.items) && m.items[m.cursor].kind == itemSession {
				m.sendKeyToSession(key)
				// Immediately refresh terminal content after sending a key
				return m, loadTerminal(m.items[m.cursor].tmuxSession)
			}
			return m, nil
		}

	case modeNewName:
		switch key {
		case "enter":
			if m.newNameInput.Value() != "" {
				m.mode = modeNewDir
				m.newDirInput.Focus()
				m.flash = "New session — enter directory (esc to cancel)"
				return m, textinput.Blink
			}
		case "esc":
			m.mode = modePassthrough
			m.flash = ""
		default:
			var cmd tea.Cmd
			m.newNameInput, cmd = m.newNameInput.Update(msg)
			return m, cmd
		}

	case modeNewDir:
		switch key {
		case "enter":
			m.mode = modeNewMission
			m.newMissionInput.SetValue("")
			m.newMissionInput.Focus()
			m.flash = "Mission statement (optional, enter to skip)"
			return m, textinput.Blink
		case "esc":
			m.mode = modePassthrough
			m.flash = ""
		case "tab":
			// Directory tab-completion
			val := m.newDirInput.Value()
			if val != "" {
				matches, _ := filepath.Glob(val + "*")
				if len(matches) == 1 {
					// Single match — complete it
					completed := matches[0]
					info, err := os.Stat(completed)
					if err == nil && info.IsDir() {
						completed += "/"
					}
					m.newDirInput.SetValue(completed)
					m.newDirInput.SetCursor(len(completed))
				} else if len(matches) > 1 {
					// Multiple matches — find common prefix
					prefix := matches[0]
					for _, match := range matches[1:] {
						for i := 0; i < len(prefix) && i < len(match); i++ {
							if prefix[i] != match[i] {
								prefix = prefix[:i]
								break
							}
						}
						if len(match) < len(prefix) {
							prefix = prefix[:len(match)]
						}
					}
					if len(prefix) > len(val) {
						m.newDirInput.SetValue(prefix)
						m.newDirInput.SetCursor(len(prefix))
					}
					// Show matches in flash
					var names []string
					for _, match := range matches {
						names = append(names, filepath.Base(match))
					}
					m.flash = strings.Join(names, "  ")
				}
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.newDirInput, cmd = m.newDirInput.Update(msg)
			return m, cmd
		}

	case modeNewMission:
		switch key {
		case "enter":
			m.newModel = 0
			m.mode = modeNewModel
			m.flash = modelPickerFlash(m.newModel)
			return m, nil
		case "esc":
			m.mode = modePassthrough
			m.flash = ""
		default:
			var cmd tea.Cmd
			m.newMissionInput, cmd = m.newMissionInput.Update(msg)
			return m, cmd
		}

	case modeNewModel:
		switch key {
		case "left", "alt+a", "h":
			if m.newModel > 0 {
				m.newModel--
				m.flash = modelPickerFlash(m.newModel)
			}
			return m, nil
		case "right", "alt+z", "l":
			if m.newModel < len(backendOptions)-1 {
				m.newModel++
				m.flash = modelPickerFlash(m.newModel)
			}
			return m, nil
		case "enter":
			m.doSpawn()
			return m, nil
		case "esc":
			m.mode = modePassthrough
			m.flash = ""
		}
		return m, nil

	case modeRename:
		switch key {
		case "enter":
			newName := m.renameInput.Value()
			if newName != "" && m.cursor < len(m.items) {
				item := m.items[m.cursor]
				if m.api != nil {
					m.api.renameSession(item.sessionID, newName)
				} else {
					renameSession(context.Background(), item.sessionID, newName)
				}
				m.flash = fmt.Sprintf("Renamed to %s", newName)
			}
			m.mode = modePassthrough
			return m, loadItemsCmd(m.api)
		case "esc":
			m.mode = modePassthrough
			m.flash = ""
		default:
			var cmd tea.Cmd
			m.renameInput, cmd = m.renameInput.Update(msg)
			return m, cmd
		}

	case modeEditMission:
		switch key {
		case "enter":
			mission := m.newMissionInput.Value()
			if m.cursor < len(m.items) {
				item := m.items[m.cursor]
				if m.api != nil {
					m.api.setMission(item.sessionID, mission)
				} else {
					updateSessionMission(context.Background(), item.sessionID, mission)
				}
				if mission == "" {
					m.flash = "Mission cleared"
				} else {
					m.flash = "Mission updated"
				}
			}
			m.mode = modePassthrough
			return m, tea.Batch(loadItemsCmd(m.api), flashClearCmd())
		case "esc":
			m.mode = modePassthrough
			m.flash = ""
		default:
			var cmd tea.Cmd
			m.newMissionInput, cmd = m.newMissionInput.Update(msg)
			return m, cmd
		}

	case modeEditProfile:
		switch key {
		case "left", "alt+a", "h":
			if m.editProfileIdx > 0 {
				m.editProfileIdx--
				m.flash = profilePickerFlash(m.editProfileIdx)
			}
			return m, nil
		case "right", "alt+z", "l":
			if m.editProfileIdx < len(backendOptions)-1 {
				m.editProfileIdx++
				m.flash = profilePickerFlash(m.editProfileIdx)
			}
			return m, nil
		case "enter":
			if m.editProfileIdx < 0 || m.editProfileIdx >= len(backendOptions) {
				m.mode = modePassthrough
				return m, nil
			}
			if m.cursor < len(m.items) && m.items[m.cursor].kind == itemSession {
				item := m.items[m.cursor]
				if m.restartingSessionIDs[item.sessionID] {
					m.flash = fmt.Sprintf("%s is already restarting", item.label)
					m.mode = modePassthrough
					return m, flashClearCmd()
				}
				opt := backendOptions[m.editProfileIdx]
				// Persist profile change; abort restart if DB write fails
				var dbErr error
				if m.api != nil {
					dbErr = m.api.updateSessionProfile(item.sessionID, opt.profile)
				} else {
					dbErr = updateSessionProfile(context.Background(), item.sessionID, opt.profile)
				}
				if dbErr != nil {
					m.flash = fmt.Sprintf("Failed to save profile: %v", dbErr)
					m.mode = modePassthrough
					return m, flashClearCmd()
				}
				// Build restart command with new profile
				happierParts := []string{"happier", "--yolo"}
				happierParts = append(happierParts, profileToHappierArgs(opt.profile)...)
				if opt.model != "" {
					happierParts = append(happierParts, "--model", opt.model)
				} else {
					happierParts = append(happierParts, "--model", effectiveModel(""))
				}
				m.restartingSessionIDs[item.sessionID] = true
				m.flash = fmt.Sprintf("Restarting %s with profile: %s...", item.label, opt.label)
				m.mode = modePassthrough
				return m, tea.Batch(
					loadItemsCmd(m.api),
					restartSessionWithProfileCmd(item.tmuxSession, item.sessionID, item.label, happierParts, opt.label),
				)
			}
			m.mode = modePassthrough
			return m, loadItemsCmd(m.api)
		case "esc":
			m.mode = modePassthrough
			m.flash = ""
		}

	case modeEditDir:
		switch key {
		case "enter":
			newDir := m.editDirInput.Value()
			if newDir != "" && m.cursor < len(m.items) {
				item := m.items[m.cursor]
				if _, err := os.Stat(newDir); err != nil {
					m.flash = fmt.Sprintf("Directory does not exist: %s", newDir)
					return m, flashClearCmd()
				}
				if m.api != nil {
					m.api.updateSessionDirectory(item.sessionID, newDir)
				} else {
					updateSessionDirectory(context.Background(), item.sessionID, newDir)
				}
				m.flash = fmt.Sprintf("Directory updated (takes effect on next restart)")
			}
			m.mode = modePassthrough
			return m, tea.Batch(loadItemsCmd(m.api), flashClearCmd())
		case "esc":
			m.mode = modePassthrough
			m.flash = ""
		case "tab":
			val := m.editDirInput.Value()
			if val != "" {
				matches, _ := filepath.Glob(val + "*")
				if len(matches) == 1 {
					completed := matches[0]
					info, err := os.Stat(completed)
					if err == nil && info.IsDir() {
						completed += "/"
					}
					m.editDirInput.SetValue(completed)
					m.editDirInput.SetCursor(len(completed))
				}
			}
			return m, nil
		default:
			var cmd tea.Cmd
			m.editDirInput, cmd = m.editDirInput.Update(msg)
			return m, cmd
		}

	case modeFeedbackType:
		switch key {
		case "left", "right":
			m.feedbackType = 1 - m.feedbackType
		case "enter":
			m.mode = modeFeedbackText
			m.feedbackInput.SetValue("")
			m.feedbackInput.Focus()
			kinds := []string{"bug", "feature"}
			m.flash = fmt.Sprintf("Describe the %s (Enter to submit, Esc to cancel)", kinds[m.feedbackType])
			return m, textinput.Blink
		case "esc":
			m.mode = m.feedbackPrevMode
			m.flash = ""
		}

	case modeFeedbackText:
		switch key {
		case "enter":
			summary := m.feedbackInput.Value()
			if summary != "" {
				kinds := []string{"bug", "feature"}
				go submitFeedback(kinds[m.feedbackType], summary, m.api, m.feedbackSnapshot)
				m.flash = fmt.Sprintf("✓ Submitted %s: %s", kinds[m.feedbackType], summary)
				m.mode = m.feedbackPrevMode
				return m, flashClearCmd()
			}
			m.mode = m.feedbackPrevMode
			return m, nil
		case "esc":
			m.mode = m.feedbackPrevMode
			m.flash = ""
		default:
			var cmd tea.Cmd
			m.feedbackInput, cmd = m.feedbackInput.Update(msg)
			return m, cmd
		}


	case modePlaneIssues:
		filtered := filteredPlaneIssues(m)
		r := handlePopupKeyShared(&m, msg, len(filtered), planeSortLabels)
		if r.action == "enter" {
			if m.popupCursor < len(filtered) {
				issue := filtered[m.popupCursor]
				m.openActionPicker(issue.Identifier+" "+issue.Title, planeIssuePrompt(issue), modePlaneIssues)
			}
		} else if r.action == "refresh" {
			m.planeIssues = nil
			m.planeReqID++
			return m, fetchPlaneIssues(m.planeReqID, m.api)
		}
		if r.handled {
			return m, r.cmd
		}
		// Plane-specific keys (not handled by shared handler)
		switch key {
		case "p": // set In Progress
			if m.popupCursor < len(filtered) {
				if m.planeStates == nil {
					m.flash = "Loading states..."
					return m, fetchPlaneStates(m.api)
				}
				issue := filtered[m.popupCursor]
				if stateID, ok := m.planeStates["started"]; ok {
					m.flash = "Setting In Progress..."
					return m, planeUpdateIssue(m.api, issue.ID, map[string]interface{}{"state": stateID})
				}
				m.flash = "No 'started' state found"
			}
		case "d": // set Done
			if m.popupCursor < len(filtered) {
				if m.planeStates == nil {
					m.flash = "Loading states..."
					return m, fetchPlaneStates(m.api)
				}
				issue := filtered[m.popupCursor]
				if stateID, ok := m.planeStates["completed"]; ok {
					m.flash = "Setting Done..."
					return m, planeUpdateIssue(m.api, issue.ID, map[string]interface{}{"state": stateID})
				}
				m.flash = "No 'completed' state found"
			}
		case "1":
			m.popupTriageMode = 1
			m.popupCursor = 0
		case "2":
			m.popupTriageMode = 2
			m.popupCursor = 0
		case "3":
			m.popupTriageMode = 3
			m.popupCursor = 0
		case "0":
			m.popupTriageMode = 0
			m.popupCursor = 0
		}

	case modeAuditLog:
		switch key {
		case "esc", "alt+l":
			m.mode = modePassthrough
		case "alt+a", "up", "k":
			if m.popupCursor > 0 {
				m.popupCursor--
				if len(m.auditEvents) > 0 {
					return m, fetchAuditScrollback(m.auditEvents[m.popupCursor].SessionID)
				}
			}
		case "alt+z", "down", "j":
			if m.popupCursor < len(m.auditEvents)-1 {
				m.popupCursor++
				if len(m.auditEvents) > 0 {
					return m, fetchAuditScrollback(m.auditEvents[m.popupCursor].SessionID)
				}
			}
		case "r":
			m.auditEvents = nil
			m.auditScrollback = ""
			m.popupCursor = 0
			m.popupErr = ""
			return m, fetchAuditEvents(m.api)
		}

	case modeIcingaAlerts:
		filtered := filteredIcingaProblems(m)
		r := handlePopupKeyShared(&m, msg, len(filtered), icingaSortLabels)
		if r.action == "enter" {
			if m.popupCursor < len(filtered) {
				problem := filtered[m.popupCursor]
				label := fmt.Sprintf("%s / %s", problem.Host, problem.Service)
				m.openActionPicker(label, icingaProblemPrompt(problem), modeIcingaAlerts)
			}
		} else if r.action == "refresh" {
			m.icingaProblems = nil
			m.icingaReqID++
			return m, fetchIcingaProblems(m.icingaReqID, m.api)
		}
		if r.handled {
			return m, r.cmd
		}
		// Icinga-specific keys
		switch key {
		case "a": // acknowledge
			if m.popupCursor < len(filtered) {
				problem := filtered[m.popupCursor]
				if problem.Acknowledged {
					m.flash = "Already acknowledged"
				} else {
					m.flash = "Acknowledging..."
					return m, icingaAcknowledge(m.api, problem.ObjectName, "Acknowledged from SwarmOps TUI")
				}
			}
		case "t": // schedule downtime (30m)
			if m.popupCursor < len(filtered) {
				problem := filtered[m.popupCursor]
				m.flash = "Scheduling 30m downtime..."
				return m, icingaScheduleDowntime(m.api, problem.ObjectName, 30*time.Minute, "Downtime from SwarmOps TUI (30m)")
			}
		case "T": // schedule downtime (2h)
			if m.popupCursor < len(filtered) {
				problem := filtered[m.popupCursor]
				m.flash = "Scheduling 2h downtime..."
				return m, icingaScheduleDowntime(m.api, problem.ObjectName, 2*time.Hour, "Downtime from SwarmOps TUI (2h)")
			}
		case "g": // toggle group by host
			m.icingaGroupByHost = !m.icingaGroupByHost
			m.popupCursor = 0
		}

	case modePopupAction:
		maxIdx := len(m.actionSessions) // last index is "new session"
		switch key {
		case "esc":
			m.mode = m.actionPrevMode
		case "alt+a", "up":
			if m.actionCursor > 0 {
				m.actionCursor--
			}
		case "alt+z", "down":
			if m.actionCursor < maxIdx {
				m.actionCursor++
			}
		case "enter":
			// Remember which session was chosen, then dispatch
			m.actionChosenIdx = m.actionCursor
			m.doDispatch(m.actionPrompt)
			m.mode = m.actionPrevMode
			return m, loadItemsCmd(m.api)
		}
	}

	return m, nil
}

// sendKeyToSession translates a Bubbletea key string to tmux send-keys.
// resizeTmuxSessions resizes all tmux session windows to match the TUI content pane.
// Skips sessions with attached clients (client size takes precedence — resize is no-op
// and would otherwise cause a SIGWINCH feedback loop that flickers the TUI width).
// Also skips sessions whose dimensions haven't changed since the last call.
func (m *tuiModel) resizeTmuxSessions(width, height int) {
	if width == m.resizeW && height == m.resizeH {
		return
	}
	log.Printf("[DEBUG] resizeTmuxSessions: w=%d h=%d (last=%d,%d)", width, height, m.resizeW, m.resizeH)
	for _, item := range m.items {
		if item.kind != itemSession || item.tmuxSession == "" {
			continue
		}
		// Skip sessions with attached clients — client terminal size takes precedence
		if out, err := exec.Command("tmux", "list-clients", "-t", item.tmuxSession).Output(); err == nil && len(out) > 0 {
			continue
		}
		exec.Command("tmux", "resize-window", "-t", item.tmuxSession,
			"-x", fmt.Sprintf("%d", width), "-y", fmt.Sprintf("%d", height)).Run()
	}
	m.resizeW = width
	m.resizeH = height
	if m.resizeDone != nil {
		close(m.resizeDone)
	}
}

func (m *tuiModel) sendKeyToSession(key string) {
	if m.cursor >= len(m.items) {
		return
	}
	item := m.items[m.cursor]
	if item.kind != itemSession || item.status != "running" {
		return
	}

	tmuxKey := key
	switch key {
	case "enter":
		tmuxKey = "Enter"
	case "tab":
		tmuxKey = "Tab"
	case "backspace":
		tmuxKey = "BSpace"
	case "delete":
		tmuxKey = "DC"
	case "up":
		tmuxKey = "Up"
	case "down":
		tmuxKey = "Down"
	case "left":
		tmuxKey = "Left"
	case "right":
		tmuxKey = "Right"
	case "home":
		tmuxKey = "Home"
	case "end":
		tmuxKey = "End"
	case "pgup":
		tmuxKey = "PPage"
	case "pgdown":
		tmuxKey = "NPage"
	case "esc":
		tmuxKey = "Escape"
	case "space":
		tmuxKey = "Space"
	case "ctrl+c":
		tmuxKey = "C-c"
	case "ctrl+l":
		tmuxKey = "C-l"
	case "ctrl+r":
		tmuxKey = "C-r"
	case "ctrl+p":
		tmuxKey = "C-p"
	case "ctrl+e":
		tmuxKey = "C-e"
	case "ctrl+w":
		tmuxKey = "C-w"
	case "ctrl+u":
		tmuxKey = "C-u"
	case "ctrl+k":
		tmuxKey = "C-k"
	}

	if len(key) == 1 {
		exec.Command("tmux", "send-keys", "-t", item.tmuxSession, "-l", key).Run()
		return
	}
	exec.Command("tmux", "send-keys", "-t", item.tmuxSession, tmuxKey).Run()
}

func (m *tuiModel) doSpawn() {
	name := m.newNameInput.Value()
	dir := m.newDirInput.Value()
	if dir == "" {
		dir = os.Getenv("HOME")
	}
	var mission *string
	if v := m.newMissionInput.Value(); v != "" {
		mission = &v
	}
	model := modelIDFromIndex(m.newModel)
	profile := profileFromIndex(m.newModel)
	envOverrides := envOverridesFromIndex(m.newModel)
	name = autoPrefixSessionName(name, model, envOverrides)
	s, err := m.spawner.Spawn(context.Background(), name, dir, mission, model, profile, envOverrides)
	if err != nil {
		m.flash = "Spawn error: " + err.Error()
	} else {
		m.flash = fmt.Sprintf("Spawned %s", s.Name)
	}
}

// doDispatch executes the dispatch action: sends prompt to a session or spawns a new one.
// Called after the context picker step. Uses m.actionChosenIdx to determine target.
func (m *tuiModel) doDispatch(prompt string) {
	if m.actionChosenIdx < len(m.actionSessions) {
		// Inject into existing session
		sess := m.actionSessions[m.actionChosenIdx]
		if sess.status == "running" {
			injectToSession(sess.tmuxSession, prompt)
			m.flash = fmt.Sprintf("Sent to %s", sess.label)
		}
	} else {
		// Spawn new session
		name := sanitizeSessionName(m.actionTarget)
		s, err := m.spawner.Spawn(context.Background(), name, os.Getenv("HOME"), nil, "", "", nil)
		if err != nil {
			m.flash = "Spawn error: " + err.Error()
		} else {
			go func() {
				time.Sleep(2 * time.Second)
				injectToSession(s.TmuxSession, prompt)
			}()
			m.flash = fmt.Sprintf("Spawned %s — injecting task", s.Name)
		}
	}
	m.mode = modePassthrough
}

// popupKeyResult holds the result of shared popup key handling.
type popupKeyResult struct {
	handled bool
	cmd     tea.Cmd
	action  string // "enter", "refresh", or ""
}

// handlePopupKeyShared processes keys common to all popup modes.
// Returns the action taken so the caller can perform popup-specific work.
func handlePopupKeyShared(m *tuiModel, msg tea.KeyMsg, filteredLen int, sortLabels []string) popupKeyResult {
	key := msg.String()

	if m.popupFilterActive {
		switch key {
		case "esc":
			m.popupFilterActive = false
			m.popupFilter.Blur()
			m.popupFilter.SetValue("")
			m.popupCursor = 0
			return popupKeyResult{handled: true}
		case "enter":
			m.popupFilterActive = false
			m.popupFilter.Blur()
			m.popupCursor = 0
			return popupKeyResult{handled: true}
		default:
			var cmd tea.Cmd
			m.popupFilter, cmd = m.popupFilter.Update(msg)
			m.popupCursor = 0
			return popupKeyResult{handled: true, cmd: cmd}
		}
	}

	switch key {
	case "q", "esc":
		m.mode = modePassthrough
		m.popupErr = ""
		m.popupFilter.SetValue("")
		m.popupSortMode = 0
		m.popupTriageMode = 0
		m.icingaGroupByHost = false
		return popupKeyResult{handled: true}
	case "alt+a", "up":
		if filteredLen > 0 && m.popupCursor > 0 {
			m.popupCursor--
		}
		return popupKeyResult{handled: true}
	case "alt+z", "down":
		if filteredLen > 0 && m.popupCursor < filteredLen-1 {
			m.popupCursor++
		}
		return popupKeyResult{handled: true}
	case "pgup":
		m.popupCursor -= 10
		if m.popupCursor < 0 {
			m.popupCursor = 0
		}
		return popupKeyResult{handled: true}
	case "pgdown":
		m.popupCursor += 10
		if m.popupCursor >= filteredLen {
			m.popupCursor = filteredLen - 1
		}
		if m.popupCursor < 0 {
			m.popupCursor = 0
		}
		return popupKeyResult{handled: true}
	case "home":
		m.popupCursor = 0
		return popupKeyResult{handled: true}
	case "end":
		if filteredLen > 0 {
			m.popupCursor = filteredLen - 1
		}
		return popupKeyResult{handled: true}
	case "/":
		m.popupFilterActive = true
		m.popupFilter.Focus()
		return popupKeyResult{handled: true, cmd: textinput.Blink}
	case "s":
		m.popupSortMode = (m.popupSortMode + 1) % len(sortLabels)
		m.popupCursor = 0
		return popupKeyResult{handled: true}
	case "enter":
		if filteredLen > 0 {
			return popupKeyResult{handled: true, action: "enter"}
		}
		return popupKeyResult{handled: true}
	case "r":
		m.popupErr = ""
		m.popupCursor = 0
		return popupKeyResult{handled: true, action: "refresh"}
	case "alt+f":
		// Capture current popup view before switching to feedback
		switch m.mode {
		case modePlaneIssues:
			m.feedbackSnapshot = renderPlanePopup(*m)
		case modeIcingaAlerts:
			m.feedbackSnapshot = renderIcingaPopup(*m)
		default:
			m.feedbackSnapshot = ""
		}
		m.feedbackPrevMode = m.mode
		m.mode = modeFeedbackType
		m.feedbackType = 0
		m.flash = "Feedback: ←/→ Bug or Feature, Enter to continue, Esc to cancel"
		return popupKeyResult{handled: true}
	}
	return popupKeyResult{}
}

// openActionPicker prepares the action picker overlay with running sessions.
func (m *tuiModel) openActionPicker(target, prompt string, prevMode tuiMode) {
	m.actionTarget = target
	m.actionPrompt = prompt
	m.actionPrevMode = prevMode
	m.actionCursor = 0
	m.mode = modePopupAction

	// Collect running sessions for the picker
	m.actionSessions = nil
	for _, item := range m.items {
		if item.kind == itemSession && item.status == "running" {
			m.actionSessions = append(m.actionSessions, item)
		}
	}
}

// sanitizeSessionName creates a safe tmux session name from a title.
func sanitizeSessionName(title string) string {
	name := strings.ToLower(title)
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, name)
	// Collapse repeated hyphens and trim
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	name = strings.Trim(name, "-")
	if len(name) > 30 {
		name = name[:30]
	}
	if name == "" {
		name = "task"
	}
	return name
}

// ─── View ────────────────────────────────────────────────────────────────────

func (m tuiModel) View() string {
	if m.w == 0 {
		return "Loading..."
	}

	// Full-screen popup modes
	switch m.mode {
	case modePlaneIssues:
		return renderPlanePopup(m)
	case modeIcingaAlerts:
		return renderIcingaPopup(m)
	case modeAuditLog:
		return renderAuditPopup(m)
	case modePopupAction:
		return renderActionPicker(m)
	}

	sidebar := m.renderSidebar()
	content := m.renderContent()

	main := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)

	var statusLine string
	switch m.mode {
	case modeNewName:
		statusLine = "Name: " + m.newNameInput.View()
	case modeNewDir:
		statusLine = "Dir: " + m.newDirInput.View()
	case modeNewMission:
		statusLine = "Mission: " + m.newMissionInput.View()
	case modeNewModel:
		statusLine = modelPickerFlash(m.newModel)
	case modeEditMission:
		statusLine = "Mission: " + m.newMissionInput.View()
	case modeRename:
		statusLine = "Rename: " + m.renameInput.View()
	case modeEditProfile:
		statusLine = profilePickerFlash(m.editProfileIdx)
	case modeEditDir:
		statusLine = "Dir: " + m.editDirInput.View()
	case modeFeedbackType:
		kinds := []string{"🐛 Bug", "✨ Feature"}
		var parts []string
		for i, k := range kinds {
			if i == m.feedbackType {
				parts = append(parts, selectedStyle.Render("> "+k))
			} else {
				parts = append(parts, dimStyle.Render("  "+k))
			}
		}
		statusLine = "Feedback: " + strings.Join(parts, "  ")
	case modeFeedbackText:
		statusLine = "Feedback: " + m.feedbackInput.View()
	default:
		if m.flash != "" {
			statusLine = dimStyle.Render(m.flash)
		} else {
			statusLine = dimStyle.Render("Alt+A/Z nav │ Alt+N new │ Alt+S start/stop │ Alt+R rename │ Alt+M mission │ Alt+K profile │ Alt+G dir │ Alt+D delete") + "\n" +
				dimStyle.Render("Alt+P plane │ Alt+I icinga │ Alt+L audit │ Alt+W close issue │ Alt+E escalations │ Alt+O pool │ Alt+F feedback │ Alt+Q quit")
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, main, statusLine)
}

func (m tuiModel) renderTopBar() string {
	barWidth := m.w - 2
	innerWidth := barWidth - 2 // account for Padding(0, 1) left+right

	// Line 1: SwarmOps (left) + time (right)
	title := topBarTitleStyle.Render("SwarmOps")
	ts := dimStyle.Render(time.Now().Format("15:04:05"))
	gap1 := innerWidth - lipgloss.Width(title) - lipgloss.Width(ts)
	if gap1 < 1 {
		gap1 = 1
	}
	line1 := title + strings.Repeat(" ", gap1) + ts

	// Line 2: session/pool summary
	running := 0
	stopped := 0
	poolSlots := 0
	poolAlive := 0
	for _, item := range m.items {
		switch item.kind {
		case itemSession:
			if item.status == "running" {
				running++
			} else {
				stopped++
			}
		case itemPoolSlot:
			poolSlots++
			if item.alive {
				poolAlive++
			}
		}
	}
	var parts []string
	if running > 0 || stopped > 0 {
		parts = append(parts, fmt.Sprintf("%d sessions (%d running)", running+stopped, running))
	}
	if poolSlots > 0 {
		parts = append(parts, fmt.Sprintf("%d/%d pool slots", poolAlive, poolSlots))
	} else {
		parts = append(parts, "pool off")
	}
	line2 := dimStyle.Render(strings.Join(parts, "  ·  "))

	content := line1 + "\n" + line2
	return topBarStyle.Width(barWidth).Render(content)
}

// sidebarStyle returns the sidebar lipgloss style with the current m.sidebarWidth.
func (m tuiModel) sidebarStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Width(m.sidebarWidth).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		Padding(1, 1)
}

// sidebarInnerWidth is the usable width inside the sidebar for content layout.
func (m tuiModel) sidebarInnerWidth() int {
	return m.sidebarWidth - 2 // subtract left+right padding
}

func (m tuiModel) renderSidebar() string {
	var lines []string

	// DEBUG: log sidebar dimensions when h seems too small
	if m.h < 20 {
		log.Printf("[DEBUG] renderSidebar LOW h: w=%d h=%d sidebarWidth=%d innerW=%d", m.w, m.h, m.sidebarWidth, m.sidebarInnerWidth())
	}

	// Top header: use plain ASCII to avoid lipgloss width mismatches with ANSI codes
	// "SwarmOps" = 8, "HH:MM:SS" = 8, gap = 6 → "SwarmOps      10:49:21" (22 visible = innerW, fits in Width(24))
	ts := time.Now().Format("15:04:05")
	header := fmt.Sprintf("%-8s %8s", "SwarmOps", ts) // left-padded to match innerW=22
	lines = append(lines, dimStyle.Render(header))
	// Version line (only show if BuildCommit is populated)
	if BuildCommit != "" {
		versionLine := fmt.Sprintf("  %-20s", "v"+BuildCommit)
		lines = append(lines, dimStyle.Render(versionLine))
	}

	// Summary line
	running := 0
	stopped := 0
	poolAlive := 0
	poolTotal := 0
	for _, item := range m.items {
		switch item.kind {
		case itemSession:
			if item.status == "running" {
				running++
			} else {
				stopped++
			}
		case itemPoolSlot:
			poolTotal++
			if item.alive {
				poolAlive++
			}
		}
	}
	var summary []string
	if running+stopped > 0 {
		summary = append(summary, fmt.Sprintf("%d sess", running+stopped))
	}
	if poolTotal > 0 {
		summary = append(summary, fmt.Sprintf("%d/%d pool", poolAlive, poolTotal))
	}
	lines = append(lines, dimStyle.Render(strings.Join(summary, " · ")))
	lines = append(lines, dimStyle.Render("────────────────────"))

	// Quota meters (from quota-proxy)
	if m.quota != nil {
		barW := m.sidebarInnerWidth() - 6 // leave room for label prefix "5h: "
		if barW < 4 {
			barW = 4
		}
		renderBar := func(label string, w *WindowData) string {
			if w == nil {
				return ""
			}
			filled := int(float64(barW) * w.Utilization)
			if filled > barW {
				filled = barW
			}
			bar := strings.Repeat("█", filled) + strings.Repeat("░", barW-filled)
			return fmt.Sprintf("%s %s %d%%", label, bar, int(w.PercentLeft))
		}
		if line := renderBar("5h:", m.quota.Session5h); line != "" {
			lines = append(lines, dimStyle.Render(line))
		}
		if line := renderBar("7d:", m.quota.Weekly7d); line != "" {
			lines = append(lines, dimStyle.Render(line))
		}
	}

	lines = append(lines, "")

	// Render session items
	now := time.Now().Unix()
	for i, item := range m.items {
		if item.kind == itemPoolSlot {
			continue // pool rendered separately below
		}
		labelLines := breakLabelAtSlashes(item.label, m.sidebarInnerWidth())
		ind := animatedIndicator(item.activity, m.animFrame)
		// SWM-11: escalated indicator after 30s of awaiting_input
		if item.activity == "awaiting_input" {
			if st, ok := m.activityStates[item.tmuxSession]; ok && st.awaitingInputSince > 0 && now-st.awaitingInputSince >= 30 {
				ind = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff4444")).Render("!")
			}
		}
		if i == m.cursor {
			lines = append(lines, selectedStyle.Render("▸")+fmt.Sprintf(" %s %s", ind, selectedLabelStyle.Render(labelLines[0])))
			for _, l := range labelLines[1:] {
				lines = append(lines, selectedStyle.Render("  ")+fmt.Sprintf("    %s", l))
			}
		} else {
			lines = append(lines, fmt.Sprintf("  %s %s", ind, labelLines[0]))
			for _, l := range labelLines[1:] {
				lines = append(lines, fmt.Sprintf("    %s", l))
			}
		}
	}

	// Pool section — collapsible (SWM-49)
	if poolTotal > 0 {
		lines = append(lines, "")
		busyCount := 0
		for _, item := range m.items {
			if item.kind == itemPoolSlot && item.state == "busy" {
				busyCount++
			}
		}
		busyStr := ""
		if busyCount > 0 {
			busyStr = fmt.Sprintf(" (%d)", busyCount)
		}
		toggleChar := "▶"
		if m.poolExpanded {
			toggleChar = "▼"
		}
		poolHeader := fmt.Sprintf("%s Pool %d/%d%s", toggleChar, poolAlive, poolTotal, busyStr)
		lines = append(lines, dimStyle.Render(poolHeader))

		if m.poolExpanded {
			for i, item := range m.items {
				if item.kind != itemPoolSlot {
					continue
				}
				label := item.label
				if len(label) > 20 {
					label = label[:17] + "..."
				}
				if i == m.cursor {
					line := fmt.Sprintf(" %s %s %s", item.indicator, selectedLabelStyle.Render(label), dimStyle.Render(item.state))
					lines = append(lines, selectedStyle.Render("▸")+line)
				} else {
					lines = append(lines, fmt.Sprintf("  %s %s %s", item.indicator, label, dimStyle.Render(item.state)))
				}
			}
		}
	}

	if len(m.items) == 0 {
		lines = append(lines, dimStyle.Render(" (no sessions)"))
	}

	for len(lines) < m.h-3 {
		lines = append(lines, "")
	}

	// Height accounts for: top bar (headerHeight), status line (2), sidebar vertical padding (2)
	sideHeight := m.h - headerHeight - 2 - 2
	if sideHeight < 3 {
		sideHeight = 3
	}
	return m.sidebarStyle().Height(sideHeight).Render(strings.Join(lines, "\n"))
}

// updateContentCache computes the right-pane content string based on current state
// and updates the viewport. Called from Update() handlers when state changes that
// affect the right pane. View() only reads from the viewport — no mutations.
func (m *tuiModel) updateContentCache() {
	if len(m.items) == 0 {
		m.contentCache = dimStyle.Render("\n  No sessions. Press Alt+N to create one.")
	} else if m.cursor < len(m.items) {
		item := m.items[m.cursor]
		switch item.kind {
		case itemPoolSlot:
			m.contentCache = m.renderPoolSlotDetail(item)
		case itemSession:
			if item.status != "running" {
				m.contentCache = dimStyle.Render("\n  Session stopped.")
			}
			// running sessions get contentCache set via terminalMsg
		}
	}

	if m.vpReady {
		m.vp.SetContent(m.contentCache)
	}
}

func (m tuiModel) renderContent() string {
	if !m.vpReady {
		return ""
	}
	contentWidth := m.w - (m.sidebarWidth + 2)
	if contentWidth < 20 {
		contentWidth = 20
	}

	// Header: session/slot name + status on line 1, separator on line 2
	var headerLine string
	if m.cursor < len(m.items) {
		item := m.items[m.cursor]
		switch item.kind {
		case itemSession:
			status := dimStyle.Render("stopped")
			if item.status == "running" {
				status = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")).Render("running")
			}
			headerLine = topBarTitleStyle.Render(item.label) + "  " + status
			if item.mission != "" {
				headerLine += "\n" + dimStyle.Render(item.mission)
			}
		case itemPoolSlot:
			state := dimStyle.Render(item.state)
			if item.state == "idle" {
				state = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")).Render("idle")
			} else if item.state == "busy" {
				state = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffaa00")).Render("busy")
			}
			headerLine = topBarTitleStyle.Render(item.model) + "  " + state
		}
	}
	if headerLine == "" {
		headerLine = dimStyle.Render("No selection")
	}

	sep := dimStyle.Render(strings.Repeat("─", contentWidth))
	header := headerLine + "\n" + sep + "\n"

	return lipgloss.NewStyle().Width(contentWidth).Render(header + m.vp.View())
}

func (m tuiModel) renderPoolSlotDetail(item sidebarItem) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("  Pool Slot: %s\n", item.slotID))
	b.WriteString(fmt.Sprintf("  Model:     %s\n", item.model))
	b.WriteString(fmt.Sprintf("  State:     %s\n", item.state))
	b.WriteString(fmt.Sprintf("  Alive:     %v\n", item.alive))
	b.WriteString(fmt.Sprintf("  Requests:  %d\n", item.requests))
	b.WriteString(fmt.Sprintf("  Cost:      $%.4f\n", item.costUSD))
	return b.String()
}

// runTUI starts the Bubbletea TUI. Database, config, and pool must be initialised by main().

// stripANSI removes ANSI escape sequences from a string.
var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// ─── Activity detector — prioritized classifier ─────────────────────────────
//
// States: "stopped" | "awaiting_input" | "working" | "idle"
// Priority order (highest wins):
//   1. Permission/menu/question prompts → awaiting_input
//   2. Prompt with user-typed text      → awaiting_input
//   3. Spinner / tool running / content changed → working
//   4. Bare prompt, stable              → idle
//
// Polled every 1s via activityTickCmd. Single-tick damper: if previous state
// was "working" and current tick is ambiguous, holds "working" for one more tick.

// activityState tracks per-session state for the activity detector.
type activityState struct {
	prevHash           uint64 // hash of previous capture for diff detection
	prevActivity       string // previous classification for 1-tick hold
	awaitingInputSince int64  // Unix timestamp when activity became awaiting_input; 0 otherwise
}

// Patterns for activity detection — compiled once.
var (
	// Spinner: Unicode spinner char + space + verb + ellipsis
	// Matches: ✶ Percolating…, ✽ Creating..., ◐ Thinking…, etc.
	spinnerRe = regexp.MustCompile(`^[✶✽✹◐◑◒◓⠋⠙⠹⠸⠼⠴⠦⠧●◆]\s+\S+[…\.]{1,3}`)

	// Tool running: "⎿  Running…" or "⎿  Running..."
	toolRunningRe = regexp.MustCompile(`^\s*⎿\s+Running[…\.]{1,3}`)

	// Permission/approval/menu prompts (blocking — highest priority).
	// Matches Claude Code's actual tool-approval UI patterns. Deliberately avoids
	// bare "allow"/"deny"/"approve" which are too broad and fire on normal output
	// like "Permission denied" or code comments (caused SWM-54 false positives).
	permissionRe = regexp.MustCompile(`(?i)(do you want to proceed|esc to cancel|yes/no|\(y/n\)|proceed\?|pick a number|[●○]\s*(allow|deny|approve)|allow this action|do you want to allow)`)

	// Prompt line: ❯ at start (with or without trailing user text)
	promptRe = regexp.MustCompile(`^❯`)

	// Bare prompt: ❯ with no user-typed text after it
	barePromptRe = regexp.MustCompile(`^❯\s*$`)

	// Shell prompt (session fell through to shell) — all alternations start-anchored
	shellPromptRe = regexp.MustCompile(`^(>\s*|\$\s*|#\s*)$`)

	// Chrome lines to skip (status bar, separators, model info)
	chromeRe = regexp.MustCompile(`^──|^\[.*\]\s+\S+.*\|.*ctx|bypass permissions|^Claude\s+(MAX|Pro|Free)|^\[.*\]\s+nuc|⏵⏵`)
)

// classifyActivity analyses captured tmux lines and session state to determine activity.
// Exported for testing. Does not call tmux — operates on pre-captured text.
func classifyActivity(capture string, state *activityState) string {
	lines := strings.Split(strings.TrimRight(capture, "\n"), "\n")

	// --- Scan bottom-up, collect meaningful lines (up to 100) ---
	var meaningful []string
	for i := len(lines) - 1; i >= 0 && len(meaningful) < 100; i-- {
		line := strings.TrimSpace(ansiRe.ReplaceAllString(lines[i], ""))
		if line == "" {
			continue
		}
		if chromeRe.MatchString(line) {
			continue
		}
		meaningful = append(meaningful, line)
	}

	// --- Hash meaningful content (not raw capture) for diff detection ---
	// This avoids false "working" from ANSI/chrome noise changes.
	// Skip hash update on empty meaningful content to prevent false diffs after capture glitches.
	isFirstSeen := state.prevHash == 0 // true on the very first call (no prior content seen)
	contentChanged := false
	if len(meaningful) > 0 {
		hash := fnvHash(strings.Join(meaningful, "\n"))
		contentChanged = !isFirstSeen && hash != state.prevHash
		state.prevHash = hash
	}

	// --- Priority 1: Permission/menu prompts → awaiting_input ---
	for _, line := range meaningful {
		if permissionRe.MatchString(line) {
			if state.prevActivity != "awaiting_input" {
				state.awaitingInputSince = time.Now().Unix()
			}
			state.prevActivity = "awaiting_input"
			return "awaiting_input"
		}
	}

	// --- Priority 2: Working — content changed ---
	// Checked BEFORE question detection: animated spinner (contentChanged=true each tick)
	// never triggers awaiting_input. Fixes SWM-52/51.
	if contentChanged {
		state.awaitingInputSince = 0
		state.prevActivity = "working"
		return "working"
	}

	// --- Priority 2b: First-seen spinner/tool → working ---
	// On the very first call prevHash was 0, so contentChanged=false even if there is
	// an active spinner in the buffer. Scan the top of the pane to detect it. Fixes
	// the case where SwarmOps starts and a session is already mid-tool-call.
	if isFirstSeen && len(meaningful) > 0 {
		limit := 15
		if len(meaningful) < limit {
			limit = len(meaningful)
		}
		for _, line := range meaningful[:limit] {
			if spinnerRe.MatchString(line) || toolRunningRe.MatchString(line) {
				state.prevActivity = "working"
				return "working"
			}
		}
	}

	// --- Priority 3: User-typed prompt (stable) → awaiting_input ---
	// ❯ followed by text means the user is composing input and hasn't submitted yet.
	// Only fires when content is stable (contentChanged=false) so transient streaming
	// output that happens to contain ❯ doesn't trigger this. Fixes SWM-46.
	for _, line := range meaningful {
		if promptRe.MatchString(line) && !barePromptRe.MatchString(line) {
			if state.prevActivity != "awaiting_input" {
				state.awaitingInputSince = time.Now().Unix()
			}
			state.prevActivity = "awaiting_input"
			return "awaiting_input"
		}
	}

	// --- Priority 4 (was 3): Question from Claude → awaiting_input ---
	// Only fires when content is STABLE (contentChanged=false) AND a bare prompt is
	// visible — meaning Claude has stopped and is waiting. Prevents false positives
	// from assistant lines that happen to end with "?". Fixes SWM-51.
	hasBarePrompt := false
	for _, line := range meaningful {
		if barePromptRe.MatchString(line) {
			hasBarePrompt = true
			break
		}
	}
	if hasBarePrompt {
		for _, line := range meaningful {
			if barePromptRe.MatchString(line) {
				continue
			}
			if promptRe.MatchString(line) || shellPromptRe.MatchString(line) {
				continue
			}
			// Spinner/tool alongside bare prompt → still working
			if spinnerRe.MatchString(line) || toolRunningRe.MatchString(line) {
				break
			}
			// First non-prompt, non-spinner assistant line: check for trailing question
			if strings.HasSuffix(strings.TrimSpace(line), "?") {
				if state.prevActivity != "awaiting_input" {
					state.awaitingInputSince = time.Now().Unix()
				}
				state.prevActivity = "awaiting_input"
				return "awaiting_input"
			}
			break
		}
	}

	// --- Priority 5: 1-tick hold ---
	// Prevents flicker when Claude pauses briefly between tool calls.
	// Also handles static spinner in buffer (SWM-46): spinner present but
	// contentChanged=false means nothing actually moved — fall through to hold/idle.
	if state.prevActivity == "working" {
		state.prevActivity = "idle"
		return "working"
	}

	// --- Priority 6: Idle ---
	state.prevActivity = "idle"
	return "idle"
}


// fnvHash computes a quick FNV-1a hash for change detection.
func fnvHash(s string) uint64 {
	h := uint64(14695981039346656037)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 16777619
	}
	return h
}

var (
	spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧"}
)

// animatedIndicator returns the indicator string for a session based on its activity and the current animation frame.
func animatedIndicator(activity string, frame int) string {
	switch activity {
	case "idle":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")).Render("●")
	case "working":
		f := spinnerFrames[frame%len(spinnerFrames)]
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#00ff00")).Render(f)
	case "awaiting_input":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#ffaa00")).Render("?")
	default:
		return statusStopped
	}
}
// backendOption describes one entry in the unified model+profile picker.
type backendOption struct {
	label        string            // display name
	model        string            // claude model override (empty = default for profile)
	profile      string            // happier backend profile (empty = anthropic)
	envOverrides map[string]string // extra env injected into the tmux session (e.g. LiteLLM routing)
}

// backendOptions is the ordered list of backend choices presented in the new-session
// wizard (modeNewModel) and the per-session profile editor (modeEditProfile).
// Indices: 0=default, 1=haiku, 2=sonnet, 3=opus, 4=deepseek, 5=openai, 6=[gpt], 7=[dseek]
//
// The `[gpt]` and `[dseek]` entries route Claude Code through LiteLLM by
// setting ANTHROPIC_BASE_URL+ANTHROPIC_API_KEY+ANTHROPIC_MODEL via env, and
// rely on autoPrefixSessionName to tag the resulting session name.
var backendOptions = []backendOption{
	{label: "default", model: "", profile: ""},
	{label: "haiku", model: "claude-haiku-4-5-20251001", profile: ""},
	{label: "sonnet", model: "claude-sonnet-4-6", profile: ""},
	{label: "opus", model: "claude-opus-4-6", profile: ""},
	{label: "deepseek", model: "", profile: "deepseek"},
	{label: "openai", model: "", profile: "openai"},
	{label: "[gpt]", model: litellmModelGPT55, profile: "", envOverrides: litellmEnvOverrides(litellmModelGPT55)},
	{label: "[dseek]", model: litellmModelDeepseek4, profile: "", envOverrides: litellmEnvOverrides(litellmModelDeepseek4)},
}

// envOverridesFromIndex returns the env override map for picker index idx,
// or nil if the entry has no overrides.
func envOverridesFromIndex(idx int) map[string]string {
	if idx >= 0 && idx < len(backendOptions) {
		return backendOptions[idx].envOverrides
	}
	return nil
}

// modelIDFromIndex returns the Claude model string for picker index idx.
func modelIDFromIndex(idx int) string {
	if idx >= 0 && idx < len(backendOptions) {
		return backendOptions[idx].model
	}
	return ""
}

// profileFromIndex returns the happier profile string for picker index idx.
func profileFromIndex(idx int) string {
	if idx >= 0 && idx < len(backendOptions) {
		return backendOptions[idx].profile
	}
	return ""
}

// profileIndexFromString returns the backendOptions index for a given profile string.
// Falls back to 0 (default) if not found.
func profileIndexFromString(profile string) int {
	for i, opt := range backendOptions {
		if opt.profile == profile {
			return i
		}
	}
	return 0
}

// modelPickerFlash returns the status bar message for the model/backend picker step.
func modelPickerFlash(idx int) string {
	var parts []string
	for i, opt := range backendOptions {
		if i == idx {
			parts = append(parts, "["+opt.label+"]")
		} else {
			parts = append(parts, opt.label)
		}
	}
	return "Backend: " + strings.Join(parts, " │ ") + "  (←/→ to pick, Enter to continue)"
}

// profilePickerFlash returns the status bar message for the profile edit mode.
func profilePickerFlash(idx int) string {
	var parts []string
	for i, opt := range backendOptions {
		if i == idx {
			parts = append(parts, "["+opt.label+"]")
		} else {
			parts = append(parts, opt.label)
		}
	}
	return "Profile: " + strings.Join(parts, " │ ") + "  (←/→ to pick, Enter to apply+restart, Esc to cancel)"
}

func runTUI(api swarmClient) error {
	// If running inside tmux, swap the outer tmux's prefix to Ctrl+\ so
	// Ctrl+b reaches the TUI instead of being intercepted by the outer tmux.
	// RestoreOriginalPrefix is called on exit.
	swapFn, restoreFn := swapNestedTmuxPrefix()
	if swapFn != nil {
		defer restoreFn()
	}

	p := tea.NewProgram(initialModel(api), tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err := p.Run()
	return err
}

// swapNestedTmuxPrefix swaps the outer tmux session's prefix key to Ctrl+\.
// This lets the TUI receive Ctrl+b (and other prefix keys) when running
// nested inside tmux. Returns (swapFn, restoreFn). swapFn is nil if not in tmux.
// The restoreFn MUST be called on exit to restore the original prefix.
func swapNestedTmuxPrefix() (func(), func()) {
	tmuxEnv := os.Getenv("TMUX")
	if tmuxEnv == "" {
		return nil, func() {}
	}

	// Get the current session name
	sessOut, err := exec.Command("tmux", "display-message", "-p", "#{session_name}").Output()
	if err != nil {
		return nil, func() {}
	}
	sess := strings.TrimSpace(string(sessOut))

	// Get current prefix key
	prefixOut, err := exec.Command("tmux", "show-options", "-t", sess, "prefix").Output()
	if err != nil {
		return nil, func() {}
	}
	// Output format: "prefix C-b" — extract the key
	origPrefix := strings.TrimSpace(string(prefixOut))
	origPrefix = strings.TrimPrefix(origPrefix, "prefix ")

	// Change outer tmux prefix to Ctrl+\
	exec.Command("tmux", "set-option", "-t", sess, "prefix", "C-\\").Run()

	restore := func() {
		exec.Command("tmux", "set-option", "-t", sess, "prefix", origPrefix).Run()
	}
	return func() {}, restore
}

// resumeClaudeCmd returns the command args for restarting a session.
// Uses happier when available; falls back to claude otherwise.
// For happier: starts a fresh wrapper each time. We don't pass
// `--existing-session` because happier needs a "session attach secret"
// to resume that we don't persist — passing the session id alone crashes
// happier ("missing session attach secret") and kills the tmux window.
// For claude: UUID IDs are resumed via --resume; non-UUID IDs (happier-era) start fresh.
//
// For happier we set ANTHROPIC_MODEL via an `env` prefix instead of passing
// `--model` directly — happier's `--model` flag triggers a bug where it
// deletes its hook settings file before claude can read it.
func resumeClaudeCmd(claudeID, model string) []string {
	if happierAvailable() {
		args := []string{"env", "ANTHROPIC_MODEL=" + effectiveModel(model), "happier", "--yolo"}
		return args
	}
	// claude fallback
	args := []string{"claude", "--dangerously-skip-permissions"}
	if claudeID != "" && isValidUUID(claudeID) {
		args = append(args, "--resume", claudeID)
	}
	if model != "" {
		args = append(args, "--model", model)
	}
	return args
}

func sidebarWidthPath() string {
	return filepath.Join(os.Getenv("HOME"), ".swarmops", "sidebar-width")
}

func loadSidebarWidth() int {
	path := sidebarWidthPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return 40 // default to max width
	}
	w := 40
	fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &w)
	if w < 18 {
		w = 18
	}
	if w > 40 {
		w = 40
	}
	return w
}

func saveSidebarWidth(w int) {
	os.MkdirAll(filepath.Dir(sidebarWidthPath()), 0755)
	os.WriteFile(sidebarWidthPath(), []byte(fmt.Sprintf("%d", w)), 0644)
}

// breakLabelAtSlashes splits a session label at "/" separators so path-like names
// (e.g. "agent/briefbox/phso-routing") render across multiple lines in the sidebar
// instead of overflowing. Continuation lines are indented 4 spaces.
func breakLabelAtSlashes(label string, innerW int) []string {
	maxLen := innerW - 12 // indicator + space + padding
	if maxLen < 4 {
		maxLen = 4
	}
	if !strings.Contains(label, "/") {
		if len(label) > maxLen {
			label = label[:maxLen-3] + "..."
		}
		return []string{label}
	}
	parts := strings.Split(label, "/")
	var lines []string
	for i, part := range parts {
		var line string
		if i == len(parts)-1 {
			line = part
		} else {
			line = part + "/"
		}
		if len(line) > maxLen {
			line = line[:maxLen-3] + "..."
		}
		lines = append(lines, line)
	}
	return lines
}

// syntheticKeyMsg converts a key string (e.g. "alt+shift+left", "enter", "q")
// into a tea.KeyMsg that handleKey's switch on msg.String() can match.
// Handles the common modifier combos used in the TUI keymap.
// Returns tea.KeyMsg{Type: 0} for unrecognised keys.
func syntheticKeyMsg(key string) tea.KeyMsg {
	switch key {
	case "alt+shift+left":
		return tea.KeyMsg{Type: tea.KeyShiftLeft, Alt: true}
	case "alt+shift+right":
		return tea.KeyMsg{Type: tea.KeyShiftRight, Alt: true}
	case "shift+alt+left":
		return tea.KeyMsg{Type: tea.KeyShiftLeft, Alt: true}
	case "shift+alt+right":
		return tea.KeyMsg{Type: tea.KeyShiftRight, Alt: true}
	case "shift+alt+up":
		return tea.KeyMsg{Type: tea.KeyShiftUp, Alt: true}
	case "shift+alt+down":
		return tea.KeyMsg{Type: tea.KeyShiftDown, Alt: true}
	case "shift+alt+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab, Alt: true}
	case "alt+left":
		return tea.KeyMsg{Type: tea.KeyLeft, Alt: true}
	case "alt+right":
		return tea.KeyMsg{Type: tea.KeyRight, Alt: true}
	case "shift+left":
		return tea.KeyMsg{Type: tea.KeyShiftLeft}
	case "shift+right":
		return tea.KeyMsg{Type: tea.KeyShiftRight}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "escape", "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "ctrl+d":
		return tea.KeyMsg{Type: tea.KeyCtrlD}
	case "q":
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	default:
		// Fallback: treat as a rune key
		if len(key) == 1 {
			return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
		}
		return tea.KeyMsg{}
	}
}
