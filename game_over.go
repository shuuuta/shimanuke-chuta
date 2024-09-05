package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	gameOverWait = 30
)

func (g *Game) updateGameOver() {
	if g.counter > gameOverWait && g.isSelectJustPressed() {
		g.counter = 0
		g.init()
		g.mode = ModeStartMenu
	}
}

func (g *Game) drawGameOver(screen *ebiten.Image) {
	title := g.location
	afterTitle := "到達"
	dist := fmt.Sprintf("%.1fkm", float64(getTravelDistance(g.y16))/1000)

	if title == g.stages[0].name {
		title = "島抜け失敗"
		afterTitle = ""
	}

	if title == g.stages[len(g.stages)-1].name {
		title = "島抜け成功!!"
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

	if g.counter > gameOverWait && g.counter%120 < 90 {
		op = &text.DrawOptions{}
		op.GeoM.Translate(screenWidth/2, titleFontSize*4+fontSize*6)
		op.ColorScale.ScaleWithColor(color.White)
		op.LineSpacing = fontSize
		op.PrimaryAlign = text.AlignCenter
		text.Draw(
			screen,
			"Tap to start menu",
			&text.GoTextFace{
				Source: misakiFont,
				Size:   fontSize,
			},
			op,
		)
	}
}
