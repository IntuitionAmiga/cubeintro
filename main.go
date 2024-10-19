package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/mix"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const (
	FPS           = 60
	numStars      = 1000
	maxTrailLen   = 5
	scrollSpeed   = 10
	fontWidth     = 32
	fontHeight    = 32
	displayWidth  = 64
	displayHeight = 64
	numBars       = 5
	barHeight     = 60
)

type Point3D struct {
	x, y, z float64
}
type Star struct {
	x, y, z, speed float64
	trail          []Point3D
}
type Edge struct {
	start, end int
}

var (
	renderer     *sdl.Renderer
	window       *sdl.Window
	err          error
	fullscreen   bool
	windowWidth  int32 = 1024
	windowHeight int32 = 768

	//Cube variables
	zoomFactor    = 0.1
	targetZoom    = 0.6
	zoomStep      = 0.002
	rotationAngle = 0.0
	cubeVertices  = []Point3D{
		{1, 1, -1}, {1, -1, -1}, {-1, -1, -1}, {-1, 1, -1},
		{1, 1, 1}, {1, -1, 1}, {-1, -1, 1}, {-1, 1, 1},
	}
	cubeEdges = []Edge{
		{0, 1}, {1, 2}, {2, 3}, {3, 0},
		{4, 5}, {5, 6}, {6, 7}, {7, 4},
		{0, 4}, {1, 5}, {2, 6}, {3, 7},
	}
	cubeFaces = [][]int{
		{0, 1, 2, 3}, // Back
		{4, 5, 6, 7}, // Front
		{0, 1, 5, 4}, // Top
		{3, 2, 6, 7}, // Bottom
		{0, 3, 7, 4}, // Left
		{1, 2, 6, 5}, // Right
	}
	faceColors = [][]uint8{
		{0, 0, 0, 0},       // Back (transparent)
		{0, 0, 0, 0},       // Front (transparent)
		{255, 0, 0, 255},   // Top (red)
		{0, 255, 0, 255},   // Bottom (green)
		{0, 0, 255, 255},   // Left (blue)
		{255, 255, 0, 255}, // Right (yellow)
	}

	// Starfield variables
	stars []Star

	// Scrolltext variables
	scrollText  = "..:INTUITION PRESENTS:..    \"I FEEL 16 AGAIN!\"    ..:PRESS THE UP AND DOWN KEYS TO ZOOM THE CUBE IN AND OUT:..    ..:\"-WIN\" ARGUMENT ON COMMANDLINE TO RUN IN WINDOWED MODE:..    ..:\"-WIN WIDTH HEIGHT\" TO SET WINDOW SIZE:..  ..:\"-DEBUG\" TO SHOW FPS:..  ..:PRESS Q OR ESC TO QUIT:..    ..:ORIGINAL COMIC BAKERY MUSIC FOR C64 BY MARTIN GALWAY IN 1984...     ..:SID TO PROTRACKER CONVERSION FOR AMIGA BY H0FFMAN (DREAMFISH OF TRSI) IN 1994:..    ..:GOLANG CODE BY INTUITION IN 2024:..    ..:FONT GRAPHICS BY UNKNOWN:..    ..:GREETS TO KARLOS AND GADGETMASTER!!!:..          "
	scrollPosX  = float64(windowWidth)
	fontTexture *sdl.Texture
	charMap     = map[rune][2]int{
		' ': {0, 0}, '!': {1, 0}, '"': {2, 0}, '@': {3, 0}, '*': {4, 0}, '£': {5, 0}, '^': {6, 0}, '\'': {7, 0}, '(': {8, 0}, ')': {9, 0},
		'&': {0, 1}, '~': {1, 1}, ',': {2, 1}, '-': {3, 1}, '.': {4, 1}, '+': {5, 1}, '0': {6, 1}, '1': {7, 1}, '2': {8, 1}, '3': {9, 1},
		'4': {0, 2}, '5': {1, 2}, '6': {2, 2}, '7': {3, 2}, '8': {4, 2}, '9': {5, 2}, ':': {6, 2}, ';': {7, 2}, '=': {8, 2}, '[': {9, 2},
		']': {0, 3}, '?': {1, 3}, '{': {2, 3}, 'A': {3, 3}, 'B': {4, 3}, 'C': {5, 3}, 'D': {6, 3}, 'E': {7, 3}, 'F': {8, 3}, 'G': {9, 3},
		'H': {0, 4}, 'I': {1, 4}, 'J': {2, 4}, 'K': {3, 4}, 'L': {4, 4}, 'M': {5, 4}, 'N': {6, 4}, 'O': {7, 4}, 'P': {8, 4}, 'Q': {9, 4},
		'R': {0, 5}, 'S': {1, 5}, 'T': {2, 5}, 'U': {3, 5}, 'V': {4, 5}, 'W': {5, 5}, 'X': {6, 5}, 'Y': {7, 5}, 'Z': {8, 5}, '`': {9, 5},
	}

	// Bouncing logo variables
	texture                 *sdl.Texture
	imageWidth, imageHeight int32
	posY, targetY           int

	// Copper bar variables
	barOffsets  = make([]int, numBars)
	barTextures []*sdl.Texture

	// Loop control
	running   = true
	startTime = time.Now()

	debug bool
)

