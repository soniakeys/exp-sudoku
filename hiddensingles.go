// Copyright 2014 Sonia Keys
// License MIT: http://www.opensource.org/licenses/MIT

package main

import (
	"fmt"
)

//--------------------------------------------------------------------
// hidden single solver/fixer

type hiddenSingleCS struct {
	// the cell to be fixed
	cell int
	// the digit that goes in the cell
	digit byte
}

func (s *hiddenSingleCS) id() int {
	return hiddenSingleSolverId
}

func solveHiddenSingles(boardCh chan *board, fixCh chan interface{}, doneCh chan *doneCS) {
	_ = fmt.Println

	var (
		done       doneCS
		digitCount [9]int
		digitCell  [9]int
	)
	done.solverId = hiddenSingleSolverId
	for {
		board := <-boardCh
		sent := make(map[int]int)

		// search rows
		for row, cell := 0, 0; row < 9; row++ {
			// clear accumulator
			for i := 0; i < 9; i++ {
				digitCount[i] = 0
			}
			for col := 0; col < 9; col, cell = col+1, cell+1 {
				for _, c := range board.candidates[cell] {
					cx := c - '1'
					digitCount[cx]++
					digitCell[cx] = cell
				}
			}
			for i, count := range digitCount {
				if count == 1 {
					// found a single
					hsCell := digitCell[i]
					if board.fixed[hsCell] {
						continue
					}
					hs := &hiddenSingleCS{hsCell, byte(i) + '1'}
					fixCh <- hs
					sent[hsCell] = 0
				}
			}
		}

		// search columns
		for col := 0; col < 9; col++ {
			// clear accumulator
			for i := 0; i < 9; i++ {
				digitCount[i] = 0
			}
			for row, cell := 0, col; row < 9; row, cell = row+1, cell+9 {
				for _, c := range board.candidates[cell] {
					cx := c - '1'
					digitCount[cx]++
					digitCell[cx] = cell
				}
			}
			for i, count := range digitCount {
				if count == 1 {
					hsCell := digitCell[i]
					if board.fixed[hsCell] {
						continue
					}
					hs := &hiddenSingleCS{hsCell, byte(i) + '1'}
					_, found := sent[hsCell]
					if found {
						continue
					}
					fixCh <- hs
					sent[hsCell] = 0
				}
			}
		}

		// search boxes
		for floor := 0; floor < 3; floor++ {
			for tower := 0; tower < 3; tower++ {
				// clear accumulator
				for i := 0; i < 9; i++ {
					digitCount[i] = 0
				}
				fr := floor * 3
				for frStop := fr + 3; fr < frStop; fr++ {
					cell := fr*9 + tower*3
					for cellStop := cell + 3; cell < cellStop; cell++ {
						for _, c := range board.candidates[cell] {
							cx := c - '1'
							digitCount[cx]++
							digitCell[cx] = cell
						}
					}
				}
				for i, count := range digitCount {
					if count == 1 {
						// found a single
						hsCell := digitCell[i]
						if board.fixed[hsCell] {
							continue
						}
						hs := &hiddenSingleCS{hsCell, byte(i) + '1'}
						_, found := sent[hsCell]
						if found {
							continue
						}
						fixCh <- hs
						sent[hsCell] = 0
					}
				}
			}
		}

		done.board = board
		done.nUpdates = len(sent)
		doneCh <- &done
	}
}

func (hs *hiddenSingleCS) fix() {
	if bMain.fixed[hs.cell] {
		return
	}
	bMain.candidates[hs.cell] = string(hs.digit)
	nakedCell(hs.cell).fix()
}
