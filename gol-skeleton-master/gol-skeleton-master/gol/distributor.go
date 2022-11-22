package gol

import (
	"fmt"
	"log"
	"net/rpc"
	"strconv"
	"time"

	"uk.ac.bris.cs/gameoflife/stubs"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

// func worker(startY, endY, startX, endX int, in [][]byte, out chan<- [][]byte, p Params) {
// 	output := calculateNextState(p, in, startY, endY)
// 	out <- output
// }

func getAliveCells(world [][]byte, p Params) []util.Cell {
	aliveCells := []util.Cell{}
	for i := 0; i < p.ImageHeight; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			if world[i][j] == 255 {
				aliveCells = append(aliveCells, util.Cell{X: j, Y: i})
			}
		}
	}
	return aliveCells
}

func makeCall(client *rpc.Client, world [][]byte, workernum int, p Params) {
	request := stubs.Request{World: world, WorkerNum: workernum, Threads: p.Threads, Turns: p.Turns}
	response := new(stubs.Response)
	client.Call(stubs.TurnHandler, request, response)
	// util.VisualiseMatrix(response.World, p.ImageWidth, p.ImageHeight)
	for i := 0; i < p.ImageHeight; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			world[i][j] = response.World[i][j]
		}
	}
}

// func copyWorld(world [][]byte) [][]byte {
// 	worldCopy := makeWorld(len(world), len(world[0]))
// 	for i := range world {
// 		for j := range world[i] {
// 			worldCopy[i][j] = world[i][j]
// 		}
// 	}
// 	return worldCopy
// }

func makeWorld(height, width int) [][]byte {
	world := make([][]byte, height)
	for i := 0; i < height; i++ {
		row := make([]byte, width)
		world[i] = row
	}
	return world
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.
	world := makeWorld(p.ImageHeight, p.ImageWidth)
	turn := 0
	c.ioCommand <- ioInput
	filename := strconv.Itoa(p.ImageHeight) + "x" + strconv.Itoa(p.ImageWidth)
	c.ioFilename <- filename
	for i := 0; i < p.ImageHeight; i++ {
		for j := 0; j < p.ImageWidth; j++ {
			world[i][j] = <-c.ioInput
		}
	}
	// TODO: Execute all turns of the Game of Life.
	ticker := time.NewTicker(2 * time.Second)
	go func() {
		select {
		case <-ticker.C:
			server := "127.0.0.1:8030"
			client, err := rpc.Dial("tcp", server)
			if err != nil {
				log.Fatal("dialing:", err)
			}
			req := stubs.Request{World: world}
			res := new(stubs.Response)
			client.Call(stubs.CellHandler, req, res)
			c.events <- AliveCellsCount{}
		default:
		}
	}()
	if p.Threads == 1 {
		server := "127.0.0.1:8030"
		// flag.Parse()
		client, err := rpc.Dial("tcp", server)
		if err != nil {
			log.Fatal("dialing:", err)
		}
		defer client.Close()
		makeCall(client, world, 0, p)
		fmt.Println("Call made")
	}
	// TODO: Report the final state using FinalTurnCompleteEvent.
	// sends event to IO saying final turn is completed
	AliveCells := getAliveCells(world, p)
	c.events <- FinalTurnComplete{CompletedTurns: turn, Alive: AliveCells}
	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{turn, Quitting}
	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}