func main() {
	parseCommandLineArgs()
	err := initSDL()
	if err != nil {
		return
	}
	defer sdl.Quit()
	defer img.Quit()
	defer mix.Quit()

	fmt.Println("Cubetro by Intuition (2024)\n")
	fmt.Println("\"-win\" argument on commandline to run in windowed mode")
	fmt.Println("\"-win width height\" to set window size (default 1024x768)")
	fmt.Println("\"-debug\" to show FPS\n")

	setupDisplay()
	defer func(window *sdl.Window) {
		err := window.Destroy()
		if err != nil {

		}
	}(window)
	defer func(renderer *sdl.Renderer) {
		err := renderer.Destroy()
		if err != nil {

		}
	}(renderer)

	// Hide the mouse cursor
	_, err = sdl.ShowCursor(sdl.DISABLE)
	if err != nil {
		return
	}

	displayKick13Image(2 * time.Second)
	playFloppySound()
	drawDecrunchEffect(2 * time.Second)
	initStars()
	playMusic()

	if err := setupFont(); err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Failed to setup font: %s\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
	defer func(fontTexture *sdl.Texture) {
		err := fontTexture.Destroy()
		if err != nil {

		}
	}(fontTexture)

	if err := setupBouncingLogo(); err != nil {
		log.Fatalf("Failed to setup bouncing logo: %s", err)
	}
	defer func(texture *sdl.Texture) {
		err := texture.Destroy()
		if err != nil {

		}
	}(texture)

	// Pre-render the copper bars into textures
	createBarTextures()

	var frameCount int
	var fps float64
	var lastTime time.Time

	// Main loop
	startTime = time.Now()
	lastTime = startTime
	for running {
		frameStart := time.Now()
		handleEvents()
		updateZoomLevel() // Zoom in/out

		err := renderer.SetDrawColor(0, 0, 0, 255)
		if err != nil {
			return
		}
		err = renderer.Clear()
		if err != nil {
			return
		}

		drawRainbowLine(50, 200, false)
		drawStarfield()

		drawCopperBars(time.Since(startTime).Seconds())
		drawCube(rotationAngle)
		rotateCube()

		drawScrollText(scrollText, scrollPosX)
		updateScrollTextPosition()

		updateBouncingLogoPosition()
		drawBouncingLogo()

		drawRainbowLine(windowHeight-50, 200, true)

		renderer.Present()

		// Calculate FPS
		frameCount++
		elapsed := time.Since(lastTime).Seconds()
		if elapsed >= 1.0 {
			fps = float64(frameCount) / elapsed
			frameCount = 0
			lastTime = time.Now()
			if debug {
				fmt.Printf("FPS: %.2f\n", fps)
				// Move cursor back up one line
				fmt.Print("\033[A")
			}
		}
		frameTime := time.Since(frameStart)
		sdl.Delay(uint32(math.Max(0, float64(1000/FPS)-float64(frameTime.Milliseconds()))))
	}
}

