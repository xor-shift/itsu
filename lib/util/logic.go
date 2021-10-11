package util

func MatCond(p, q bool) bool {
	return !p || q
}

type Relation uint8

var (
	RelFalse = Relation(0b0000) //false
	RelAnd   = Relation(0b0001) //p && q

	RelP = Relation(0b0011) //p

	RelQ   = Relation(0b0101) //q
	RelXor = Relation(0b0110) //p != q

	RelBicond    = Relation(0b1001) //p == q
	RelNotQ      = Relation(0b1010) //!q
	RelConvImply = Relation(0b1011) //p <= q
	RelNotP      = Relation(0b1100) //!p
	RelImply     = Relation(0b1101) //p -> q
	RelOr        = Relation(0b1110) //p || q
	RelTrue      = Relation(0b1111) //true

	RelNEQ = RelXor
	RelEQ  = RelBicond
)

func TTableEval(p, q bool, rel Relation) bool {
	idx := uint8(0)

	if p {
		idx |= 0b01
	}
	if q {
		idx |= 0b10
	}

	idx = 3 - idx

	return (rel>>idx)&1 == 1
}
