package main

import (
	"fmt"
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

type transition struct {
	active    bool
	elapsed   time.Duration
	duration  time.Duration
	fromScene sceneMode
	toScene   sceneMode
	direction int
}

type projectInfo struct {
	title       string
	asciiArt    []string
	description []string
}

type (
	starTickMsg       time.Time
	revealTickMsg     time.Time
	typeTickMsg       time.Time
	shimmerTickMsg    time.Time
	transitionTickMsg time.Time
)

const (
	typeTickInterval  = 15 * time.Millisecond // Faster typing
	startupTotalTicks = 40                    // Faster total startup animation
)

type model struct {
	width  int
	height int

	themes     []theme
	themeIndex int

	scene sceneMode

	transition transition

	stars []star

	portraitOriginal []string
	portraitFitted   []string
	revealLines      int
	revealDone       bool

	shimmerActive bool
	shimmerX      int
	shimmerWait   int
	shimmerPhase  int

	introLines   []string
	introText    string
	introRunes   []rune
	introVisible int
	startupTicks int
	aboutLines   []string

	projects    []projectInfo
	contact     []string
	contactHead []string
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

	// "PASTRY" in Alligator/Doom style
	pastryArt := []string{
		`:::::::::      :::      ::::::::  ::::::::::: :::::::::  :::   ::: `,
		`:+:    :+:   :+: :+:   :+:    :+:     :+:     :+:    :+: :+:   :+: `,
		`+:+    +:+  +:+   +:+  +:+            +:+     +:+    +:+  +:+ +:+  `,
		`+#++:++#+  +#++:++#++: +#++:++#++     +#+     +#++:++#:    +#++:   `,
		`+#+        +#+     +#+        +#+     +#+     +#+    +#+    +#+    `,
		`#+#        #+#     #+# #+#    #+#     #+#     #+#    #+#    #+#    `,
		`###        ###     ###  ########      ###     ###    ###    ###    `,
	}

	// "GYRO CAR" in Alligator/Doom style
	gyroArt := []string{
		`  ::::::::  :::   ::: :::::::::   ::::::::        ::::::::      :::     ::::::::: `,
		` :+:    :+: :+:   :+: :+:    :+: :+:    :+:      :+:    :+:   :+: :+:   :+:    :+:`,
		`+:+         +:+   +:+ +:+    +:+ +:+    +:+      +:+         +:+   +:+  +:+    +:+`,
		`:#:          +#++:++  +#++:++#:  +#+    +:+      +#+        +#++:++#++: +#++:++#: `,
		`+#+   +#+#    +#+     +#+    +#+ +#+    +#+      +#+        +#+     +#+ +#+    +#+`,
		`#+#    #+#    #+#     #+#    #+# #+#    #+#      #+#    #+# #+#     #+# #+#    #+#`,
		` ########     ###     ###    ###  ########        ########  ###     ### ###    ###`,
	}

	// "CONTACT" in Alligator/Doom style - Cleaner version
	contactArt := []string{
		` ::::::::  ::::::::  ::::    ::: :::::::::::     :::     :::::::: :::::::::::`,
		`:+:    :+: :+:    :+: :+:+:   :+:     :+:       :+: :+:  :+:    :+:    :+:    `,
		`+:+        +:+    +:+ :+:+:+  +:+     +:+      +:+   +:+ +:+           +:+    `,
		`+#+        +#+    +:+ +#+ +:+ +#+     +#+     +#++:++#++ :+#+          +#+    `,
		`+#+        +#+    +#+ +#+  +#+#+#     +#+     +#+     +#+ +#+   +#+#    +#+   `,
		`#+#    #+# #+#    #+# #+#   #+#+#     #+#     #+#     #+# #+#    #+#    #+#   `,
		` ########   ::::::::  ###    ####     ###     ###     ###  ########     ###   `,
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
	introJoined := strings.Join(introLines, "\n")

	m := model{
		themes: []theme{
			{name: "Cyber Green", primary: "#A4FFB0", accent: "#43FF75", highlight: "#D8FFE0", particle: "#4EFF78", scanline: "#1B5E20", enableScanline: true},
			{name: "Ocean Blue", primary: "#79D7FF", accent: "#4EA8FF", highlight: "#A8FFEC", particle: "#6EC5FF", scanline: "#0D3C66", enableScanline: true},
			{name: "Amber Retro", primary: "#FFDCA3", accent: "#FF9F1C", highlight: "#FFF0A3", particle: "#FF6A3D", scanline: "#5A3D00", enableScanline: true},
			{name: "Minimal White", primary: "#F5F5F5", accent: "#9EC5FF", highlight: "#FFE9A8", particle: "#BEE7D3", scanline: "#7A7A7A", enableScanline: false},
		},
		themeIndex:       1,
		scene:            sceneHome,
		portraitOriginal: portrait,
		portraitFitted:   portrait,
		revealLines:      0,
		revealDone:       false,
		shimmerActive:    false,
		shimmerX:         -5,
		shimmerWait:      20,
		introLines:       introLines,
		introText:        introJoined,
		introRunes:       []rune(introJoined),
		aboutLines: []string{
			"ABOUT ME",
			"I build fast, break limits, and ship clean.",
			"Web x Cybersecurity is my home turf.",
			"I made a Pastebin/Rentry-style clone.",
			"I built a gyro car and my own control app.",
			"I turn rough ideas into working systems.",
			"Currently grinding hard for internships.",
			"Always building. Always learning. Always shipping.",
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
	return tea.Batch(starTick(), revealTick(), typeTick(), shimmerTick())
}

func starTick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(t time.Time) tea.Msg { return starTickMsg(t) })
}

func revealTick() tea.Cmd {
	return tea.Tick(40*time.Millisecond, func(t time.Time) tea.Msg { return revealTickMsg(t) })
}

func typeTick() tea.Cmd {
	return tea.Tick(typeTickInterval, func(t time.Time) tea.Msg { return typeTickMsg(t) })
}

func shimmerTick() tea.Cmd {
	return tea.Tick(160*time.Millisecond, func(t time.Time) tea.Msg { return shimmerTickMsg(t) })
}

func transitionTick() tea.Cmd {
	return tea.Tick(20*time.Millisecond, func(t time.Time) tea.Msg { return transitionTickMsg(t) })
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
		case "left":
			if m.transition.active {
				return m, nil
			}
			to := m.scene
			switch m.scene {
			case sceneHome:
				to = sceneContact
			case sceneProject:
				to = sceneHome
			case sceneContact:
				to = sceneProject
			}
			if to != m.scene {
				m.startTransition(to, -1)
				return m, transitionTick()
			}
			return m, nil
		case "right":
			if m.transition.active {
				return m, nil
			}
			to := m.scene
			switch m.scene {
			case sceneHome:
				to = sceneProject
			case sceneProject:
				to = sceneContact
			case sceneContact:
				to = sceneHome
			}
			if to != m.scene {
				m.startTransition(to, 1)
				return m, transitionTick()
			}
			return m, nil
		}
		return m, nil

	case starTickMsg:
		m.updateStars()
		return m, starTick()

	case revealTickMsg:
		if !m.revealDone {
			m.revealLines += 3 + rand.Intn(2)
			if m.revealLines >= len(m.portraitFitted) {
				m.revealLines = len(m.portraitFitted)
				m.revealDone = true
			}
		}
		return m, revealTick()

	case typeTickMsg:
		if m.scene == sceneHome && m.introVisible < len(m.introRunes) {
			m.startupTicks++
			remainingTicks := max(1, startupTotalTicks-m.startupTicks+1)
			remainingChars := len(m.introRunes) - m.introVisible
			step := (remainingChars + remainingTicks - 1) / remainingTicks
			if step < 1 {
				step = 1
			}
			m.introVisible += step
			if m.introVisible > len(m.introRunes) {
				m.introVisible = len(m.introRunes)
			}
		}
		return m, typeTick()

	case shimmerTickMsg:
		m.updateShimmer()
		return m, shimmerTick()

	case transitionTickMsg:
		if !m.transition.active {
			return m, nil
		}
		m.transition.elapsed += 20 * time.Millisecond
		if m.transition.elapsed >= m.transition.duration {
			m.scene = m.transition.toScene
			m.transition.active = false
			m.transition.elapsed = 0
		}
		if m.transition.active {
			return m, transitionTick()
		}
		return m, nil
	}

	return m, nil
}