func loadKick13Texture(renderer *sdl.Renderer) (*sdl.Texture, error) {
	rwops, err := sdl.RWFromMem(kick13)
	if err != nil {
		return nil, fmt.Errorf("could not create RWops from bytes: %v", err)
	}
	imgSurface, err := img.LoadPNGRW(rwops)
	if err != nil {
		return nil, fmt.Errorf("could not load image from RWops: %v", err)
	}
	defer imgSurface.Free()

	texture, err := renderer.CreateTextureFromSurface(imgSurface)
	if err != nil {
		return nil, fmt.Errorf("could not create texture: %v", err)
	}

	return texture, nil
}
func displayKick13Image(duration time.Duration) {
	texture, err := loadKick13Texture(renderer)
	if err != nil {
		log.Fatalf("Failed to load kick13 image: %s", err)
	}
	defer func(texture *sdl.Texture) {
		err := texture.Destroy()
		if err != nil {

		}
	}(texture)

	startTime := time.Now()
	for time.Since(startTime) < duration {
		err := renderer.SetDrawColor(0, 0, 0, 255)
		if err != nil {
			return
		}
		err = renderer.Clear()
		if err != nil {
			return
		}

		dstRect := sdl.Rect{X: 0, Y: 0, W: windowWidth, H: windowHeight}
		err = renderer.Copy(texture, nil, &dstRect)
		if err != nil {
			return
		}
		renderer.Present()
		sdl.Delay(16) // Limit frame rate to about 60 FPS
	}
	// Turn the entire screen white
	err = renderer.SetDrawColor(255, 255, 255, 255)
	if err != nil {
		return
	}
	err = renderer.Clear()
	if err != nil {
		return
	}
	renderer.Present()
}

func loadMp3FromBytes(data []byte) (*mix.Music, error) {
	rwops, err := sdl.RWFromMem(data)
	if err != nil {
		return nil, fmt.Errorf("could not create RWops from bytes: %v", err)
	}

	music, err := mix.LoadMUSRW(rwops, 0)
	if err != nil {
		return nil, fmt.Errorf("could not load MP3 from RWops: %v", err)
	}

	return music, nil
}

func playMp3(music *mix.Music) error {
	if err := music.Play(1); err != nil {
		return fmt.Errorf("failed to play music: %v", err)
	}

	// Wait for the audio to finish playing
	for mix.PlayingMusic() != false {
		sdl.Delay(100)
	}

	return nil
}

func playFloppySound() {
	music, err := loadMp3FromBytes(floppySound)
	if err != nil {
		log.Fatalf("Failed to load floppysound MP3: %s", err)
	}

	err = playMp3(music)
	if err != nil {
		log.Fatalf("Failed to play floppysound MP3: %s", err)
	}
	music.Free()
}

