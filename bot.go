package main

import (
	"fmt"
	"log"
	"math"
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

func (v *Vector2d) Opposite() Vector2d {
	return Vector2d{-v.x, -v.y}
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
	notSeenSinceXRound              int
	useSpeedAbility                 bool
	debugText                       string
}

func NewPacman(pacId, x, y int, typeId string) *Pacman {
	pacman := Pacman{}
	pacman.pacId = pacId
	pacman.position = Vector2d{x, y}
	pacman.typeId = typeId
	pacman.speedTurnsLeft = 0
	pacman.abilityCooldown = 0
	pacman.positionHistory = make([]Vector2d, 0, 200)
	pacman.notSeenSinceXRound = 0
	pacman.useSpeedAbility = false
	return &pacman
}

type BattleResult int

const (
	BATTLE_WIN BattleResult = iota + 1
	BATTLE_DRAW
	BATTLE_LOOSE
)

func (pacman *Pacman) Beat(pacman2 *Pacman) BattleResult {
	if pacman.typeId == pacman2.typeId {
		return BATTLE_DRAW
	} else if pacman.typeId == "ROCK" && pacman2.typeId == "SCISSORS" {
		return BATTLE_WIN
	} else if pacman.typeId == "SCISSORS" && pacman2.typeId == "PAPER" {
		return BATTLE_WIN
	} else if pacman.typeId == "PAPER" && pacman2.typeId == "ROCK" {
		return BATTLE_WIN
	} else {
		return BATTLE_LOOSE
	}
}

func (pacman *Pacman) UpdatePosition(x, y int) {
	if pacman.position.x != x || pacman.position.y != y {
		pacman.positionHistory = append(pacman.positionHistory, pacman.position)
	}
	pacman.position = Vector2d{x, y}
}

func (pacman *Pacman) Update(x, y int, typeId string, speedTurnsLeft, abilityCooldown int) {
	pacman.UpdatePosition(x, y)
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

//func (pacman *Pacman) PastDirection(index int) Vector2d {
//
//}

func (pacman *Pacman) Move(x, y int) {
	pacman.Command(fmt.Sprintf("MOVE %d %d %d", pacman.pacId, x, y))
}

func (pacman *Pacman) SpeedUp() {
	pacman.Command(fmt.Sprintf("SPEED %v", pacman.pacId))
}

func (pacman *Pacman) Command(command string) {
	debug := fmt.Sprintf("%v", pacman.debugText)
	//debug := fmt.Sprintf("%d %s", pacman.pacId, pacman.typeId)
	fmt.Printf("%s %s|", command, debug)
}

func (pacman *Pacman) String() string {
	direction := pacman.Direction()
	return fmt.Sprintf("id: %d, position: %s, notSeenSinceXRound: %v, direction: %s, type: %s, cooldown: %d, speedTurnsLeft: %d", pacman.pacId, pacman.position.String(), pacman.notSeenSinceXRound, direction.String(), pacman.typeId, pacman.abilityCooldown, pacman.speedTurnsLeft)
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
	for _, pacman := range pacmanList.enemies {
		pacman.notSeenSinceXRound += 1
		if pacman.abilityCooldown > 0 {
			pacman.abilityCooldown--
		}
		if pacman.speedTurnsLeft > 0 {
			pacman.speedTurnsLeft--
		}
	}
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
			pacmanList.enemies[pacId].notSeenSinceXRound = 0
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
	roundCount    int
	groundMap     [][]MapItem
	pelletMap     [][]float64
	alliesIdMap   [][]int
	enemiesIdMap  [][]int
	pacmanList    *PacmanList
}

func (gameMap *GameMap) PlayRound(scanner *bufio.Scanner) {
	var myScore, opponentScore int
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &myScore, &opponentScore)
	// visiblePacCount: all your pacs and enemy pacs in sight
	gameMap.pacmanList.UpdateFromInput(scanner, gameMap)
	gameMap.UpdatePacmansPositions(gameMap.pacmanList)
	gameMap.UpdatePalletsFromInput(scanner)
	//logger.Println(gameMap.StringAlliesMap())
	//logger.Println(gameMap.StringPacmanMap(gameMap.enemiesIdMap))
	//logger.Println(gameMap.StringPalletMap())
	start := time.Now()
	for _, pacman := range gameMap.pacmanList.allies {
		gameMap.MovePacmanCleverly(pacman)
	}
	logger.Println(gameMap.pacmanList)
	fmt.Println()
	gameMap.roundCount += 1
	elapsed := time.Since(start)
	log.Printf("Round took %s", elapsed)
}

func (gameMap *GameMap) StringPalletMap() string {
	var description string
	for y := 0; y < gameMap.height; y++ {
		for x := 0; x < gameMap.width; x++ {
			if math.Trunc(gameMap.pelletMap[x][y]) == 10 {
				description += "O"
			} else if gameMap.pelletMap[x][y] > 0.5 {
				description += "."
			} else {
				description += " "
			}
		}
		description += "\n"
	}
	return description
}

func (gameMap *GameMap) StringPacmanMap(pacmanMap [][]int) string {
	var description string
	for y := 0; y < gameMap.height; y++ {
		for x := 0; x < gameMap.width; x++ {
			if pacmanMap[x][y] != -1 {
				description += "O"
			} else {
				description += " "
			}
		}
		description += "\n"
	}
	return description
}

func NewGameMapFromInput(scanner *bufio.Scanner) *GameMap {
	gameMap := GameMap{}
	gameMap.pacmanList = NewPacmanList()
	gameMap.roundCount = 0
	// width: size of the grid
	// height: top left corner is (x=0, y=0)
	scanner.Scan()
	fmt.Sscan(scanner.Text(), &gameMap.width, &gameMap.height)
	gameMap.groundMap = make([][]MapItem, gameMap.width)
	for i := range gameMap.groundMap {
		gameMap.groundMap[i] = make([]MapItem, gameMap.height)
	}

	gameMap.pelletMap = make([][]float64, gameMap.width)
	for i := range gameMap.pelletMap {
		gameMap.pelletMap[i] = make([]float64, gameMap.height)
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
		if pacman.typeId != "DEAD" {
			gameMap.enemiesIdMap[pacman.position.x][pacman.position.y] = pacId
		}
	}
}

func (gameMap *GameMap) UpdatePalletsFromInput(scanner *bufio.Scanner) {
	if gameMap.roundCount == 0 {
		for x := 0; x < gameMap.width; x++ {
			for y := 0; y < gameMap.height; y++ {
				if gameMap.groundMap[x][y] == FloorItem && gameMap.enemiesIdMap[x][y] == -1 && gameMap.alliesIdMap[x][y] == -1 {
					gameMap.pelletMap[x][y] = 1
				} else {
					gameMap.pelletMap[x][y] = 0
				}
			}
		}
	} else {
		for x := 0; x < gameMap.width; x++ {
			for y := 0; y < gameMap.height; y++ {
				if gameMap.pelletMap[x][y] > 0 {
					gameMap.pelletMap[x][y] = 0.99 * gameMap.pelletMap[x][y]
				}
			}
		}
	}

	for _, pacman := range gameMap.pacmanList.allies {
		for _, direction := range ALL_DIRECTIONS {
			currentItem := FloorItem
			position := pacman.position
			for currentItem != WallItem {
				gameMap.pelletMap[position.x][position.y] = 0
				position = direction.Add(&position)
				gameMap.MapBorderPositionCorrection(&position)
				currentItem = gameMap.groundMap[position.x][position.y]
			}
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
		gameMap.pelletMap[x][y] = float64(value)
	}
}

func (gameMap *GameMap) MapBorderPositionCorrection(v *Vector2d) {
	if v.x == -1 {
		v.x = gameMap.width - 1
	} else if v.x == gameMap.width {
		v.x = 0
	} else if v.y == -1 {
		v.y = gameMap.height - 1
	} else if v.y == gameMap.height {
		v.y = 0
	}
}

func (gameMap *GameMap) ComputePossibleDirectionsFromPosition(position *Vector2d) []Vector2d {
	noWallDirection := make([]Vector2d, 0, 4)
	for _, direction := range ALL_DIRECTIONS {
		nextPosition := position.Add(&direction)
		gameMap.MapBorderPositionCorrection(&nextPosition)
		if gameMap.groundMap[nextPosition.x][nextPosition.y] != WallItem {
			noWallDirection = append(noWallDirection, direction)
		}
	}
	return noWallDirection
}

func DecreaseScoreWithDepth(score, depthPenalty float64, depth, maxDepth int) float64 {
	return score * math.Max((float64(maxDepth+1)-math.Pow(float64(depth+1), depthPenalty))/float64(maxDepth+1), 0)
}

func (gameMap *GameMap) ComputePositionScore(position *Vector2d, pacman *Pacman, depth int, maxDepth int, positionHistory []*Vector2d, isDeadEnd bool) float64 {
	for i := len(positionHistory) - 1; i > 0; i-- {
		if *positionHistory[i] == *position {
			return 0
		}
	}
	if gameMap.alliesIdMap[position.x][position.y] != -1 && gameMap.alliesIdMap[position.x][position.y] != pacman.pacId {
		return DecreaseScoreWithDepth(-20, 2, depth, maxDepth) - 2
	} else if gameMap.enemiesIdMap[position.x][position.y] != -1 {
		if pacman.Beat(gameMap.pacmanList.enemies[gameMap.enemiesIdMap[position.x][position.y]]) == BATTLE_WIN {
			if gameMap.pacmanList.enemies[gameMap.enemiesIdMap[position.x][position.y]].abilityCooldown > 0 {
				if isDeadEnd && gameMap.pacmanList.enemies[gameMap.enemiesIdMap[position.x][position.y]].abilityCooldown > depth {
					pacman.debugText = "TRAP"
					return DecreaseScoreWithDepth(100, 3, depth, maxDepth)
				} else if pacman.speedTurnsLeft > 0 {
					return DecreaseScoreWithDepth(15, 3, depth, maxDepth)
				} else {
					return DecreaseScoreWithDepth(-1, 1, depth, maxDepth)
				}
			} else {
				return DecreaseScoreWithDepth(-20, 2, depth, maxDepth)
			}
		} else if pacman.Beat(gameMap.pacmanList.enemies[gameMap.enemiesIdMap[position.x][position.y]]) == BATTLE_LOOSE {
			return DecreaseScoreWithDepth(-10, 2, depth, maxDepth) * (1.0 / (1.0 + 3*float64(gameMap.pacmanList.enemies[gameMap.enemiesIdMap[position.x][position.y]].notSeenSinceXRound))) //* (1.0 + float64(gameMap.pacmanList.enemies[gameMap.enemiesIdMap[position.x][position.y]].speedTurnsLeft))
		} else if gameMap.pacmanList.enemies[gameMap.enemiesIdMap[position.x][position.y]].abilityCooldown == 0 {
			return DecreaseScoreWithDepth(-20, 2, depth, maxDepth)
		}
		return DecreaseScoreWithDepth(-20, 2, depth, maxDepth)
	}
	if gameMap.pelletMap[position.x][position.y] != 0 {
		return DecreaseScoreWithDepth(gameMap.pelletMap[position.x][position.y], 1.1, depth, maxDepth) + gameMap.pelletMap[position.x][position.y]/2
	}
	return 0
}

func (gameMap *GameMap) ComputeDirectionScore(pacman *Pacman, startPosition *Vector2d, direction *Vector2d, depth int, maxDepth int, positionHistory []*Vector2d) (float64, *Vector2d, bool) {
	if depth == maxDepth {
		return 0, nil, false
	}
	nextPosition := startPosition.Add(direction)
	gameMap.MapBorderPositionCorrection(&nextPosition)

	bestScore := 0.0
	meanScore := 0.0
	directionCount := 0.0
	var bestNextPosition *Vector2d
	positionHistory = append(positionHistory, &nextPosition)
	possibleDirections := gameMap.ComputePossibleDirectionsFromPosition(&nextPosition)
	possibleDirectionsCount := len(possibleDirections)
	isDeadEnd := false
	for _, newDirection := range possibleDirections {
		if direction.Opposite() == newDirection && possibleDirectionsCount != 1 {
			continue
		}
		var localScore float64
		var nextPosition2 *Vector2d
		localScore, nextPosition2, isDeadEnd = gameMap.ComputeDirectionScore(pacman, &nextPosition, &newDirection, depth+1, maxDepth, positionHistory)
		meanScore += localScore
		if localScore > bestScore {
			bestScore = localScore
			bestNextPosition = nextPosition2
		}
		directionCount += 1
	}
	if possibleDirectionsCount == 1 {
		isDeadEnd = true
	} else if possibleDirectionsCount > 2 {
		isDeadEnd = false
	}
	positionHistory = positionHistory[:len(positionHistory)-1]
	score := gameMap.ComputePositionScore(&nextPosition, pacman, depth, maxDepth, positionHistory, isDeadEnd)
	if depth == 1 {
		bestNextPosition = &nextPosition
	}
	if depth < 2 {
		return score + meanScore/directionCount, bestNextPosition, isDeadEnd
	}
	return score + bestScore, bestNextPosition, isDeadEnd
}

func (gameMap *GameMap) MovePacmanCleverly(pacman *Pacman) {
	pacman.debugText = fmt.Sprintf("%v", pacman.pacId)
	var bestDirection Vector2d
	bestScore := float64(math.MinInt32)
	var bestBestNextPosition *Vector2d
	logger.Printf("Pacman %v\n", pacman.pacId)
	for _, direction := range gameMap.ComputePossibleDirectionsFromPosition(&pacman.position) {
		maxDepth := 25
		score, bestNextPosition, _ := gameMap.ComputeDirectionScore(pacman, &pacman.position, &direction, 0, maxDepth, make([]*Vector2d, 0, maxDepth))
		pacmanDirection := pacman.Direction()
		if direction == pacmanDirection.Opposite() {
			score -= math.Abs(score * 0.2)
		}
		logger.Printf("Direction %v : %v\n", direction.Direction(), score)
		if score > bestScore {
			bestScore = score
			bestBestNextPosition = bestNextPosition
			bestDirection = direction
		}
		//logger.Printf("%v : %v", direction.Direction(), gameMap.ComputeDirectionScore(&pacman.position, &direction, 0, 10))
	}
	logger.Printf("Best next position %v", bestBestNextPosition)
	nextPosition := bestDirection.Add(&pacman.position)
	if bestBestNextPosition != nil && pacman.position != *bestBestNextPosition && pacman.speedTurnsLeft != 0 {
		pacman.UpdatePosition(nextPosition.x, nextPosition.y)
		nextPosition = *bestBestNextPosition
	}
	gameMap.MapBorderPositionCorrection(&nextPosition)
	if pacman.abilityCooldown == 0 {
		pacman.SpeedUp()
	} else {
		pacman.Move(nextPosition.x, nextPosition.y)
	}
}

var logger = log.New(os.Stderr, "", 0)

func main() {
	rand.Seed(0)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1000000), 1000000)

	gameMap := NewGameMapFromInput(scanner)
	for {
		gameMap.PlayRound(scanner)
	}
}
