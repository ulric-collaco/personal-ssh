package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type theme struct {
	name           string
	primary        lipgloss.Color
	accent         lipgloss.Color
	highlight      lipgloss.Color
	particle       lipgloss.Color
	scanline       lipgloss.Color
	enableScanline bool
}

type particle struct {
	x     int
	y     int
	glyph rune
	phase int
	drift int
}

type matrixDrop struct {
	x      int
	y      float64
	speed  float64
	length int
}

type pageMode int

const (
	modeHome pageMode = iota
	modeProjects
	modeAbout
	modeContact
	modeProjectDetail
)

type project struct {
	title       string
	assetPath   string
	description string
}

type transition struct {
	active    bool
	direction int
	elapsed   time.Duration
	duration  time.Duration
	fromMode  pageMode
	toMode    pageMode
}

type (
	particleTickMsg   time.Time
	revealTickMsg     time.Time
	typeTickMsg       time.Time
	scanlineTickMsg   time.Time
	matrixTickMsg     time.Time
	transitionTickMsg time.Time
)

type model struct {
	width  int
	height int

	themes     []theme
	themeIndex int
	matrixMode bool

	portrait       []string
	revealLines    int
	scaledPortrait []string

	introLines    []string
	introProgress int
	bodyProgress  int
	navbarPhase   int
	uiPhase       int
	hintOffset    int
	colorPhase    int
	colorDelay    int

	navItems    []string
	selectedNav int
	currentMode pageMode

	projects        []project
	selectedProject int
	projectASCII    map[string][]string

	particles []particle
	scanlineY int
	matrix    []matrixDrop

	transition transition
}

func main() {
	rand.Seed(time.Now().UnixNano())
	program := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fmt.Println("error running program:", err)
		os.Exit(1)
	}
}

func newModel() model {
	portrait := loadASCII("assets/me_ascii.txt", []string{"(portrait missing: assets/me_ascii.txt)"})
	projects := []project{
		{title: "Gyro Controlled Car", assetPath: "assets/projects/gyro_car.txt", description: "A gyro-stabilized RC platform with embedded control loops and responsive remote steering."},
		{title: "ChatRTX Clone", assetPath: "assets/projects/chatrtx.txt", description: "A local retrieval-augmented chat interface optimized for low-latency terminal workflows."},
	}

	projectASCII := make(map[string][]string, len(projects))
	for _, p := range projects {
		projectASCII[p.assetPath] = loadASCII(p.assetPath, []string{"(project art missing)"})
	}

	m := model{
		themes: []theme{
			{name: "Cyber Green", primary: "#A4FFB0", accent: "#43FF75", highlight: "#D8FFE0", particle: "#4EFF78", scanline: "#1B5E20", enableScanline: true},
			{name: "Ocean Blue", primary: "#79D7FF", accent: "#4EA8FF", highlight: "#A8FFEC", particle: "#6EC5FF", scanline: "#0D3C66", enableScanline: true},
			{name: "Amber Retro", primary: "#FFDCA3", accent: "#FF9F1C", highlight: "#FFF0A3", particle: "#FF6A3D", scanline: "#5A3D00", enableScanline: true},
			{name: "Minimal White", primary: "#F5F5F5", accent: "#9EC5FF", highlight: "#FFE9A8", particle: "#BEE7D3", scanline: "#7A7A7A", enableScanline: false},
		},
		portrait:       portrait,
		revealLines:    len(portrait),
		scaledPortrait: portrait,
		introLines: []string{
			"Hello, I'm Ulric Collaco",
			"Computer Engineering Student",
			"Builder of interesting systems and tools",
		},
		navItems:     []string{"Home", "Projects", "About", "Contact"},
		selectedNav:  0,
		currentMode:  modeHome,
		projects:     projects,
		projectASCII: projectASCII,
		scanlineY:    0,
		colorDelay:   8,
	}

	m.resetTyping()
	return m
}

func loadASCII(path string, fallback []string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return fallback
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(text, "\n")
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return fallback
	}
	return lines
}

func (m model) Init() tea.Cmd {
	return tea.Batch(particleTick(), revealTick(), typeTick(), scanlineTick(), matrixTick())
}

