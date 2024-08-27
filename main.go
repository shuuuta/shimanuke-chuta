package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth  = 480
	screenHeight = 640

	waveAreaWidth  = screenWidth
	waveAreaHeight = screenHeight

	playerWidth  = 32
	playerHeight = 32

	tileSize = 32

	surfStartOffset = 24
)

var (
	//go:embed resources/tiles.png
	Tiles_png  []byte
	TilesImage *ebiten.Image

	//go:embed resources/player.png
	Player_png  []byte
	PlayerImage *ebiten.Image
	PlayerRow   int
)

func init() {
	timg, _, err := image.Decode(bytes.NewReader(Tiles_png))
	if err != nil {
		log.Fatal(err)
	}
	TilesImage = ebiten.NewImageFromImage(timg)

	pimg, _, err := image.Decode(bytes.NewReader(Player_png))
	if err != nil {
		log.Fatal(err)
	}
	PlayerImage = ebiten.NewImageFromImage(pimg)
	PlayerRow = 0
}

type waveType int

const (
	waveToLeft waveType = iota
	waveToRight
)

type waveArea = struct {
	Y        int
	WaveType waveType
}

type Stage struct {
	name         string
	dist         int
	speed        int
	surfGap      int
	surfInterval int
}

type Stages []Stage

type Mode int

const (
	ModeTitle Mode = iota
	ModeGame
	ModeGameOver
)

type Game struct {
	counter  int
	mode     Mode
	stages   Stages
	location string

	speed int

	// Counter
	countAfterClick int

	// Camera
	cameraX int
	cameraY int

	// The player's position
	x16  int
	y16  int
	vx16 int

	shipDir int

	//waves
	waveAreas    []*waveArea
	surfInterval int
	surfGap      int
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func (g *Game) isSelectJustPressed() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		return true
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		return true
	}
	return false
}

func (g *Game) isRightJustPressed() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		return true
	}
	x, _ := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && x >= screenWidth/2 {
		return true
	}
	return false
}

func (g *Game) isLeftJustPressed() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		return true
	}
	x, _ := ebiten.CursorPosition()
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) && x < screenWidth/2 {
		return true
	}
	return false
}

func (g *Game) init() {
	g.counter = 0
	g.x16 = (screenWidth/2 - playerWidth/2) * 16
	g.y16 = (screenHeight - playerHeight - 96) * 16
	g.cameraX = 0
	g.cameraY = 0

	//init Stage
	g.stages = Stages{
		Stage{
			name:         "Hachijojima",
			dist:         0,
			speed:        4,
			surfGap:      8,
			surfInterval: 7,
		}, Stage{
			name:         "Mikurajima",
			dist:         83,
			speed:        6,
			surfGap:      6,
			surfInterval: 6,
		}, Stage{
			name:         "Miyakejima",
			dist:         106,
			speed:        9,
			surfGap:      4,
			surfInterval: 5,
		}, Stage{
			name:         "Kouzushima",
			dist:         133,
			speed:        11,
			surfGap:      2,
			surfInterval: 3,
		},
	}
	g.setStage()

	//init waves
	for i := 0; i < 2; i++ {
		t := waveToLeft
		if i%2 == 0 {
			t = waveToRight
		}
		g.waveAreas = append(g.waveAreas, &waveArea{
			Y:        -waveAreaHeight * i,
			WaveType: t,
		})
	}
	//init surfs
	for i := 0; i < (screenHeight*3/tileSize-surfStartOffset)/(g.surfInterval+1); i++ {
		surfs = append(surfs, &surf{
			Y:         screenHeight - tileSize - (surfStartOffset*tileSize + i*(g.surfInterval+1)*tileSize),
			LeftWidth: genSurfLeftWidth(g.surfGap),
			Gap:       g.surfGap,
		})
	}
}

func NewGame() ebiten.Game {
	g := &Game{}
	g.init()
	return g
}