func drawDecrunchEffect(duration time.Duration) {
	startTime := time.Now()
	speed := int32(20)
	barThickness := int32(10) // Adjust this value to change the thickness of the bars

	colors := [][3]uint8{
		{0, 0, 0}, {255, 255, 255}, {136, 0, 0}, {170, 255, 238},
		{204, 68, 204}, {0, 204, 85}, {0, 0, 170}, {238, 238, 119},
		{221, 136, 85}, {102, 68, 0}, {255, 119, 119}, {51, 51, 51},
		{119, 119, 119}, {170, 255, 102}, {0, 136, 255}, {187, 187, 187},
		{0, 0, 0}, {255, 255, 255}, {136, 0, 0}, {170, 255, 238},
		{204, 68, 204}, {0, 204, 85}, {0, 0, 170}, {238, 238, 119},
		{221, 136, 85}, {102, 68, 0}, {255, 119, 119}, {51, 51, 51},
		{119, 119, 119}, {170, 255, 102}, {0, 136, 255}, {187, 187, 187},
		{0, 0, 0}, {255, 255, 255}, {136, 0, 0}, {170, 255, 238},
		{204, 68, 204}, {0, 204, 85}, {0, 0, 170}, {238, 238, 119},
		{221, 136, 85}, {102, 68, 0}, {255, 119, 119}, {51, 51, 51},
		{119, 119, 119}, {170, 255, 102}, {0, 136, 255}, {187, 187, 187},
		{0, 0, 0}, {255, 255, 255}, {136, 0, 0}, {170, 255, 238},
		{204, 68, 204}, {0, 204, 85}, {0, 0, 170}, {238, 238, 119},
		{221, 136, 85}, {102, 68, 0}, {255, 119, 119}, {51, 51, 51},
		{119, 119, 119}, {170, 255, 102}, {0, 136, 255}, {187, 187, 187},
		{0, 0, 0}, {255, 255, 255}, {136, 0, 0}, {170, 255, 238},
		{204, 68, 204}, {0, 204, 85}, {0, 0, 170}, {238, 238, 119},
		{221, 136, 85}, {102, 68, 0}, {255, 119, 119}, {51, 51, 51},
		{119, 119, 119}, {170, 255, 102}, {0, 136, 255}, {187, 187, 187},
		{0, 0, 0}, {255, 255, 255}, {136, 0, 0}, {170, 255, 238},
		{204, 68, 204}, {0, 204, 85}, {0, 0, 170}, {238, 238, 119},
		{221, 136, 85}, {102, 68, 0}, {255, 119, 119}, {51, 51, 51},
		{119, 119, 119}, {170, 255, 102}, {0, 136, 255}, {187, 187, 187},
	}

	for time.Since(startTime) < duration {
		t := int32(time.Since(startTime).Milliseconds() / int64(speed))
		for y := int32(0); y < windowHeight; y += barThickness {
			colorIndex := (y + t) % int32(len(colors))
			color := colors[colorIndex]

			err := renderer.SetDrawColor(color[0], color[1], color[2], 255)
			if err != nil {
				return
			}
			for i := int32(0); i < barThickness; i++ {
				err := renderer.DrawLine(0, y+i, windowWidth, y+i)
				if err != nil {
					return
				}
			}
		}
		renderer.Present()
		sdl.Delay(16) // Limit frame rate to about 60 FPS
	}
}

func drawRainbowLine(y int32, speed int32, reverse bool) {
	colors := [][3]uint8{
		{255, 0, 0}, {255, 127, 0}, {255, 255, 0}, {127, 255, 0},
		{0, 255, 0}, {0, 255, 127}, {0, 255, 255}, {0, 127, 255},
		{0, 0, 255}, {127, 0, 255}, {255, 0, 255}, {255, 0, 127},
	}

	numColors := int32(len(colors))
	lineWidth := windowWidth / numColors

	// Calculate the current time offset for cycling the colors
	t := int32(time.Since(startTime).Milliseconds() / int64(speed))
	if reverse {
		t = -t
	}

	for i := int32(0); i < windowWidth; i++ {
		// Calculate the fractional position within the current color segment
		pos := float32(i%lineWidth) / float32(lineWidth)

		// Get the current and next color index, adjusted to cycle in the correct direction
		colorIndex := (i/lineWidth + t) % numColors
		if colorIndex < 0 {
			colorIndex += numColors
		}
		nextColorIndex := (colorIndex + 1) % numColors

		// Interpolate between the current and next color
		color := interpolateColor(colors[colorIndex], colors[nextColorIndex], pos)

		err := renderer.SetDrawColor(color[0], color[1], color[2], 255)
		if err != nil {
			return
		}
		err = renderer.DrawLine(i, y, i, y+5)
		if err != nil {
			return
		}
	}
}

func interpolateColor(c1, c2 [3]uint8, t float32) [3]uint8 {
	return [3]uint8{
		uint8(float32(c1[0])*(1-t) + float32(c2[0])*t),
		uint8(float32(c1[1])*(1-t) + float32(c2[1])*t),
		uint8(float32(c1[2])*(1-t) + float32(c2[2])*t),
	}
}

