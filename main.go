package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
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

type sceneMode int

const (
	sceneHome sceneMode = iota
	sceneProject
	sceneContact
)

type star struct {
	x          int
	y          int
	glyph      rune
	brightness int
	dx         int
	dy         int
}

type projectInfo struct {
	title       string
	asciiArt    []string
	description []string
}

type (
	starTickMsg    time.Time
	shimmerTickMsg time.Time
)

type model struct {
	width  int
	height int

	themes     []theme
	themeIndex int

	scene sceneMode

	stars []star

	portraitOriginal []string
	portraitFitted   []string

	shimmerActive bool
	shimmerX      int
	shimmerWait   int
	shimmerPhase  int

	introLines []string
	aboutLines []string

	projects    []projectInfo
	contact     []string
	contactHead []string
}

func main() {
	rand.Seed(time.Now().UnixNano())

	program := tea.NewProgram(
		newModel(),
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stdout),
		tea.WithAltScreen(),
	)

	if _, err := program.Run(); err != nil {
		fmt.Println("error running program:", err)
		os.Exit(1)
	}
}

func newModel() model {
	portrait := loadASCII("assets/me_ascii.txt", []string{"(portrait missing)"})

	pastryArt := []string{
		`:::::::::      :::      ::::::::  ::::::::::: :::::::::  :::   ::: `,
		`:+:    :+:   :+: :+:   :+:    :+:     :+:     :+:    :+: :+:   :+: `,
		`+:+    +:+  +:+   +:+  +:+            +:+     +:+    +:+  +:+ +:+  `,
		`+#++:++#+  +#++:++#++: +#++:++#++     +#+     +#++:++#:    +#++:   `,
		`+#+        +#+     +#+        +#+     +#+     +#+    +#+    +#+    `,
		`#+#        #+#     #+# #+#    #+#     #+#     #+#    #+#    #+#    `,
		`###        ###     ###  ########      ###     ###    ###    ###    `,
	}

	gyroArt := []string{
		`  ::::::::  :::   ::: :::::::::   ::::::::        ::::::::      :::     ::::::::: `,
		` :+:    :+: :+:   :+: :+:    :+: :+:    :+:      :+:    :+:   :+: :+:   :+:    :+:`,
		`+:+         +:+   +:+ +:+    +:+ +:+    +:+      +:+         +:+   +:+  +:+    +:+`,
		`:#:          +#++:++  +#++:++#:  +#+    +:+      +#+        +#++:++#++: +#++:++#: `,
		`+#+   +#+#    +#+     +#+    +#+ +#+    +#+      +#+        +#+     +#+ +#+    +#+`,
		`#+#    #+#    #+#     #+#    #+# #+#    #+#      #+#    #+# #+#     #+# #+#    #+#`,
		` ########     ###     ###    ###  ########        ########  ###     ### ###    ###`,
	}

	contactArt := []string{
		` ::::::::    ::::::::   ::::    :::  ::::::::::::     :::       ::::::::   :::::::::::`,
		`:+:    :+:  :+:    :+:  :+:+:   :+:      :+:        :+: :+:    :+:    :+:      :+:    `,
		`+:+         +:+    +:+  :+:+:+  +:+      +:+       +:+   +:+   +:+             +:+    `,
		`+#+         +#+    +:+  +#+ +:+ +#+      +#+      +#++:++#++   :+#             +#+    `,
		`+#+         +#+    +#+  +#+  +#+#+#      +#+      +#+     +#+  +#+    #+#      +#+   `,
		`#+#    #+#  #+#    #+#  #+#   #+#+#      #+#      #+#     #+#  #+#    #+#      #+#   `,
		` ########    ::::::::   ###    ####      ###      ###     ###   ########       ###   `,
	}

	projects := []projectInfo{
		{
			title:    "Pastry",
			asciiArt: pastryArt,
			description: []string{
				"A PASTEBIN / RENTRY CLONE.",
				"SIMPLE, FAST, AND CONTENT-FOCUSED.",
			},
		},
		{
			title:    "Gyro Car",
			asciiArt: gyroArt,
			description: []string{
				"ESP32 BASED CAR CONTROLLED VIA GYRO.",
				"CUSTOM MOBILE APP FOR GESTURE CONTROL.",
			},
		},
	}

	introLines := []string{
		"     :::    ::: :::        :::::::::  ::::::::::: :::::::: ",
		"    :+:    :+: :+:        :+:    :+:     :+:    :+:    :+:",
		"   +:+    +:+ +:+        +:+    +:+     +:+    +:+",
		"  +#+    +:+ +#+        +#++:++#:      +#+    +#+",
		" +#+    +#+ +#+        +#+    +#+     +#+    +#+",
		"#+#    #+# #+#        #+#    #+#     #+#    #+#    #+#",
		"########  ########## ###    ### ########### ########",
		"",
		"      ::::::::   ::::::::  :::        :::            :::      ::::::::   ::::::::",
		"    :+:    :+: :+:    :+: :+:        :+:          :+: :+:   :+:    :+: :+:    :+:",
		"   +:+        +:+    +:+ +:+        +:+         +:+   +:+  +:+        +:+    +:+",
		"  +#+        +#+    +:+ +#+        +#+        +#++:++#++: +#+        +#+    +:+",
		" +#+        +#+    +#+ +#+        +#+        +#+     +#+ +#+        +#+    +#+",
		"#+#    #+# #+#    #+# #+#        #+#        #+#     #+# #+#    #+# #+#    #+#",
		"########   ########  ########## ########## ###     ###  ########   ########",
	}

	m := model{
		themes: []theme{
			{name: "Cyber Green", primary: "#A4FFB0", accent: "#A4FFB0", highlight: "#43FF75", particle: "#4EFF78", scanline: "#1B5E20", enableScanline: true},
			{name: "Ocean Blue", primary: "#79D7FF", accent: "#79D7FF", highlight: "#4EA8FF", particle: "#6EC5FF", scanline: "#0D3C66", enableScanline: true},
			{name: "Amber Retro", primary: "#FFDCA3", accent: "#FFDCA3", highlight: "#FF9F1C", particle: "#FF6A3D", scanline: "#5A3D00", enableScanline: true},
		},
		themeIndex:       0,
		scene:            sceneHome,
		portraitOriginal: portrait,
		portraitFitted:   portrait,
		shimmerActive:    false,
		shimmerX:         -5,
		shimmerWait:      20,
		introLines:       introLines,
		aboutLines: []string{
			"Student at Fr. Conceicao Rodrigues College of Engineering",
			"Frontend Engineer | UI/UX Designer",
			"Building web apps and exploring cybersecurity",
		},
		projects:    projects,
		contactHead: contactArt,
		contact: []string{
			"GitHub:    https://github.com/ulric-collaco",
			"Email:     collacou@gmail.com",
			"LinkedIn:  https://www.linkedin.com/in/ulric-collaco/",
			"Instagram: https://www.instagram.com/ulric_collaco/",
		},
	}

	return m
}