func particleTick() tea.Cmd {
	return tea.Tick(150*time.Millisecond, func(t time.Time) tea.Msg { return particleTickMsg(t) })
}

func revealTick() tea.Cmd {
	return tea.Tick(35*time.Millisecond, func(t time.Time) tea.Msg { return revealTickMsg(t) })
}

func typeTick() tea.Cmd {
	return tea.Tick(28*time.Millisecond, func(t time.Time) tea.Msg { return typeTickMsg(t) })
}

func scanlineTick() tea.Cmd {
	return tea.Tick(110*time.Millisecond, func(t time.Time) tea.Msg { return scanlineTickMsg(t) })
}

func matrixTick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg { return matrixTickMsg(t) })
}

func transitionTick() tea.Cmd {
	return tea.Tick(20*time.Millisecond, func(t time.Time) tea.Msg { return transitionTickMsg(t) })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.refreshScaledPortrait()
		m.reseedParticles()
		m.reseedMatrix()
		return m, nil

	case tea.KeyMsg:
		s := msg.String()
		switch s {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "ctrl+m":
			m.matrixMode = !m.matrixMode
			if m.matrixMode {
				m.themeIndex = 0
				m.reseedMatrix()
			}
			m.reseedParticles()
			return m, nil
		case "t":
			m.themeIndex = (m.themeIndex + 1) % len(m.themes)
			return m, nil
		case "left", "h":
			m.selectedNav = (m.selectedNav - 1 + len(m.navItems)) % len(m.navItems)
			return m, nil
		case "right", "l":
			m.selectedNav = (m.selectedNav + 1) % len(m.navItems)
			return m, nil
		case "up", "k":
			if m.currentMode == modeProjects || m.currentMode == modeProjectDetail {
				m.selectedProject = (m.selectedProject - 1 + len(m.projects)) % len(m.projects)
				if m.currentMode == modeProjectDetail {
					m.startTransition(modeProjectDetail, 1)
				}
			}
			return m, nil
		case "down", "j":
			if m.currentMode == modeProjects || m.currentMode == modeProjectDetail {
				m.selectedProject = (m.selectedProject + 1) % len(m.projects)
				if m.currentMode == modeProjectDetail {
					m.startTransition(modeProjectDetail, -1)
				}
			}
			return m, nil
		case "esc", "backspace":
			if m.currentMode == modeProjectDetail {
				m.startTransition(modeProjects, -1)
				return m, transitionTick()
			}
			return m, nil
		case "enter":
			if m.currentMode == modeProjects && navToMode(m.selectedNav) == modeProjects {
				m.startTransition(modeProjectDetail, 1)
				return m, transitionTick()
			}
			target := navToMode(m.selectedNav)
			if target != m.currentMode {
				direction := 1
				if m.selectedNav < navIndexForMode(m.currentMode) {
					direction = -1
				}
				m.startTransition(target, direction)
				return m, transitionTick()
			}
			return m, nil
		}
		return m, nil

	case particleTickMsg:
		m.updateParticles()
		return m, particleTick()

	case revealTickMsg:
		m.revealLines = len(m.scaledPortrait)
		return m, revealTick()

	case typeTickMsg:
		if m.currentMode == modeHome {
			totalIntro := len([]rune(strings.Join(m.introLines, "\n")))
			if m.introProgress < totalIntro {
				m.introProgress++
			}
		}
		maxBody := len([]rune(m.currentBodyText()))
		if m.bodyProgress < maxBody {
			m.bodyProgress++
		}
		m.navbarPhase = (m.navbarPhase + 1) % 8
		m.uiPhase = (m.uiPhase + 1) % 120
		m.hintOffset++
		m.updateColorPhase()
		return m, typeTick()

	case scanlineTickMsg:
		if m.height > 0 {
			m.scanlineY = (m.scanlineY + 1) % m.height
		}
		return m, scanlineTick()

	case matrixTickMsg:
		if m.matrixMode {
			m.updateMatrix()
		}
		return m, matrixTick()

	case transitionTickMsg:
		if !m.transition.active {
			return m, nil
		}
		m.transition.elapsed += 20 * time.Millisecond
		if m.transition.elapsed >= m.transition.duration {
			m.currentMode = m.transition.toMode
			m.transition.active = false
			m.resetTyping()
			return m, nil
		}
		return m, transitionTick()
	}

	return m, nil
}

