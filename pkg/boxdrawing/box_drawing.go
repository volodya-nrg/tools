package boxdrawing

import (
	"slices"
	"strings"
	"unicode/utf8"
)

var (
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[37m"
	ColorWhite  = "\033[97m"
	colorReset  = "\033[0m"
)

type BoxDrawing struct {
	title         string
	columns       uint
	minDots       uint
	horizPadding  uint
	blocks        [][2]string
	colored       bool
	titleColor    string
	otherColor    string
	horizSymbol   string
	vertSymbol    string
	angleTLSymbol string
	angleTRSymbol string
	angleDLSymbol string
	angleDRSymbol string
	dotSymbol     string
	splitSymbol   string
	spaceSymbol   string
}

func (b *BoxDrawing) AddBlock(strLeft, strRight string) {
	b.blocks = append(b.blocks, [2]string{strLeft, strRight})
}

func (b *BoxDrawing) Draw() []string {
	contentWidth := 0
	blocks := b.createBlocks()
	blockLines := make([]string, 0, len(blocks))
	prepare := make([]string, 0, len(blocks))

	for chunk := range slices.Chunk(blocks, int(b.columns)) {
		x := strings.Join(chunk, b.spaceSymbol+b.splitSymbol+b.spaceSymbol)
		blockLines = append(blockLines, x)
	}

	titleAndBlocks := make([]string, 0, len(blockLines)+1)
	for i, v := range append([]string{b.title}, blockLines...) {
		vLoc := v
		titleAndBlocks = append(titleAndBlocks, vLoc)
		vLen := utf8.RuneCountInString(vLoc)

		if i == 0 && b.colored {
			vLen -= utf8.RuneCountInString(b.titleColor)
		}
		if vLen > contentWidth {
			contentWidth = vLen
		}
	}

	amountHorizSymbols := contentWidth + (int(b.horizPadding) * utf8.RuneCountInString(b.spaceSymbol) * 2)

	// add first file
	prepare = append(prepare, strings.Repeat(b.horizSymbol, amountHorizSymbols))
	// add title and blocks
	for i, v := range titleAndBlocks {
		prepare = append(prepare, b.addSpaces(contentWidth, v, i == 0))
	}
	// add last line
	prepare = append(prepare, strings.Repeat(b.horizSymbol, amountHorizSymbols))

	return b.wrapBox(prepare)
}

func (b *BoxDrawing) createBlocks() []string {
	// found max content
	var maxSymbols int
	for _, arBlock := range b.blocks {
		lenFirst := utf8.RuneCountInString(arBlock[0])
		lenSecond := utf8.RuneCountInString(arBlock[1])

		if rCount := lenFirst + lenSecond; rCount > maxSymbols {
			maxSymbols = rCount
		}
	}

	// build rows
	result := make([]string, len(b.blocks))
	for k, arBlock := range b.blocks {
		lenFirst := utf8.RuneCountInString(arBlock[0])
		lenSecond := utf8.RuneCountInString(arBlock[1])
		amountNeedDots := maxSymbols - (lenFirst + lenSecond)
		firstPart := arBlock[0] + strings.Repeat(b.dotSymbol, int(b.minDots))
		secondPart := strings.Repeat(b.dotSymbol, amountNeedDots)
		thirdPart := arBlock[1]

		result[k] = firstPart + secondPart + thirdPart
	}

	return result
}

func (b *BoxDrawing) addSpaces(contentWidth int, str string, isTitle bool) string {
	lenStr := utf8.RuneCountInString(str)

	if isTitle && b.colored {
		lenStr -= utf8.RuneCountInString(b.titleColor)
	}

	totalSpace := contentWidth - lenStr
	spaces := strings.Repeat(b.spaceSymbol, totalSpace)

	if !isTitle {
		return str + spaces
	}

	lenHalfSpaces := totalSpace / 2 // example: if 23 then 11
	halfSpaces := strings.Repeat(b.spaceSymbol, lenHalfSpaces)

	if currentContentWidth := lenHalfSpaces*2 + lenStr; currentContentWidth < contentWidth { // add lost spaces
		str += strings.Repeat(b.spaceSymbol, contentWidth-currentContentWidth)
	}

	return halfSpaces + str + halfSpaces
}

func (b *BoxDrawing) wrapBox(rows []string) []string {
	var (
		colorLoc      string
		colorResetLoc string
	)

	if b.colored {
		colorLoc = b.otherColor
		colorResetLoc = colorReset
	}

	result := make([]string, len(rows))
	for i, row := range rows {
		if i == 0 {
			result[i] = colorLoc + b.angleTLSymbol + row + b.angleTRSymbol + colorResetLoc
			continue
		}
		if i == len(rows)-1 {
			result[i] = colorLoc + b.angleDLSymbol + row + b.angleDRSymbol + colorResetLoc
			continue
		}

		padding := strings.Repeat(b.spaceSymbol, int(b.horizPadding))
		result[i] = colorLoc + b.vertSymbol + padding + row + colorLoc + padding + b.vertSymbol + colorResetLoc
	}

	return result
}

func NewBoxDrawing(title string, cfg Config) *BoxDrawing {
	b := BoxDrawing{
		title:         title,
		horizSymbol:   cfg.horizSymbol,
		vertSymbol:    cfg.vertSymbol,
		angleTLSymbol: cfg.angleTLSymbol,
		angleTRSymbol: cfg.angleTRSymbol,
		angleDLSymbol: cfg.angleDLSymbol,
		angleDRSymbol: cfg.angleDRSymbol,
		dotSymbol:     cfg.dotSymbol,
		splitSymbol:   cfg.splitSymbol,
		spaceSymbol:   cfg.spaceSymbol,
		minDots:       cfg.minDots,
		horizPadding:  cfg.horizPadding,
		columns:       cfg.columns,
		colored:       cfg.colored,
		titleColor:    cfg.titleColor,
		otherColor:    cfg.otherColor,
	}

	if b.colored {
		b.title = b.titleColor + b.title
	}

	return &b
}
