package ecdh

/*

import "math/big"

type ec struct {
	p, a, b  *big.Int
	g        *ep
	n, h     *big.Int
	size     *big.Int
	packSize *big.Int
}

func newEllipticCurve(P, A, B *big.Int, G *ep, N, H, Size, PackSize *big.Int) *ec {
	return &ec{p: P, a: A, b: B, g: G, n: N, h: H, size: Size, packSize: PackSize}
}

// 此函数验证没有问题
func (c *ec) checkOn(point *ep) bool {
	// (pow(point.y, 2) - pow(point.x, 3) - self._A * point.x - self._B) % self._P == 0
	// return (point.y*point.y-point.x*point.x*point.x-c.a*point.x-c.b)%c.p == 0
	// 计算 point.y*point.y-point.x*point.x*point.x-c.a*point.x-c.b
	left := new(big.Int).Mod(new(big.Int).Sub(new(big.Int).Exp(point.y, big.NewInt(2), nil),
		new(big.Int).Add(new(big.Int).Exp(point.x, big.NewInt(3), nil),
			new(big.Int).Add(new(big.Int).Mul(c.a, point.x),
				c.b,
			),
		),
	),
		c.p,
	)
	// left == 0
	return left.Cmp(big.NewInt(0)) == 0
}

func newP256Curve() *ec {
	// SetString方法接收纯16进制字符串，需要去掉0x前缀
	P256P, _ := new(big.Int).SetString("ffffffff00000001000000000000000000000000ffffffffffffffffffffffff", 16)
	P256A, _ := new(big.Int).SetString("ffffffff00000001000000000000000000000000fffffffffffffffffffffffc", 16)
	P256B, _ := new(big.Int).SetString("5ac635d8aa3a93e7b3ebbd55769886bc651d06b0cc53b0f63bce3c3e27d2604b", 16)
	P256Gx, _ := new(big.Int).SetString("6b17d1f2e12c4247f8bce6e563a440f277037d812deb33a0f4a13945d898c296", 16)
	P256Gy, _ := new(big.Int).SetString("4fe342e2fe1a7f9b8ee7eb4a7c0f9e162bce33576b315ececbb6406837bf51f5", 16)
	P256N, _ := new(big.Int).SetString("ffffffff00000000ffffffffffffffffbce6faada7179e84f3b9cac2fc632551", 16)

	return newEllipticCurve(
		P256P,
		P256A,
		P256B,
		&ep{
			P256Gx,
			P256Gy,
		},
		P256N,
		big.NewInt(1),
		big.NewInt(32),
		big.NewInt(16),
	)
}

func newS192Curve() *ec {
	// SetString方法接收纯16进制字符串，需要去掉0x前缀
	S192P, _ := new(big.Int).SetString("fffffffffffffffffffffffffffffffffffffffeffffee37", 16)
	S192Gx, _ := new(big.Int).SetString("db4ff10ec057e9ae26b07d0280b7f4341da5d1b1eae06c7d", 16)
	S192Gy, _ := new(big.Int).SetString("9b2f2f6d9c5628a7844163d015be86344082aa88d95e2f9d", 16)
	S192N, _ := new(big.Int).SetString("fffffffffffffffffffffffe26f2fc170f69466a74defd8d", 16)
	return newEllipticCurve(
		S192P,
		big.NewInt(0),
		big.NewInt(3),
		&ep{
			S192Gx,
			S192Gy,
		},
		S192N,
		big.NewInt(1),
		big.NewInt(24),
		big.NewInt(24),
	)
}

*/
