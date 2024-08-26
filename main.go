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
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	speed = 2
	//speed = 5

	screenWidth  = 480
	screenHeight = 640

	waveAreaWidth  = screenWidth
	waveAreaHeight = screenHeight

	playerWidth  = 32
	playerHeight = 32

	tileSize = 32

	surfOffset = 12
)

var (
	surfInterval int
	surfGap      int

	//go:embed resources/tiles.png
	Tiles_png  []byte
	TilesImage *ebiten.Image
)

func init() {
	surfInterval = 7
	surfGap = 8

	img, _, err := image.Decode(bytes.NewReader(Tiles_png))
	if err != nil {
		log.Fatal(err)
	}
	TilesImage = ebiten.NewImageFromImage(img)
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

var (
	playerImage *ebiten.Image
	tilesImage  *ebiten.Image
	waveImage   *ebiten.Image
	waveAreas   []*waveArea
)

func init() {
	playerImage = ebiten.NewImage(playerWidth, playerHeight)
	playerImage.Fill(color.RGBA{0xa0, 0xc0, 0x80, 0xff})
	vector.DrawFilledCircle(playerImage, 32, 32, 16, color.RGBA{0xa0, 0xc0, 0x80, 0xff}, true)

	tilesImage = ebiten.NewImage(tileSize, tileSize)
	tilesImage.Fill(color.RGBA{0xa0, 0x80, 0xc0, 0xff})
	vector.DrawFilledCircle(tilesImage, 0, 0, 16, color.RGBA{0xa0, 0xc0, 0x80, 0xff}, true)

	initWaveImage()

	//init waves
	for i := 0; i < 2; i++ {
		t := waveToLeft
		if i%2 == 0 {
			t = waveToRight
		}
		waveAreas = append(waveAreas, &waveArea{
			Y:        -waveAreaHeight * i,
			WaveType: t,
		})
	}

	//init surfs
	for i := 0; i < (screenHeight*3/tileSize-surfOffset)/(surfInterval+1); i++ {
		surfs = append(surfs, &surf{
			Y:         screenHeight - tileSize - (surfOffset*tileSize + i*(surfInterval+1)*tileSize),
			LeftWidth: genSurfLeftWidth(),
		})
	}
}

func initWaveImage() {
	waveImage = ebiten.NewImage(tileSize, tileSize)
	vector.DrawFilledCircle(waveImage, tileSize, tileSize, tileSize, color.RGBA{26, 106, 204, 255}, true)
	vector.DrawFilledCircle(waveImage, tileSize, tileSize, tileSize*2/3, color.RGBA{51, 131, 229, 255}, true)
	vector.DrawFilledCircle(waveImage, tileSize, tileSize, tileSize/3, color.RGBA{97, 159, 235, 255}, true)
}

type Mode int

const (
	ModeTitle Mode = iota
	ModeGame
	ModeGameOver
)

type Game struct {
	mode Mode

	// Camera
	cameraX int
	cameraY int

	// The player's position
	x16  int
	y16  int
	vx16 int
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
	g.x16 = (screenWidth/2 - playerWidth/2) * 16
	g.y16 = (screenHeight - playerHeight - 64) * 16
	g.cameraX = 0
	g.cameraY = 0
}

func NewGame() ebiten.Game {
	g := &Game{}
	g.init()
	return g
}

func (g *Game) Update() error {
	switch g.mode {
	case ModeTitle:
		if g.isSelectJustPressed() {
			g.mode = ModeGame
		}
	case ModeGame:
		g.cameraY += speed
		g.y16 += speed * 16

		if g.isRightJustPressed() {
			g.vx16 = 96
		}
		if g.isLeftJustPressed() {
			g.vx16 = -96
		}

		g.x16 += g.vx16
		//Player moves off screen
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

		if g.cameraY%screenHeight == 0 {
			//Add wave
			t := waveToLeft
			if rand.IntN(2)%2 == 0 {
				t = waveToRight
			}
			waveAreas = append(waveAreas, &waveArea{
				Y:        -screenHeight - g.cameraY,
				WaveType: t,
			})
			waveAreas = waveAreas[1:]
		}

		if g.cameraY%((surfInterval+1)*tileSize) < speed {
			//Add surfs
			lastY := surfs[len(surfs)-1].Y
			surfs = append(surfs, &surf{
				Y:         lastY - (surfInterval+1)*tileSize,
				LeftWidth: genSurfLeftWidth(),
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

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawWaves(screen)

	g.drawSurfs(screen)
	if g.mode == ModeGame {
		g.drawPlayer(screen)
	}

	sampleLog(screen,
		fmt.Sprintf(
			"Hit: %v, "+
				"Y:%v, vx: %v\n"+
				"waves: %v, surfs: %v",
			g.hit(),
			g.cameraY,
			g.vx16,
			len(waveAreas),
			len(surfs),
		),
	)
}

func (g *Game) hit() bool {
	if g.mode != ModeGame {
		return false
	}

	x0 := int(math.Floor(float64(g.x16 / 16)))
	x1 := x0 + playerImage.Bounds().Dx()
	y0 := int(math.Floor(float64(g.y16/16))) - g.cameraY
	y1 := y0 + playerImage.Bounds().Dy()

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
		lx0 := rx1 + surfGap*tileSize
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
	op := &ebiten.DrawImageOptions{}
	w, h := playerImage.Bounds().Dx(), playerImage.Bounds().Dy()
	op.GeoM.Translate(-float64(w)/2.0, -float64(h)/2.0)
	op.GeoM.Rotate(float64(g.vx16) / 96.0 * math.Pi / 6)
	op.GeoM.Translate(float64(w)/2.0, float64(h)/2.0)
	op.GeoM.Translate(float64(g.x16/16)-float64(g.cameraX), float64(g.y16/16)-float64(g.cameraY))
	op.Filter = ebiten.FilterLinear

	screen.DrawImage(playerImage, op)
}

func (g *Game) drawTiles(screen *ebiten.Image) {
	const (
		nx           = screenWidth / tileSize
		ny           = screenHeight / tileSize
		pipeTileSrcX = 128
		pipeTileSrcY = 192
	)

	op := &ebiten.DrawImageOptions{}

	for i := -2; i < nx+1; i++ {
		// ground
		op.GeoM.Reset()
		//op.GeoM.Translate(float64(i*tileSize-floorMod(g.cameraX, tileSize)),
		//	float64((ny-1)*tileSize-floorMod(g.cameraY, tileSize)))
		op.GeoM.Translate(float64(i*tileSize),
			float64((ny-1)*tileSize))

		screen.DrawImage(tilesImage, op)
	}
}

func (g *Game) drawWaves(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op2 := &ebiten.DrawImageOptions{}
	for _, w := range waveAreas {
		op.GeoM.Reset()
		op.GeoM.Translate(0, float64(w.Y+g.cameraY))
		nw := ebiten.NewImage(waveAreaWidth, waveAreaHeight)
		for i := 0; i < waveAreaWidth/tileSize; i++ {
			for j := 0; j < waveAreaHeight/tileSize; j++ {
				op2.GeoM.Reset()
				op2.GeoM.Translate(float64(i*tileSize), float64(j*tileSize))
				switch w.WaveType {
				case waveToLeft:
					//nw.Fill(color.RGBA{142, 186, 241, 255})
					if g.cameraY/speed%120 < 60 {
						nw.DrawImage(TilesImage.SubImage(image.Rect(0, 0, tileSize, tileSize)).(*ebiten.Image), op2)
					} else {
						nw.DrawImage(TilesImage.SubImage(image.Rect(0, tileSize, tileSize, tileSize*2)).(*ebiten.Image), op2)
					}
				case waveToRight:
					//nw.Fill(color.RGBA{187, 214, 246, 255})
					if g.cameraY/speed%120 < 60 {
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
	for _, w := range waveAreas {
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
		//sl.Fill(color.RGBA{232, 241, 252, 255})
		for i := 0; i < s.LeftWidth; i++ {
			op2.GeoM.Reset()
			op2.GeoM.Translate(float64(i*tileSize), 0)
			if g.cameraY/speed%60 < 30 {
				sl.DrawImage(TilesImage.SubImage(image.Rect(tileSize*2, 0, tileSize*3, tileSize)).(*ebiten.Image), op2)
			} else {
				sl.DrawImage(TilesImage.SubImage(image.Rect(tileSize*2, tileSize, tileSize*3, tileSize*2)).(*ebiten.Image), op2)
			}
		}
		screen.DrawImage(sl, op)

		op.GeoM.Reset()
		op.GeoM.Translate(float64(s.LeftWidth*tileSize+surfGap*tileSize), y)
		sr := ebiten.NewImage(screenWidth-s.LeftWidth*tileSize+surfGap*tileSize, tileSize)
		//sr.Fill(color.RGBA{232, 241, 252, 255})
		for i := 0; i < screenWidth/tileSize-s.LeftWidth-surfGap; i++ {
			op2.GeoM.Reset()
			op2.GeoM.Translate(float64(i*tileSize), 0)
			if g.cameraY/speed%60 < 30 {
				sr.DrawImage(TilesImage.SubImage(image.Rect(tileSize*2, 0, tileSize*3, tileSize)).(*ebiten.Image), op2)
			} else {
				sr.DrawImage(TilesImage.SubImage(image.Rect(tileSize*2, tileSize, tileSize*3, tileSize*2)).(*ebiten.Image), op2)
			}
		}
		screen.DrawImage(sr, op)
	}
}

func genSurfLeftWidth() int {
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
