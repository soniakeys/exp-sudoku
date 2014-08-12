# Experimental sudoku program

I wrote this when I was just learning Go and exploring concurrency.
It's a solver that implements "human" solving techniques.  Each technique
runs as a goroutine that scans the puzzle and reports findings to a central
goroutine that maintains the board.  The program runs until the puzzle is
solved or until none of the solver goroutines see any more moves.

The program can be made smarter just by implementing new solving techniques.

It was a fun exercise but in it's current state is far short of an interesting
or useful program for sudoku enthusiasts.
