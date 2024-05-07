package ecdh

import (
	"crypto/md5"
	"crypto/rand"
	"errors"
	"math/big"
)

var (
	ErrPubKeyLenMismatch = errors.New("public key len mismatch")
	ErrInvalidPubKey     = errors.New("invalid public key")
	ErrECCheckFailed     = errors.New("ec check failed")
	ErrPointUnexist      = errors.New("points is not on the curve")
	ErrInverseUnexist    = errors.New("inverse does not exist")
)

type provider struct {
	curve  *ec
	secret *big.Int
	public *ep
}

func newProvider(curve *ec) (p *provider, err error) {
	p = &provider{
		curve:  curve,
		secret: big.NewInt(0),
		public: &ep{},
	}

	p.secret = p.createSecret()
	p.public, err = p.createPublic(p.secret)

	return
}

func (p *provider) keyExchange(bobPub []byte, hashed bool) ([]byte, error) {
	unpacked, err := p.unpackPublic(bobPub)
	if err != nil {
		return nil, err
	}
	shared, err := p.createShared(p.secret, unpacked)
	if err != nil {
		return nil, err
	}
	return p.packShared(shared, hashed), nil
}

func (p *provider) unpackPublic(pub []byte) (*ep, error) {
	length := uint64(len(pub))
	// if length != p.curve.size*2+1 && length != p.curve.size+1
	if length != p.curve.size.Uint64()*2+1 && length != p.curve.size.Uint64()+1 {
		return nil, ErrPubKeyLenMismatch
	}

	x := append(make([]byte, 1), pub[1:p.curve.size.Uint64()+1]...)

	if pub[0] == 0x04 {
		y := append(make([]byte, 1), pub[p.curve.size.Uint64()+1:p.curve.size.Uint64()*2+1]...)
		gx := new(big.Int).SetBytes(x)
		gy := new(big.Int).SetBytes(y)
		return &ep{
			x: gx,
			y: gy,
		}, nil
	}

	px := new(big.Int).SetBytes(x)
	// x3 := (px * px * px) % p.curve.p
	x3 := new(big.Int).Mod(new(big.Int).Exp(px, big.NewInt(3), nil), p.curve.p)
	// ax := px * p.curve.p
	ax := new(big.Int).Mul(px, p.curve.p)
	// right := (x3 + ax + p.curve.b) % p.curve.p
	right := new(big.Int).Mod(new(big.Int).Add(x3, new(big.Int).Add(ax, p.curve.p)), p.curve.p)

	// tmp := (p.curve.p + 1) >> 2
	tmp := new(big.Int).Rsh(new(big.Int).Add(p.curve.p, big.NewInt(1)), 2)
	// py := pow(right, tmp, p.curve.p)
	py := new(big.Int).Exp(right, tmp, p.curve.p)

	// if py%2 == 0
	if new(big.Int).Mod(py, big.NewInt(2)).Cmp(big.NewInt(0)) == 0 {
		tmp = p.curve.p
		// tmp -= py
		tmp.Sub(tmp, py)
		py = tmp
	}

	return &ep{
		x: px,
		y: py,
	}, nil
}

func (p *provider) packPublic(compress bool) (result []byte) {
	if compress {
		result = append(make([]byte, int(p.curve.size.Uint64())-len(p.public.x.Bytes())), p.public.x.Bytes()...)
		result = append(make([]byte, 1), result...)
		// result[0] = 0x02 if (((self._public.y % 2) == 0) ^ ((self._public.y > 0) < 0)) else 0x03
		// 乱七八糟的，实际上就是 (self._public.y % 2) == 0
		if new(big.Int).Mod(p.public.y, big.NewInt(2)).Cmp(big.NewInt(0)) == 0 {
			result[0] = 0x02
		} else {
			result[0] = 0x03
		}
		return result
	}
	x := append(make([]byte, int(p.curve.size.Uint64())-len(p.public.x.Bytes())), p.public.x.Bytes()...)
	y := append(make([]byte, int(p.curve.size.Uint64())-len(p.public.y.Bytes())), p.public.y.Bytes()...)

	result = append(append(make([]byte, 1), x...), y...)
	result[0] = 0x04
	return result
}

func (p *provider) packShared(shared *ep, hashed bool) (x []byte) {
	x = append(make([]byte, int(p.curve.size.Uint64())-len(shared.x.Bytes())), shared.x.Bytes()...)
	if hashed {
		hash := md5.Sum(x[0:p.curve.packSize.Uint64()])
		x = hash[:]
	}
	return x
}

func (p *provider) createPublic(sec *big.Int) (*ep, error) {
	return p.createShared(sec, p.curve.g)
}