func loadASCII(path string, fallback []string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		if alt := assetPath(path); alt != path {
			data, err = os.ReadFile(alt)
		}
		if err != nil {
			return fallback
		}
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

func assetPath(rel string) string {
	exe, err := os.Executable()
	if err != nil {
		return rel
	}
	base := filepath.Dir(exe)
	return filepath.Join(base, rel)
}

func (m model) Init() tea.Cmd {
	return tea.Batch(starTick(), shimmerTick())
}

func starTick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg { return starTickMsg(t) })
}

func shimmerTick() tea.Cmd {
	return tea.Tick(160*time.Millisecond, func(t time.Time) tea.Msg { return shimmerTickMsg(t) })
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.refreshLayout()
		m.reseedStars()
		return m, nil

	case tea.KeyMsg:
		s := msg.String()
		switch s {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "t":
			m.themeIndex = (m.themeIndex + 1) % len(m.themes)
			return m, nil
		case "left", "h":
			switch m.scene {
			case sceneHome:
				m.scene = sceneContact
			case sceneProject:
				m.scene = sceneHome
			case sceneContact:
				m.scene = sceneProject
			}
			return m, nil
		case "right", "l":
			switch m.scene {
			case sceneHome:
				m.scene = sceneProject
			case sceneProject:
				m.scene = sceneContact
			case sceneContact:
				m.scene = sceneHome
			}
			return m, nil
		}
		return m, nil

	case starTickMsg:
		m.updateStars()
		return m, starTick()

	case shimmerTickMsg:
		m.updateShimmer()
		return m, shimmerTick()
	}

	return m, nil
}