func (m *model) startTransition(to pageMode, direction int) {
	m.transition = transition{
		active:    true,
		direction: direction,
		elapsed:   0,
		duration:  320 * time.Millisecond,
		fromMode:  m.currentMode,
		toMode:    to,
	}
}

func navToMode(index int) pageMode {
	switch index {
	case 0:
		return modeHome
	case 1:
		return modeProjects
	case 2:
		return modeAbout
	default:
		return modeContact
	}
}

func navIndexForMode(mode pageMode) int {
	switch mode {
	case modeHome:
		return 0
	case modeProjects, modeProjectDetail:
		return 1
	case modeAbout:
		return 2
	default:
		return 3
	}
}

func (m *model) resetTyping() {
	m.introProgress = 0
	m.bodyProgress = 0
}

func (m *model) updateColorPhase() {
	m.colorDelay--
	if m.colorDelay > 0 {
		return
	}

	// Slower, non-linear palette motion for a less mechanical look.
	m.colorDelay = 6 + rand.Intn(10)
	step := rand.Intn(5) - 2
	if step == 0 {
		if rand.Intn(2) == 0 {
			return
		}
		step = 1
	}

	m.colorPhase = (m.colorPhase + step + 10000) % 10000
}

func (m *model) refreshScaledPortrait() {
	if m.height <= 0 || m.width <= 0 {
		m.scaledPortrait = m.portrait
		return
	}

	// Preserve source glyph fidelity by cropping instead of resampling.
	targetHeight := max(8, int(float64(m.height)*0.62))
	targetWidth := max(24, int(float64(m.width)*0.96))
	m.scaledPortrait = cropPortraitToFit(m.portrait, targetWidth, targetHeight)

	if m.revealLines > len(m.scaledPortrait) {
		m.revealLines = len(m.scaledPortrait)
	}
	m.revealLines = len(m.scaledPortrait)
}

func (m model) portraitBounds() (int, int, int, int, bool) {
	portraitVisible := m.scaledPortrait[:min(m.revealLines, len(m.scaledPortrait))]
	if len(portraitVisible) == 0 || m.width <= 0 || m.height <= 0 {
		return 0, 0, 0, 0, false
	}

	topBreath := max(2, m.height/10)
	portraitTop := topBreath
	if portraitTop+len(portraitVisible) > m.height-6 {
		portraitTop = max(2, m.height-len(portraitVisible)-6)
	}

	maxW := 0
	for _, line := range portraitVisible {
		w := lipgloss.Width(line)
		if w > maxW {
			maxW = w
		}
	}
	if maxW <= 0 {
		return 0, 0, 0, 0, false
	}

	left := (m.width - maxW) / 2
	right := left + maxW - 1
	top := portraitTop
	bottom := portraitTop + len(portraitVisible) - 1

	if left < 0 {
		left = 0
	}
	if right >= m.width {
		right = m.width - 1
	}
	if top < 0 {
		top = 0
	}
	if bottom >= m.height {
		bottom = m.height - 1
	}

	return left, right, top, bottom, true
}

func cropPortraitToFit(lines []string, maxWidth, maxHeight int) []string {
	if len(lines) == 0 || maxWidth <= 0 || maxHeight <= 0 {
		return lines
	}

	startY := 0
	endY := len(lines)
	if len(lines) > maxHeight {
		startY = (len(lines) - maxHeight) / 2
		endY = startY + maxHeight
	}
	visible := lines[startY:endY]

	result := make([]string, 0, len(visible))
	for _, line := range visible {
		r := []rune(line)
		if len(r) <= maxWidth {
			result = append(result, line)
			continue
		}
		startX := (len(r) - maxWidth) / 2
		endX := startX + maxWidth
		result = append(result, string(r[startX:endX]))
	}

	return result
}

func (m *model) reseedParticles() {
	if m.width <= 0 || m.height <= 0 {
		return
	}
	target := max(180, (m.width*m.height)/45)
	if m.matrixMode {
		target = max(260, (m.width*m.height)/30)
	}
	m.particles = make([]particle, 0, target)
	for i := 0; i < target; i++ {
		m.particles = append(m.particles, m.newParticle())
	}
}

