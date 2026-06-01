package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	xansi "github.com/charmbracelet/x/ansi"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const screenshotsDir = "/home/sbarker/swarmops/screenshots"

// charWidth/charHeight match basicfont.Face7x13 glyph cell dimensions.
const (
	charWidth  = 7
	charHeight = 13
	linePad    = 2 // extra vertical pixels between lines
)

type screenshotMsg struct {
	path string
	err  error
}

func screenshotCmd(viewContent string) tea.Cmd {
	return func() tea.Msg {
		path, err := takeScreenshot(viewContent)
		return screenshotMsg{path: path, err: err}
	}
}

func takeScreenshot(viewContent string) (string, error) {
	if err := os.MkdirAll(screenshotsDir, 0755); err != nil {
		return "", fmt.Errorf("create screenshots dir: %w", err)
	}

	plain := xansi.Strip(viewContent)
	lines := strings.Split(plain, "\n")

	// Determine canvas size.
	maxCols := 0
	for _, line := range lines {
		n := utf8.RuneCountInString(line)
		if n > maxCols {
			maxCols = n
		}
	}
	if maxCols == 0 {
		maxCols = 80
	}

	lineH := charHeight + linePad
	width := maxCols * charWidth
	height := len(lines) * lineH
	if height == 0 {
		height = lineH
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	bg := color.RGBA{0x18, 0x18, 0x18, 0xff}
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

	fg := image.NewUniform(color.RGBA{0xcc, 0xcc, 0xcc, 0xff})
	d := &font.Drawer{
		Dst:  img,
		Src:  fg,
		Face: basicfont.Face7x13,
	}

	for i, line := range lines {
		// baseline is charHeight-2 pixels from top of cell (Face7x13 convention)
		d.Dot = fixed.Point26_6{
			X: fixed.I(0),
			Y: fixed.I(i*lineH + charHeight - 2),
		}
		d.DrawString(line)
	}

	ts := time.Now().Format("20060102-150405")
	path := filepath.Join(screenshotsDir, fmt.Sprintf("screenshot-%s.png", ts))

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return "", fmt.Errorf("encode PNG: %w", err)
	}

	return path, nil
}
