// Copyright 2014 Sonia Keys
// License MIT: http://www.opensource.org/licenses/MIT

package main

import (
	"fmt"
)

type nakedPairCS struct {
	pair  string
	cells []int
}

func (s *nakedPairCS) id() int {
	return nakedPairSolverId
}

// naked pair solver.  runs as goroutine.
//
// blindly looks at each cell on the board, looking for number of candidates
// = 2.  when found, looks for the same pair in later in the houses, then
// calls for eliminations where found.
func solveNakedPairs(boardCh chan *board, fixCh chan interface{}, doneCh chan *doneCS) {
	_ = fmt.Println

	var done doneCS
	done.solverId = nakedPairSolverId
	var foundBuffer [21]int
	for {
		board := <-boardCh
		done.board = board
		done.nUpdates = 0
		for i := 0; i < 80; i++ {
			pair := board.candidates[i]
			if len(board.candidates[i]) != 2 {
				continue
			}
			nElim := 0

			// search across row
			cell := i + 1
			for cellStop := (i/9 + 1) * 9; cell < cellStop; cell++ {
				if board.candidates[cell] == pair {
					//fmt.Printf("np row candidate %s cells %d and %d\n",pair,i,cell)
					// store cells with possible eliminations
					for cp := cellStop - 9; cp < cellStop; cp++ {
						if cp != i && cp != cell && len(bMain.candidates[cp]) > 1 {
							foundBuffer[nElim] = cp
							nElim++
						}
					}
				}
			}
			// search down column
			for cell := i + 9; cell < 81; cell += 9 {
				if board.candidates[cell] == pair {
					//fmt.Printf("np col candidate %s cells %d and %d\n",pair,i,cell)
					for cp := i % 9; cp < 81; cp += 9 {
						if cp != i && cp != cell && len(bMain.candidates[cp]) > 1 {
							foundBuffer[nElim] = cp
							nElim++
						}
					}
				}
			}
			// search forward in block
			//fmt.Println("searching block forward from cell",i)
			rowStop := (i/27 + 1) * 3
			//fmt.Println("rowStop",rowStop)
			cell = i + 1
			if cell%3 == 0 {
				cell += 6
			}
			//fmt.Println("first cell to consider:",cell)
			for row := cell / 9; row < rowStop; row++ {
				for cellStop := (cell%3 + 1) * 3; cell < cellStop; cell++ {
					if board.candidates[cell] == pair {
						//fmt.Printf("np box candidate %s cells %d and %d\n",pair,i,cell)
						for row := i % 27 * 3; row < rowStop; row++ {
							cp := row*9 + i%9/3*3
							for cpStop := cp + 3; cp < cpStop; cp++ {
								if cp != i && cp != cell && len(bMain.candidates[cp]) > 1 {
									foundBuffer[nElim] = cp
									nElim++
								}
							}
						}
					}
				}
			}

			// check that potential eliminations would really eliminate
			// candidates and that they are unique.
			if nElim > 0 {
				m := make(map[int]int)
				c1 := rune(pair[0])
				c2 := rune(pair[1])
				for i := 0; i < nElim; i++ {
					cell := foundBuffer[i]
					for _, c := range board.candidates[cell] {
						if c == c1 || c == c2 {
							m[cell] = 0
						}
					}
				}
				if len(m) > 0 {
					s := make([]int, len(m))
					i := 0
					for cell := range m {
						s[i] = cell
						i++
					}
					//fmt.Println("found naked pair", pair)
					//fmt.Println("eliminations in cells", s)
					//printBoard(board, true)
					fixCh <- &nakedPairCS{pair, s}
					done.nUpdates++
				}
			}
		}
		doneCh <- &done
	}
}

// naked pair fixer
//
// called by main goroutine (not solver) so access to main board
// stays serialized.
func (np *nakedPairCS) fix() {
	bMain.revision++
	for _, cell := range np.cells {
		eliminate(cell, np.pair[0])
		eliminate(cell, np.pair[1])
	}
}