func (m *model) startTransition(to sceneMode, direction int) {
	m.transition = transition{
		active:    true,
		elapsed:   0,
		duration:  220 * time.Millisecond,
		fromScene: m.scene,
		toScene:   to,
		direction: direction,
	}
}

func (m *model) refreshLayout() {
	if m.width <= 0 || m.height <= 0 {
		return
	}

	contentTop := max(2, m.height/10)
	contentBottom := m.height - 4
	maxH := max(6, (contentBottom-contentTop)*60/100)
	maxW := max(18, m.width*34/100)
	fitted := fitASCIIToBox(m.portraitOriginal, maxW, maxH)
	m.portraitFitted = fitted
	if m.revealDone {
		m.revealLines = len(m.portraitFitted)
	} else if m.revealLines > len(m.portraitFitted) {
		m.revealLines = len(m.portraitFitted)
	}
}

func fitASCIIToBox(lines []string, maxW, maxH int) []string {
	if len(lines) == 0 {
		return lines
	}
	if maxW <= 0 || maxH <= 0 {
		return []string{}
	}

	// Downsample rows when needed, then center-crop remaining height.
	rows := lines
	if len(rows) > maxH {
		step := float64(len(rows)) / float64(maxH)
		tmp := make([]string, 0, maxH)
		for i := 0; i < maxH; i++ {
			idx := int(float64(i) * step)
			if idx < 0 {
				idx = 0
			}
			if idx >= len(rows) {
				idx = len(rows) - 1
			}
			tmp = append(tmp, rows[idx])
		}
		rows = tmp
	}
	if len(rows) > maxH {
		start := (len(rows) - maxH) / 2
		rows = rows[start : start+maxH]
	}

	out := make([]string, 0, len(rows))
	for _, line := range rows {
		r := []rune(line)
		if len(r) <= maxW {
			out = append(out, line)
			continue
		}
		start := (len(r) - maxW) / 2
		out = append(out, string(r[start:start+maxW]))
	}
	return out
}

