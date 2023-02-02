package assets

import (
	"bytes"
	"embed"
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

func Init(width, height int) error {
	for _, f := range []func(int, int) error{
		initImages,
		initFonts,
		initOther,
	} {
		if err := f(width, height); err != nil {
			return err
		}
	}
	return nil
}

func initImages(width, height int) error {
	for _, v := range []struct {
		name string
		img  **ebiten.Image
	}{
		{"avatar-circle.png", &AvatarCircle},
		{"avatar-cross.png", &AvatarCross},
		{"avatar-placeholder.png", &AvatarPlaceholder},
		{"bg.png", &Bg},
		{"board.png", &Board},
		{"btn.png", &Btn},
		{"btn-active.png", &BtnActive},
		{"logo-a.png", &LogoA},
		{"logo-c.png", &LogoC},
		{"logo-e.png", &LogoE},
		{"logo-i.png", &LogoI},
		{"logo-o.png", &LogoO},
		{"logo-t.png", &LogoT},
		{"vs.png", &Vs},
		{"circle.png", &Circle},
		{"cross.png", &Cross},
	} {
		buf, err := files.ReadFile(v.name)
		if err != nil {
			return err
		}
		img, _, err := image.Decode(bytes.NewReader(buf))
		if err != nil {
			return err
		}
		*v.img = ebiten.NewImageFromImage(img)
	}
	return nil
}

func initFonts(width, height int) error {
	buf, err := files.ReadFile("DINRoundPro-Bold.otf")
	if err != nil {
		return err
	}
	ff, err := opentype.Parse(buf)
	if err != nil {
		return err
	}
	for _, v := range []struct {
		size float64
		face *font.Face
	}{
		{8, &Din8},
		{16, &Din16},
		{24, &Din24},
		{48, &Din48},
		{72, &Din72},
	} {
		var err error
		if *v.face, err = opentype.NewFace(ff, &opentype.FaceOptions{
			Size:    v.size,
			DPI:     72,
			Hinting: font.HintingFull,
		}); err != nil {
			return err
		}
	}
	return nil
}

func initOther(width, height int) error {
	BgOpts = new(ebiten.DrawImageOptions)
	b := Bg.Bounds()
	BgOpts.GeoM.Scale(float64(width)/float64(b.Dx()), float64(height)/float64(b.Dy()))
	return nil
}

var (
	Din8              font.Face
	Din16             font.Face
	Din24             font.Face
	Din48             font.Face
	Din72             font.Face
	AvatarCircle      *ebiten.Image
	AvatarCross       *ebiten.Image
	AvatarPlaceholder *ebiten.Image
	Bg                *ebiten.Image
	Board             *ebiten.Image
	Btn               *ebiten.Image
	BtnActive         *ebiten.Image
	LogoA             *ebiten.Image
	LogoC             *ebiten.Image
	LogoE             *ebiten.Image
	LogoI             *ebiten.Image
	LogoO             *ebiten.Image
	LogoT             *ebiten.Image
	Vs                *ebiten.Image
	Circle            *ebiten.Image
	Cross             *ebiten.Image
	BgOpts            *ebiten.DrawImageOptions
)

//go:embed *.png
//go:embed *.otf
var files embed.FS
