package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

type Direction int

const (
	Up Direction = iota
	Down
	Left
	Right
)

type Action struct {
	Type string
	Dir  Direction
}

type Point struct {
	x, y int
}

type Snake []Point

var boardWidth int
var boardHeight int

var snake Snake
var foods []Point
var dir Direction
var gameOver bool
var paused bool
var score int
var highScore int
var difficulty string
var speed time.Duration
var foodCount int
var oldTerm syscall.Termios

func init() {
	rand.Seed(time.Now().Unix())
}

func loadHighScore() int {
	data, err := os.ReadFile("highscore.txt")
	if err != nil {
		return 0
	}
	hs, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0
	}
	return hs
}

func saveHighScore(hs int) {
	err := os.WriteFile("highscore.txt", []byte(strconv.Itoa(hs)), 0644)
	if err != nil {
		fmt.Println("Error saving high score:", err)
	}
}

func reset() {
	snake = Snake{{boardWidth / 2, boardHeight / 2}}
	dir = Right
	gameOver = false
	paused = false
	score = 0
	foods = []Point{}
	for i := 0; i < foodCount; i++ {
		placeOneFood()
	}
}

func placeOneFood() {
	for {
		p := Point{rand.Intn(boardWidth), rand.Intn(boardHeight)}
		if !containsSnake(p) && !containsFood(p) {
			foods = append(foods, p)
			break
		}
	}
}

func containsFood(p Point) bool {
	for _, f := range foods {
		if f == p {
			return true
		}
	}
	return false
}

func containsSnake(p Point) bool {
	for _, pp := range snake {
		if pp == p {
			return true
		}
	}
	return false
}

func move() {
	head := snake[0]
	switch dir {
	case Up:
		head.y--
	case Down:
		head.y++
	case Left:
		head.x--
	case Right:
		head.x++
	}
	if head.x < 0 || head.x >= boardWidth || head.y < 0 || head.y >= boardHeight {
		gameOver = true
		return
	}
	for _, p := range snake {
		if p == head {
			gameOver = true
			return
		}
	}
	snake = append([]Point{head}, snake...)
	ate := false
	for i, f := range foods {
		if head == f {
			score++
			// remove this food
			foods = append(foods[:i], foods[i+1:]...)
			ate = true
			break
		}
	}
	if ate {
		placeOneFood()
	} else {
		snake = snake[:len(snake)-1]
	}
}

func draw() {
	fmt.Print("\033[2J\033[1;1H") // clear screen and move to top
	// Top border
	fmt.Print("┌")
	for i := 0; i < boardWidth; i++ {
		fmt.Print("─")
	}
	fmt.Println("┐")
	// Rows
	for y := 0; y < boardHeight; y++ {
		fmt.Print("│")
		for x := 0; x < boardWidth; x++ {
			p := Point{x, y}
			if containsSnake(p) {
				fmt.Print("\033[32m◆\033[0m")
			} else if containsFood(p) {
				fmt.Print("\033[31m★\033[0m")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println("│")
	}
	// Bottom border
	fmt.Print("└")
	for i := 0; i < boardWidth; i++ {
		fmt.Print("─")
	}
	fmt.Println("┘")
	fmt.Printf("Difficulty: %s | Score: %d | High Score: %d\n", difficulty, score, highScore)
}

func setRawMode() {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(os.Stdin.Fd()), uintptr(syscall.TCGETS), uintptr(unsafe.Pointer(&oldTerm)))
	if err != 0 {
		panic(err)
	}
	term := oldTerm
	term.Lflag &^= syscall.ICANON | syscall.ECHO
	term.Cc[syscall.VMIN] = 1
	term.Cc[syscall.VTIME] = 0
	_, _, err = syscall.Syscall(syscall.SYS_IOCTL, uintptr(os.Stdin.Fd()), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&term)))
	if err != 0 {
		panic(err)
	}
}

func restoreMode() {
	_, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(os.Stdin.Fd()), uintptr(syscall.TCSETS), uintptr(unsafe.Pointer(&oldTerm)))
	if err != 0 {
		panic(err)
	}
}

func drawStartScreen() {
	width := 40
	fmt.Print("┌")
	for i := 0; i < width-2; i++ {
		fmt.Print("─")
	}
	fmt.Println("┐")

	// Title
	title := "SNAKE CLI"
	padding := (width - 2 - len(title)) / 2
	fmt.Print("│")
	for i := 0; i < padding; i++ {
		fmt.Print(" ")
	}
	fmt.Print(title)
	for i := 0; i < width-2-padding-len(title); i++ {
		fmt.Print(" ")
	}
	fmt.Println("│")

	// Empty line
	fmt.Print("│")
	for i := 0; i < width-2; i++ {
		fmt.Print(" ")
	}
	fmt.Println("│")

	// Instructions
	lines := []string{
		"Instructions:",
		"- Use arrow keys or WASD to move",
		"- P to pause/resume",
		"- F to restart, Q to quit",
		"",
		fmt.Sprintf("High Score: %d", highScore),
		"",
		"Press any key to start",
	}

	for _, line := range lines {
		fmt.Print("│ ")
		fmt.Print(line)
		for i := 0; i < width-4-len(line); i++ {
			fmt.Print(" ")
		}
		fmt.Println(" │")
	}

	// Bottom border
	fmt.Print("└")
	for i := 0; i < width-2; i++ {
		fmt.Print("─")
	}
	fmt.Println("┘")
}

