package itsu_forth

import (
	"errors"
	"example.com/itsuMain/lib/vm"
	"strconv"
	"strings"
)

func CompileFORTH(builder *vm.ProgramBuilder, str string) error {
	//builder := NewProgramBuilder()
	tokens := vm.TokenizeString(str)

	singleByteTokens := map[string]byte{
		"0":     vm.OpNCONST_0,
		"1":     vm.OpNCONST_1,
		"2":     vm.OpNCONST_2,
		"false": vm.OpBCONST_0,
		"true":  vm.OpBCONST_1,
		"F":     vm.OpBCONST_0,
		"T":     vm.OpBCONST_1,

		"NIL":   vm.OpNILCONST,
		"ISNIL": vm.OpISNIL,
		"KIND":  vm.OpKIND,

		"DUP":  vm.OpSDUP,
		"DROP": vm.OpSDROP,
		"SWAP": vm.OpSSWAP,
		"OVER": vm.OpSOVER,
		"ROT":  vm.OpSROT,

		"CMP": vm.OpCMP,
		"LT":  vm.OpLT,
		"LE":  vm.OpLE,
		"EQ":  vm.OpEQ,
		"GE":  vm.OpGE,
		"GT":  vm.OpGT,
		"NE":  vm.OpNE,
		"<":   vm.OpLT,
		"<=":  vm.OpLE,
		"==":  vm.OpEQ,
		">=":  vm.OpGE,
		">":   vm.OpGT,
		"!=":  vm.OpNE,
		"AND": vm.OpLAND,
		"OR":  vm.OpLOR,
		"XOR": vm.OpLXOR,
		"NOT": vm.OpLNOT,
		"&&":  vm.OpLAND,
		"||":  vm.OpLOR,
		"^^":  vm.OpLXOR,
		"!":   vm.OpLNOT,

		"ADD":  vm.OpNADD,
		"SUB":  vm.OpNSUB,
		"MUL":  vm.OpNMUL,
		"DIV":  vm.OpNDIV,
		"FMOD": vm.OpNFMOD,
		"+":    vm.OpNADD,
		"-":    vm.OpNSUB,
		"*":    vm.OpNMUL,
		"/":    vm.OpNDIV,
		".%":   vm.OpNFMOD,
		"POW":  vm.OpNPOW,
		"^":    vm.OpNPOW,
		"**":   vm.OpNPOW,

		"SQRT":  vm.OpNSQRT,
		"TRUNC": vm.OpNTRUNC,
		"FLOOR": vm.OpNFLOOR,
		"CEIL":  vm.OpNCEIL,
		"SHL":   vm.OpNSHL,
		"SHR":   vm.OpNSHR,
		"<<":    vm.OpNSHL,
		">>":    vm.OpNSHR,

		"HLT": vm.OpHLT,
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

			idx := builder.AddConstant(vm.MakeValue(strings.TrimSuffix(strings.TrimPrefix(s, "\""), "\"")))
			builder.EmitCLoad(idx)

			return true
		}, func(s string) bool {
			if v, err := strconv.ParseFloat(s, 64); err != nil {
				return false
			} else {
				builder.EmitConst(vm.MakeValue(v))
				return true
			}
		}, func(s string) bool {
			if ttbl, ok := parsePrefixedTTBL(s, "LTTBLB_"); !ok {
				return false
			} else {
				builder.EmitByte(vm.OpLTTBLB)
				builder.EmitByte(ttbl)
				return true
			}
		}, func(s string) bool {
			if ttbl, ok := parsePrefixedTTBL(s, "LTTBLU_"); !ok {
				return false
			} else {
				builder.EmitByte(vm.OpLTTBLU)
				builder.EmitByte(ttbl)
				return true
			}
		}, func(s string) bool {
			if base, index, valid := parseIndexed(s); !valid {
				return false
			} else {
				if base != "CLOAD" {
					return false
				}

				builder.EmitCLoad(index)
				return true
			}
		}, func(s string) bool {
			if base, rhs, valid := parseSplit(s); !valid {
				return false
			} else {
				if base != "CNAMED" {
					return false
				}

				builder.EmitCLoad(builder.ReserveConstant(rhs))
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
