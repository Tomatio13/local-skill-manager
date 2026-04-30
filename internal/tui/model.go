package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"localskillmanager/internal/config"
	"localskillmanager/internal/deploy"
	"localskillmanager/internal/discovery"
)

type pane int

const (
	skillsPane pane = iota
	targetsPane
)

type Model struct {
	cfg         config.Config
	manager     *deploy.Manager
	skills      []discovery.Skill
	filtered    []int
	skillIndex  int
	targetIndex int
	activePane  pane
	pending     map[string]map[string]bool
	searchMode  bool
	searchQuery string
	status      string
	width       int
	height      int
}

var ui = newUIStyles()

func NewModel(cfg config.Config, skills []discovery.Skill, manager *deploy.Manager) (Model, error) {
	return Model{
		cfg:        cfg,
		manager:    manager,
		skills:     skills,
		filtered:   buildFilteredIndices(skills, ""),
		pending:    make(map[string]map[string]bool),
		searchMode: false,
		status:     "Ready",
	}, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if m.searchMode {
			m.handleSearchInput(msg)
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "/":
			if m.activePane == skillsPane {
				m.searchMode = true
				m.status = "Search skills by name or description. Enter to apply, Esc to cancel."
			}
		case "c":
			if m.activePane == skillsPane {
				m.clearSearch()
			}
		case "tab", "left", "right", "h", "l":
			if m.activePane == skillsPane {
				m.activePane = targetsPane
			} else {
				m.activePane = skillsPane
			}
		case "up":
			m.moveUp()
		case "down":
			m.moveDown()
		case " ":
			if m.activePane == targetsPane {
				m.toggleTarget()
			}
		case "enter":
			if m.activePane == skillsPane {
				m.activePane = targetsPane
				m.status = "Target pane active. Use Space or Enter to toggle, a to apply."
			} else {
				m.toggleTarget()
			}
		case "a":
			m.applySelectedSkill()
		case "r":
			m.reloadSkills()
		}
	}
	return m, nil
}

func (m Model) View() string {
	totalWidth := m.width
	if totalWidth <= 0 {
		totalWidth = 100
	}
	totalHeight := m.height
	if totalHeight <= 0 {
		totalHeight = 28
	}

	contentWidth := totalWidth - ui.app.GetHorizontalFrameSize()
	if contentWidth < 40 {
		contentWidth = 40
	}

	header := m.renderHeader(contentWidth)
	footer := m.renderFooter(contentWidth)
	status := m.renderStatus(contentWidth)
	chromeHeight := lipgloss.Height(header) + lipgloss.Height(footer) + lipgloss.Height(status)
	paneHeight := totalHeight - ui.app.GetVerticalFrameSize() - chromeHeight
	if paneHeight < 8 {
		paneHeight = 8
	}

	if len(m.skills) == 0 {
		empty := ui.panel.
			Width(contentWidth).
			Height(paneHeight).
			Render("No skills found\n\nStore: " + m.cfg.StoreSkillsPath)
		content := ui.app.Render(lipgloss.JoinVertical(lipgloss.Left, header, empty, footer, status))
		return paintWindow(content, totalWidth, totalHeight, ui.windowFill)
	}

	leftWidth := contentWidth / 2
	rightWidth := contentWidth - leftWidth - 1
	if leftWidth < 28 {
		leftWidth = 28
	}
	if rightWidth < 28 {
		rightWidth = 28
	}

	leftInnerWidth := leftWidth - paneInnerWidth()
	rightInnerWidth := rightWidth - paneInnerWidth()
	innerHeight := paneHeight - paneInnerHeight()
	if innerHeight < 4 {
		innerHeight = 4
	}

	left := paneStyle(m.activePane == skillsPane).
		Width(leftWidth).
		Height(paneHeight).
		Render(fillHeight(m.renderSkills(leftInnerWidth, innerHeight), innerHeight))
	right := paneStyle(m.activePane == targetsPane).
		Width(rightWidth).
		Height(paneHeight).
		Render(fillHeight(m.renderTargets(rightInnerWidth, innerHeight), innerHeight))

	content := ui.app.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			lipgloss.JoinHorizontal(lipgloss.Top, left, ui.paneGap.Render(" "), right),
			footer,
			status,
		),
	)
	return paintWindow(content, totalWidth, totalHeight, ui.windowFill)
}

func (m *Model) moveUp() {
	if m.activePane == skillsPane && m.skillIndex > 0 {
		m.skillIndex--
		m.targetIndex = 0
		return
	}
	if m.activePane == targetsPane && m.targetIndex > 0 {
		m.targetIndex--
	}
}

func (m *Model) moveDown() {
	if m.activePane == skillsPane && m.skillIndex < len(m.filtered)-1 {
		m.skillIndex++
		m.targetIndex = 0
		return
	}

	targets := m.manager.Targets()
	if m.activePane == targetsPane && m.targetIndex < len(targets)-1 {
		m.targetIndex++
	}
}