func (m *model) refreshLayout() {
	if m.width <= 0 || m.height <= 0 {
		return
	}

	// Portrait sizing - target 110x40, clamp with room for nav and padding.
	maxW := min(110, m.width)
	availableH := m.height - 4
	if availableH < 1 {
		availableH = 1
	}
	maxH := min(40, availableH)

	m.portraitFitted = fitASCIIToBox(m.portraitOriginal, maxW, maxH)
}

func fitASCIIToBox(lines []string, maxW, maxH int) []string {
	if len(lines) == 0 {
		return lines
	}
	if maxW <= 0 || maxH <= 0 {
		return []string{}
	}

	rows := lines
	if len(rows) > maxH {
		rows = rows[:maxH]
	}

	out := make([]string, 0, len(rows))
	for _, line := range rows {
		r := []rune(line)
		if len(r) <= maxW {
			out = append(out, line)
			continue
		}
		out = append(out, string(r[:maxW]))
	}
	return out
}

func (m *model) reseedStars() {
	if m.width <= 0 || m.height <= 0 {
		return
	}
	target := min(100, (m.width*m.height)/25)
	m.stars = make([]star, 0, target)
	glyphs := []rune{'.', '*', '·', '+'}
	for i := 0; i < target; i++ {
		m.stars = append(m.stars, star{
			x:          rand.Intn(max(1, m.width)),
			y:          rand.Intn(max(1, m.height)),
			glyph:      glyphs[rand.Intn(len(glyphs))],
			brightness: rand.Intn(3),
			dx:         rand.Intn(3) - 1,
			dy:         rand.Intn(3) - 1,
		})
	}
}

func (m *model) updateStars() {
	for i := range m.stars {
		if rand.Intn(4) == 0 {
			m.stars[i].brightness = rand.Intn(3)
		}
		if rand.Intn(8) == 0 {
			m.stars[i].x += m.stars[i].dx
		}
		if rand.Intn(12) == 0 {
			m.stars[i].y += m.stars[i].dy
		}

		if m.stars[i].x < 0 {
			m.stars[i].x = m.width - 1
		}
		if m.stars[i].x >= m.width {
			m.stars[i].x = 0
		}
		if m.stars[i].y < 0 {
			m.stars[i].y = m.height - 1
		}
		if m.stars[i].y >= m.height {
			m.stars[i].y = 0
		}
	}
}

func (m *model) updateShimmer() {
	m.shimmerPhase++
	if !m.shimmerActive {
		m.shimmerWait--
		if m.shimmerWait <= 0 {
			m.shimmerActive = true
			m.shimmerX = -4
		}
		return
	}
	m.shimmerX++
	if m.shimmerX > m.maxPortraitWidth()+4 {
		m.shimmerActive = false
		m.shimmerWait = 18 + rand.Intn(16)
	}
}

