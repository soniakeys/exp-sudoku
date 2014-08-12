// Copyright 2014 Sonia Keys
// License MIT: http://www.opensource.org/licenses/MIT

package main

import (
	"fmt"
)

// type to send over channel and have fix method
type nakedCell int

/*
func (s nakedCell) id() int {
	return nakedSingleSolverId
}
*/
// naked single solver.  runs as goroutine.
//
// blindly looks at each cell on the board, fixing any unfixed naked singles.
func solveNakedSingles(boardCh chan *board, fixCh chan interface{}, doneCh chan *doneCS) {
	_ = fmt.Println

	var done doneCS
	//	done.solverId = nakedSingleSolverId
	for {
		board := <-boardCh
		done.nUpdates = 0
		for i := 0; i < 81; i++ {
			if !board.fixed[i] && len(board.candidates[i]) == 1 {
				fixCh <- nakedCell(i)
				done.nUpdates++
			}
		}
		done.board = board
		doneCh <- &done
	}
}

// naked single fixer
//
// called by main goroutine (not solver) so access to main board
// stays serialized.
func (fixed nakedCell) fix() {
	if bMain.fixed[fixed] {
		return
	}
	bMain.revision++
	bMain.fixed[fixed] = true
	bMain.nFixed++
	digit := bMain.candidates[fixed][0]

	// eliminate candidate from houses
	cell0 := int(fixed)

	// row
	row0 := cell0 / 9
	cell := row0 * 9
	for stop := cell + 9; cell < stop; cell++ {
		if cell != cell0 {
			eliminate(cell, digit)
		}
	}
	// column
	col0 := cell0 % 9
	cell = col0
	for ; cell < 81; cell += 9 {
		if cell != cell0 {
			eliminate(cell, digit)
		}
	}
	// box
	row := row0 / 3 * 3
	col := col0 / 3 * 3
	cell = row*9 + col
	rowStop := row + 3
	colStop := col + 3
	for ; row < rowStop; row, col, cell = row+1, col-3, cell+6 {
		for ; col < colStop; col, cell = col+1, cell+1 {
			if col != col0 && row != row0 {
				eliminate(cell, digit)
			}
		}
	}
}