func (m *Model) toggleTarget() {
	skill := m.selectedSkill()
	if skill.Name == "" {
		return
	}
	state := m.desiredState(skill)
	targets := m.manager.Targets()
	if len(targets) == 0 || m.targetIndex >= len(targets) {
		return
	}

	target := targets[m.targetIndex]
	state[target.RootPath] = !state[target.RootPath]
	m.pending[skill.Name] = state
	m.status = fmt.Sprintf("Pending %s for %s", onOff(state[target.RootPath]), targetLabel(target))
}

func (m *Model) applySelectedSkill() {
	skill := m.selectedSkill()
	if skill.Name == "" {
		return
	}
	desired := m.desiredState(skill)
	results := m.manager.Apply(skill, desired)

	var errors []string
	for _, result := range results {
		if result.Err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", result.Target.Name, result.Err))
		}
	}

	if len(errors) > 0 {
		m.status = "Apply failed: " + strings.Join(errors, "; ")
		return
	}

	delete(m.pending, skill.Name)
	m.status = fmt.Sprintf("Applied %s to %d targets", skill.Name, len(results))
}

func (m *Model) reloadSkills() {
	skills, err := discovery.ListSkills(m.cfg.StoreSkillsPath)
	if err != nil {
		m.status = "Reload failed: " + err.Error()
		return
	}

	m.skills = skills
	m.applyFilter(false)
	if len(m.skills) == 0 {
		m.skillIndex = 0
		m.targetIndex = 0
		m.status = "Reloaded 0 skills"
		return
	}
	if m.skillIndex >= len(m.skills) {
		m.skillIndex = len(m.skills) - 1
	}
	if m.targetIndex >= len(m.manager.Targets()) {
		m.targetIndex = 0
	}
	m.status = fmt.Sprintf("Reloaded %d skills", len(m.skills))
}

func (m Model) renderSkills(width int, height int) string {
	selected := m.selectedSkill()
	visibleCount := max(1, height-5)
	start, end := windowRange(len(m.filtered), m.skillIndex, visibleCount)
	searchLabel := "/ search"
	if m.searchQuery != "" {
		searchLabel = "/ " + m.searchQuery
	}
	if m.searchMode {
		searchLabel += " _"
	}
	lines := []string{
		ui.title.Render("SKILLS"),
		ui.meta.Render(fmt.Sprintf("%d skills  |  %d hits  |  %d-%d visible", len(m.skills), len(m.filtered), visibleRangeStart(start, end), end)),
		ui.meta.Render("store: " + truncate(m.cfg.StoreSkillsPath, max(10, width-7))),
		ui.search.Render(truncate(searchLabel, max(8, width))),
		"",
	}

	if len(m.filtered) == 0 {
		lines = append(lines, ui.meta.Render("No matching skills"))
		lines = append(lines, "", ui.meta.Render("Selected: none"))
		return strings.Join(lines, "\n")
	}

	for i := start; i < end; i++ {
		skill := m.skills[m.filtered[i]]
		marker := " "
		style := ui.row
		if i == m.skillIndex {
			marker = "▸"
			style = ui.rowSelected
		}

		nameWidth := max(8, width/3)
		name := truncate(skill.Name, nameWidth)
		descWidth := max(8, width-lipgloss.Width(name)-6)
		description := truncate(skill.Description, descWidth)
		line := fmt.Sprintf("%s %s", marker, name)
		if description != "" {
			line += "  " + ui.meta.Render(description)
		}
		lines = append(lines, style.Render(line))
	}

	lines = append(lines, "", ui.meta.Render("Selected: "+truncate(selected.Name, max(8, width-10))))
	return strings.Join(lines, "\n")
}

