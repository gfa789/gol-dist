package gol

import (
	"flag"
	"fmt"
	"net/rpc"
	"strconv"

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
	fmt.Println("Calling")
	client.Call(stubs.TurnHandler, request, response)
	fmt.Println("Responded")
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {

	// TODO: Create a 2D slice to store the world.
	var world = make([][]byte, p.ImageHeight)
	for i := 0; i < p.ImageHeight; i++ {
		row := make([]byte, p.ImageWidth)
		world[i] = row
	}
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
	//creates an array of output channels for the workers to put their chunks in
	// chanoutarray := []chan [][]byte{}
	// for i := 0; i < p.Threads; i++ {
	// 	chanout := make(chan [][]byte)
	// 	chanoutarray = append(chanoutarray, chanout)
	// }
	if p.Threads == 1 {
		server := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
		flag.Parse()
		client, _ := rpc.Dial("tcp", *server)
		defer client.Close()
		fmt.Println("Making call")
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
