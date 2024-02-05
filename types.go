package imgo

import "github.com/anton2920/imgo/gr"

/* NOTE(anton2920): I defined this separately, because if I ever decide to get rid of 'gr', I will still need these types. */
type (
	Color = gr.Color
	Font  = gr.Font
	Rect  = gr.Rect
)
