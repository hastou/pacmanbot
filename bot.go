package main

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"
)
import "os"
import "bufio"

type Vector2d struct {
	x, y int
}

var (
	UP             = Vector2d{0, -1}
	DOWN           = Vector2d{0, 1}
	RIGHT          = Vector2d{1, 0}
	LEFT           = Vector2d{-1, 0}
	ALL_DIRECTIONS = [4]Vector2d{UP, DOWN, LEFT, RIGHT}
)

func (v *Vector2d) Add(d *Vector2d) Vector2d {
	return Vector2d{v.x + d.x, v.y + d.y}
}

func (v *Vector2d) Subtract(d *Vector2d) Vector2d {
	return Vector2d{v.x - d.x, v.y - d.y}
}

func (v *Vector2d) String() string {
	return fmt.Sprintf("{%v, %v}", v.x, v.y)
}

func (v *Vector2d) Direction() string {
	switch *v {
	case UP:
		return "UP"
	case DOWN:
		return "DOWN"
	case RIGHT:
		return "RIGHT"
	case LEFT:
		return "LEFT"
	}
	return "NOT A DIRECTION"
}

type Pacman struct {
	pacId                           int
	position                        Vector2d
	typeId                          string
	speedTurnsLeft, abilityCooldown int
	positionHistory                 []Vector2d
}

func NewPacman(pacId, x, y int, typeId string) *Pacman {
	pacman := Pacman{}
	pacman.pacId = pacId
	pacman.position = Vector2d{x, y}
	pacman.typeId = typeId
	pacman.speedTurnsLeft = 0
	pacman.abilityCooldown = 0
	pacman.positionHistory = make([]Vector2d, 0, 200)
	return &pacman
}

func (pacman *Pacman) Update(x, y int, typeId string, speedTurnsLeft, abilityCooldown int) {
	pacman.positionHistory = append(pacman.positionHistory, pacman.position)
	pacman.position = Vector2d{x, y}
	pacman.typeId = typeId
	pacman.speedTurnsLeft = speedTurnsLeft
	pacman.abilityCooldown = abilityCooldown
}

func (pacman *Pacman) LastPosition() (bool, Vector2d) {
	if len(pacman.positionHistory) == 0 {
		return false, Vector2d{}
	}
	return true, pacman.positionHistory[len(pacman.positionHistory)-1]
}

func (pacman *Pacman) Direction() Vector2d {
	ok, lastPosition := pacman.LastPosition()
	if !ok {
		return Vector2d{0, 0}
	}
	return pacman.position.Subtract(&lastPosition)
}

func (pacman *Pacman) Move(x, y int) {
	pacman.Command(fmt.Sprintf("MOVE %d %d %d", pacman.pacId, x, y))
}

func (pacman *Pacman) Command(command string) {
	debug := fmt.Sprintf("%d %s", pacman.pacId, pacman.typeId)
	fmt.Printf("%s %s|", command, debug)
}

func (pacman *Pacman) String() string {
	direction := pacman.Direction()
	return fmt.Sprintf("id: %d, position: %s, direction: %s, type: %s, cooldown: %d, speedTurnsLeft: %d", pacman.pacId, pacman.position.String(), direction.String(), pacman.typeId, pacman.abilityCooldown, pacman.speedTurnsLeft)
}

type PacmanList struct {
	visiblePacmanCount int
	totalCount         int
	allies             map[int]*Pacman
	enemies            map[int]*Pacman
}

func NewPacmanList() *PacmanList {
	pacmanList := PacmanList{}
	pacmanList.totalCount = 0
	pacmanList.visiblePacmanCount = 0
	pacmanList.enemies = make(map[int]*Pacman, 5)
	pacmanList.allies = make(map[int]*Pacman, 5)
	return &pacmanList
}