func (m *model) newParticle() particle {
	glyphs := []rune{'✦', '✧', '⋆', '✶', '*', '·'}
	if m.matrixMode {
		glyphs = []rune{'0', '1', '·', '.'}
	}
	return particle{
		x:     rand.Intn(max(1, m.width)),
		y:     rand.Intn(max(1, m.height)),
		glyph: glyphs[rand.Intn(len(glyphs))],
		phase: rand.Intn(4),
		drift: rand.Intn(3) - 1,
	}
}

func (m *model) updateParticles() {
	if len(m.particles) == 0 {
		m.reseedParticles()
	}
	for i := range m.particles {
		m.particles[i].phase = (m.particles[i].phase + 1) % 4
		if rand.Intn(5) == 0 {
			m.particles[i].x += m.particles[i].drift
		}
		if rand.Intn(12) == 0 {
			m.particles[i].y++
		}
		if rand.Intn(20) == 0 {
			m.particles[i].y--
		}
		if m.particles[i].x < 0 {
			m.particles[i].x = m.width - 1
		}
		if m.particles[i].x >= m.width {
			m.particles[i].x = 0
		}
		if m.particles[i].y < 0 {
			m.particles[i].y = m.height - 1
		}
		if m.particles[i].y >= m.height {
			m.particles[i].y = 0
		}
	}
}

func (m *model) reseedMatrix() {
	if m.width <= 0 || m.height <= 0 {
		return
	}
	m.matrix = m.matrix[:0]
	if !m.matrixMode {
		return
	}
	columns := max(10, m.width/3)
	for i := 0; i < columns; i++ {
		m.matrix = append(m.matrix, matrixDrop{
			x:      rand.Intn(max(1, m.width)),
			y:      -float64(rand.Intn(max(1, m.height))),
			speed:  0.6 + rand.Float64()*1.2,
			length: 3 + rand.Intn(10),
		})
	}
}

func (m *model) updateMatrix() {
	if len(m.matrix) == 0 {
		m.reseedMatrix()
	}
	for i := range m.matrix {
		m.matrix[i].y += m.matrix[i].speed
		if int(m.matrix[i].y)-m.matrix[i].length > m.height {
			m.matrix[i].x = rand.Intn(max(1, m.width))
			m.matrix[i].y = -float64(rand.Intn(max(1, m.height/2)))
			m.matrix[i].speed = 0.6 + rand.Float64()*1.2
			m.matrix[i].length = 3 + rand.Intn(10)
		}
	}
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing terminal..."
	}

	t := m.activeTheme()
	if m.matrixMode {
		t = m.matrixTheme()
	}

	base := m.renderCanvas(t)
	overlay := m.renderOverlay(t)
	return base + overlay
}

func (m model) renderCanvas(t theme) string {
	lines := make([]string, m.height)
	for i := range lines {
		lines[i] = strings.Repeat(" ", max(1, m.width))
	}

	if m.height > 0 {
		lines[0] = centerLine(m.renderStatusBar(t), m.width)
	}
	if m.transition.active && m.height > 1 {
		lines[1] = centerLine(m.renderTransitionBar(t), m.width)
	}

	portraitVisible := m.scaledPortrait[:min(m.revealLines, len(m.scaledPortrait))]
	palette := themePalette(t)
	portraitStyle := lipgloss.NewStyle().Foreground(t.accent).Bold(true)

	topBreath := max(2, m.height/10)
	portraitTop := topBreath
	if portraitTop+len(portraitVisible) > m.height-6 {
		portraitTop = max(2, m.height-len(portraitVisible)-6)
	}
	for i, raw := range portraitVisible {
		y := portraitTop + i
		if y >= 0 && y < m.height {
			lines[y] = centerLine(portraitStyle.Render(raw), m.width)
		}
	}

	contentLines := m.renderPageContent(t)
	contentLines = m.applyTransition(contentLines)
	contentStart := portraitTop + len(portraitVisible) + 1
	contentLimit := m.height - 3
	for i, raw := range contentLines {
		y := contentStart + i
		if y >= 0 && y < contentLimit {
			if strings.Contains(raw, "\x1b[") {
				lines[y] = centerLine(raw, m.width)
				continue
			}
			c := palette[(i+m.colorPhase)%len(palette)]
			style := lipgloss.NewStyle().Foreground(c)
			lines[y] = centerLine(style.Render(raw), m.width)
		}
	}

	hintY := m.height - 3
	if hintY >= 2 && hintY < m.height {
		lines[hintY] = centerLine(m.renderHintTicker(t, max(24, m.width-8)), m.width)
	}

	navY := m.height - 2
	if navY >= 0 && navY < m.height {
		lines[navY] = centerLine(m.renderNavbar(t), m.width)
	}

	return strings.Join(lines, "\n")
}