func (p *provider) createSecret() *big.Int {
	result := big.NewInt(0)
	for result.Cmp(big.NewInt(1)) == -1 || result.Cmp(p.curve.n) != -1 {
		buffer := make([]byte, p.curve.size.Uint64()+1)
		_, _ = rand.Read(buffer)
		buffer[p.curve.size.Uint64()] = 0
		result = new(big.Int).SetBytes(reverseBytes(buffer))
	}
	return result
}

// TODO 上次看到这里
func (p *provider) createShared(sec *big.Int, pub *ep) (*ep, error) {
	// if sec % p.curve.n == 0 || pub.IsDefault():
	if new(big.Int).Mod(sec, p.curve.n).Cmp(big.NewInt(0)) == 0 || pub.IsDefault() {
		return newEllipticPoint(big.NewInt(0), big.NewInt(0)), nil
	}
	// if sec < 0:
	if sec.Cmp(big.NewInt(0)) == -1 {
		return p.createShared(new(big.Int).Neg(sec), pub.Negate())
	}

	if !p.curve.checkOn(pub) {
		return nil, ErrInvalidPubKey
	}

	pr := newEllipticPoint(big.NewInt(0), big.NewInt(0))
	pa := pub
	var err error
	for sec.Cmp(big.NewInt(0)) == 1 {
		// if (sec & 1) > 0
		if new(big.Int).And(sec, big.NewInt(1)).Cmp(big.NewInt(0)) == 1 {
			pr, err = pointAdd(p.curve, pr, pa)
			if err != nil {
				return nil, err
			}
		}
		pa, err = pointAdd(p.curve, pa, pa)
		if err != nil {
			return nil, err
		}
		// sec >>= 1
		sec = new(big.Int).Rsh(sec, 1)
	}

	if !p.curve.checkOn(pr) {
		return nil, ErrECCheckFailed
	}

	return pr, nil
}

func pointAdd(curve *ec, p1, p2 *ep) (*ep, error) {
	if p1.IsDefault() {
		return p2, nil
	}
	if p2.IsDefault() {
		return p1, nil
	}
	if !(curve.checkOn(p1) && curve.checkOn(p2)) {
		return nil, ErrPointUnexist
	}

	var m *big.Int
	if p1.x.Cmp(p2.x) == 0 {
		if p1.y.Cmp(p2.y) == 0 {
			inv, err := modInverse(new(big.Int).Lsh(p1.y, 1), curve.p)
			if err != nil {
				return nil, err
			}
			m = new(big.Int).Mul(new(big.Int).Add(new(big.Int).Mul(
				big.NewInt(3), new(big.Int).Exp(p1.x, big.NewInt(2), nil)), curve.a),
				inv,
			)
		} else {
			return newEllipticPoint(big.NewInt(0), big.NewInt(0)), nil
		}
	} else {
		inv, err := modInverse(new(big.Int).Sub(p1.x, p2.x), curve.p)
		if err != nil {
			return nil, err
		}
		m = new(big.Int).Mul(new(big.Int).Sub(p1.y, p2.y), inv)
	}

	// xr = _mod(m * m - p1.x - p2.x, curve.P)
	xr := mod(new(big.Int).Sub(new(big.Int).Exp(m, big.NewInt(2), nil), new(big.Int).Add(p1.x, p2.x)), curve.p)
	// yr = _mod(m * (p1.x - xr) - p1.y, curve.P)
	yr := mod(new(big.Int).Sub(new(big.Int).Mul(m, new(big.Int).Sub(p1.x, xr)), p1.y), curve.p)
	pr := newEllipticPoint(xr, yr)

	if !curve.checkOn(pr) {
		return nil, ErrPointUnexist
	}

	return pr, nil
}

func mod(a, b *big.Int) (result *big.Int) {
	result = new(big.Int).Mod(a, b)
	if result.Cmp(big.NewInt(0)) == -1 {
		result.Add(result, b)
	}
	return result
}

func modInverse(a, p *big.Int) (*big.Int, error) {
	if a.Cmp(big.NewInt(0)) == -1 {
		inv, err := modInverse(a.Neg(a), p)
		if err != nil {
			return nil, err
		}
		return new(big.Int).Sub(p, inv), nil
	}

	g := new(big.Int).GCD(nil, nil, a, p)
	if g.Cmp(big.NewInt(1)) != 0 {
		return nil, ErrInverseUnexist
	}

	return new(big.Int).Exp(a, new(big.Int).Sub(p, big.NewInt(2)), p), nil
}

func reverseBytes(bytes []byte) []byte {
	reversed := make([]byte, len(bytes))
	for i := range bytes {
		reversed[i] = bytes[len(bytes)-i-1]
	}
	return reversed
}
