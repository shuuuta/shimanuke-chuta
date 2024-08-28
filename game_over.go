package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func (g *Game) drawGameOver(screen *ebiten.Image) {
	title := g.location
	afterTitle := "到達"
	dist := fmt.Sprintf("%.1fkm", float64(g.travelDistance)/1000)

	if title == "八丈島" {
		title = "GAME OVER"
		afterTitle = ""
	}

	op := &text.DrawOptions{}
	op.GeoM.Translate(screenWidth/2, titleFontSize*3)
	op.ColorScale.ScaleWithColor(color.White)
	op.LineSpacing = titleFontSize
	op.PrimaryAlign = text.AlignCenter
	text.Draw(
		screen,
		title,
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
		afterTitle,
		&text.GoTextFace{
			Source: misakiFont,
			Size:   fontSize,
		},
		op,
	)

	op = &text.DrawOptions{}
	op.GeoM.Translate(screenWidth/2, titleFontSize*4+fontSize*3)
	op.ColorScale.ScaleWithColor(color.White)
	op.LineSpacing = fontSize
	op.PrimaryAlign = text.AlignCenter
	text.Draw(
		screen,
		dist,
		&text.GoTextFace{
			Source: misakiFont,
			Size:   fontSize,
		},
		op,
	)
}
