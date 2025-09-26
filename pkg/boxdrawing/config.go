package boxdrawing

type Config struct {
	columns       uint
	minDots       uint
	horizPadding  uint
	colored       bool
	horizSymbol   string
	vertSymbol    string
	angleTLSymbol string
	angleTRSymbol string
	angleDLSymbol string
	angleDRSymbol string
	dotSymbol     string
	splitSymbol   string
	spaceSymbol   string
	titleColor    string
	otherColor    string
}

func (c Config) WithCustomSymbols(horiz, vert, angleTL, angleTR, angleDL, angleDR, dot, split, space string) Config {
	c.horizSymbol = horiz
	c.vertSymbol = vert
	c.angleTLSymbol = angleTL
	c.angleTRSymbol = angleTR
	c.angleDLSymbol = angleDL
	c.angleDRSymbol = angleDR
	c.dotSymbol = dot
	c.splitSymbol = split
	c.spaceSymbol = space
	return c
}

func (c Config) WithTitleColor(color string) Config {
	c.titleColor = color
	return c
}

func (c Config) WithOtherColor(color string) Config {
	c.otherColor = color
	return c
}

func NewConfig(columns, horizPadding, minDots uint, colored bool) Config {
	return Config{
		minDots:       minDots,
		horizPadding:  horizPadding,
		columns:       columns,
		colored:       colored,
		horizSymbol:   "─",
		vertSymbol:    "│",
		angleTLSymbol: "╭",
		angleTRSymbol: "╮",
		angleDLSymbol: "╰",
		angleDRSymbol: "╯",
		dotSymbol:     ".",
		splitSymbol:   "┊",
		spaceSymbol:   " ",
		titleColor:    ColorWhite,
		otherColor:    ColorGray,
	}
}
