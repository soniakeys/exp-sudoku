// Copyright 2014 Sonia Keys
// License MIT: http://www.opensource.org/licenses/MIT

package main

import (
	"fmt"
)

//--------------------------------------------------------------------
// locked candidates solver/fixer

const (
	lcType1Row = iota
	lcType1Col
	lcType2Row
	lcType2Col
)

type lockedCS struct {
	lcType int
	line   int
	chute  int
	digit  byte
}

func (s *lockedCS) id() int {
	return lockedSolverId
}

func solveLocked(boardCh chan *board, fixCh chan interface{}, doneCh chan *doneCS) {
	_ = fmt.Println
	var done doneCS
	done.solverId = lockedSolverId
	for {
		type digitCount [9]int
		var boxCount [9]digitCount

		board := <-boardCh
		done.board = board
		done.nUpdates = 0

		// total all boxes first, since the totals are used for both
		// row interactions and column interactions
		for box := 0; box < 9; box++ {
			fr := box / 3 * 3 // fr = first row number on this floor
			for frStop := fr + 3; fr < frStop; fr++ {
				cell := fr*9 + box%3*3
				for cellStop := cell + 3; cell < cellStop; cell++ {
					for _, c := range board.candidates[cell] {
						boxCount[box][c-'1']++
					}
				}
			}
		}

		// row interactions
		for row := 0; row < 9; row++ {
			var rowCount digitCount
			var rowletCount [3]digitCount
			// total rows and rowlets
			for tower, col, cell := 0, 0, row*9; tower < 3; tower++ {
				for colStop := col + 3; col < colStop; col, cell = col+1, cell+1 {
					for _, c := range board.candidates[cell] {
						cx := c - '1'
						rowCount[cx]++
						rowletCount[tower][cx]++
					}
				}
			}
			// look for interactions
			for tower, box := 0, row/3*3; tower < 3; tower, box = tower+1, box+1 {
				for cx := 0; cx < 9; cx++ {
					rc := rowletCount[tower][cx]
					if rc > 1 {
						switch {
						case rowCount[cx] == rc && boxCount[box][cx] > rc:
							fixCh <- &lockedCS{lcType2Row, row, tower, byte(cx) + '1'}
							done.nUpdates++
						case boxCount[box][cx] == rc && rowCount[cx] > rc:
							fixCh <- &lockedCS{lcType1Row, row, tower, byte(cx) + '1'}
							done.nUpdates++
						}
					}
				}
			}
		}

		// column interactions
		for col := 0; col < 9; col++ {
			var colCount digitCount
			var colletCount [3]digitCount
			// total cols and collets
			for floor, row, cell := 0, 0, col; floor < 3; floor++ {
				for rowStop := row + 3; row < rowStop; row, cell = row+1, cell+9 {
					for _, c := range board.candidates[cell] {
						cx := c - '1'
						colCount[cx]++
						colletCount[floor][cx]++
					}
				}
			}
			// look for interactions
			for floor, box := 0, col/3; floor < 3; floor, box = floor+1, box+3 {
				for cx := 0; cx < 9; cx++ {
					cc := colletCount[floor][cx]
					if cc > 1 {
						switch {
						case colCount[cx] == cc && boxCount[box][cx] > cc:
							fixCh <- &lockedCS{lcType2Col, col, floor, byte(cx) + '1'}
							done.nUpdates++
						case boxCount[box][cx] == cc && colCount[cx] > cc:
							fixCh <- &lockedCS{lcType1Col, col, floor, byte(cx) + '1'}
							done.nUpdates++
						}
					}
				}
			}
		}

		doneCh <- &done
	}
}

func (lc *lockedCS) fix() {
	bMain.revision++
	switch lc.lcType {
	case lcType1Row:
		for col, cell := 0, lc.line*9; col < 9; col, cell = col+1, cell+1 {
			if col/3 != lc.chute {
				eliminate(cell, lc.digit)
			}
		}
	case lcType2Row:
		fr := lc.line / 3 * 3 // fr = first row of floor
		for frStop := fr + 3; fr < frStop; fr++ {
			if fr != lc.line {
				cell := fr*9 + lc.chute*3
				for cellStop := cell + 3; cell < cellStop; cell++ {
					eliminate(cell, lc.digit)
				}
			}
		}
	case lcType1Col:
		for row, cell := 0, lc.line; row < 9; row, cell = row+1, cell+9 {
			if row/3 != lc.chute {
				eliminate(cell, lc.digit)
			}
		}
	case lcType2Col:
		// lc.line is column
		// lc.chute is floor
		tc := lc.line / 3 * 3 //  = first column of tower
		for tcStop := tc + 3; tc < tcStop; tc++ {
			if tc != lc.line {
				cell := lc.chute*27 + tc
				for cellStop := cell + 27; cell < cellStop; cell += 9 {
					eliminate(cell, lc.digit)
				}
			}
		}
	}
}