func rotatePoint(point Point3D, angle float64) Point3D {
	cosA := math.Cos(angle)
	sinA := math.Sin(angle)

	// Rotation around the Y axis
	x := point.x*cosA - point.z*sinA
	z := point.x*sinA + point.z*cosA
	point.x, point.z = x, z

	// Rotation around the X axis
	y := point.y*cosA - point.z*sinA
	z = point.y*sinA + point.z*cosA
	point.y, point.z = y, z

	// Rotation around the Z axis
	x = point.x*cosA - point.y*sinA
	y = point.x*sinA + point.y*cosA
	point.x, point.y = x, y

	return point
}
func projectPoint(point Point3D) Point3D {
	distance := 3.0
	factor := distance / (distance + point.z) * zoomFactor
	x := point.x * factor * float64(windowWidth) / 2
	y := point.y * factor * float64(windowHeight) / 2
	return Point3D{x, y, point.z}
}
func fillPolygon(renderer *sdl.Renderer, points []sdl.Point, indices []int) {
	var sdlPoints []sdl.Point
	for _, index := range indices {
		sdlPoints = append(sdlPoints, points[index])
	}

	// Find the bounds of the polygon
	minY, maxY := sdlPoints[0].Y, sdlPoints[0].Y
	for _, p := range sdlPoints {
		if p.Y < minY {
			minY = p.Y
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	// Scanline algorithm to fill the polygon
	for y := minY; y <= maxY; y++ {
		var intersections []int32
		for i := 0; i < len(sdlPoints); i++ {
			j := (i + 1) % len(sdlPoints)
			if (sdlPoints[i].Y <= y && sdlPoints[j].Y > y) || (sdlPoints[j].Y <= y && sdlPoints[i].Y > y) {
				x := sdlPoints[i].X + (y-sdlPoints[i].Y)*(sdlPoints[j].X-sdlPoints[i].X)/(sdlPoints[j].Y-sdlPoints[i].Y)
				intersections = append(intersections, x)
			}
		}
		if len(intersections) > 1 {
			for i := 0; i < len(intersections)-1; i += 2 {
				if intersections[i] > intersections[i+1] {
					intersections[i], intersections[i+1] = intersections[i+1], intersections[i]
				}
				err := renderer.DrawLine(intersections[i], y, intersections[i+1], y)
				if err != nil {
					return
				}
			}
		}
	}
}
func drawCube(angle float64) {
	projectedPoints := make([]sdl.Point, len(cubeVertices))

	for i, vertex := range cubeVertices {
		rotated := rotatePoint(vertex, angle)
		projected := projectPoint(rotated)
		projectedPoints[i] = sdl.Point{
			X: int32(projected.x + float64(windowWidth)/2),
			Y: int32(projected.y + float64(windowHeight)/2),
		}
	}

	for i, face := range cubeFaces {
		if faceColors[i][3] > 0 { // Only draw non-transparent faces
			err := renderer.SetDrawColor(faceColors[i][0], faceColors[i][1], faceColors[i][2], faceColors[i][3])
			if err != nil {
				return
			}
			fillPolygon(renderer, projectedPoints, face)
		}
	}

	err := renderer.SetDrawColor(255, 255, 255, 255)
	if err != nil {
		return
	}
	for _, edge := range cubeEdges {
		start := projectedPoints[edge.start]
		end := projectedPoints[edge.end]
		err := renderer.DrawLine(start.X, start.Y, end.X, end.Y)
		if err != nil {
			return
		}
	}
}
func rotateCube() {
	// Rotate the cube
	rotationAngle += 0.01
}
func updateZoomLevel() {
	// Smoothly adjust the zoom factor
	if zoomFactor < targetZoom {
		zoomFactor += zoomStep
		if zoomFactor > targetZoom {
			zoomFactor = targetZoom
			//if !zoomReversed {
			//	targetZoom = 0.2
			//	zoomReversed = true
			//}
			// Optimized Implementation
			zoomFactor += zoomStep
		}
	} else if zoomFactor > targetZoom {
		zoomFactor -= zoomStep
		if zoomFactor < targetZoom {
			zoomFactor = targetZoom
		}
	}
}
func initStars() {
	// Initialize stars
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < numStars; i++ {
		stars = append(stars, Star{
			x:     rng.Float64()*2 - 1,
			y:     rng.Float64()*2 - 1,
			z:     rng.Float64()*2 - 1,
			speed: rng.Float64()*0.05 + 0.01,
			trail: make([]Point3D, 0, maxTrailLen),
		})
	}
}

func drawStarfield() {
	for i := range stars {
		stars[i].z -= stars[i].speed
		if stars[i].z <= 0 {
			stars[i] = Star{
				x:     rand.Float64()*2 - 1,
				y:     rand.Float64()*2 - 1,
				z:     1,
				speed: rand.Float64()*0.05 + 0.01,
				trail: make([]Point3D, 0, maxTrailLen),
			}
		}

		factor := 3.0 / stars[i].z
		x := stars[i].x * factor * float64(windowWidth) / 2
		y := stars[i].y * factor * float64(windowHeight) / 2

		if len(stars[i].trail) >= maxTrailLen {
			stars[i].trail = stars[i].trail[1:]
		}
		stars[i].trail = append(stars[i].trail, Point3D{x, y, stars[i].z})

		for j := len(stars[i].trail) - 1; j > 0; j-- {
			alpha := uint8(255 * float64(j) / float64(len(stars[i].trail)))
			err := renderer.SetDrawColor(alpha, alpha, alpha, alpha)
			if err != nil {
				return
			}
			err = renderer.DrawLine(
				int32(stars[i].trail[j-1].x)+windowWidth/2, int32(stars[i].trail[j-1].y)+windowHeight/2,
				int32(stars[i].trail[j].x)+windowWidth/2, int32(stars[i].trail[j].y)+windowHeight/2,
			)
			if err != nil {
				return
			}
		}

		err := renderer.SetDrawColor(255, 255, 255, 255)
		if err != nil {
			return
		}
		err = renderer.DrawPoint(int32(x)+windowWidth/2, int32(y)+windowHeight/2)
		if err != nil {
			return
		}
	}
}

func createBarTextures() {
	for i := 0; i < numBars; i++ {
		texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, windowWidth, barHeight)
		if err != nil {
			log.Fatalf("Failed to create texture: %s", err)
		}

		err = renderer.SetRenderTarget(texture)
		if err != nil {
			return
		}
		for y := 0; y < barHeight; y++ {
			ratio := math.Abs(float64(y-barHeight/2) / float64(barHeight/2))
			var r, g, b uint8
			switch i % 3 {
			case 0:
				r = uint8(255 * (1 - ratio))
				g = 0
				b = 0
			case 1:
				r = 0
				g = uint8(255 * (1 - ratio))
				b = 0
			case 2:
				r = 0
				g = 0
				b = uint8(255 * (1 - ratio))
			}

			err := renderer.SetDrawColor(r, g, b, 255)
			if err != nil {
				return
			}
			err = renderer.DrawLine(0, int32(y), windowWidth, int32(y))
			if err != nil {
				return
			}
		}
		err = renderer.SetRenderTarget(nil)
		if err != nil {
			return
		}
		barTextures = append(barTextures, texture)
	}

	for i := 0; i < numBars; i++ {
		barOffsets[i] = i * (int(windowHeight) / numBars)
	}
}
func drawCopperBars(elapsedTime float64) {
	for i := 0; i < numBars; i++ {
		// Calculate the new y position based on the sine function
		amplitude := float64(barHeight)
		frequency := 2.0
		baseY := (windowHeight / 2) + int32(i*barHeight/2)
		offsetY := amplitude * math.Sin(frequency*elapsedTime+float64(i)*0.4)

		dstRect := sdl.Rect{X: 0, Y: baseY + int32(offsetY), W: windowWidth, H: barHeight}

		// Render the bar
		err := renderer.Copy(barTextures[i], nil, &dstRect)
		if err != nil {
			return
		}
	}
}

func setupFont() error {
	fontTexture, err = loadTextureFromBytes(fontPng, renderer)
	if err != nil {
		return fmt.Errorf("failed to load font texture: %v", err)
	}
	return nil
}

func drawScrollText(text string, posX float64) {
	textLength := len(text)
	totalTextWidth := textLength * displayWidth

	for i := 0; i <= int(windowWidth)/displayWidth+1; i++ {
		for j, c := range text {
			charPos, ok := charMap[c]
			if !ok {
				continue
			}

			x := int32(posX) + int32(i*totalTextWidth) + int32(j*displayWidth)
			if x > windowWidth {
				break
			}

			srcRect := sdl.Rect{X: int32(charPos[0] * fontWidth), Y: int32(charPos[1] * fontHeight), W: fontWidth, H: fontHeight}
			offsetY := int32(20 * math.Sin(float64(x)/100))

			dstRect := sdl.Rect{X: x, Y: windowHeight/2 + offsetY, W: displayWidth, H: displayHeight}
			err := renderer.Copy(fontTexture, &srcRect, &dstRect)
			if err != nil {
				return
			}

			mirroredOffsetY := int32(-20 * math.Sin(float64(x)/100))
			mirroredDstRect := sdl.Rect{X: x, Y: windowHeight/2 + displayHeight + mirroredOffsetY, W: displayWidth, H: displayHeight}
			err = renderer.CopyEx(fontTexture, &srcRect, &mirroredDstRect, 0, nil, sdl.FLIP_VERTICAL)
			if err != nil {
				return
			}
		}
	}
}

func updateScrollTextPosition() {
	scrollPosX -= scrollSpeed * (60.0 / FPS)
	if scrollPosX <= -float64(len(scrollText)*displayWidth) {
		scrollPosX += float64(len(scrollText) * displayWidth)
	}
}

func handleEvents() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.QuitEvent:
			introQuit()
			//running = false
		case *sdl.KeyboardEvent:
			if e.State == sdl.PRESSED {
				switch e.Keysym.Sym {
				case sdl.K_UP:
					targetZoom = math.Min(targetZoom+0.1, 0.6)
				case sdl.K_DOWN:
					targetZoom = math.Max(targetZoom-0.1, 0.1)
				case sdl.K_q, sdl.K_ESCAPE:
					introQuit()
				}
			}
		}
	}
}
func parseCommandLineArgs() {
	fullscreen = true // Default to fullscreen
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if arg == "-win" {
			fullscreen = false
			// Check if custom width and height are provided
			if len(os.Args) > i+2 {
				if w, err := strconv.Atoi(os.Args[i+1]); err == nil {
					windowWidth = int32(w)
				}
				if h, err := strconv.Atoi(os.Args[i+2]); err == nil {
					windowHeight = int32(h)
				}
				i += 2 // Skip the width and height arguments
			} else {
				windowWidth = 1024
				windowHeight = 768
			}
		} else if arg == "-debug" {
			debug = true
		}
	}
}

