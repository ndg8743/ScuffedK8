// Package game implements the core Battleship game logic
package game

import (
	"fmt"
	"strings"
)

const (
	BoardSize = 10
)

// CellState represents the state of a cell on the board
type CellState int

const (
	Empty   CellState = iota // Empty water
	HasShip                  // Ship present (not hit)
	Hit                      // Ship hit
	Miss                     // Missed shot
)

// Coordinate represents a position on the board
type Coordinate struct {
	Row int `json:"row"` // 0-9
	Col int `json:"col"` // 0-9
}

// Ship represents a ship on the board
type Ship struct {
	Name      string       `json:"name"`
	Size      int          `json:"size"`
	Positions []Coordinate `json:"positions"`
	Hits      int          `json:"hits"`
}

// IsDestroyed checks if the ship is completely destroyed
func (s *Ship) IsDestroyed() bool {
	return s.Hits >= s.Size
}

// ShipType constants for standard Battleship ships
var (
	ShipCarrier    = ShipType{Name: "Carrier", Size: 5}
	ShipBattleship = ShipType{Name: "Battleship", Size: 4}
	ShipCruiser    = ShipType{Name: "Cruiser", Size: 3}
	ShipSubmarine  = ShipType{Name: "Submarine", Size: 3}
	ShipDestroyer  = ShipType{Name: "Destroyer", Size: 2}
)

// ShipType defines a type of ship
type ShipType struct {
	Name string
	Size int
}

// GetStandardShips returns the standard set of 5 ships for Battleship
func GetStandardShips() []ShipType {
	return []ShipType{
		ShipCarrier,
		ShipBattleship,
		ShipCruiser,
		ShipSubmarine,
		ShipDestroyer,
	}
}

// Board represents a player's game board
type Board struct {
	Grid  [BoardSize][BoardSize]CellState `json:"-"` // Don't expose internal grid in JSON
	Ships []Ship                          `json:"ships"`
}

// NewBoard creates a new empty board
func NewBoard() *Board {
	return &Board{
		Grid:  [BoardSize][BoardSize]CellState{},
		Ships: []Ship{},
	}
}

// IsValidCoordinate checks if a coordinate is within board bounds
func IsValidCoordinate(coord Coordinate) bool {
	return coord.Row >= 0 && coord.Row < BoardSize && coord.Col >= 0 && coord.Col < BoardSize
}

// PlaceShip attempts to place a ship on the board
// Returns error if placement is invalid
func (b *Board) PlaceShip(shipType ShipType, start Coordinate, horizontal bool) error {
	positions := []Coordinate{}

	// Calculate all positions the ship would occupy
	for i := 0; i < shipType.Size; i++ {
		var pos Coordinate
		if horizontal {
			pos = Coordinate{Row: start.Row, Col: start.Col + i}
		} else {
			pos = Coordinate{Row: start.Row + i, Col: start.Col}
		}

		// Check if position is valid
		if !IsValidCoordinate(pos) {
			return fmt.Errorf("ship placement out of bounds")
		}

		// Check if position is already occupied
		if b.Grid[pos.Row][pos.Col] == HasShip {
			return fmt.Errorf("position already occupied by another ship")
		}

		positions = append(positions, pos)
	}

	// Place the ship
	for _, pos := range positions {
		b.Grid[pos.Row][pos.Col] = HasShip
	}

	b.Ships = append(b.Ships, Ship{
		Name:      shipType.Name,
		Size:      shipType.Size,
		Positions: positions,
		Hits:      0,
	})

	return nil
}

// Attack processes an attack at the given coordinate
// Returns (hit, shipDestroyed, error)
func (b *Board) Attack(coord Coordinate) (bool, *Ship, error) {
	if !IsValidCoordinate(coord) {
		return false, nil, fmt.Errorf("invalid coordinate")
	}

	cell := b.Grid[coord.Row][coord.Col]

	// Already attacked this position
	if cell == Hit || cell == Miss {
		return false, nil, fmt.Errorf("position already attacked")
	}

	// Miss
	if cell == Empty {
		b.Grid[coord.Row][coord.Col] = Miss
		return false, nil, nil
	}

	// Hit
	b.Grid[coord.Row][coord.Col] = Hit

	// Find which ship was hit
	for i := range b.Ships {
		ship := &b.Ships[i]
		for _, pos := range ship.Positions {
			if pos.Row == coord.Row && pos.Col == coord.Col {
				ship.Hits++
				if ship.IsDestroyed() {
					return true, ship, nil
				}
				return true, nil, nil
			}
		}
	}

	return true, nil, nil
}

// AllShipsDestroyed checks if all ships on the board are destroyed
func (b *Board) AllShipsDestroyed() bool {
	for _, ship := range b.Ships {
		if !ship.IsDestroyed() {
			return false
		}
	}
	return len(b.Ships) > 0 // Return false if no ships placed yet
}

// GetCellState returns the state of a cell (for viewing own board)
func (b *Board) GetCellState(coord Coordinate) CellState {
	if !IsValidCoordinate(coord) {
		return Empty
	}
	return b.Grid[coord.Row][coord.Col]
}

// GetOpponentView returns a board view for the opponent (hides unhit ships)
func (b *Board) GetOpponentView() [BoardSize][BoardSize]CellState {
	view := [BoardSize][BoardSize]CellState{}
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			cell := b.Grid[i][j]
			// Hide ships that haven't been hit
			if cell == HasShip {
				view[i][j] = Empty
			} else {
				view[i][j] = cell
			}
		}
	}
	return view
}

// String returns a string representation of the board
func (b *Board) String() string {
	var sb strings.Builder
	sb.WriteString("  0 1 2 3 4 5 6 7 8 9\n")
	for i := 0; i < BoardSize; i++ {
		sb.WriteString(fmt.Sprintf("%d ", i))
		for j := 0; j < BoardSize; j++ {
			cell := b.Grid[i][j]
			switch cell {
			case Empty:
				sb.WriteString("~ ")
			case HasShip:
				sb.WriteString("S ")
			case Hit:
				sb.WriteString("X ")
			case Miss:
				sb.WriteString("O ")
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
