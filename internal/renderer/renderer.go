package renderer

import "regexp"

var leadingHeadingPattern = regexp.MustCompile(`\A(?s)#{1,6}[ \t]+([^\n]+)\n+`)

type Renderer struct {
	opts   Options
	styles styles
}

func New(opts Options) Renderer {
	if opts.Width == 0 {
		opts.Width = defaultWidth
	}
	return Renderer{
		opts:   opts,
		styles: newStyles(),
	}
}