func (pacmanList *PacmanList) String() string {
	var description string
	description += fmt.Sprintf("visiblePacmanCount: %d\n", pacmanList.visiblePacmanCount)
	description += fmt.Sprintf("totalCount: %d\n", pacmanList.totalCount)
	description += fmt.Sprintf("allies :\n")
	alliesId := make([]int, 0, len(pacmanList.allies))
	for key := range pacmanList.allies {
		alliesId = append(alliesId, key)
	}
	sort.Ints(alliesId)
	for _, pacId := range alliesId {
		description += fmt.Sprintf("\t- %s\n", pacmanList.allies[pacId])
	}
	description += fmt.Sprintf("enemies :\n")
	enemiesId := make([]int, 0, len(pacmanList.enemies))
	for key := range pacmanList.enemies {
		enemiesId = append(enemiesId, key)
	}
	sort.Ints(enemiesId)
	for _, pacId := range enemiesId {
		description += fmt.Sprintf("\t- %s\n", pacmanList.enemies[pacId])
	}
	return description
}

func (pacmanList *PacmanList) UpdateFromInput(scanner *bufio.Scanner, gameMap *GameMap) {
	var visiblePacCount int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &visiblePacCount)
	pacmanList.visiblePacmanCount = visiblePacCount

	isAlivePacAllies := make(map[int]bool, 5)
	for i := 0; i < visiblePacCount; i++ {
		var pacId int
		var mine int
		var x, y int
		var typeId string
		var speedTurnsLeft, abilityCooldown int
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &pacId, &mine, &x, &y, &typeId, &speedTurnsLeft, &abilityCooldown)
		if mine == 1 {
			isAlivePacAllies[pacId] = true
		}
		if mine == 1 && pacmanList.allies[pacId] != nil {
			pacmanList.allies[pacId].Update(x, y, typeId, speedTurnsLeft, abilityCooldown)
		} else if mine == 0 && pacmanList.enemies[pacId] != nil {
			pacmanList.enemies[pacId].Update(x, y, typeId, speedTurnsLeft, abilityCooldown)
		} else {
			pacmanList.totalCount += 1
			if mine == 1 {
				pacmanList.allies[pacId] = NewPacman(pacId, x, y, typeId)
				pacmanList.enemies[pacId] = NewPacman(pacId, gameMap.width-1-x, y, typeId)
			}
		}
	}
	for i := range pacmanList.allies {
		if !isAlivePacAllies[i] {
			delete(pacmanList.allies, i)
		}
	}
}

func (pacmanList *PacmanList) GetPacman(pacId int) *Pacman {
	if pacmanList.enemies[pacId] != nil {
		return pacmanList.enemies[pacId]
	} else if pacmanList.allies[pacId] != nil {
		return pacmanList.allies[pacId]
	}
	return nil
}

type MapItem int8

const (
	FloorItem MapItem = iota + 1
	WallItem
)

type GameMap struct {
	width, height int
	groundMap     [][]MapItem
	pelletMap     [][]int
	alliesIdMap   [][]int
	enemiesIdMap  [][]int
}

func NewGameMapFromInput(scanner *bufio.Scanner) *GameMap {
	gameMap := GameMap{}
	// width: size of the grid
	// height: top left corner is (x=0, y=0)
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &gameMap.width, &gameMap.height)
	gameMap.groundMap = make([][]MapItem, gameMap.width)
	for i := range gameMap.groundMap {
		gameMap.groundMap[i] = make([]MapItem, gameMap.height)
	}

	gameMap.pelletMap = make([][]int, gameMap.width)
	for i := range gameMap.pelletMap {
		gameMap.pelletMap[i] = make([]int, gameMap.height)
	}

	gameMap.alliesIdMap = make([][]int, gameMap.width)
	for i := range gameMap.alliesIdMap {
		gameMap.alliesIdMap[i] = make([]int, gameMap.height)
	}

	gameMap.enemiesIdMap = make([][]int, gameMap.width)
	for i := range gameMap.enemiesIdMap {
		gameMap.enemiesIdMap[i] = make([]int, gameMap.height)
	}

	for y := 0; y < gameMap.height; y++ {
		scanner.Scan()
		for x, c := range scanner.Text() {
			if c == '#' {
				gameMap.groundMap[x][y] = WallItem
			} else {
				gameMap.groundMap[x][y] = FloorItem
			}
		}
	}
	return &gameMap
}