func centerLine(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	left := (width - w) / 2
	right := width - w - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

func (m model) renderPageContent(t theme) []string {
	switch m.currentMode {
	case modeHome:
		return m.renderHomeContent()
	case modeProjects:
		return m.renderProjectsContent(t)
	case modeProjectDetail:
		return m.renderProjectDetailContent(t)
	case modeAbout:
		about := "Ulric is a computer engineering student focused on immersive terminal interfaces, systems programming, and building practical tools with strong UX detail."
		return []string{typed(about, m.bodyProgress)}
	case modeContact:
		return []string{
			typed("Github: github.com/ulric", m.bodyProgress),
			typed("Email: ulric@example.com", m.bodyProgress),
			typed("LinkedIn: linkedin.com/in/ulric", m.bodyProgress),
		}
	default:
		return []string{}
	}
}

func (m model) renderHomeContent() []string {
	joined := strings.Join(m.introLines, "\n")
	visible := typed(joined, m.introProgress)
	return strings.Split(visible, "\n")
}

func (m model) renderProjectsContent(t theme) []string {
	lines := []string{"Projects"}
	selectedStyle := lipgloss.NewStyle().Foreground(t.highlight).Bold(m.navbarPhase < 4)
	normalStyle := lipgloss.NewStyle().Foreground(t.primary)
	for i, p := range m.projects {
		if i == m.selectedProject {
			lines = append(lines, selectedStyle.Render("[ "+p.title+" ]"))
		} else {
			lines = append(lines, normalStyle.Render(p.title))
		}
	}
	lines = append(lines, "Enter to open project")
	return lines
}

func (m model) renderProjectDetailContent(t theme) []string {
	p := m.projects[m.selectedProject]
	art := m.projectASCII[p.assetPath]
	titleStyle := lipgloss.NewStyle().Foreground(t.highlight).Bold(true)
	lines := []string{titleStyle.Render(p.title), ""}
	for _, line := range art {
		lines = append(lines, line)
	}
	lines = append(lines, "")
	lines = append(lines, typed(p.description, m.bodyProgress))
	lines = append(lines, "Esc to go back")
	return lines
}

func (m model) currentBodyText() string {
	switch m.currentMode {
	case modeAbout:
		return "Ulric is a computer engineering student focused on immersive terminal interfaces, systems programming, and building practical tools with strong UX detail."
	case modeContact:
		return "Github: github.com/ulric\nEmail: ulric@example.com\nLinkedIn: linkedin.com/in/ulric"
	case modeProjectDetail:
		return m.projects[m.selectedProject].description
	default:
		return ""
	}
}

func typed(text string, progress int) string {
	r := []rune(text)
	if progress <= 0 {
		return ""
	}
	if progress >= len(r) {
		return text
	}
	return string(r[:progress])
}

func (m model) applyTransition(lines []string) []string {
	if !m.transition.active {
		return lines
	}
	p := float64(m.transition.elapsed) / float64(m.transition.duration)
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	shift := int((1 - p) * 10)
	trailCount := int((1 - p) * 4)
	trail := strings.Repeat("·", trailCount)
	out := make([]string, len(lines))
	for i, line := range lines {
		if m.transition.direction > 0 {
			out[i] = trail + strings.Repeat(" ", shift) + line
		} else {
			out[i] = line + strings.Repeat(" ", shift) + trail
		}
	}
	return out
}

func (m model) renderNavbar(t theme) string {
	parts := make([]string, 0, len(m.navItems))
	highlightColor := t.highlight
	if m.navbarPhase < 4 {
		highlightColor = t.accent
	}
	sel := lipgloss.NewStyle().Foreground(highlightColor).Bold(true)
	left := "⟪"
	right := "⟫"
	if (m.uiPhase/4)%2 == 0 {
		left = "⟨"
		right = "⟩"
	}

	for i, item := range m.navItems {
		if i == m.selectedNav {
			parts = append(parts, rainbowText(left+" "+item+" "+right, themePalette(t), m.colorPhase+i*3, true))
		} else {
			parts = append(parts, rainbowText("• "+item+" •", themePalette(t), m.colorPhase+i*5, false))
		}
	}
	bar := "✧ " + strings.Join(parts, "   ") + " ✧"
	return sel.Render(bar)
}

func (m model) renderStatusBar(t theme) string {
	spinner := []string{"◜", "◝", "◞", "◟"}
	icon := spinner[(m.uiPhase/3)%len(spinner)]
	mode := strings.ToUpper(modeLabel(m.currentMode))
	badge := "LIVE"
	if m.matrixMode {
		badge = "MATRIX"
	}
	bar := fmt.Sprintf("%s ◈ %s ◈ THEME %s ◈ %s", icon, mode, strings.ToUpper(t.name), badge)
	return rainbowText(bar, themePalette(t), m.colorPhase, true)
}

func (m model) renderHintTicker(t theme, width int) string {
	hints := []string{
		"h/l or arrows: switch tabs",
		"j/k: move project",
		"enter: open",
		"esc: back",
		"t: next theme",
		"ctrl+m: matrix mode",
	}
	joined := strings.Join(hints, "  ✦  ")
	ticker := marquee(joined, width, m.hintOffset)
	return rainbowText("✦ "+ticker+" ✦", themePalette(t), m.colorPhase/2, false)
}

func (m model) renderTransitionBar(t theme) string {
	if !m.transition.active {
		return ""
	}
	p := float64(m.transition.elapsed) / float64(m.transition.duration)
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	width := max(12, m.width/3)
	filled := int(float64(width) * p)
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	text := fmt.Sprintf("◉ TRANSITION %s ◉", bar)
	return rainbowText(text, themePalette(t), m.colorPhase/2, true)
}

func marquee(text string, width, offset int) string {
	r := []rune(text)
	if len(r) == 0 || width <= 0 {
		return ""
	}
	if len(r) <= width {
		return text
	}
	padding := []rune("   ")
	loop := append(append([]rune{}, r...), padding...)
	loop = append(loop, r...)
	start := offset % (len(r) + len(padding))
	end := start + width
	if end > len(loop) {
		end = len(loop)
	}
	return string(loop[start:end])
}

func modeLabel(mode pageMode) string {
	switch mode {
	case modeHome:
		return "Home"
	case modeProjects:
		return "Projects"
	case modeProjectDetail:
		return "Project"
	case modeAbout:
		return "About"
	default:
		return "Contact"
	}
}

func (m model) renderOverlay(t theme) string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	var builder strings.Builder
	particleStyle := lipgloss.NewStyle().Foreground(t.particle)
	matrixStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3BFF63"))
	scanStyle := lipgloss.NewStyle().Foreground(t.scanline)
	accentStyle := lipgloss.NewStyle().Foreground(t.accent)
	highlightStyle := lipgloss.NewStyle().Foreground(t.highlight).Bold(true)

	portraitLeft, portraitRight, portraitTop, portraitBottom, hasPortrait := m.portraitBounds()

	for _, p := range m.particles {
		if hasPortrait && p.x >= portraitLeft && p.x <= portraitRight && p.y >= portraitTop && p.y <= portraitBottom {
			continue
		}
		if p.x < 0 || p.x >= m.width || p.y < 0 || p.y >= m.height {
			continue
		}
		builder.WriteString(cursorAt(p.x+1, p.y+1))
		builder.WriteString(particleStyle.Render(string(p.glyph)))
	}

	cornerGlyphs := []string{"·", "•", "✦", "•"}
	corner := cornerGlyphs[(m.uiPhase/2)%len(cornerGlyphs)]
	corners := [][2]int{{2, 2}, {m.width - 1, 2}, {2, m.height - 1}, {m.width - 1, m.height - 1}}
	for _, c := range corners {
		x := c[0]
		y := c[1]
		if x <= 0 || y <= 0 || x > m.width || y > m.height {
			continue
		}
		builder.WriteString(cursorAt(x, y))
		builder.WriteString(accentStyle.Render(corner))
	}

	railRange := max(1, m.height-4)
	railY := 3 + (m.uiPhase % railRange)
	if railY > 2 && railY < m.height {
		builder.WriteString(cursorAt(2, railY))
		builder.WriteString(accentStyle.Render("▌"))
		builder.WriteString(cursorAt(max(2, m.width-1), railY))
		builder.WriteString(accentStyle.Render("▐"))
	}

	if t.enableScanline && m.scanlineY >= 0 && m.scanlineY < m.height {
		for x := 1; x <= m.width; x += 2 {
			if hasPortrait {
				px := x - 1
				if px >= portraitLeft && px <= portraitRight && m.scanlineY >= portraitTop && m.scanlineY <= portraitBottom {
					continue
				}
			}
			builder.WriteString(cursorAt(x, m.scanlineY+1))
			builder.WriteString(scanStyle.Render("·"))
		}
	}

	if m.matrixMode {
		chars := []rune("01アイウカキクケコサシスセソ")
		for _, drop := range m.matrix {
			headY := int(drop.y)
			for i := 0; i < drop.length; i++ {
				y := headY - i
				if y < 0 || y >= m.height || drop.x < 0 || drop.x >= m.width {
					continue
				}
				builder.WriteString(cursorAt(drop.x+1, y+1))
				if i == 0 {
					builder.WriteString(matrixStyle.Render("█"))
				} else {
					builder.WriteString(matrixStyle.Render(string(chars[rand.Intn(len(chars))])))
				}
			}
		}
	}

	if m.transition.active {
		p := float64(m.transition.elapsed) / float64(m.transition.duration)
		if p < 0 {
			p = 0
		}
		if p > 1 {
			p = 1
		}
		cx := m.width / 2
		cy := m.height / 2
		glyphs := []rune{'*', '+', '·', 'x'}
		radius := 2 + int((1-p)*10)
		for i := 0; i < 24; i++ {
			a := float64(i)*0.52 + float64(m.uiPhase)/5
			x := cx + int(math.Cos(a)*float64(radius))
			y := cy + int(math.Sin(a)*float64(radius)/2)
			if x < 1 || x > m.width || y < 1 || y > m.height {
				continue
			}
			builder.WriteString(cursorAt(x, y))
			builder.WriteString(highlightStyle.Render(string(glyphs[i%len(glyphs)])))
		}
	}

	builder.WriteString(cursorAt(1, m.height))
	builder.WriteString("\x1b[0m")
	return builder.String()
}

func (m model) activeTheme() theme {
	if m.themeIndex < 0 || m.themeIndex >= len(m.themes) {
		return m.themes[0]
	}
	return m.themes[m.themeIndex]
}

func (m model) matrixTheme() theme {
	return theme{
		name:           "Matrix",
		primary:        "#7DFFB3",
		accent:         "#39FF14",
		highlight:      "#B3FFD9",
		particle:       "#76FF03",
		scanline:       "#0E4D20",
		enableScanline: true,
	}
}

func themePalette(t theme) []lipgloss.Color {
	return []lipgloss.Color{t.primary, t.accent, t.highlight, t.particle}
}

func rainbowText(text string, palette []lipgloss.Color, phase int, bold bool) string {
	if len(palette) == 0 || text == "" {
		return text
	}
	r := []rune(text)
	var b strings.Builder
	for i, ch := range r {
		if ch == ' ' {
			b.WriteRune(ch)
			continue
		}
		color := palette[(i+phase)%len(palette)]
		style := lipgloss.NewStyle().Foreground(color).Bold(bold)
		b.WriteString(style.Render(string(ch)))
	}
	return b.String()
}

func cursorAt(x, y int) string {
	return fmt.Sprintf("\x1b[%d;%dH", y, x)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
