package main

import (
	"flag"
	"log"
	"math"
	"math/rand"
	"net"
	"net/rpc"
	"time"

	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

var turn int
var world [][]byte

func mod(x int, y int) int {
	a := x % y
	if a < 0 {
		a += y
	}
	return a
}

func calcLiveNeighbours(world [][]byte, x int, y int) int {
	liveNeighbours := 0
	height := len(world)
	width := len(world[0])
	xp := mod(x+1, width)
	xm := mod(x-1, width)
	yp := mod(y+1, height)
	ym := mod(y-1, height)
	xs := []int{x, xp, xm}
	ys := []int{y, yp, ym}
	for _, xi := range xs {
		for _, yi := range ys {
			if (world[yi][xi] == 255) && (!((xi == x) && (yi == y))) {
				liveNeighbours++
			}
		}
	}

	return liveNeighbours
}

func calculateNextState(world [][]byte, starty, endy int) [][]byte {
	var change []util.Cell
	newWorld := make([][]byte, endy-starty)
	width := len(world[0])
	for i := range newWorld {
		newWorld[i] = make([]byte, len(world[0]))
	}
	for h := starty; h < endy; h++ {
		for w := 0; w < width; w++ {
			newWorld[h-starty][w] = world[h][w]
			ln := calcLiveNeighbours(world, w, h)
			if world[h][w] == 255 {
				if ln < 2 || ln > 3 {
					change = append(change, util.Cell{X: w, Y: h})
				}
			} else {
				if ln == 3 {
					change = append(change, util.Cell{X: w, Y: h})
				}
			}
		}
	}
	for i := 0; i < len(change); i++ {
		if world[change[i].Y][change[i].X] == 0 {
			newWorld[change[i].Y-starty][change[i].X] = 255
		} else if world[change[i].Y][change[i].X] == 255 {
			newWorld[change[i].Y-starty][change[i].X] = 0
		}
	}
	return newWorld
}

func getAliveCells(world [][]byte, width, height int) int {
	aliveCells := 0
	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			if world[i][j] == 255 {
				aliveCells++
			}
		}
	}
	return aliveCells
}

func copyWorld(world [][]byte) [][]byte {
	worldCopy := makeWorld(len(world), len(world[0]))
	for i := range world {
		for j := range world[i] {
			worldCopy[i][j] = world[i][j]
		}
	}
	return worldCopy
}

func makeWorld(height, width int) [][]byte {
	world := make([][]byte, height)
	for i := 0; i < height; i++ {
		row := make([]byte, width)
		world[i] = row
	}
	return world
}

type BoardOperations struct{}

func (s *BoardOperations) CalculateNextBoard(req stubs.Request, res *stubs.Response) (err error) {

	height := len(req.World)
	// width := len(req.World[0])
	starth := int(math.Ceil(float64(req.WorkerNum) * (float64(height) / float64(req.Threads))))
	endh := int(math.Ceil(float64(req.WorkerNum+1) * (float64(height) / float64(req.Threads))))
	newWorld := copyWorld(req.World)
	for i := 0; i < req.Turns; i++ {
		newWorld = calculateNextState(newWorld, starth, endh)
		turn++
		world = copyWorld(newWorld)
	}
	*res = stubs.Response{World: newWorld}
	return
}

func (s *BoardOperations) GetAliveCells(req stubs.Request, res *stubs.Response) (err error) {
	aliveCount := 0
	for i := range world {
		for j := range world[0] {
			if world[i][j] == 255 {
				aliveCount++
			}
		}
	}
	*res = stubs.Response{World: world, AliveCells: aliveCount, Turn: turn}
	return
}

func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&BoardOperations{})
	listener, err := net.Listen("tcp", ":"+*pAddr)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer listener.Close()
	rpc.Accept(listener)
}
