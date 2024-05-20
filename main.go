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
	"time"
)

const (
	windowWidth   = 1024
	windowHeight  = 768
	FPS           = 60
	numStars      = 1000
	maxTrailLen   = 5
	scrollSpeed   = 10
	fontWidth     = 32
	fontHeight    = 32
	displayWidth  = 64
	displayHeight = 64
	imagePath     = "intuitiontextlogo.png"
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
	renderer *sdl.Renderer
	window   *sdl.Window
	err      error

	//Cube variables
	zoomFactor    = 0.1
	targetZoom    = 0.6
	zoomStep      = 0.002
	zoomReversed  = false
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
	scrollText  = "..:INTUITION PRESENTS:..    \"I FEEL 16 AGAIN!\"    ..:PRESS THE UP AND DOWN KEYS TO ZOOM THE CUBE IN AND OUT:..    ..:PRESS Q OR ESC TO QUIT:..    ..:ORIGINAL COMIC BAKERY MUSIC FOR C64 BY MARTIN GALWAY IN 1984...     ..:SID TO PROTRACKER CONVERSION FOR AMIGA BY H0FFMAN (DREAMFISH/TRSI) IN 1994:..    ..:GOLANG CODE BY INTUITION IN 2024:..    ..:FONT GRAPHICS BY UNKNOWN:..    ..:GREETS TO KARLOS AND GADGETMASTER!!!:..          "
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
	barSpeed    = 4
	barOffsets  = make([]int, numBars)
	barTextures []*sdl.Texture

	// Loop control
	running   = true
	startTime = time.Now()
)

func main() {
	initSDL()
	defer sdl.Quit()
	defer img.Quit()
	defer mix.Quit()

	fmt.Println("Cubetro by Intuition (2024)")
	setupDisplay()
	defer window.Destroy()
	defer renderer.Destroy()

	// Hide the mouse cursor
	sdl.ShowCursor(sdl.DISABLE)

	initStars()
	playMusic()

	if err := setupFont(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup font: %s\n", err)
		os.Exit(1)
	}
	defer fontTexture.Destroy()

	if err := setupBouncingLogo(); err != nil {
		log.Fatalf("Failed to setup bouncing logo: %s", err)
	}
	defer texture.Destroy()

	// Pre-render the copper bars into textures
	createBarTextures()

	// Main loop
	startTime = time.Now()
	for running {
		handleEvents()
		updateZoomLevel() // Zoom in/out

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		drawStarfield()

		drawCopperBars(time.Since(startTime).Seconds())
		drawCube(rotationAngle)
		rotateCube()

		drawScrollText(scrollText, scrollPosX)
		updateScrollTextPosition()

		updateBouncingLogoPosition()
		drawBouncingLogo()

		renderer.Present()
		sdl.Delay(1000 / FPS)
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
	x := point.x * factor * windowWidth / 2
	y := point.y * factor * windowHeight / 2
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
				renderer.DrawLine(intersections[i], y, intersections[i+1], y)
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
			X: int32(projected.x + windowWidth/2),
			Y: int32(projected.y + windowHeight/2),
		}
	}

	for i, face := range cubeFaces {
		if faceColors[i][3] > 0 { // Only draw non-transparent faces
			renderer.SetDrawColor(faceColors[i][0], faceColors[i][1], faceColors[i][2], faceColors[i][3])
			fillPolygon(renderer, projectedPoints, face)
		}
	}

	renderer.SetDrawColor(255, 255, 255, 255)
	for _, edge := range cubeEdges {
		start := projectedPoints[edge.start]
		end := projectedPoints[edge.end]
		renderer.DrawLine(start.X, start.Y, end.X, end.Y)
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
			if !zoomReversed {
				targetZoom = 0.1
				zoomReversed = true
			}
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
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < numStars; i++ {
		stars = append(stars, Star{
			x:     rand.Float64()*2 - 1,
			y:     rand.Float64()*2 - 1,
			z:     rand.Float64()*2 - 1,
			speed: rand.Float64()*0.05 + 0.01,
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
		x := stars[i].x * factor * windowWidth / 2
		y := stars[i].y * factor * windowHeight / 2

		if len(stars[i].trail) >= maxTrailLen {
			stars[i].trail = stars[i].trail[1:]
		}
		stars[i].trail = append(stars[i].trail, Point3D{x, y, stars[i].z})

		for j := len(stars[i].trail) - 1; j > 0; j-- {
			alpha := uint8(255 * float64(j) / float64(len(stars[i].trail)))
			renderer.SetDrawColor(alpha, alpha, alpha, alpha)
			renderer.DrawLine(
				int32(stars[i].trail[j-1].x)+windowWidth/2, int32(stars[i].trail[j-1].y)+windowHeight/2,
				int32(stars[i].trail[j].x)+windowWidth/2, int32(stars[i].trail[j].y)+windowHeight/2,
			)
		}

		renderer.SetDrawColor(255, 255, 255, 255)
		renderer.DrawPoint(int32(x)+windowWidth/2, int32(y)+windowHeight/2)
	}
}

func createBarTextures() {
	for i := 0; i < numBars; i++ {
		texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, windowWidth, barHeight)
		if err != nil {
			log.Fatalf("Failed to create texture: %s", err)
		}

		renderer.SetRenderTarget(texture)
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

			renderer.SetDrawColor(r, g, b, 255)
			renderer.DrawLine(0, int32(y), windowWidth, int32(y))
		}
		renderer.SetRenderTarget(nil)
		barTextures = append(barTextures, texture)
	}

	for i := 0; i < numBars; i++ {
		barOffsets[i] = i * (windowHeight / numBars)
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
		renderer.Copy(barTextures[i], nil, &dstRect)
	}
}