func (m *model) reseedStars() {
	if m.width <= 0 || m.height <= 0 {
		return
	}
	target := max(140, (m.width*m.height)/20)
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
	if len(m.stars) == 0 {
		m.reseedStars()
	}
	for i := range m.stars {
		if rand.Intn(3) == 0 {
			m.stars[i].brightness = rand.Intn(3)
		}
		if rand.Intn(6) == 0 {
			m.stars[i].x += m.stars[i].dx
		}
		if rand.Intn(10) == 0 {
			m.stars[i].y += m.stars[i].dy
		}
		if rand.Intn(20) == 0 {
			m.stars[i].dx = rand.Intn(3) - 1
			m.stars[i].dy = rand.Intn(3) - 1
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
	if !m.revealDone {
		return
	}
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
	lines := make([][]rune, m.height)
	for y := 0; y < m.height; y++ {
		lines[y] = []rune(strings.Repeat(" ", m.width))
	}

	colorMap := make(map[int]lipgloss.Color, m.width*m.height)
	boldMap := make(map[int]bool, m.width*m.height)

	m.paintStars(lines, colorMap, boldMap, t)

	var page []string
	if m.transition.active {
		page = m.renderTransitionFrame(t)
	} else {
		page = m.renderScene(t, m.scene)
	}

	pageTop := max(1, (m.height-len(page))/2)
	for i, raw := range page {
		y := pageTop + i
		if y < 0 || y >= m.height {
			continue
		}

		// Use t.primary as default, but try to respect existing ANSI if possible?
		// Given we don't have a full parser, let's revert to a simpler stable rendering:
		// 1. Strip ANSI for layout.
		// 2. Use a heuristic that is STABLE (checking first few chars? checking length?)
		// 3. BUT the user specifically wants the portrait to be ONE color (t.accent/t.primary).
		// Currently it's mixed because my previous heuristic was per-line.
		
		plain := stripANSI(raw)
		
		// Improved Heuristic:
		// If it's a link, highlight it.
		// If it's ASCII art (Project Title, Contact Art, Portrait), accent/highlight it.
		// If it's normal text, primary.
		
		lineColor := t.primary
		isBold := false
		
		if strings.Contains(plain, "http") || strings.Contains(plain, "www.") || strings.Contains(plain, "@") {
			lineColor = t.highlight
			isBold = true 
		} else if isAsciiArtLine(plain) {
			lineColor = t.accent
		}
		
		// Specific override for "About Me" headers or known text logic? 
		// "ABOUT ME" is all caps, short. 
		if plain == "ABOUT ME" {
			lineColor = t.highlight
			isBold = true
		}

		m.blitCenteredLine(lines, colorMap, boldMap, plain, y, lineColor, isBold)
	}

	nav := m.renderNav(t)
	navY := m.height - 2
	if navY >= 0 && navY < m.height {
		m.blitCenteredLine(lines, colorMap, boldMap, stripANSI(nav), navY, t.highlight, true)
	}

	return buildFrame(lines, colorMap, boldMap, m.width, m.height)
}

func isAsciiArtLine(s string) bool {
	// A line is likely ASCII art if it has extensive use of specific chars
	// and low use of normal alphabet.
	// But "Project Title" has alphabet.
	// Let's check for density of art chars.
	artChars := ":+*#_/-|\\."
	count := 0
	for _, r := range s {
		if strings.ContainsRune(artChars, r) {
			count++
		}
	}
	// If more than 30% or just strictly > 3 chars? 
	// The portrait has lines with just "   :::   " -> 3+ chars.
	// Normal text "I build things" -> 0 art chars.
	return count >= 3
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

func (m model) paintStars(lines [][]rune, colorMap map[int]lipgloss.Color, boldMap map[int]bool, t theme) {
	palette := []lipgloss.Color{t.scanline, t.particle, t.accent}
	for _, s := range m.stars {
		if s.x < 0 || s.x >= m.width || s.y < 0 || s.y >= m.height {
			continue
		}
		lines[s.y][s.x] = s.glyph
		idx := s.brightness
		if idx < 0 {
			idx = 0
		}
		if idx >= len(palette) {
			idx = len(palette) - 1
		}
		key := s.y*m.width + s.x
		colorMap[key] = palette[idx]
		boldMap[key] = idx == 2
	}
}

func (m model) renderTransitionFrame(t theme) []string {
	from := m.renderScene(t, m.transition.fromScene)
	to := m.renderScene(t, m.transition.toScene)

	height := max(len(from), len(to))
	width := max(linesWidth(from), linesWidth(to))
	from = padLines(from, width, height)
	to = padLines(to, width, height)

	p := float64(m.transition.elapsed) / float64(m.transition.duration)
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	cut := int(float64(width) * p)

	out := make([]string, height)
	for i := 0; i < height; i++ {
		if m.transition.direction >= 0 {
			out[i] = wipeStyled(from[i], to[i], cut, false)
		} else {
			out[i] = wipeStyled(from[i], to[i], cut, true)
		}
	}
	return out
}

func (m model) renderScene(t theme, scene sceneMode) []string {
	switch scene {
	case sceneProject:
		return m.renderProjectScene(t)
	case sceneContact:
		return m.renderContactScene(t)
	default:
		return m.renderHomeScene(t)
	}
}

func (m model) renderHomeScene(t theme) []string {
	portrait := m.renderPortrait(t)
	intro := m.renderIntroText(t)
	about := m.aboutLines

	portraitPlain := make([]string, 0, len(portrait))
	portraitW := 0
	for _, l := range portrait {
		p := stripANSI(l)
		portraitPlain = append(portraitPlain, p)
		w := len([]rune(p))
		if w > portraitW {
			portraitW = w
		}
	}

	introPlain := make([]string, 0, len(intro)+1+len(about))
	introW := 0
	for _, l := range intro {
		p := stripANSI(l)
		introPlain = append(introPlain, p)
		w := len([]rune(p))
		if w > introW {
			introW = w
		}
	}
	if len(introPlain) > 0 {
		introPlain = append(introPlain, "")
	}
	for _, l := range about {
		introPlain = append(introPlain, l)
		w := len([]rune(l))
		if w > introW {
			introW = w
		}
	}

	gap := 4
	totalW := portraitW + gap + introW
	if totalW > m.width-2 {
		gap = 2
		totalW = portraitW + gap + introW
	}
	leftX := max(1, (m.width-totalW)/2)
	rightX := leftX + portraitW + gap

	blockH := max(len(portraitPlain), len(introPlain))
	if blockH <= 0 {
		blockH = 1
	}
	lines := make([]string, blockH)
	for i := 0; i < blockH; i++ {
		row := []rune(strings.Repeat(" ", max(1, m.width)))
		if i < len(portraitPlain) {
			for j, ch := range []rune(portraitPlain[i]) {
				x := leftX + j
				if x >= 0 && x < len(row) {
					row[x] = ch
				}
			}
		}
		if i < len(introPlain) {
			for j, ch := range []rune(introPlain[i]) {
				x := rightX + j
				if x >= 0 && x < len(row) {
					row[x] = ch
				}
			}
		}
		lines[i] = string(row)
	}

	return lines
}

func (m model) renderProjectScene(t theme) []string {
	p1 := m.renderProjectBlock(t, m.projects[0])
	p2 := m.renderProjectBlock(t, m.projects[1])

	w1 := m.blockWidth(p1)
	w2 := m.blockWidth(p2)
	gap := 8 // Increased gap
	totalW := w1 + gap + w2

	var combined []string
	if totalW <= m.width {
		h := max(len(p1), len(p2))
		combined = make([]string, h)
		// Calculate simple centering
		startX := (m.width - totalW) / 2
		if startX < 0 {
			startX = 0
		}
		leftPad := strings.Repeat(" ", startX)

		// Separator line
		sepStyle := lipgloss.NewStyle().Foreground(t.scanline)

		// Adjust gap for separator
		// Layout: [LeftPad][P1][Gap/2][Sep][Gap/2][P2]
		// Actually let's just use gap spaces and put sep in middle?
		// Gap is 8. P1 ... 3 spaces ... SEP ... 3 spaces ... P2
		// Total gap width logic: 3 + 3 + 3 = 9? No.
		// Let's do: P1 + "   │    " + P2.
		// Gap string:
		gapStr := sepStyle.Render("   │    ") // 3 spaces, bar, 4 spaces?
		// Or just centered bar.

		for i := 0; i < h; i++ {
			line1 := ""
			if i < len(p1) {
				line1 = p1[i]
			}
			line2 := ""
			if i < len(p2) {
				line2 = p2[i]
			}
			// Pad line1 to w1, preserving ANSI
			l1W := len([]rune(stripANSI(line1)))
			if l1W < w1 {
				line1 += strings.Repeat(" ", w1-l1W)
			}
			combined[i] = leftPad + line1 + gapStr + line2
		}
	} else {
		// Stacked
		combined = append(combined, p1...)
		combined = append(combined, "") // spacer
		combined = append(combined, "") // spacer
		// Horizontal separator
		sepLine := strings.Repeat("─", min(40, m.width-4))
		combined = append(combined, centerStyled(lipgloss.NewStyle().Foreground(t.scanline).Render(sepLine), m.width))
		combined = append(combined, "")
		combined = append(combined, "")
		combined = append(combined, p2...)
		// Center horizontally
		for i := range combined {
			combined[i] = centerStyled(combined[i], m.width)
		}
	}

	// Footer: "for more projects visit" + ASCII URL
	footerText := "FOR MORE PROJECTS VISIT:"
	footerStyle := lipgloss.NewStyle().Foreground(t.primary).Bold(true)

	// Regular big font - Styled
	siteLink := "https://ulriccollaco.me"
	siteStyle := lipgloss.NewStyle().
		Foreground(t.highlight).
		Background(t.scanline).
		Bold(true).
		Padding(0, 2).
		MarginTop(1)

	renderedLink := siteStyle.Render(siteLink)

	combined = append(combined, "", "", centerStyled(footerStyle.Render(footerText), m.width))
	combined = append(combined, centerStyled(renderedLink, m.width))
	return combined
}

func (m model) blockWidth(lines []string) int {
	w := 0
	for _, l := range lines {
		lw := len([]rune(stripANSI(l)))
		if lw > w {
			w = lw
		}
	}
	return w
}

func (m model) renderProjectBlock(t theme, p projectInfo) []string {
	artStyle := lipgloss.NewStyle().Foreground(t.accent)
	descStyle := lipgloss.NewStyle().Foreground(t.primary).Bold(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(t.scanline).Padding(0, 1)

	out := make([]string, 0, len(p.asciiArt)+len(p.description)+3)
	for _, l := range p.asciiArt {
		out = append(out, artStyle.Render(l))
	}
	out = append(out, "")

	// Render description as a block
	descText := strings.Join(p.description, "\n")
	descRendered := descStyle.Render(descText)
	descLines := strings.Split(descRendered, "\n")
	out = append(out, descLines...)

	// Center content within the block's own width
	currW := m.blockWidth(out)
	for i := range out {
		out[i] = lipgloss.PlaceHorizontal(currW, lipgloss.Center, out[i])
	}
	return out
}

func (m model) renderContactScene(t theme) []string {
	artStyle := lipgloss.NewStyle().Foreground(t.accent)
	detailStyle := lipgloss.NewStyle().Foreground(t.primary)

	var art []string
	for _, l := range m.contactHead {
		art = append(art, artStyle.Render(l))
	}

	out := append(art, "", "")
	for _, c := range m.contact {
		out = append(out, centerStyled(detailStyle.Render(c), m.width))
	}

	for i := range out {
		out[i] = centerStyled(out[i], m.width)
	}

	return out
}

func (m model) renderPortrait(t theme) []string {
	visible := min(m.revealLines, len(m.portraitFitted))
	if visible < 0 {
		visible = 0
	}
	art := m.portraitFitted[:visible]

	base := lipgloss.NewStyle().Foreground(t.accent)
	bright := lipgloss.NewStyle().Foreground(t.highlight).Bold(true)

	out := make([]string, 0, len(art))
	for y, line := range art {
		if !m.revealDone {
			out = append(out, base.Render(line))
			continue
		}

		var b strings.Builder
		for x, ch := range []rune(line) {
			if ch == ' ' {
				b.WriteRune(ch)
				continue
			}
			dx := x - m.shimmerX
			if m.shimmerActive && dx >= -1 && dx <= 1 && pseudoRand(x, y, m.shimmerPhase)%3 == 0 {
				b.WriteString(bright.Render(string(ch)))
				continue
			}
			b.WriteString(base.Render(string(ch)))
		}
		out = append(out, b.String())
	}
	return out
}

func (m model) renderIntroText(t theme) []string {
	visible := ""
	if m.introVisible > 0 {
		visible = string(m.introRunes[:min(m.introVisible, len(m.introRunes))])
	}
	lines := strings.Split(visible, "\n")
	style := lipgloss.NewStyle().Foreground(t.primary)
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		out = append(out, style.Render(l))
	}
	return out
}

func (m model) renderNav(t theme) string {
	leftArrow := lipgloss.NewStyle().Foreground(t.accent).Render("◄")
	rightArrow := lipgloss.NewStyle().Foreground(t.accent).Render("►")

	var leftLabel, currentLabel, rightLabel string

	// Navigation map based on left/right logic
	// Logic:
	// Home: Left->Contact, Right->Projects
	// Projects: Left->Home, Right->Contact
	// Contact: Left->Projects, Right->Home

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

	leftStyle := lipgloss.NewStyle().Foreground(t.primary).Faint(true)
	currStyle := lipgloss.NewStyle().Foreground(t.highlight).Bold(true)
	rightStyle := lipgloss.NewStyle().Foreground(t.primary).Faint(true)

	// Format:  Contact ◄ HOME ► Projects
	line := fmt.Sprintf("%s %s %s %s %s",
		leftStyle.Render(leftLabel),
		leftArrow,
		currStyle.Render(currentLabel),
		rightArrow,
		rightStyle.Render(rightLabel),
	)

	return centerStyled(line, m.width)
}

func buildFrame(lines [][]rune, colorMap map[int]lipgloss.Color, boldMap map[int]bool, width, height int) string {
	var b strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			key := y*width + x
			ch := string(lines[y][x])
			if c, ok := colorMap[key]; ok {
				style := lipgloss.NewStyle().Foreground(c).Bold(boldMap[key])
				b.WriteString(style.Render(ch))
			} else {
				b.WriteString(ch)
			}
		}
		if y < height-1 {
			b.WriteRune('\n')
		}
	}
	return b.String()
}

func wipeStyled(from, to string, cut int, reverse bool) string {
	rFrom := []rune(stripANSI(from))
	rTo := []rune(stripANSI(to))
	w := max(len(rFrom), len(rTo))
	rFrom = padRunes(rFrom, w)
	rTo = padRunes(rTo, w)
	if cut < 0 {
		cut = 0
	}
	if cut > w {
		cut = w
	}
	out := make([]rune, w)
	if !reverse {
		copy(out[:cut], rTo[:cut])
		copy(out[cut:], rFrom[cut:])
	} else {
		split := w - cut
		copy(out[:split], rFrom[:split])
		copy(out[split:], rTo[split:])
	}
	return string(out)
}

func linesWidth(lines []string) int {
	w := 0
	for _, l := range lines {
		lw := len([]rune(stripANSI(l)))
		if lw > w {
			w = lw
		}
	}
	return w
}

func padLines(lines []string, width, height int) []string {
	out := make([]string, height)
	for i := 0; i < height; i++ {
		if i >= len(lines) {
			out[i] = strings.Repeat(" ", width)
			continue
		}
		r := []rune(stripANSI(lines[i]))
		if len(r) > width {
			r = r[:width]
		}
		if len(r) < width {
			r = append(r, []rune(strings.Repeat(" ", width-len(r)))...)
		}
		out[i] = string(r)
	}
	return out
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

func centerStyled(s string, width int) string {
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, s)
}

func padRunes(in []rune, width int) []rune {
	if len(in) >= width {
		return in
	}
	out := make([]rune, width)
	copy(out, in)
	for i := len(in); i < width; i++ {
		out[i] = ' '
	}
	return out
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
