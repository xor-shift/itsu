package vm

import (
	"errors"
	"strconv"
	"strings"
)

func CompileFORTH(builder *ProgramBuilder, str string) error {
	//builder := NewProgramBuilder()
	tokens := tokenizeString(str)

	singleByteTokens := map[string]byte{
		"0":     OpNCONST_0,
		"1":     OpNCONST_1,
		"2":     OpNCONST_2,
		"false": OpBCONST_0,
		"true":  OpBCONST_1,
		"F":     OpBCONST_0,
		"T":     OpBCONST_1,

		"NIL":   OpNILCONST,
		"ISNIL": OpISNIL,
		"KIND":  OpKIND,

		"DUP":  OpSDUP,
		"DROP": OpSDROP,
		"SWAP": OpSSWAP,
		"OVER": OpSOVER,
		"ROT":  OpSROT,

		"CMP":    OpNCMP,
		"STRCMP": OpSCMP,
		"LT":     OpLT,
		"LE":     OpLE,
		"EQ":     OpEQ,
		"GE":     OpGE,
		"GT":     OpGT,
		"NE":     OpNE,
		"<":      OpLT,
		"<=":     OpLE,
		"==":     OpEQ,
		">=":     OpGE,
		">":      OpGT,
		"!=":     OpNE,
		"AND":    OpLAND,
		"OR":     OpLOR,
		"XOR":    OpLXOR,
		"NOT":    OpLNOT,
		"&&":     OpLAND,
		"||":     OpLOR,
		"^^":     OpLXOR,
		"!":      OpLNOT,

		"ADD":  OpNADD,
		"SUB":  OpNSUB,
		"MUL":  OpNMUL,
		"DIV":  OpNDIV,
		"FMOD": OpNFMOD,
		"+":    OpNADD,
		"-":    OpNSUB,
		"*":    OpNMUL,
		"/":    OpNDIV,
		".%":   OpNFMOD,
		"POW":  OpNPOW,
		"^":    OpNPOW,
		"**":   OpNPOW,

		"SQRT":  OpNSQRT,
		"TRUNC": OpNTRUNC,
		"FLOOR": OpNFLOOR,
		"CEIL":  OpNCEIL,
		"SHL":   OpNSHL,
		"SHR":   OpNSHR,
		"<<":    OpNSHL,
		">>":    OpNSHR,

		"HLT": OpHLT,
	}

	parseTTBL := func(s string) (v uint8, valid bool) {
		if !strings.HasPrefix(s, "LTTBLB_") {
			return
		}

		ttbl := strings.TrimPrefix(s, "LTTBLB_")

		if len(ttbl) != 1 && len(ttbl) != 4 {
			return
		}

		bits := 2
		if len(ttbl) == 1 {
			bits = 16
		}

		if v, err := strconv.ParseUint(ttbl, bits, 8); err != nil {
			return 0, false
		} else {
			return uint8(v), true
		}
	}

	parsePrefixedTTBL := func(s string, prefix string) (v uint8, valid bool) {
		if !strings.HasPrefix(s, "LTTBLB_") {
			return
		}

		return parseTTBL(strings.TrimPrefix(s, "LTTBLB_"))
	}

	parseSplit := func(s string) (lhs, rhs string, valid bool) {
		if idx := strings.Index(s, "_"); idx == -1 {
			return
		} else {
			lhs = s[:idx]
			rhs = s[idx+1:]
			valid = true
			return
		}
	}

	parseIndexed := func(s string) (base string, index uint32, valid bool) {
		if lhs, rhs, ok := parseSplit(s); !ok {
			return
		} else {
			base = lhs
			if i64, err := strconv.ParseUint(rhs, 10, 32); err != nil {
				return
			} else {
				index = uint32(i64)
				valid = true
			}
		}

		return
	}

	generators := []func(string) bool{
		func(s string) bool {
			if !strings.HasPrefix(s, "\"") || !strings.HasSuffix(s, "\"") {
				return false
			}

			idx := builder.AddConstant(MakeValue(strings.TrimSuffix(strings.TrimPrefix(s, "\""), "\"")))
			builder.EmitLoad(idx, KindString)

			return true
		}, func(s string) bool {
			if v, err := strconv.ParseFloat(s, 64); err != nil {
				return false
			} else {
				builder.EmitConst(MakeValue(v))
				return true
			}
		}, func(s string) bool {
			if ttbl, ok := parsePrefixedTTBL(s, "LTTBLB_"); !ok {
				return false
			} else {
				builder.EmitByte(OpLTTBLB)
				builder.EmitByte(ttbl)
				return true
			}
		}, func(s string) bool {
			if ttbl, ok := parsePrefixedTTBL(s, "LTTBLU_"); !ok {
				return false
			} else {
				builder.EmitByte(OpLTTBLU)
				builder.EmitByte(ttbl)
				return true
			}
		}, func(s string) bool {
			if base, index, valid := parseIndexed(s); !valid {
				return false
			} else {
				switch base {
				case "NLOAD":
					builder.EmitLoad(index, KindNumber)
					break
				case "BLOAD":
					builder.EmitLoad(index, KindBool)
					break
				case "STRLOAD":
					builder.EmitLoad(index, KindString)
					break
				default:
					return false
				}
				return true
			}
		}, func(s string) bool {
			if lhs, rhs, valid := parseSplit(s); !valid {
				return false
			} else {
				switch lhs {
				case "CNUM":
					builder.EmitLoad(builder.ReserveConstant(rhs, KindNumber), KindNumber)
					break
				case "CBOOL":
					builder.EmitLoad(builder.ReserveConstant(rhs, KindBool), KindBool)
					break
				case "CSTR":
					builder.EmitLoad(builder.ReserveConstant(rhs, KindString), KindString)
					break
				default:
					return false
				}
				return true
			}
		},
	}

	for _, token := range tokens {
		if b, ok := singleByteTokens[token]; ok {
			builder.EmitByte(b)
			continue
		}

		done := false
		for _, gen := range generators {
			if gen(token) {
				done = true
				break
			}
		}
		if done {
			continue
		}

		return errors.New("bad token: " + token)
	}

	return nil
}

func StrictExprToFORTH(str string) []string {
	//((const0>=1)&&(const0<=3))||(const1=="asdasdasd")
	//compiles to:
	//CNUM_const0 1 CMP GE CNUM_const0 3 CMP LE AND CSTR_const1 "asdasdasd" CMP EQ OR

	return nil
}