func setupFont() error {
	fontSurface, err := img.Load("32X32-FM.png")
	if err != nil {
		return err
	}
	defer fontSurface.Free()

	// Set color key to make black transparent
	if err := fontSurface.SetColorKey(true, sdl.MapRGB(fontSurface.Format, 0, 0, 0)); err != nil {
		return err
	}

	fontTexture, err = renderer.CreateTextureFromSurface(fontSurface)
	if err != nil {
		return err
	}
	return nil
}
func drawScrollText(text string, posX float64) {
	textLength := len(text)
	totalTextWidth := textLength * displayWidth

	for i := 0; i <= windowWidth/displayWidth+1; i++ {
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

			// Draw the original text
			dstRect := sdl.Rect{X: x, Y: windowHeight/2 + offsetY, W: displayWidth, H: displayHeight}
			renderer.Copy(fontTexture, &srcRect, &dstRect)

			// Draw the mirrored text
			mirroredOffsetY := int32(-20 * math.Sin(float64(x)/100))
			mirroredDstRect := sdl.Rect{X: x, Y: windowHeight/2 + displayHeight + mirroredOffsetY, W: displayWidth, H: displayHeight}
			renderer.CopyEx(fontTexture, &srcRect, &mirroredDstRect, 0, nil, sdl.FLIP_VERTICAL)
		}
	}
}
func updateScrollTextPosition() {
	// Update the position of the scrolling text
	scrollPosX -= scrollSpeed
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
	window, renderer, err = sdl.CreateWindowAndRenderer(windowWidth, windowHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window and renderer: %s\n", err)
		os.Exit(1)
	}
}
func playMusic() {
	if music, errLoadingMusic := mix.LoadMUS("comicbakery3.mod"); errLoadingMusic != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to load music from disk: %s\n", errLoadingMusic)
		os.Exit(1)
	} else {
		_ = music.Play(-1)
	}
}

func loadTexture(file string, renderer *sdl.Renderer) (*sdl.Texture, error) {
	imgSurface, err := img.Load(file)
	if err != nil {
		return nil, fmt.Errorf("could not load image: %v", err)
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
	texture, err = loadTexture(imagePath, renderer)
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
	targetY = (windowHeight - int(imageHeight)) / 7

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
	renderer.Copy(texture, nil, &dstRect)
}

func introQuit() {
	// Enable blending mode
	renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

	// Fade out the music and quit
	for i := 0; i <= 255; i++ {
		// Reduce the volume of the music
		mix.VolumeMusic(255 - i)

		// Update animations
		rotateCube()
		updateScrollTextPosition()

		// Draw the current screen content
		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		// Draw the existing scene
		drawStarfield()
		drawCube(rotationAngle)
		drawScrollText(scrollText, scrollPosX)

		// Draw a full screen semi-transparent rectangle
		renderer.SetDrawColor(0, 0, 0, uint8(i))
		renderer.FillRect(&sdl.Rect{X: 0, Y: 0, W: windowWidth, H: windowHeight})

		// Present the renderer
		renderer.Present()
		sdl.Delay(400 / FPS)
	}
	running = false
	fmt.Println("\nThanks for watching...\n")
}
