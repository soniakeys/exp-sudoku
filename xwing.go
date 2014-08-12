// Copyright 2014 Sonia Keys
// License MIT: http://www.opensource.org/licenses/MIT

package main

import (
	"fmt"
	"strings"
)

//--------------------------------------------------------------------
// x-wing solver/fixer

type xWingCS struct {
	digit byte
	cells []int
}

func (s *xWingCS) id() int {
	return xWingSolverId
}

func solveXWings(boardCh chan *board, fixCh chan interface{}, doneCh chan *doneCS) {
	_ = fmt.Println
	var done doneCS
	done.solverId = xWingSolverId
	for {
		type digitCount [9]int

		board := <-boardCh
		done.board = board
		done.nUpdates = 0

		// row type:  candidates for two digits confined to two columns
		// in each of two rows.
		var rowCount [9]digitCount
		for row := 0; row < 9; row++ {
			cell := row * 9
			for cellStop := cell + 9; cell < cellStop; cell++ {
				for _, c := range board.candidates[cell] {
					rowCount[row][c-'1']++
				}
			}
		}

		// look for two rows with count=2
		for r1 := 0; r1 < 8; r1++ {
			for r2 := r1 + 1; r2 < 9; r2++ {
			rdLoop:
				for digit := 0; digit < 9; digit++ {
					db := byte(digit) + '1'
					ds := string(db)
					if rowCount[r1][digit] == 2 && rowCount[r2][digit] == 2 {
						// it's a start--now go back and see if the digits
						//  were in the same columns.
						for col, cell1, cell2 := 0, r1*9, r2*9; col < 9; col, cell1, cell2 = col+1, cell1+1, cell2+1 {
							if strings.Index(board.candidates[cell1], ds) >= 0 &&
								strings.Index(board.candidates[cell2], ds) < 0 {
								continue rdLoop // nope, different columns
							}
						}
						// yes! found x-wing.  looking for eliminations
						var fix *xWingCS
						var nFix int
						// one more time...
						for col, cell1, cell2 := 0, r1*9, r2*9; col < 9; col, cell1, cell2 = col+1, cell1+1, cell2+1 {
							if strings.Index(board.candidates[cell1], ds) >= 0 {
								for cell := cell1 % 9; cell < 81; cell += 9 {
									if cell != cell1 && cell != cell2 &&
										strings.Index(board.candidates[cell], ds) >= 0 {
										// found elimination
										if fix == nil {
											fix = &xWingCS{db, make([]int, 14)}
											nFix = 0
										}
										fix.cells[nFix] = cell
										nFix++
									}
								}
							}
						}
						if nFix > 0 {
							fix.cells = fix.cells[:nFix]
							fixCh <- fix
							done.nUpdates++
						}
					}
				}
			}
		}

		// col type:  candidates for two digits confined to two rows
		// in each of two columns.
		var colCount [9]digitCount
		for col := 0; col < 9; col++ {
			for cell := col % 9; cell < 81; cell += 9 {
				for _, c := range board.candidates[cell] {
					colCount[col][c-'1']++
				}
			}
		}

		// look for two columns with count=2
		for c1 := 0; c1 < 8; c1++ {
			for c2 := c1 + 1; c2 < 9; c2++ {
			cdLoop:
				for digit := 0; digit < 9; digit++ {
					db := byte(digit) + '1'
					ds := string(db)
					if colCount[c1][digit] == 2 && colCount[c2][digit] == 2 {
						// it's a start--now go back and see if the digits
						//  were in the same rows.
						for cell1, cell2 := c1%9, c2%9; cell1 < 81; cell1, cell2 = cell1+9, cell2+9 {
							if strings.Index(board.candidates[cell1], ds) >= 0 &&
								strings.Index(board.candidates[cell2], ds) < 0 {
								continue cdLoop // nope, different rows
							}
						}
						// yes! found x-wing.  looking for eliminations
						var fix *xWingCS
						var nFix int
						// one more time...
						for cell1, cell2 := c1%9, c2%9; cell1 < 81; cell1, cell2 = cell1+9, cell2+9 {
							if strings.Index(board.candidates[cell1], ds) >= 0 {
								cell := cell1 / 9 * 9
								for cellStop := cell + 9; cell < cellStop; cell++ {
									if cell != cell1 && cell != cell2 &&
										strings.Index(board.candidates[cell], ds) >= 0 {
										// found elimination
										if fix == nil {
											fix = &xWingCS{db, make([]int, 14)}
											nFix = 0
										}
										fix.cells[nFix] = cell
										nFix++
									}
								}
							}
						}
						if nFix > 0 {
							fix.cells = fix.cells[:nFix]
							fixCh <- fix
							done.nUpdates++
						}
					}
				}
			}
		}
		doneCh <- &done
	}
}

// X wing fixer
//
// called by main goroutine (not solver) so access to main board
// stays serialized.
func (xw *xWingCS) fix() {
	bMain.revision++
	for _, cell := range xw.cells {
		eliminate(cell, xw.digit)
	}
}