func (m model) View() string {
	if m.width <= 0 || m.height <= 0 {
		return ""
	}

	t := m.activeTheme()
	var b strings.Builder
	minW := m.minHomeWidth()
	minH := m.minHomeHeight()
	if m.width < minW || m.height < minH {
		lines := make([][]rune, m.height)
		colorMap := make(map[int]lipgloss.Color)
		boldMap := make(map[int]bool)

		for y := 0; y < m.height; y++ {
			lines[y] = []rune(strings.Repeat(" ", m.width))
		}

		box := renderTooSmallMessage(m.width, m.height, minW, minH)
		boxTop := max(0, (m.height-len(box))/2)
		for i, line := range box {
			y := boxTop + i
			if y < 0 || y >= m.height {
				continue
			}
			m.blitCenteredLine(lines, colorMap, boldMap, line, y, t.highlight, true)
		}

		for y := 0; y < m.height; y++ {
			for x := 0; x < m.width; x++ {
				key := y*m.width + x
				ch := string(lines[y][x])
				if c, ok := colorMap[key]; ok {
					style := lipgloss.NewStyle().Foreground(c).Bold(boldMap[key])
					b.WriteString(style.Render(ch))
				} else {
					b.WriteString(ch)
				}
			}
			if y < m.height-1 {
				b.WriteRune('\n')
			}
		}

		return b.String()
	}

	// Create canvas
	lines := make([][]rune, m.height)
	colorMap := make(map[int]lipgloss.Color)
	boldMap := make(map[int]bool)

	for y := 0; y < m.height; y++ {
		lines[y] = []rune(strings.Repeat(" ", m.width))
	}

	// Paint stars first
	m.paintStars(lines, colorMap, boldMap, t)

	// Render content
	page := m.renderScene(t)
	pageTop := max(0, (m.height-len(page))/2)
	if m.scene == sceneProject {
		blockW := 0
		for _, line := range page {
			w := len([]rune(stripANSI(line)))
			if w > blockW {
				blockW = w
			}
		}
		m.clearBlock(lines, colorMap, boldMap, pageTop, len(page), blockW)
	}

	for i, line := range page {
		y := pageTop + i
		if y < 0 || y >= m.height {
			continue
		}

		plain := stripANSI(line)
		lineColor := t.primary
		isBold := false

		// Highlight links
		if strings.Contains(plain, "http") || strings.Contains(plain, "@") {
			lineColor = t.highlight
			isBold = true
		}

		m.blitCenteredLine(lines, colorMap, boldMap, plain, y, lineColor, isBold)
	}

	// Nav
	nav := m.renderNav(t)
	navY := m.height - 2
	if navY >= 0 {
		m.blitCenteredLine(lines, colorMap, boldMap, stripANSI(nav), navY, t.highlight, true)
	}

	// Build final output
	for y := 0; y < m.height; y++ {
		for x := 0; x < m.width; x++ {
			key := y*m.width + x
			ch := string(lines[y][x])
			if c, ok := colorMap[key]; ok {
				style := lipgloss.NewStyle().Foreground(c).Bold(boldMap[key])
				b.WriteString(style.Render(ch))
			} else {
				b.WriteString(ch)
			}
		}
		if y < m.height-1 {
			b.WriteRune('\n')
		}
	}

	return b.String()
}

func renderTooSmallMessage(width, height, minW, minH int) []string {
	content := []string{
		"WINDOW TOO SMALL",
		"",
		fmt.Sprintf("Please resize to at least %dx%d", minW, minH),
		"",
		"FULLSCREEN TO VIEW",
		"",
		"Maximize or expand your terminal window",
		"",
		fmt.Sprintf("Current size: %dx%d", width, height),
		"",
		"Press 'q' to quit",
	}

	maxW := 0
	for _, line := range content {
		w := len([]rune(line))
		if w > maxW {
			maxW = w
		}
	}

	innerW := maxW + 2
	borderTop := "┌" + strings.Repeat("─", innerW) + "┐"
	borderBottom := "└" + strings.Repeat("─", innerW) + "┘"

	box := make([]string, 0, len(content)+2)
	box = append(box, borderTop)
	for _, line := range content {
		pad := innerW - len([]rune(line))
		left := pad / 2
		right := pad - left
		box = append(box, "│"+strings.Repeat(" ", left)+line+strings.Repeat(" ", right)+"│")
	}
	box = append(box, borderBottom)

	return box
}

func (m model) minHomeWidth() int {
	portraitW := 0
	for _, l := range m.portraitOriginal {
		w := len([]rune(l))
		if w > portraitW {
			portraitW = w
		}
	}
	introW := 0
	for _, l := range m.introLines {
		w := len([]rune(l))
		if w > introW {
			introW = w
		}
	}
	gap := 4
	return portraitW + gap + introW
}

func (m model) minHomeHeight() int {
	portraitH := len(m.portraitOriginal)
	introH := len(m.introLines)
	aboutH := len(m.aboutLines)
	contentH := introH + 1 + aboutH
	pageH := max(portraitH, contentH)
	return pageH + 2
}

