package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

const windowWidth = 800
const windowHeight = 600

const startingPlayerLives = 3

const ballStartPositionX = 50
const ballStartPositionY = 450

const ballStartSpeedX = 3
const ballStartSpeedY = -3

const brickRowCount = 5
const brickColumnCount = 8

type Rect struct {
	sdl.Rect
}

type Point struct {
	sdl.Point
}

type Ball struct {
	rect  Rect
	speed Point
}

type Brick struct {
	rect   Rect
	active bool
}

func main() {

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("GoKanoid",
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		windowWidth, windowHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()

	var playerLives int32 = startingPlayerLives

	paddleRect := Rect{sdl.Rect{X: 50, Y: 530, W: 64, H: 8}}
	var paddleSpeed int32 = 12

	ball := Ball{
		rect:  Rect{sdl.Rect{X: ballStartPositionX, Y: ballStartPositionY, W: 8, H: 8}},
		speed: Point{sdl.Point{X: ballStartSpeedX, Y: ballStartSpeedY}},
	}

	var bricks = [brickRowCount][brickColumnCount] Brick{}
	var brickCount int32 = brickRowCount * brickColumnCount

	const brickBoardStartX = 180
	const brickBoardStartY = 50

	const brickOffsetX = 32
	const brickOffsetY = 12

	const brickWidth = 32
	const brickHeight = 12

	for row := 0; row < brickRowCount; row++ {
		for column := 0; column < brickColumnCount; column++ {
			bricks[row][column] = Brick{Rect{
				sdl.Rect{X: brickBoardStartX + int32(column*(brickWidth+brickOffsetX)),
					Y: brickBoardStartY + int32(row*(brickHeight+brickOffsetY)),
					W: brickWidth,
					H: brickHeight}},
				true}
		}
	}

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch eventType := event.(type) {
			case *sdl.QuitEvent:
				println("Exited Game")
				running = false
				break

			case *sdl.KeyboardEvent:
				if eventType.State == sdl.PRESSED {
					switch event.(*sdl.KeyboardEvent).Keysym.Sym {
					case sdl.K_ESCAPE:
						running = false
						break
					case sdl.K_LEFT:
						paddleRect.X -= paddleSpeed
						paddleRect.checkForPaddleWallCollisions()
						break
					case sdl.K_RIGHT:
						paddleRect.X += paddleSpeed
						paddleRect.checkForPaddleWallCollisions()
						break
					}

				}
			}
		}

		ball.updateBall()
		ball.resolveBallWallCollisions(&playerLives)
		ball.resolveBallPaddleCollisions(paddleRect)

		// Check for ball brick collision
		for row := 0; row < brickRowCount; row++ {
			for column := 0; column < brickColumnCount; column++ {
				if bricks[row][column].active {
					ball.resolveBallBrickCollisions(&bricks[row][column], &brickCount)
				}
			}
		}

		if checkForVictoryConditions(playerLives, brickCount) {

			ball.restartPlayerAndBall(&playerLives)

			for row := 0; row < brickRowCount; row++ {
				for column := 0; column < brickColumnCount; column++ {
					bricks[row][column].active = true
				}
			}

			brickCount = brickRowCount * brickColumnCount
		}

		renderer.SetDrawColor(100, 149, 237, 255)
		renderer.Clear()

		renderer.SetDrawColor(0, 255, 0, 255)
		renderer.FillRect(&paddleRect.Rect)

		renderer.SetDrawColor(0, 0, 255, 255)
		renderer.FillRect(&ball.rect.Rect)

		renderer.SetDrawColor(255, 100, 0, 0)
		for row := 0; row < brickRowCount; row++ {
			for column := 0; column < brickColumnCount; column++ {
				if bricks[row][column].active {
					renderer.FillRect(&bricks[row][column].rect.Rect)
				}
			}
		}

		// The rects have been drawn, now it is time to tell the renderer to show
		// what has been draw to the screen
		renderer.Present()

		sdl.Delay(16)
	}

}

func (ball *Ball) updateBall() {
	ball.rect.X += ball.speed.X
	ball.rect.Y += ball.speed.Y
}

func (paddleRect *Rect) checkForPaddleWallCollisions() {

	if paddleRect.X < 0 {
		paddleRect.X = 0
	}

	if paddleRect.X > (windowWidth - paddleRect.W) {
		paddleRect.X = windowWidth - paddleRect.W
	}
}

func (ball *Ball) resolveBallWallCollisions(playerLives *int32) {
	if ball.rect.X < 0 {
		ball.rect.X = 0
		ball.speed.X = -ball.speed.X
	}
	if ball.rect.right() > windowWidth {
		ball.rect.setRight(windowWidth)
		ball.speed.X = -ball.speed.X
	}
	if ball.rect.Y < 0 {
		ball.rect.Y = 0
		ball.speed.Y = -ball.speed.Y
	}
	if ball.rect.bottom() > windowHeight {
		ball.rect.setBottom(windowHeight)
		ball.speed.Y = -ball.speed.Y

		*playerLives--
		ball.resetBall()
	}
}

func (ball *Ball) resolveBallPaddleCollisions(paddleRect Rect) {
	if ball.rect.HasIntersection(&paddleRect.Rect) {
		if (ball.rect.bottom() >= paddleRect.Y) && ball.speed.Y > 0 {
			ball.speed.Y = -ball.speed.Y
			// TODO: Make the ball change direction based on the position of the paddle (experiment also with speed change)
		}
	}
}

func (ball *Ball) resolveBallBrickCollisions(brick *Brick, brickCount *int32) {
	collided := false

	if ball.rect.HasIntersection(&brick.rect.Rect) {
		if (ball.rect.bottom() >= brick.rect.Rect.Y) && ball.speed.Y > 0 ||
			(ball.rect.Y >= brick.rect.bottom()) && ball.speed.Y < 0 {
			ball.speed.Y = -ball.speed.Y
			collided = true
		}
		if (ball.rect.right() >= brick.rect.Rect.X) && ball.speed.X > 0 ||
			(ball.rect.X >= brick.rect.Rect.X) && ball.speed.X < 0 {
			ball.speed.X = -ball.speed.X
			collided = true
		}
	}

	if collided {
		brick.active = false
		*brickCount--
	}
}

func checkForVictoryConditions(playerLives int32, brickCount int32) bool{
	needsRestart := false

	if playerLives == 0 {
		println("GameOver!")
		needsRestart = true
	}
	if brickCount == 0 {
		println("Victory!")
		needsRestart = true
	}

	return needsRestart
}

func (ball *Ball)restartPlayerAndBall(playerLives *int32){
		*playerLives = startingPlayerLives

		ball.resetBall()
}

func (ball *Ball)resetBall(){

	ball.rect.X = ballStartPositionX
	ball.rect.Y = ballStartPositionY

	ball.speed.X = ballStartSpeedX
	ball.speed.Y = ballStartSpeedY
}


// Rect extensions
func (rect *Rect) right() int32 {
	return rect.X + rect.W
}

func (rect *Rect) setRight(rightPosition int32) {
	rect.X = rightPosition - rect.W
}

func (rect *Rect) bottom() int32 {
	return rect.Y + rect.H
}

func (rect *Rect) setBottom(bottomPosition int32) {
	rect.Y = bottomPosition - rect.H
}