func initSDL() error {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return err
	}
	if err := img.Init(img.INIT_PNG); err != nil {
		return err
	}
	if err := sdl.Init(sdl.INIT_AUDIO); err != nil {
		return err
	}
	if err := mix.OpenAudio(44100, mix.DEFAULT_FORMAT, 2, 4096); err != nil {
		return err
	}
	if err := mix.Init(mix.INIT_MOD); err != nil {
		return err
	}
	return nil
}

func setupDisplay() {
	var flags uint32 = sdl.WINDOW_SHOWN
	if fullscreen {
		flags |= sdl.WINDOW_FULLSCREEN_DESKTOP
		dm, err := sdl.GetCurrentDisplayMode(0)
		if err != nil {
			log.Fatalf("Failed to get display mode: %s", err)
		}
		windowWidth = dm.W
		windowHeight = dm.H
	} else {
		flags |= sdl.WINDOW_BORDERLESS
	}

	window, renderer, err = sdl.CreateWindowAndRenderer(windowWidth, windowHeight, flags)
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Failed to create window and renderer: %s\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}

func loadModFromBytes(data []byte) (*mix.Music, error) {
	rwops, err := sdl.RWFromMem(data)
	if err != nil {
		return nil, fmt.Errorf("could not create RWops from bytes: %v", err)
	}

	music, err := mix.LoadMUSRW(rwops, 0)
	if err != nil {
		return nil, fmt.Errorf("could not load MOD from RWops: %v", err)
	}

	return music, nil
}
func playMusic() {
	music, err := loadModFromBytes(comicbakeryMod)
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Failed to load music: %s\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}

	if err := music.Play(-1); err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Failed to play music: %s\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}