func (gameMap *GameMap) UpdatePacmansPositions(pacmanList *PacmanList) {
	for x := 0; x < gameMap.width; x++ {
		for y := 0; y < gameMap.height; y++ {
			gameMap.enemiesIdMap[x][y] = -1
			gameMap.alliesIdMap[x][y] = -1
		}
	}
	for pacId, pacman := range pacmanList.allies {
		gameMap.alliesIdMap[pacman.position.x][pacman.position.y] = pacId
	}
	for pacId, pacman := range pacmanList.enemies {
		gameMap.enemiesIdMap[pacman.position.x][pacman.position.y] = pacId
	}
}

func (gameMap *GameMap) UpdatePalletsFromInput(scanner *bufio.Scanner) {
	for x := 0; x < gameMap.width; x++ {
		for y := 0; y < gameMap.height; y++ {
			gameMap.pelletMap[x][y] = 0
		}
	}
	// visiblePelletCount: all pellets in sight
	var visiblePelletCount int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &visiblePelletCount)

	for i := 0; i < visiblePelletCount; i++ {
		// value: amount of points this pellet is worth
		var x, y, value int
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &x, &y, &value)
		gameMap.pelletMap[x][y] = value
	}
}

func (gameMap *GameMap) MapBorderPositionCorrection(v *Vector2d) {
	if v.x == -1 {
		v.x = gameMap.width - 1
	} else if v.x == gameMap.width {
		v.x = 0
	} else if v.y == -1 {
		v.y = gameMap.height - 1
	} else if nextPosition.y == gameMap.height {
		v.y = 0
	}
}

func (gameMap *GameMap) ComputePossibleDirectionsFromPosition(position *Vector2d) []Vector2d {
	noWallDirection := make([]Vector2d, 0, 4)
	for _, direction := range ALL_DIRECTIONS {
		nextPosition := position.Add(&direction)
		if gameMap.groundMap[nextPosition.x][nextPosition.y] != WallItem {
			noWallDirection = append(noWallDirection, direction)
		}
	}
	return noWallDirection
}

func (gameMap *GameMap) ComputePositionScore(position Vector2d) float64 {
	return float64(gameMap.pelletMap[position.x][position.y])
}

func (gameMap *GameMap) ComputeDirectionScore(startPosition *Vector2d, direction *Vector2d, depth int, maxDepth int) float64 {
	if depth == maxDepth {
		return 0
	}
	nextPosition := startPosition.Add(direction)
	gameMap.MapBorderPositionCorrection(&nextPosition)
	score := gameMap.ComputePositionScore(nextPosition) * float64(maxDepth-depth) / float64(maxDepth)
	for _, newDirection := range gameMap.ComputePossibleDirectionsFromPosition(&nextPosition) {
		score += gameMap.ComputeDirectionScore(&nextPosition, &newDirection, depth+1, maxDepth)
	}
	return score
}

var logger = log.New(os.Stderr, "", 0)

func main() {
	rand.Seed(0)
	pacmanList := NewPacmanList()
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1000000), 1000000)

	gameMap := NewGameMapFromInput(scanner)
	tourCount := 0
	for {
		start := time.Now()
		var myScore, opponentScore int
		scanner.Scan()
		fmt.Sscan(scanner.Text(), &myScore, &opponentScore)
		// visiblePacCount: all your pacs and enemy pacs in sight
		pacmanList.UpdateFromInput(scanner, gameMap)
		gameMap.UpdatePacmansPositions(pacmanList)
		gameMap.UpdatePalletsFromInput(scanner)
		if tourCount == 0 {
			pacman := pacmanList.allies[0]
			for _, direction := range gameMap.ComputePossibleDirectionsFromPosition(&pacman.position) {
				logger.Printf("%v : %v", direction.Direction(), gameMap.ComputeDirectionScore(&pacman.position, &direction, 0, 10))
			}
		}
		for _, pacman := range pacmanList.allies {
			pacman.Move(rand.Intn(gameMap.width), rand.Intn(gameMap.height))
		}
		logger.Println(pacmanList)
		fmt.Println()
		tourCount += 1
		elapsed := time.Since(start)
		log.Printf("Round took %s", elapsed)
	}
}
