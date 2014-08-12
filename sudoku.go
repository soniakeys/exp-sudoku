// Copyright 2014 Sonia Keys
// License MIT: http://www.opensource.org/licenses/MIT

package main

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// board representation.
// main keeps master copy plus a working copy for each solver.
type board struct {
	candidates    []string // use arrays here?
	fixed         []bool
	nFixed        int
	revision      int
	contradiction bool
}

func newBoard() *board {
	return &board{
		candidates: make([]string, 81),
		fixed:      make([]bool, 81),
	}
}

const (
	nakedSingleSolverId = iota
	hiddenSingleSolverId
	lockedSolverId
	nakedPairSolverId
	xWingSolverId
	nSolvers
)

type solver struct {
	strategy

	// goroutine ownership indication.  nil means the solver has ownership at
	// the moment.  non-nil means main has it, with the revsion number
	// indicating the last revision evaluated by the solver.  main will send
	// it back to the solver immediately if there is a new revision worth
	// working on.  otherwise the solver stays on the bench.  it stays
	// until some new revision is available.
	bench *board

	// a channel to each solver
	boardCh chan *board

	fixesFound int
}

type strategy interface {
	//	id() int
	fix()
	solve(boardCh chan *board, fixCh chan interface{}, doneCh chan *doneCS)
}

/* nix package vars
var (
	bMain       *board // allocated in main
	solversBusy int   // incremented in sendBoard(), decremented on <-doneCh
)
*/

