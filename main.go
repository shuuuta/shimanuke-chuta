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
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

func getTravelDistance(y16 int) int {
	return y16 / 16 * 20
}
func pxToTravelDistance(y int) int {
	return y * 20
}

var (
	muteki = false
	dev    = true
)

const (
	screenWidth  = 480.0
	screenHeight = 640.0

	waveAreaWidth  = screenWidth
	waveAreaHeight = screenHeight

	tileSize = 32

	playerWidth  = 64
	playerHeight = 64

	surfWidth  = 64
	surfHeight = 64

	surfStartOffset = 48
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

const (
	fontSize      = 24
	titleFontSize = fontSize * 3
)

var (
	//go:embed resources/misaki_gothic_2nd.ttf
	misakiFontTTF []byte
	misakiFont    *text.GoTextFaceSource

	//go:embed resources/k8x12S.ttf
	k8x12sFontTTF []byte
	k8x12sFont    *text.GoTextFaceSource
)

func init() {
	m, err := text.NewGoTextFaceSource(bytes.NewReader(misakiFontTTF))
	if err != nil {
		log.Fatal(err)
	}
	misakiFont = m

	k, err := text.NewGoTextFaceSource(bytes.NewReader(k8x12sFontTTF))
	if err != nil {
		log.Fatal(err)
	}
	k8x12sFont = k
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
	ModeStartMenu Mode = iota
	ModeGame
	ModeGameOver
)

type Game struct {
	counter  int
	mode     Mode
	stages   Stages
	location string

	speed int

	// Input
	touchIDs []ebiten.TouchID

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
	surfs        []*surf
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
	if len(g.touchIDs) > 0 {
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
	if len(g.touchIDs) > 0 {
		x, _ := ebiten.TouchPosition(g.touchIDs[0])
		if x >= screenWidth/2 {
			return true
		}
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
	if len(g.touchIDs) > 0 {
		x, _ := ebiten.TouchPosition(g.touchIDs[0])
		if x < screenWidth/2 {
			return true
		}
	}
	return false
}

func (g *Game) init() {
	g.counter = 0
	g.x16 = (screenWidth/2 - playerWidth/2) * 16
	g.y16 = (screenHeight - playerHeight - tileSize*4) * 16
	g.cameraX = 0
	g.cameraY = 0

	//init Stage
	g.stages = Stages{
		Stage{
			name:         "八丈島",
			dist:         0,
			speed:        2,
			surfGap:      9,
			surfInterval: 12,
		}, Stage{
			name:         "御蔵島",
			dist:         83,
			speed:        2,
			surfGap:      8,
			surfInterval: 12,
		}, Stage{
			name:         "三宅島",
			dist:         106,
			speed:        3,
			surfGap:      8,
			surfInterval: 12,
		}, Stage{
			name:         "神津島",
			dist:         133,
			speed:        3,
			surfGap:      8,
			surfInterval: 11,
		}, Stage{
			name:         "式根島",
			dist:         143,
			speed:        4,
			surfGap:      8,
			surfInterval: 11,
		}, Stage{
			name:         "新島",
			dist:         150,
			speed:        4,
			surfGap:      8,
			surfInterval: 10,
		}, Stage{
			name:         "利島",
			dist:         160,
			speed:        5,
			surfGap:      8,
			surfInterval: 10,
		}, Stage{
			name:         "大島",
			dist:         176,
			speed:        5,
			surfGap:      7,
			surfInterval: 10,
			//}, Stage{
			//	name:         "千葉",
			//	dist:         197,
			//	speed:        5,
			//	surfGap:      7,
			//	surfInterval: 9,
			//}, Stage{
			//	name:         "神奈川",
			//	dist:         225,
			//	speed:        5,
			//	surfGap:      7,
			//	surfInterval: 8,
		}, Stage{
			name:         "東京",
			dist:         280,
			speed:        6,
			surfGap:      7,
			surfInterval: 8,
		},
	}
	g.setStage()

	//init waves
	g.waveAreas = []*waveArea{}
	for i := 0; i < 3; i++ {
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
	g.surfs = []*surf{}

	initSurfsNum := int(math.Floor(float64(screenHeight / ((g.surfGap + 1) * surfHeight))))
	if initSurfsNum < 0 {
		initSurfsNum = 1
	}
	for i := 0; i < initSurfsNum; i++ {
		s := g.stages[0]
		y := -surfStartOffset * tileSize

		if i > 0 {
			lastY := g.surfs[len(g.surfs)-1].Y
			for _, v := range g.stages {
				if v.dist*1000 < pxToTravelDistance(-lastY) {
					s = v
				}
			}
			y = lastY - surfHeight - s.surfGap*tileSize
		}

		g.surfs = append(g.surfs, &surf{
			Y:         y,
			LeftWidth: genSurfLeftWidth(s.surfGap),
			Gap:       s.surfGap,
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
	g.touchIDs = inpututil.AppendJustPressedTouchIDs(g.touchIDs[:0])

	switch g.mode {
	case ModeStartMenu:
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
			lastY := g.surfs[len(g.surfs)-1].Y
			s := g.stages[0]
			for _, v := range g.stages {
				if v.dist*1000 < pxToTravelDistance(-lastY) {
					s = v
				}
			}
			g.surfs = append(g.surfs, &surf{
				Y:         lastY - surfHeight - s.surfInterval*tileSize,
				LeftWidth: genSurfLeftWidth(s.surfGap),
				Gap:       s.surfGap,
			})

			rmCount := 0
			for _, s := range g.surfs {
				if s.Y+g.cameraY > screenHeight {
					rmCount++
				}
			}
			g.surfs = g.surfs[rmCount:]
		}

		if g.hit() && !muteki {
			g.counter = 0
			g.mode = ModeGameOver
		}

	case ModeGameOver:
		if g.counter > 30 && g.isSelectJustPressed() {
			g.counter = 0
			g.init()
			g.mode = ModeStartMenu
		}
	}
	return nil
}

func (g *Game) setStage() {
	var s Stage
	for _, v := range g.stages {
		if v.dist*1000 <= getTravelDistance(g.y16) {
			s = v
		}
	}
	g.speed = s.speed
	g.surfInterval = s.surfInterval
	g.surfGap = s.surfGap
	g.location = s.name
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawWaves(screen)
	g.drawSurfs(screen)

	if g.mode == ModeStartMenu {
		g.drawStartMenu(screen)
	}

	if g.mode == ModeGame {
		g.drawPlayer(screen)
	}

	if g.mode == ModeGameOver {
		g.drawPlayer(screen)
		g.drawGameOver(screen)
	}

	if dev {
		sampleLog(screen,
			fmt.Sprintf(
				"Hit: %v, "+
					"Y:%v, vx: %v\n"+
					"dist: %vm, "+
					"waves: %v, surfs: %v",
				g.hit(),
				g.cameraY,
				g.vx16,
				fmt.Sprintf("%.2fkm", float64(getTravelDistance(g.y16)/1000)),
				len(g.waveAreas),
				len(g.surfs),
			),
		)
		ebitenutil.DebugPrint(screen, fmt.Sprintf("TPS: %0.1f", ebiten.ActualTPS()))
	}
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
	for _, s := range g.surfs {
		sy0 := s.Y + g.cameraY
		sy1 := sy0 + surfHeight

		rx0 := 0
		rx1 := s.LeftWidth * tileSize
		lx0 := rx1 + s.Gap*tileSize
		lx1 := screenWidth

		if y0 < sy1 && sy0 < y1 {
			if x0 < rx1 && rx0 < x1 {
				return true
			}
			if x0 < int(lx1) && lx0 < x1 {
				return true
			}
		}
	}
	return false
}

func (g *Game) drawPlayer(screen *ebiten.Image) {
	img := ebiten.NewImage(playerWidth, playerHeight)
	op := &ebiten.DrawImageOptions{}

	px0 := 0
	py0 := 0
	if g.countAfterClick < 30 {
		px0 = int(math.Floor(float64(g.cameraY/g.speed/5%4))) * playerWidth
		py0 = g.shipDir * playerHeight
	} else {
		px0 = int(math.Floor(float64(g.cameraY/g.speed/10%4))) * playerWidth
	}
	img.DrawImage(PlayerImage.SubImage(image.Rect(px0, py0, px0+playerWidth, py0+playerHeight)).(*ebiten.Image), nil)

	op.GeoM.Translate(-float64(playerWidth)/2.0, -float64(playerHeight)/2.0)
	op.GeoM.Rotate(float64(g.vx16) / 96.0 * math.Pi / 6)
	op.GeoM.Translate(float64(playerWidth)/2.0, float64(playerHeight)/2.0)
	op.GeoM.Translate(float64(g.x16/16)-float64(g.cameraX), float64(g.y16/16)-float64(g.cameraY))
	op.Filter = ebiten.FilterLinear

	screen.DrawImage(img, op)
}

func (g *Game) drawWaves(screen *ebiten.Image) {
	const waveWidth = 64.0
	const waveHeight = 64.0
	const waveX1 = 0
	const waveX2 = waveWidth

	op := &ebiten.DrawImageOptions{}
	for _, w := range g.waveAreas {
		areaY := w.Y + g.cameraY
		if areaY < -waveAreaHeight || areaY > screenHeight {
			continue
		}

		for i := 0; i < int(math.Ceil(waveAreaWidth/waveWidth)); i++ {
			for j := 0; j < int(math.Ceil(waveAreaHeight/waveHeight)); j++ {
				posX := i * waveWidth
				posY := areaY + j*waveHeight
				op.GeoM.Reset()
				op.GeoM.Translate(float64(posX), float64(posY))
				x := waveX1
				y := 0

				if w.WaveType == waveToRight {
					x = waveX2
				}

				if g.counter%180 < 60 {
					y = waveHeight
				} else if g.counter%180 < 120 {
					y = waveHeight * 2
				}
				screen.DrawImage(TilesImage.SubImage(image.Rect(x, y, x+waveWidth, y+waveHeight)).(*ebiten.Image), op)
			}
		}
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

func (g *Game) drawSurfs(screen *ebiten.Image) {
	const surfX1 = 128
	const surfX2 = 160
	const surfX3 = 224

	op := &ebiten.DrawImageOptions{}

	sy := 0
	if g.counter%40 > 20 {
		sy = surfHeight
	}

	for _, s := range g.surfs {
		y := float64(s.Y + g.cameraY)
		if y < -surfHeight || y > screenHeight {
			continue
		}

		for i := s.LeftWidth; i >= -1; i = i - 2 {
			sx := surfX2
			if i == s.LeftWidth {
				sx = surfX3
			}
			op.GeoM.Reset()
			op.GeoM.Translate(float64(i*tileSize), y)
			screen.DrawImage(TilesImage.SubImage(image.Rect(sx, sy, sx+surfWidth, sy+surfHeight)).(*ebiten.Image), op)
		}

		for i := 0; i < screenWidth/tileSize-s.LeftWidth-s.Gap; i++ {
			sx := surfX2
			if i == 0 {
				sx = surfX1
			}
			op.GeoM.Reset()
			op.GeoM.Translate(float64(i*tileSize+s.LeftWidth*tileSize+s.Gap*tileSize), y)
			screen.DrawImage(TilesImage.SubImage(image.Rect(sx, sy, sx+surfWidth, sy+surfHeight)).(*ebiten.Image), op)
		}
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
