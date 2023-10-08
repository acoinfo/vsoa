package position

import (
	"net"
)

type Position struct {
	Name     string `json:"name"`
	Domain   int    `json:"domain"`
	IP       string `json:"addr"`
	Port     int    `json:"port"`
	Security bool   `json:"security"`
}

type PositionList []Position

func (pl PositionList) Len() int { return len(pl) }

func NewPositionList() *PositionList {
	return &PositionList{}
}

func NewPosition(name string, domain int, ip string, port int, security bool) *Position {
	return &Position{
		Name:     name,
		Domain:   domain,
		IP:       ip,
		Port:     port,
		Security: security,
	}
}

// Add adds a Position to the PositionList.
// It updates the PositionList if the Position already exists.
// If Position.IP is not a valid IP address, it does nothing.
// It takes a Position as a parameter and does not return anything.
func (pl *PositionList) Add(p Position) {
	if *pl == nil {
		*pl = make([]Position, 0)
	}

	if net.ParseIP(p.IP) == nil {
		return
	}

	for i, op := range *pl {
		if op.Name == p.Name {
			(*pl)[i] = p
			return
		}
	}

	*pl = append(*pl, p)
}

// Remove removes an element from the PositionList based on the provided name.
//
// Parameters:
// - name: the name of the element to be removed.
func (pl *PositionList) Remove(name string) {
	for i, p := range *pl {
		if p.Name == name {
			(*pl)[i] = (*pl)[len(*pl)-1]
			*pl = (*pl)[:len(*pl)-1]
			return
		}
	}
}

func (pl PositionList) lookUp(name string) *Position {
	for _, p := range pl {
		if p.Name == name {
			return &p
		}
	}
	return nil
}