func main() {
	bMain := newBoard()
	sum := 0

	// start solvers
	startSolvers()

	// read puzzle
	b, err := ioutil.ReadFile("sudoku.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	lines := strings.Split(string(b), "\n")
	for line := 0; line < len(lines); line += 10 {
		boardName := strings.TrimSpace(lines[line])
		givens := ""
		for row := line + 1; row <= line+9; row++ {
			givens += strings.TrimSpace(lines[row])
		}
		bMain.nFixed = 0
		bMain.contradiction = false
		bMain.revision = 1
		var nGivens int
		for i := 0; i < 81; i++ {
			g := givens[i]
			if g == '0' {
				bMain.candidates[i] = "123456789"
			} else {
				bMain.candidates[i] = string(g)
				nGivens++
			}
			bMain.fixed[i] = false
		}

		fmt.Printf("%s.  %d givens.\n", boardName, nGivens)
		printBoard(bMain, false)

		solve()

		var heading string
		switch {
		case bMain.contradiction:
			heading = "contradiction encountered solving " + boardName
			nContradictions++
		case bMain.nFixed == 81:
			heading = boardName + " solved!"
			nSolved++
		default:
			var sFixed string
			if bMain.nFixed == 0 {
				sFixed = "none"
			} else {
				sFixed = fmt.Sprintf("only %d", bMain.nFixed)
			}
			heading = fmt.Sprintf("%s not solved.  %s fixed.", boardName, sFixed)
			nFailed++
		}
		fmt.Println(heading)
		printBoard(bMain, true)

		if bMain.nFixed == 81 && !bMain.contradiction {
			sum += int(bMain.candidates[0][0]-'0') * 100
			sum += int(bMain.candidates[1][0]-'0') * 10
			sum += int(bMain.candidates[2][0] - '0')
		}
	}

	fmt.Println("\nStatistics:")
	for i := 0; i < nSolvers; i++ {
		fmt.Println("solver", i, "fixes", fixesFound[i])
	}
	fmt.Println("")
	if nFailed == 0 && nContradictions == 0 {
		fmt.Println("all solved!")
		fmt.Println("sum: ", sum)
	} else {
		fmt.Println("Solved", nSolved)
		fmt.Println("Failed", nFailed)
		fmt.Println("Contradictions", nContradictions)
	}
}

// channel structs -- structs sent over channels

type doneCS struct {
	// solver id, so all solvers can send to the same channel and main
	// can still figure out where it came from
	solverId int

	// number of updates messages sent evaluating this revision
	nUpdates int

	// board ownership returned to main goroutine
	board *board
}

// more package variables, setup by startSolvers() and used by solve()
var (
	// common channels back to main
	doneCh chan *doneCS
	fixCh  chan interface{}
)

// gathered statistics
var (
	nSolved, nContradictions, nFailed int
)

func startSolvers() {
	for i := 0; i < nSolvers; i++ {
		bench[i] = newBoard()
		boardCh[i] = make(chan *board)
	}

	// all solvers send to this common notification channel.
	// solvers should send blocking, unbuffered.
	// solvers should then loop and wait for new board.
	doneCh = make(chan *doneCS)
	fixCh = make(chan interface{}, nSolvers)

	go solveNakedSingles(boardCh[nakedSingleSolverId], fixCh, doneCh)
	go solveHiddenSingles(boardCh[hiddenSingleSolverId], fixCh, doneCh)
	go solveLocked(boardCh[lockedSolverId], fixCh, doneCh)
	go solveNakedPairs(boardCh[nakedPairSolverId], fixCh, doneCh)
	go solveXWings(boardCh[xWingSolverId], fixCh, doneCh)
}

func sendBoard(solverId int) {
	if bMain.nFixed == 81 || bMain.contradiction {
		return
	}
	var b *board
	b, bench[solverId] = bench[solverId], nil
	b.revision = bMain.revision
	b.nFixed = bMain.nFixed
	copy(b.fixed, bMain.fixed)
	copy(b.candidates, bMain.candidates)
	solversBusy++
	boardCh[solverId] <- b
}

func wakeBenched() {
	for i, benched := range bench {
		if benched != nil {
			if benched.revision < bMain.revision {
				sendBoard(i)
			}
		}
	}
}

func solve() {
	var (
		updatesSent     [nSolvers]int
		updatesReceived [nSolvers]int
	)
	// give initial board to all solvers
	for i := 0; i < nSolvers; i++ {
		sendBoard(i)
	}
wait:
	for {
		select {
		case s := <-fixCh:
			s.fix()
			updatesReceived[s.id()]++
			wakeBenched()
		case done := <-doneCh:
			solversBusy--
			solver := done.solverId
			bench[solver] = done.board
			if done.nUpdates > 0 {
				//fmt.Printf("solver %d sent %d updates.\n",solver,done.nUpdates)
				// this is the local value to check that all updates are in
				updatesSent[solver] += done.nUpdates
				// this is a package value accumulated as a statistic
				fixesFound[solver] += done.nUpdates

				if bMain.revision > done.board.revision {
					sendBoard(solver)
				}
			}
		}

		if solversBusy > 0 {
			continue
		}
		for i := 0; i < nSolvers; i++ {
			if updatesReceived[i] != updatesSent[i] {
				continue wait
			}
		}
		// safe to return at this point
		if bMain.nFixed == 81 {
			return // success
		}
		if bMain.contradiction {
			return // complaint
		}
		for i := 0; i < nSolvers; i++ {
			if bench[i].revision < bMain.revision {
				// but wait, look again
				sendBoard(i)
				continue wait
			}
		}
		return // failure
	}
}

func printBoard(b *board, printCandidates bool) {
	var max int
	if printCandidates == false {
		max = 1
	} else {
		for _, s := range b.candidates {
			if len(s) > max {
				max = len(s)
			}
		}
	}
	cellFmt := fmt.Sprintf("%%%ds ", max)
	boxLine := strings.Repeat("-", 1+(max+1)*3)
	hLine := "." + strings.Repeat(boxLine+".", 3)
	fmt.Println(hLine)
	for floor, row, cell := 0, 0, 0; floor < 3; floor++ {
		for fr := 0; fr < 3; fr, row = fr+1, row+1 {
			for tower := 0; tower < 3; tower++ {
				fmt.Print("| ")
				for tc := 0; tc < 3; tc, cell = tc+1, cell+1 {
					c := board.candidates[cell]
					if printCandidates || len(c) == 1 {
						fmt.Printf(cellFmt, c)
					} else {
						fmt.Print("  ")
					}
				}
			}
			fmt.Println("|")
		}
		fmt.Println(hLine)
	}
}

// eliminate digit from cell
func eliminate(cell int, digit byte) {
	c := bMain.candidates[cell]
	x := strings.Index(c, string(digit))
	if x < 0 {
		return
	}
	r := c[:x] + c[x+1:]
	bMain.candidates[cell] = r
}
