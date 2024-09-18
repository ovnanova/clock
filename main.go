package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	width        = 80
	height       = 24
	centerX      = width / 2
	centerY      = height / 2
	maxSpeed     = 20 * time.Millisecond
	initialSpeed = 100 * time.Millisecond
	acceleration = 20 * time.Millisecond
)

type Piece struct {
	x, y    float64 // Positions
	vx, vy  float64 // Velocities
	char    rune    // Character to display
	stopped bool    // Whether the piece has stopped moving
}

func main() {
	hideCursor()
	defer showCursor()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		showCursor()
		clearScreen()
		os.Exit(0)
	}()

	speed := initialSpeed
	spinCount := 0
	minuteAngle := 0.0
	hourAngle := 0.0

	for {
		for step := 0; step < 24; step++ {
			clearScreen()

			grid := make([][]rune, height)
			for i := range grid {
				grid[i] = make([]rune, width)
				for j := range grid[i] {
					grid[i][j] = ' '
				}
			}

			drawCircle(grid)
			drawMarkers(grid)

			minuteAngle += 15 * (math.Pi / 180)
			hourAngle += 1.25 * (math.Pi / 180)

			drawHand(grid, hourAngle, height/4, 'h')
			drawHand(grid, minuteAngle, height/2-4, 'm')

			for _, line := range grid {
				fmt.Println(string(line))
			}

			time.Sleep(speed)
		}

		if speed > maxSpeed {
			speed -= acceleration
		} else {
			spinCount++
			if spinCount > 5 {
				breakApartAnimation(minuteAngle, hourAngle)
				return
			}
		}
	}
}

func hideCursor() {
	fmt.Print("\033[?25l")
}

func showCursor() {
	fmt.Print("\033[?25h")
}

func clearScreen() {
	fmt.Print("\033[H\033[2J\033[3J")
}

func drawCircle(grid [][]rune) {
	radius := float64(height/2) - 2
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			dx := (float64(x - centerX)) * 0.5
			dy := float64(y - centerY)
			distance := math.Sqrt(dx*dx + dy*dy)
			if distance >= radius-0.5 && distance <= radius+0.5 {
				grid[y][x] = '*'
			}
		}
	}
}

func drawMarkers(grid [][]rune) {
	for hour := 1; hour <= 12; hour++ {
		angle := float64(hour) * 30 * (math.Pi / 180)
		x := int(float64(centerX) + (float64(height/2)-3)*math.Sin(angle)*2)
		y := int(float64(centerY) - (float64(height/2)-3)*math.Cos(angle))
		if x >= 0 && x < width && y >= 0 && y < height {
			hourStr := fmt.Sprintf("%d", hour)
			if len(hourStr) == 1 {
				grid[y][x] = rune(hourStr[0])
			} else if len(hourStr) == 2 {
				if x-1 >= 0 && x < width {
					grid[y][x-1] = rune(hourStr[0])
					grid[y][x] = rune(hourStr[1])
				}
			}
		}
	}
}

func drawHand(grid [][]rune, angle float64, length int, symbol rune) {
	for i := 1; i <= length; i++ {
		x := int(math.Round(float64(centerX) + float64(i)*math.Sin(angle)*2))
		y := int(math.Round(float64(centerY) - float64(i)*math.Cos(angle)))
		if x >= 0 && x < width && y >= 0 && y < height {
			grid[y][x] = symbol
		}
	}
}

func breakApartAnimation(minuteAngle, hourAngle float64) {
	var pieces []Piece

	grid := make([][]rune, height)
	for i := range grid {
		grid[i] = make([]rune, width)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	drawCircle(grid)
	drawMarkers(grid)

	drawHand(grid, hourAngle, height/4, 'h')
	drawHand(grid, minuteAngle, height/2-4, 'm')

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if grid[y][x] != ' ' {
				dx := float64(x - centerX)
				dy := float64(y - centerY)
				distance := math.Sqrt(dx*dx + dy*dy)
				if distance == 0 {
					distance = 1
				}
				speed := 1.5
				vx := (dx / distance) * speed
				vy := (dy / distance) * speed
				pieces = append(pieces, Piece{
					x:       float64(x),
					y:       float64(y),
					vx:      vx,
					vy:      vy,
					char:    grid[y][x],
					stopped: false,
				})
			}
		}
	}

	simulatePieces(pieces)
}

func simulatePieces(pieces []Piece) {
	ground := make([][]bool, height)
	for i := range ground {
		ground[i] = make([]bool, width)
	}

	gravity := 0.2
	for {
		clearScreen()
		grid := make([][]rune, height)
		for i := range grid {
			grid[i] = make([]rune, width)
			for j := range grid[i] {
				grid[i][j] = ' '
			}
		}

		allStopped := true
		for i := range pieces {
			if pieces[i].stopped {
				x := int(pieces[i].x)
				y := int(pieces[i].y)
				if x >= 0 && x < width && y >= 0 && y < height {
					grid[y][x] = pieces[i].char
				}
				continue
			}

			pieces[i].vy += gravity

			newX := pieces[i].x + pieces[i].vx
			newY := pieces[i].y + pieces[i].vy

			x := int(newX)
			y := int(newY)

			if x < 0 || x >= width {
				pieces[i].vx = -pieces[i].vx
				newX = pieces[i].x + pieces[i].vx
				x = int(newX)
			}

			if y >= height-1 || (y+1 >= 0 && y+1 < height && ground[y+1][x]) {
				pieces[i].stopped = true
				if y >= 0 && y < height && x >= 0 && x < width {
					ground[y][x] = true
					grid[y][x] = pieces[i].char
				}
			} else {
				if y >= 0 && y < height && x >= 0 && x < width {
					grid[y][x] = pieces[i].char
				}
				pieces[i].x = newX
				pieces[i].y = newY
				allStopped = false
			}
		}

		for _, line := range grid {
			fmt.Println(string(line))
		}

		if allStopped {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	showCursor()
}
