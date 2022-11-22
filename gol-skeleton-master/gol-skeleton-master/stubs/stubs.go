package stubs

var TurnHandler = "BoardOperations.CalculateNextBoard"

var CellHandler = "BoardOperations.GetAliveCells"

// var PremiumReverseHandler = "SecretStringOperations.FastReverse"

type Response struct {
	World      [][]byte
	AliveCells int
	Turn       int
}

type Request struct {
	World     [][]byte
	WorkerNum int
	Threads   int
	Turns     int
}