func (m model) blitCenteredLine(lines [][]rune, colorMap map[int]lipgloss.Color, boldMap map[int]bool, text string, y int, color lipgloss.Color, bold bool) {
	r := []rune(text)
	if len(r) == 0 || y < 0 || y >= m.height {
		return
	}
	x0 := (m.width - len(r)) / 2
	if x0 < 0 {
		x0 = 0
	}
	for i, ch := range r {
		x := x0 + i
		if x < 0 || x >= m.width {
			continue
		}
		lines[y][x] = ch
		key := y*m.width + x
		colorMap[key] = color
		boldMap[key] = bold
	}
}

func (m model) clearBlock(lines [][]rune, colorMap map[int]lipgloss.Color, boldMap map[int]bool, top, height, width int) {
	if height <= 0 || width <= 0 {
		return
	}
	x0 := (m.width - width) / 2
	if x0 < 0 {
		x0 = 0
	}
	x1 := min(m.width, x0+width)
	for y := top; y < top+height; y++ {
		if y < 0 || y >= m.height {
			continue
		}
		for x := x0; x < x1; x++ {
			lines[y][x] = ' '
			key := y*m.width + x
			delete(colorMap, key)
			delete(boldMap, key)
		}
	}
}

func (m model) paintStars(lines [][]rune, colorMap map[int]lipgloss.Color, boldMap map[int]bool, t theme) {
	palette := []lipgloss.Color{t.scanline, t.particle, t.accent}
	for _, s := range m.stars {
		if s.x < 0 || s.x >= m.width || s.y < 0 || s.y >= m.height {
			continue
		}
		lines[s.y][s.x] = s.glyph
		idx := min(s.brightness, len(palette)-1)
		key := s.y*m.width + s.x
		colorMap[key] = palette[idx]
		boldMap[key] = idx == 2
	}
}

func (m model) renderScene(t theme) []string {
	switch m.scene {
	case sceneProject:
		return m.renderProjectScene(t)
	case sceneContact:
		return m.renderContactScene(t)
	default:
		return m.renderHomeScene(t)
	}
}

func centerLines(lines []string, width int) []string {
	if width <= 0 {
		return lines
	}
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		pad := width - len([]rune(line))
		if pad <= 0 {
			out = append(out, line)
			continue
		}
		left := pad / 2
		out = append(out, strings.Repeat(" ", left)+line)
	}
	return out
}

func interleaveBlank(lines []string) []string {
	if len(lines) == 0 {
		return lines
	}
	out := make([]string, 0, len(lines)*2-1)
	for i, line := range lines {
		out = append(out, line)
		if i < len(lines)-1 {
			out = append(out, "")
		}
	}
	return out
}

func padRightLines(lines []string, width int) []string {
	if width <= 0 {
		return lines
	}
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		pad := width - len([]rune(line))
		if pad > 0 {
			out = append(out, line+strings.Repeat(" ", pad))
		} else {
			out = append(out, line)
		}
	}
	return out
}

func alignContactLines(lines []string) []string {
	maxLabel := 0
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}
		label := strings.TrimSpace(parts[0])
		w := len([]rune(label))
		if w > maxLabel {
			maxLabel = w
		}
	}

	out := make([]string, 0, len(lines))
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			out = append(out, line)
			continue
		}
		label := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		pad := maxLabel - len([]rune(label))
		if pad < 0 {
			pad = 0
		}
		out = append(out, label+":"+strings.Repeat(" ", pad+2)+value)
	}
	return out
}

func (m model) renderHomeScene(t theme) []string {
	portrait := m.renderPortrait(t)
	intro := m.introLines
	about := m.aboutLines

	portraitW := 0
	for _, l := range portrait {
		w := len([]rune(stripANSI(l)))
		if w > portraitW {
			portraitW = w
		}
	}

	introW := 0
	for _, l := range intro {
		w := len([]rune(l))
		if w > introW {
			introW = w
		}
	}
	for _, l := range about {
		w := len([]rune(l))
		if w > introW {
			introW = w
		}
	}

	about = centerLines(about, introW)
	about = interleaveBlank(about)

	gap := 4
	totalW := portraitW + gap + introW

	// If too wide, stack vertically
	if totalW > m.width-4 {
		var out []string
		out = append(out, portrait...)
		out = append(out, "")
		out = append(out, intro...)
		out = append(out, "")
		out = append(out, about...)
		return out
	}

	// Side by side
	leftX := (m.width - totalW) / 2
	rightX := leftX + portraitW + gap

	allIntro := append(intro, "")
	allIntro = append(allIntro, about...)

	blockH := max(len(portrait), len(allIntro))
	lines := make([]string, blockH)

	for i := 0; i < blockH; i++ {
		row := []rune(strings.Repeat(" ", m.width))

		if i < len(portrait) {
			for j, ch := range []rune(stripANSI(portrait[i])) {
				x := leftX + j
				if x >= 0 && x < m.width {
					row[x] = ch
				}
			}
		}

		if i < len(allIntro) {
			for j, ch := range []rune(allIntro[i]) {
				x := rightX + j
				if x >= 0 && x < m.width {
					row[x] = ch
				}
			}
		}

		lines[i] = string(row)
	}

	return lines
}