func (g *Game) Update() error {
	g.counter++
	switch g.mode {
	case ModeTitle:
		if g.isSelectJustPressed() {
			g.mode = ModeGame
		}
	case ModeGame:
		g.setStage()

		g.countAfterClick += 1
		g.cameraY += g.speed
		g.y16 += g.speed * 16

		if g.isRightJustPressed() {
			g.shipDir = 1
			g.countAfterClick = 0
			g.vx16 = 96
		}
		if g.isLeftJustPressed() {
			g.shipDir = 2
			g.countAfterClick = 0
			g.vx16 = -96
		}

		g.x16 += g.vx16
		//Check is player moves off screen
		if g.x16 < 0 {
			g.x16 = 0
		}
		if g.x16 > (screenWidth-playerWidth)*16 {
			g.x16 = (screenWidth - playerWidth) * 16
		}

		g.vx16 += g.getWaveDirection()

		if g.vx16 > 96 {
			g.vx16 = 96
		}
		if g.vx16 < -96 {
			g.vx16 = -96
		}

		//Add wave
		if g.cameraY%screenHeight < g.speed {
			t := waveToLeft
			if rand.IntN(2)%2 == 0 {
				t = waveToRight
			}
			g.waveAreas = append(g.waveAreas, &waveArea{
				Y:        g.waveAreas[len(g.waveAreas)-1].Y - screenHeight,
				WaveType: t,
			})
			g.waveAreas = g.waveAreas[1:]
		}

		//Add surfs
		if g.cameraY%((g.surfInterval+1)*tileSize) < g.speed {
			lastY := surfs[len(surfs)-1].Y
			surfs = append(surfs, &surf{
				Y:         lastY - (g.surfInterval+1)*tileSize,
				LeftWidth: genSurfLeftWidth(g.surfGap),
				Gap:       g.surfGap,
			})

			rmCount := 0
			for _, s := range surfs {
				if s.Y+g.cameraY > screenHeight {
					rmCount++
				}
			}
			surfs = surfs[rmCount:]
		}

		if g.hit() {
			//g.mode = ModeGameOver
		}

	case ModeGameOver:
		if g.isSelectJustPressed() {
			g.init()
			g.mode = ModeTitle
		}
	}
	return nil
}

func (g *Game) setStage() {
	var s Stage
	for _, v := range g.stages {
		if v.dist*10 <= g.cameraY {
			s = v
		}
	}
	g.speed = s.speed
	//g.speed = 5
	g.surfInterval = s.surfInterval
	g.surfGap = s.surfGap
	g.location = s.name
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawWaves(screen)

	g.drawSurfs(screen)
	if g.mode == ModeGame {
		g.drawPlayer(screen)
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("location: %v", g.location))

	sampleLog(screen,
		fmt.Sprintf(
			"Hit: %v, "+
				"Y:%v, vx: %v\n"+
				"waves: %v, surfs: %v",
			g.hit(),
			g.cameraY,
			g.vx16,
			len(g.waveAreas),
			len(surfs),
		),
	)
}

func (g *Game) hit() bool {
	if g.mode != ModeGame {
		return false
	}

	x0 := int(math.Floor(float64(g.x16 / 16)))
	x1 := x0 + playerWidth
	y0 := int(math.Floor(float64(g.y16/16))) - g.cameraY
	y1 := y0 + playerHeight

	//out of screen
	if x0 <= 0 {
		return true
	}
	if x1 >= screenWidth {
		return true
	}

	//hit surf
	for _, s := range surfs {
		sy0 := s.Y + g.cameraY
		sy1 := sy0 + tileSize

		rx0 := 0
		rx1 := s.LeftWidth * tileSize
		lx0 := rx1 + g.surfGap*tileSize
		lx1 := screenWidth

		if y0 < sy1 && sy0 < y1 {
			if x0 < rx1 && rx0 < x1 {
				return true
			}
			if x0 < lx1 && lx0 < x1 {
				return true
			}
		}
	}
	return false
}

func (g *Game) drawPlayer(screen *ebiten.Image) {
	img := ebiten.NewImage(playerWidth, playerHeight)
	op := &ebiten.DrawImageOptions{}

	if g.countAfterClick < 30 {
		px := int(math.Floor(float64(g.cameraY / g.speed / 2 % 4)))
		img.DrawImage(PlayerImage.SubImage(image.Rect(tileSize*px, tileSize*g.shipDir, tileSize*(px+1), tileSize*(g.shipDir+1))).(*ebiten.Image), nil)
	} else {
		px := int(math.Floor(float64(g.cameraY / g.speed / 10 % 4)))
		img.DrawImage(PlayerImage.SubImage(image.Rect(tileSize*px, 0, tileSize*(px+1), tileSize)).(*ebiten.Image), nil)
	}

	op.GeoM.Translate(-float64(playerWidth)/2.0, -float64(playerHeight)/2.0)
	op.GeoM.Rotate(float64(g.vx16) / 96.0 * math.Pi / 6)
	op.GeoM.Translate(float64(playerWidth)/2.0, float64(playerHeight)/2.0)
	op.GeoM.Translate(float64(g.x16/16)-float64(g.cameraX), float64(g.y16/16)-float64(g.cameraY))
	op.Filter = ebiten.FilterLinear

	screen.DrawImage(img, op)
}

