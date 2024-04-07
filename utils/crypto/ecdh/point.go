package ecdh

import "math/big"

type EllipticPoint struct {
	x, y *big.Int
}

func newEllipticPoint(x, y *big.Int) *EllipticPoint {
	return &EllipticPoint{x: x, y: y}
}

func (p *EllipticPoint) Equals(other *EllipticPoint) bool {
	// return p.x == other.x && p.y == other.y
	return (p.x.Cmp(other.x) == 0) && (p.y.Cmp(other.y) == 0)
}

func (p *EllipticPoint) Negate() *EllipticPoint {
	// return &EllipticPoint{-p.x, -p.y}
	return &EllipticPoint{
		x: new(big.Int).Neg(p.x),
		y: new(big.Int).Neg(p.y),
	}
}

func (p *EllipticPoint) IsDefault() bool {
	// return p.x == 0 && p.y == 0
	return (p.x.Cmp(big.NewInt(0)) == 0) && (p.y.Cmp(big.NewInt(0)) == 0)
}