func main() {
	highScore = loadHighScore()

	setRawMode()
	defer restoreMode()

	var r = bufio.NewReader(os.Stdin)

	drawStartScreen()
	r.ReadByte() // wait for any key

	fmt.Println("Welcome to Snake CLI!")
	fmt.Printf("Current High Score: %d\n", highScore)
	fmt.Println("Choose board size:")
	fmt.Println("1. 10x10")
	fmt.Println("2. 15x15")
	fmt.Println("3. 20x20")
	fmt.Print("Enter choice (1-3): ")

	var b byte
	var err error
	// Read choice
	for {
		b, err = r.ReadByte()
		if err != nil {
			continue
		}
		if b == '1' {
			boardWidth, boardHeight = 20, 10
			break
		} else if b == '2' {
			boardWidth, boardHeight = 30, 15
			break
		} else if b == '3' {
			boardWidth, boardHeight = 40, 20
			break
		}
	}

	reset()

	fmt.Println("Choose difficulty:")
	fmt.Println("1. Easy")
	fmt.Println("2. Medium")
	fmt.Println("3. Hard")
	fmt.Print("Enter choice (1-3): ")
	for {
		b, err = r.ReadByte()
		if err != nil {
			continue
		}
		if b == '1' {
			difficulty = "Easy"
			speed = 500 * time.Millisecond
			foodCount = 1
			break
		} else if b == '2' {
			difficulty = "Medium"
			speed = 300 * time.Millisecond
			foodCount = 2
			break
		} else if b == '3' {
			difficulty = "Hard"
			speed = 150 * time.Millisecond
			foodCount = 3
			break
		}
	}

	actionChan := make(chan Action)

	ticker := time.NewTicker(speed)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			actionChan <- Action{Type: "tick"}
		}
	}()
	go func() {
		for {
			b, err := r.ReadByte()
			if err != nil {
				return
			}
			if b == 'p' || b == 'P' {
				actionChan <- Action{Type: "pause"}
			} else if b == 'w' || b == 'W' {
				if dir != Down {
					actionChan <- Action{Type: "move", Dir: Up}
				}
			} else if b == 's' || b == 'S' {
				if dir != Up {
					actionChan <- Action{Type: "move", Dir: Down}
				}
			} else if b == 'd' || b == 'D' {
				if dir != Left {
					actionChan <- Action{Type: "move", Dir: Right}
				}
			} else if b == 'a' || b == 'A' {
				if dir != Right {
					actionChan <- Action{Type: "move", Dir: Left}
				}
			} else if b == 27 { // escape sequence
				b2, err := r.ReadByte()
				if err != nil {
					continue
				}
				if b2 == 91 {
					var b3 byte
					b3, err = r.ReadByte()
					if err != nil {
						continue
					}
					switch b3 {
					case 65: // up
						if dir != Down {
							actionChan <- Action{Type: "move", Dir: Up}
						}
					case 66: // down
						if dir != Up {
							actionChan <- Action{Type: "move", Dir: Down}
						}
					case 67: // right
						if dir != Left {
							actionChan <- Action{Type: "move", Dir: Right}
						}
					case 68: // left
						if dir != Right {
							actionChan <- Action{Type: "move", Dir: Left}
						}
					}
				}
			}
		}
	}()

	draw() // initial draw
	for {
		for !gameOver {
			var action = <-actionChan
			if action.Type == "pause" {
				paused = !paused
				if paused {
					fmt.Println("\033[33mGame paused. Press P to resume.\033[0m")
				} else {
					fmt.Println("\033[33mGame resumed.\033[0m")
				}
			} else if action.Type == "move" {
				dir = action.Dir
			} else if action.Type == "tick" && !paused {
				move()
				draw()
			}
		}

		// Check and save high score immediately when game ends
		if score > highScore {
			highScore = score
			saveHighScore(highScore)
		}

		// Game over
		fmt.Println("\033[33mGame Over! Press F to restart, Q to quit.\033[0m")
		restart := false
		for {
			b, err = r.ReadByte()
			if err != nil {
				continue
			}
			if b == 'f' || b == 'F' {
				reset()
				draw()
				restart = true
				break
			} else if b == 'q' || b == 'Q' {
				return
			}
		}
		if !restart {
			break
		}
	}
}