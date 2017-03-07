package main

import (
	"errors"
	"fmt"
)

// OutOfRangeError tells if the given position is out of range.
type OutOfRangeError error

// N is the amount of cells per row and amount of columns.
const N int = 6

// Cell contains the state of one board cell.
type Cell int

const (
	// Empty indicates that the cell is empty.
	Empty Cell = -1
	// OO is an O.
	OO Cell = 0
	// XX is an X.
	XX Cell = 1
)

// GameRole indicates the type the player is
type GameRole int

const (
	// Order gamerole
	Order GameRole = 0
	// Chaos gamerole
	Chaos GameRole = 1
)

func (role GameRole) String() string {
	if role == Order {
		return "order"
	}
	return "chaos"
}

// Board contains the state of the game board
type Board struct {
	Cells  [N * N]Cell
	Onturn GameRole
}

// MakeBoard creates a new board
func MakeBoard(startPlayer GameRole) *Board {
	bd := &Board{
		Onturn: startPlayer,
	}

	for i := 0; i < N*N; i++ {
		bd.Cells[i] = Empty
	}

	return bd
}

// ApplyMove applies the given move
func (bd *Board) ApplyMove(stone Cell, pos int) error {
	if pos < 0 || pos >= N*N {
		return OutOfRangeError(errors.New("Pos out of range in ApplyMove"))
	}
	if stone == Empty {
		return errors.New("Cannot set cell to empty in ApplyMove")
	}
	if bd.Cells[pos] != Empty {
		return errors.New("Target cell not empty in ApplyMove")
	}

	bd.Cells[pos] = stone
	if bd.Onturn == Order {
		bd.Onturn = Chaos
	} else {
		bd.Onturn = Order
	}

	return nil
}

// IsEmpty checks if the cell at the given position is empty.
func (bd *Board) IsEmpty(pos int) (bool, error) {
	if pos < 0 || pos >= N*N {
		return false, OutOfRangeError(errors.New("Pos out of range in IsEmpty"))
	}

	return bd.Cells[pos] == Empty, nil
}

// CheckWin checks if the game is over and if so who has won the game.
func (bd *Board) CheckWin() (winner GameRole, gameOver bool) {
	full := true
	for i := 0; i < N*N; i++ {
		if bd.Cells[i] == Empty {
			full = false
			break
		}
	}
	if full {
		return Chaos, true
	}

	var k int
	for i := 0; i < N; i++ {
		for j := 0; j < 2; j++ {
			stone := bd.Cells[N*i+j] // Horizontal
			if stone != Empty {
				for k = 1; k < N-1; k++ {
					if bd.Cells[N*i+j+k] != stone {
						break
					}
				}
				if k == N-1 {
					return Order, true
				}
			}

			stone = bd.Cells[N*j+i] // Vertical
			if stone != Empty {
				for k = 1; k < N-1; k++ {
					if bd.Cells[N*j+i+k] != stone {
						break
					}
				}
				if k == N-1 {
					return Order, true
				}
			}
		}
	}

	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			stone := bd.Cells[N*i+j] // Diagonal \
			if stone != Empty {
				for k = 1; k < N-1; k++ {
					if bd.Cells[N*(i+j)+j+k] != stone {
						break
					}
				}
				if k == N-1 {
					return Order, true
				}
			}

			stone = bd.Cells[N*i+N-1-j] // Diagonal /
			if stone != Empty {
				for k = 1; k < N-1; k++ {
					if bd.Cells[N*(i+j)+N-1-j-k] != stone {
						break
					}
				}
				if k == N-1 {
					return Order, true
				}
			}
		}
	}

	return GameRole(-1), false // Nobody
}

func (bd *Board) printBoard() {
	for y := 0; y < N; y++ {
		for x := 0; x < N; x++ {
			if bd.Cells[N*y+x] == OO {
				fmt.Print("O ")
			} else if bd.Cells[N*y+x] == XX {
				fmt.Print("X ")
			} else {
				fmt.Print(". ")
			}
		}
		fmt.Print("\n")
	}
}