func (g *Game) drawWaves(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op2 := &ebiten.DrawImageOptions{}
	for _, w := range g.waveAreas {
		op.GeoM.Reset()
		op.GeoM.Translate(0, float64(w.Y+g.cameraY))
		nw := ebiten.NewImage(waveAreaWidth, waveAreaHeight)
		for i := 0; i < waveAreaWidth/tileSize; i++ {
			for j := 0; j < waveAreaHeight/tileSize; j++ {
				op2.GeoM.Reset()
				op2.GeoM.Translate(float64(i*tileSize), float64(j*tileSize))
				switch w.WaveType {
				case waveToLeft:
					if g.counter%120 < 60 {
						nw.DrawImage(TilesImage.SubImage(image.Rect(0, 0, tileSize, tileSize)).(*ebiten.Image), op2)
					} else {
						nw.DrawImage(TilesImage.SubImage(image.Rect(0, tileSize, tileSize, tileSize*2)).(*ebiten.Image), op2)
					}
				case waveToRight:
					if g.counter%120 < 60 {
						nw.DrawImage(TilesImage.SubImage(image.Rect(tileSize, 0, tileSize*2, tileSize)).(*ebiten.Image), op2)
					} else {
						nw.DrawImage(TilesImage.SubImage(image.Rect(tileSize, tileSize, tileSize*2, tileSize*2)).(*ebiten.Image), op2)
					}
				}
			}
		}
		screen.DrawImage(nw, op)
	}
}

func (g *Game) getWaveDirection() int {
	y := g.y16/16 - g.cameraY
	for _, w := range g.waveAreas {
		wy0 := w.Y + g.cameraY
		wy1 := w.Y + g.cameraY + waveAreaHeight
		if y > wy0 && y <= wy1 {
			switch w.WaveType {
			case waveToLeft:
				return -4
			case waveToRight:
				return 4
			default:
				return 0
			}
		}
	}
	return 0
}

type surf = struct {
	Y         int
	LeftWidth int
	Gap       int
}

var (
	surfs []*surf
)

func (g *Game) drawSurfs(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op2 := &ebiten.DrawImageOptions{}
	for _, s := range surfs {
		y := float64(s.Y + g.cameraY)

		op.GeoM.Reset()
		op.GeoM.Translate(0, y)
		sl := ebiten.NewImage(s.LeftWidth*tileSize, tileSize)

		for i := 0; i < s.LeftWidth; i++ {
			op2.GeoM.Reset()
			op2.GeoM.Translate(float64(i*tileSize), 0)
			if g.counter%60 < 30 {
				sl.DrawImage(TilesImage.SubImage(image.Rect(tileSize*2, 0, tileSize*3, tileSize)).(*ebiten.Image), op2)
			} else {
				sl.DrawImage(TilesImage.SubImage(image.Rect(tileSize*2, tileSize, tileSize*3, tileSize*2)).(*ebiten.Image), op2)
			}
		}
		screen.DrawImage(sl, op)

		op.GeoM.Reset()
		op.GeoM.Translate(float64(s.LeftWidth*tileSize+s.Gap*tileSize), y)
		sr := ebiten.NewImage(screenWidth-s.LeftWidth*tileSize+s.Gap*tileSize, tileSize)

		for i := 0; i < screenWidth/tileSize-s.LeftWidth-s.Gap; i++ {
			op2.GeoM.Reset()
			op2.GeoM.Translate(float64(i*tileSize), 0)
			if g.counter%60 < 30 {
				sr.DrawImage(TilesImage.SubImage(image.Rect(tileSize*2, 0, tileSize*3, tileSize)).(*ebiten.Image), op2)
			} else {
				sr.DrawImage(TilesImage.SubImage(image.Rect(tileSize*2, tileSize, tileSize*3, tileSize*2)).(*ebiten.Image), op2)
			}
		}
		screen.DrawImage(sr, op)
	}
}

func genSurfLeftWidth(surfGap int) int {
	maxLeftWidth := screenWidth/tileSize - surfGap - 1
	return rand.IntN(maxLeftWidth) + 1
}

func sampleLog(screen *ebiten.Image, message string) {
	const (
		mWidth  = 128 * 2
		row     = 2
		mHeight = 16 * row
		marginB = 8 * row
		mRight  = 8
	)

	m := ebiten.NewImage(mWidth, mHeight)
	m.Fill(color.RGBA{90, 90, 90, 90})
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenWidth-mWidth-mRight, screenHeight-mHeight-marginB)
	ebitenutil.DebugPrint(m, message)
	screen.DrawImage(m, op)
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Hello, world!")
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