func (m Model) renderTargets(width int, height int) string {
	skill := m.selectedSkill()
	if skill.Name == "" {
		return strings.Join([]string{
			ui.title.Render("TARGETS"),
			ui.meta.Render("No skill selected"),
		}, "\n")
	}

	state := m.desiredState(skill)
	current := m.manager.CurrentState(skill)
	targets := m.manager.Targets()

	enabledCount := 0
	pendingCount := 0
	for _, target := range targets {
		if state[target.RootPath] {
			enabledCount++
		}
		if state[target.RootPath] != current[target.RootPath] {
			pendingCount++
		}
	}

	lines := []string{
		ui.title.Render("TARGETS"),
		ui.meta.Render(fmt.Sprintf("%s  |  %d on  |  %d pending", truncate(skill.Name, max(8, width-24)), enabledCount, pendingCount)),
		"",
	}

	visibleItems := max(1, (height-3)/2)
	start, end := windowRange(len(targets), m.targetIndex, visibleItems)
	lines[1] = ui.meta.Render(fmt.Sprintf("%s  |  %d on  |  %d pending  |  %d-%d visible", truncate(skill.Name, max(8, width-38)), enabledCount, pendingCount, start+1, end))

	for i := start; i < end; i++ {
		target := targets[i]
		marker := " "
		style := ui.row
		if i == m.targetIndex {
			marker = "▸"
			style = ui.rowSelected
		}

		status := "OFF"
		badge := ui.off
		if state[target.RootPath] {
			status = "ON "
			badge = ui.on
		}

		syncState := ui.meta.Render("live")
		if state[target.RootPath] != current[target.RootPath] {
			syncState = ui.pending.Render("pending")
		}

		label := truncate(targetLabel(target), max(6, width-28))
		deployPath := truncate(target.DeployPath, max(8, width-4))
		top := fmt.Sprintf("%s %s %s %s", marker, badge.Render(status), label, syncState)
		lines = append(lines, style.Render(top))
		lines = append(lines, "  "+ui.path.Render(deployPath))
	}

	return strings.Join(lines, "\n")
}

func (m Model) desiredState(skill discovery.Skill) map[string]bool {
	if pending, ok := m.pending[skill.Name]; ok {
		clone := make(map[string]bool, len(pending))
		for k, v := range pending {
			clone[k] = v
		}
		return clone
	}
	return m.manager.CurrentState(skill)
}

func (m Model) selectedSkill() discovery.Skill {
	if len(m.filtered) == 0 || m.skillIndex >= len(m.filtered) {
		return discovery.Skill{}
	}
	return m.skills[m.filtered[m.skillIndex]]
}

func (m Model) renderHeader(width int) string {
	left := ui.headerTitle.Render("Local Skill Manager")
	meta := ui.headerMeta.Render("Tab switch  Space toggle  Enter apply")
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", meta),
		ui.headerSub.Render("Deploy local skills from store_path/skills into agent-specific targets"),
	)
	return ui.headerBlock.Width(width).Render(content)
}

func (m Model) renderFooter(width int) string {
	text := strings.Join([]string{
		"↑/↓ move",
		"/ search",
		"c clear",
		"Tab/←/→ pane",
		"Space/Enter toggle",
		"a apply",
		"r reload",
		"q quit",
	}, "  ·  ")
	return ui.footer.Width(width).Render(text)
}

func (m Model) renderStatus(width int) string {
	return ui.status.Width(width).Render(truncate(m.status, width-2))
}

func paneStyle(active bool) lipgloss.Style {
	style := ui.panel
	if active {
		style = style.BorderForeground(lipgloss.Color("#C68A1F"))
	}
	return style
}

func paneInnerWidth() int {
	return ui.panel.GetHorizontalFrameSize()
}

func paneInnerHeight() int {
	return ui.panel.GetVerticalFrameSize()
}

