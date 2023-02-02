package ebxoxo

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

func inButton(img *ebiten.Image, x, y int) bool {
	return img.RGBA64At(x, y).A != 0
}

type Button struct {
	label       string
	x           int
	y           int
	w           int
	h           int
	color       color.Color
	colorActive color.Color
	img         *ebiten.Image
	imgOpts     *ebiten.DrawImageOptions
	active      *ebiten.Image
	activeOpts  *ebiten.DrawImageOptions
}

func NewButton(label string, x, y, w, h int, color, colorActive color.Color, img, active *ebiten.Image) *Button {
	return &Button{
		label:       label,
		x:           x,
		y:           y,
		w:           w,
		h:           h,
		color:       color,
		colorActive: colorActive,
		img:         img,
		active:      active,
	}
}

func (b *Button) Draw(screen *ebiten.Image, x, y, tick int) {
	screen.DrawImage(b.img, b.imgOpts)
}

type Board struct{}

func (b *Board) Draw(screen *ebiten.Image, tick int) {
}

func (b *Board) ClickHandler(x, y int) {
}