func loadTextureFromBytes(data []byte, renderer *sdl.Renderer) (*sdl.Texture, error) {
	rwops, err := sdl.RWFromMem(data)
	if err != nil {
		return nil, fmt.Errorf("could not create RWops from bytes: %v", err)
	}
	imgSurface, err := img.LoadPNGRW(rwops)
	if err != nil {
		return nil, fmt.Errorf("could not load image from RWops: %v", err)
	}
	defer imgSurface.Free()

	// Set color key to make black transparent
	if err := imgSurface.SetColorKey(true, sdl.MapRGB(imgSurface.Format, 0, 0, 0)); err != nil {
		return nil, fmt.Errorf("could not set color key: %v", err)
	}

	texture, err := renderer.CreateTextureFromSurface(imgSurface)
	if err != nil {
		return nil, fmt.Errorf("could not create texture: %v", err)
	}

	return texture, nil
}
func setupBouncingLogo() error {
	var err error
	texture, err = loadTextureFromBytes(intuitiontextlogoPng, renderer)
	if err != nil {
		return err
	}

	// Get the dimensions of the texture
	_, _, imageWidth, imageHeight, err = texture.Query()
	if err != nil {
		return err
	}

	// Initial position and target position
	posY = -int(imageHeight)
	targetY = (int(windowHeight) - int(imageHeight)) / 7

	return nil
}
func updateBouncingLogoPosition() {
	// Calculate the elapsed time
	elapsed := time.Since(startTime).Seconds()

	// Animate the position of the image
	if posY < targetY {
		posY += int(float64(imageHeight) * elapsed * 0.05) // Adjust logo sliding speed
		if posY > targetY {
			posY = targetY
		}
	} else {
		posY = targetY + int(4*math.Sin(elapsed*20)) // Adjust bounce here
	}
}
func drawBouncingLogo() {
	// Draw the image
	dstRect := sdl.Rect{
		X: (windowWidth - imageWidth) / 2,
		Y: int32(posY),
		W: imageWidth,
		H: imageHeight,
	}
	err := renderer.Copy(texture, nil, &dstRect)
	if err != nil {
		return
	}
}

func introQuit() {
	// Enable blending mode
	err := renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		return
	}

	// Fade out the music and quit
	for i := 0; i <= 255; i++ {
		// Reduce the volume of the music
		mix.VolumeMusic(255 - i)

		// Update animations
		rotateCube()
		updateScrollTextPosition()

		// Draw the current screen content
		err := renderer.SetDrawColor(0, 0, 0, 255)
		if err != nil {
			return
		}
		err = renderer.Clear()
		if err != nil {
			return
		}

		// Draw the existing scene
		drawStarfield()
		drawCube(rotationAngle)
		drawScrollText(scrollText, scrollPosX)

		// Draw a full screen semi-transparent rectangle
		err = renderer.SetDrawColor(0, 0, 0, uint8(i))
		if err != nil {
			return
		}
		err = renderer.FillRect(&sdl.Rect{X: 0, Y: 0, W: windowWidth, H: windowHeight})
		if err != nil {
			return
		}

		// Present the renderer
		renderer.Present()
		sdl.Delay(400 / FPS)
	}
	running = false
	fmt.Println("\n\nThanks for watching...\n")
}