func onOff(v bool) string {
	if v {
		return "ON"
	}
	return "OFF"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func truncate(value string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(value)
	if lipgloss.Width(value) <= width {
		return value
	}
	if width == 1 {
		return "…"
	}
	for len(runes) > 0 && lipgloss.Width(string(runes)) > width-1 {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

func targetLabel(target deploy.ResolvedTarget) string {
	base := filepath.Base(target.RootPath)
	if strings.HasPrefix(base, ".") {
		base = strings.TrimPrefix(base, ".")
	}
	if base == "" {
		return target.RootPath
	}
	return base
}

func fillHeight(content string, height int) string {
	current := lipgloss.Height(content)
	if current >= height {
		return content
	}
	if content == "" {
		return strings.Repeat("\n", height-1)
	}
	return content + strings.Repeat("\n", height-current)
}

func windowRange(total int, selected int, visible int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	if visible >= total {
		return 0, total
	}
	if selected < 0 {
		selected = 0
	}
	if selected >= total {
		selected = total - 1
	}

	start := selected - visible/2
	if start < 0 {
		start = 0
	}
	end := start + visible
	if end > total {
		end = total
		start = end - visible
	}
	return start, end
}

type uiStyles struct {
	app         lipgloss.Style
	panel       lipgloss.Style
	paneGap     lipgloss.Style
	headerBlock lipgloss.Style
	headerTitle lipgloss.Style
	headerSub   lipgloss.Style
	headerMeta  lipgloss.Style
	title       lipgloss.Style
	meta        lipgloss.Style
	search      lipgloss.Style
	row         lipgloss.Style
	rowSelected lipgloss.Style
	path        lipgloss.Style
	on          lipgloss.Style
	off         lipgloss.Style
	pending     lipgloss.Style
	footer      lipgloss.Style
	status      lipgloss.Style
	windowFill  lipgloss.Style
}

func newUIStyles() uiStyles {
	return uiStyles{
		app: lipgloss.NewStyle().
			Padding(1, 2).
			Background(lipgloss.Color("#11182A")).
			Foreground(lipgloss.Color("#E8ECF3")),
		panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3B465B")).
			BorderBackground(lipgloss.Color("#101725")).
			Padding(1, 2).
			Background(lipgloss.Color("#11182A")),
		paneGap: lipgloss.NewStyle().
			Background(lipgloss.Color("#11182A")),
		headerBlock: lipgloss.NewStyle().
			Background(lipgloss.Color("#11182A")).
			PaddingBottom(1),
		headerTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F3E6C2")).
			Bold(true),
		headerSub: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#95A0B5")),
		headerMeta: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#95A0B5")),
		title: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F3E6C2")).
			Bold(true),
		meta: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#95A0B5")),
		search: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E8ECF3")).
			Background(lipgloss.Color("#1A2236")).
			Padding(0, 1),
		row: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E8ECF3")),
		rowSelected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF8E8")).
			Background(lipgloss.Color("#1B2540")),
		path: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#73809B")),
		on: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#08150E")).
			Background(lipgloss.Color("#7AD3A7")).
			Padding(0, 1).
			Bold(true),
		off: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E8ECF3")).
			Background(lipgloss.Color("#475167")).
			Padding(0, 1).
			Bold(true),
		pending: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#1A1203")).
			Background(lipgloss.Color("#E5AE4A")).
			Padding(0, 1).
			Bold(true),
		footer: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAB4C8")).
			Background(lipgloss.Color("#11182A")).
			PaddingTop(1),
		status: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0B1020")).
			Background(lipgloss.Color("#D7DDE8")).
			Padding(0, 1),
		windowFill: lipgloss.NewStyle().
			Background(lipgloss.Color("#11182A")),
	}
}

func (m *Model) handleSearchInput(msg tea.KeyMsg) {
	switch msg.String() {
	case "esc":
		m.searchMode = false
		m.status = "Search canceled"
	case "enter":
		m.searchMode = false
		m.applyFilter(true)
	case "backspace":
		if len(m.searchQuery) > 0 {
			runes := []rune(m.searchQuery)
			m.searchQuery = string(runes[:len(runes)-1])
			m.applyFilter(false)
		}
	case "ctrl+u":
		m.searchQuery = ""
		m.applyFilter(false)
		m.status = "Search input cleared"
	default:
		if msg.Type == tea.KeyRunes {
			text := string(msg.Runes)
			if len(msg.Runes) == 0 {
				return
			}
			if unicode.IsControl(msg.Runes[0]) {
				return
			}
			m.searchQuery += text
			m.applyFilter(false)
		}
	}
}

func (m *Model) applyFilter(final bool) {
	m.filtered = buildFilteredIndices(m.skills, m.searchQuery)
	if len(m.filtered) == 0 {
		m.skillIndex = 0
		if final {
			m.status = "No matching skills"
		}
		return
	}
	if m.skillIndex >= len(m.filtered) {
		m.skillIndex = len(m.filtered) - 1
	}
	if final {
		m.status = fmt.Sprintf("Search matched %d skills", len(m.filtered))
	}
}

func (m *Model) clearSearch() {
	m.searchQuery = ""
	m.searchMode = false
	m.filtered = buildFilteredIndices(m.skills, "")
	m.skillIndex = 0
	m.targetIndex = 0
	m.status = fmt.Sprintf("Search cleared. Showing %d skills", len(m.filtered))
}

func buildFilteredIndices(skills []discovery.Skill, query string) []int {
	query = strings.TrimSpace(strings.ToLower(query))
	indices := make([]int, 0, len(skills))
	for i, skill := range skills {
		if query == "" {
			indices = append(indices, i)
			continue
		}
		haystack := strings.ToLower(skill.Name + "\n" + skill.Description)
		if strings.Contains(haystack, query) {
			indices = append(indices, i)
		}
	}
	return indices
}

func visibleRangeStart(start int, end int) int {
	if end == 0 {
		return 0
	}
	return start + 1
}

func paintWindow(content string, width int, height int, fill lipgloss.Style) string {
	if width <= 0 {
		width = lipgloss.Width(content)
	}
	if height <= 0 {
		height = lipgloss.Height(content)
	}

	lines := strings.Split(content, "\n")
	painted := make([]string, 0, max(height, len(lines)))
	for _, line := range lines {
		lineWidth := lipgloss.Width(line)
		if lineWidth < width {
			line += fill.Render(strings.Repeat(" ", width-lineWidth))
		}
		painted = append(painted, line)
	}

	for len(painted) < height {
		painted = append(painted, fill.Render(strings.Repeat(" ", width)))
	}

	return strings.Join(painted, "\n")
}
