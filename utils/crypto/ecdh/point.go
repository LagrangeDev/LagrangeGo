package ecdh

import "math/big"

type ep struct {
	x, y *big.Int
}

func newEllipticPoint(x, y *big.Int) *ep {
	return &ep{x: x, y: y}
}

func (p *ep) Equals(other *ep) bool {
	// return p.x == other.x && p.y == other.y
	return (p.x.Cmp(other.x) == 0) && (p.y.Cmp(other.y) == 0)
}

func (p *ep) Negate() *ep {
	// return &EllipticPoint{-p.x, -p.y}
	return &ep{
		x: new(big.Int).Neg(p.x),
		y: new(big.Int).Neg(p.y),
	}
}

func (p *ep) IsDefault() bool {
	// return p.x == 0 && p.y == 0
	return (p.x.Cmp(big.NewInt(0)) == 0) && (p.y.Cmp(big.NewInt(0)) == 0)
}
