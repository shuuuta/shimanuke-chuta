package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func (g *Game) drawGameOver(screen *ebiten.Image) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(screenWidth/2, titleFontSize*3)
	op.ColorScale.ScaleWithColor(color.White)
	op.LineSpacing = titleFontSize
	op.PrimaryAlign = text.AlignCenter
	text.Draw(
		screen,
		g.location,
		&text.GoTextFace{
			Source: k8x12sFont,
			Size:   titleFontSize,
		},
		op,
	)

	op = &text.DrawOptions{}
	op.GeoM.Translate(screenWidth/2, titleFontSize*4+fontSize)
	op.ColorScale.ScaleWithColor(color.White)
	op.LineSpacing = fontSize
	op.PrimaryAlign = text.AlignCenter
	text.Draw(
		screen,
		"到達",
		&text.GoTextFace{
			Source: misakiFont,
			Size:   fontSize,
		},
		op,
	)
}