func (m model) renderProjectScene(t theme) []string {
	var out []string

	for i, p := range m.projects {
		for _, l := range p.asciiArt {
			out = append(out, l)
		}
		out = append(out, "")
		for _, d := range p.description {
			out = append(out, d)
		}

		if i < len(m.projects)-1 {
			out = append(out, "", "")
		}
	}

	out = append(out, "", "", "FOR MORE PROJECTS VISIT:")
	out = append(out, "https://ulriccollaco.me")

	return out
}

func (m model) renderContactScene(t theme) []string {
	var out []string
	alignedContacts := alignContactLines(m.contact)
	spacedContacts := interleaveBlank(alignedContacts)
	all := make([]string, 0, len(m.contactHead)+len(spacedContacts)+2)
	all = append(all, m.contactHead...)
	all = append(all, "", "")
	all = append(all, spacedContacts...)
	maxW := 0
	for _, line := range all {
		w := len([]rune(line))
		if w > maxW {
			maxW = w
		}
	}
	all = padRightLines(all, maxW)
	out = append(out, all...)
	return out
}

func (m model) renderPortrait(t theme) []string {
	art := m.portraitFitted
	baseColor := t.primary
	base := lipgloss.NewStyle().Foreground(baseColor)
	bright := lipgloss.NewStyle().Foreground(t.highlight).Bold(true)

	out := make([]string, 0, len(art))
	for y, line := range art {
		var b strings.Builder
		for x, ch := range []rune(line) {
			if ch == ' ' {
				b.WriteRune(ch)
				continue
			}
			dx := x - m.shimmerX
			if m.shimmerActive && dx >= -1 && dx <= 1 && pseudoRand(x, y, m.shimmerPhase)%3 == 0 {
				b.WriteString(bright.Render(string(ch)))
			} else {
				b.WriteString(base.Render(string(ch)))
			}
		}
		out = append(out, b.String())
	}
	return out
}

func (m model) renderNav(t theme) string {
	leftArrow := "◄"
	rightArrow := "►"

	var leftLabel, currentLabel, rightLabel string

	switch m.scene {
	case sceneHome:
		leftLabel = "Contact"
		currentLabel = "HOME"
		rightLabel = "Projects"
	case sceneProject:
		leftLabel = "Home"
		currentLabel = "PROJECTS"
		rightLabel = "Contact"
	case sceneContact:
		leftLabel = "Projects"
		currentLabel = "CONTACT"
		rightLabel = "Home"
	}

	return fmt.Sprintf("%s %s %s %s %s  |  %s/%s navigate  t theme  q quit",
		leftLabel, leftArrow, currentLabel, rightArrow, rightLabel, leftArrow, rightArrow)
}

func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inEsc {
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inEsc = false
			}
			continue
		}
		if ch == 0x1b {
			inEsc = true
			continue
		}
		b.WriteByte(ch)
	}
	return b.String()
}

func (m model) maxPortraitWidth() int {
	w := 0
	for _, line := range m.portraitFitted {
		lw := len([]rune(line))
		if lw > w {
			w = lw
		}
	}
	return w
}

func (m model) activeTheme() theme {
	if m.themeIndex < 0 || m.themeIndex >= len(m.themes) {
		return m.themes[0]
	}
	return m.themes[m.themeIndex]
}

func pseudoRand(x, y, phase int) int {
	v := x*73 + y*131 + phase*29 + 97
	if v < 0 {
		v = -v
	}
	return v
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
