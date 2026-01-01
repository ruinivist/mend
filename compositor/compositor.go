/*
I asked for very little ( a dialog box ) from life ( bubbletea ), and even this little was denied me.
I guess asking for an functional overlay was too much in 2025 so here is my own compositor layer.
Idea is simple really I just draw the old ui and then on top draw the new ui, only overwriting what's needed
The main problem is with handling of characters and styles which is "decent" enough here.
- ruinivist, 1Jan26
*/

package compositor

import (
	"regexp"
	"strings"

	"github.com/mattn/go-runewidth"
)

/*
a colored character is like some color followed by one or more "runes" aka characters
since I'm overwriting arbitrary parts of screen, to keep color intact and not have it "split up"
every rune has it's own style now
*/
type Pixel struct {
	Char  rune
	Style string
}

type Grid struct {
	Width  int
	Height int
	Rows   [][]Pixel
}

func NewGrid(w, h int) *Grid {
	rows := make([][]Pixel, h)
	for i := range rows {
		rows[i] = make([]Pixel, w)
		for j := range rows[i] {
			rows[i][j] = Pixel{Char: ' ', Style: ""}
		}
	}
	return &Grid{
		Width:  w,
		Height: h,
		Rows:   rows,
	}
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`) // regex to match ANSI SGR sequences
// what even is ANSI SGR aka escape codes for colors and styles for terminals

// Write content string starting at (startX, startY)
func (g *Grid) Write(startX, startY int, content string) {
	x, y := startX, startY
	currentStyle := ""

	runes := []rune(content) // note that this is now a one dimensional slice
	i := 0
	for i < len(runes) {
		if y >= g.Height {
			break
		}

		// Check for ANSI escape sequence and apply them to the end
		if runes[i] == '\x1b' {
			// Find the end of the sequence 'm'
			end := -1
			for j := i; j < len(runes); j++ {
				if runes[j] == 'm' {
					end = j
					break
				}
			}

			if end != -1 {
				seq := string(runes[i : end+1])
				// this should be a valid ansi sgr seq now
				if ansiRegex.MatchString(seq) {
					currentStyle = seq
					i = end + 1
					continue
				}
			}
		}

		// line splits
		if runes[i] == '\n' {
			y++
			x = startX
			i++
			continue
		}

		// Handle regular character
		r := runes[i]
		rw := runewidth.RuneWidth(r)

		if x+rw <= g.Width {
			g.Rows[y][x] = Pixel{Char: r, Style: currentStyle}

			// null pad wide chars that span across multiple cells
			for k := 1; k < rw; k++ {
				if x+k < g.Width {
					g.Rows[y][x+k] = Pixel{Char: 0, Style: currentStyle}
				}
			}
		}

		x += rw
		i++
	}
}

// Render gives me back the string representation for terminal
func (g *Grid) Render() string {
	var b strings.Builder
	currentActiveStyle := ""

	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			p := g.Rows[y][x]

			// wide chars, these must have been handled already
			if p.Char == 0 {
				continue
			}

			// Optimizing state changes
			if p.Style != currentActiveStyle {
				// when a style changes, the simplest way is to print a "reset" first and
				// then the new style
				b.WriteString("\x1b[0m") // this means "reset" style of following chars

				if p.Style != "" {
					b.WriteString(p.Style)
				}
				currentActiveStyle = p.Style
			}

			b.WriteRune(p.Char)
		}

		b.WriteString("\x1b[0m") // end of line reset for wrap artifacts on \n next
		currentActiveStyle = ""

		if y < g.Height-1 {
			b.WriteRune('\n')
		}
	}

	return b.String()
}
