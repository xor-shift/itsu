package comparison

import (
	"errors"
	"example.com/itsuMain/lib/util"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type CType uint8
type CField uint64

type TokenID uint8

const (
	TypeLt  = CType(0)
	TypeLE  = CType(1)
	TypeEq  = CType(2)
	TypeGE  = CType(3)
	TypeGt  = CType(4)
	TypeNeq = CType(5)

	TokenIDPush        = TokenID(0x00)
	TokenIDDrop        = TokenID(0x01)
	TokenIDDup         = TokenID(0x02)
	TokenIDField       = TokenID(0x03)
	TokenIDRoll        = TokenID(0x04)
	TokenIDCmp         = TokenID(0x20)
	TokenIDBinaryLogic = TokenID(0x21)
	TokenIDUnaryLogic  = TokenID(0x22)

	stackSize = 64
)

var (
	comparisonToStringMap = map[CType]string{
		TypeLt:  "<",
		TypeLE:  "<=",
		TypeEq:  "==",
		TypeGE:  ">=",
		TypeGt:  ">",
		TypeNeq: "!=",
	}
)

var (
	ErrorBadTokenID        = errors.New("token with unknown id")
	ErrorBadTokenImmediate = errors.New("token with inappropriate immediate")
	ErrorStackUnderflow    = errors.New("stack underflow")
	ErrorStackOverflow     = errors.New("stack overflow")
	ErrorUnknownField      = errors.New("unknown field")
	ErrorIncomparable      = errors.New("tried to compare two incomparable values")
	ErrorBadStackContent   = errors.New("bad stack content for the given token")
	ErrorBadFiledType      = errors.New("popped value is not a valid field type")
)

type Token struct {
	ID  TokenID
	Imm interface{}
}

var (
	TokenDROP = Token{
		ID:  TokenIDDrop,
		Imm: nil,
	}
	TokenDUP = Token{
		ID:  TokenIDDup,
		Imm: nil,
	}
	TokenROL = Token{
		ID:  TokenIDRoll,
		Imm: nil,
	}

	//TokenFieldResolve is the standalone field resolver, pops an integer or string value from the stack and resolves it
	TokenFieldResolve = Token{
		ID:  TokenIDField,
		Imm: nil,
	}

	TokenAND = Token{
		ID:  TokenIDBinaryLogic,
		Imm: uint8(util.RelAnd),
	}
	TokenOR = Token{
		ID:  TokenIDBinaryLogic,
		Imm: uint8(util.RelOr),
	}
	TokenXOR = Token{
		ID:  TokenIDBinaryLogic,
		Imm: uint8(util.RelXor),
	}
	TokenNOT = Token{
		ID:  TokenIDUnaryLogic,
		Imm: uint8(util.RelNotP),
	}

	TokenCmpLt = Token{
		ID:  TokenIDCmp,
		Imm: uint8(TypeLt),
	}
	TokenCmpLE = Token{
		ID:  TokenIDCmp,
		Imm: uint8(TypeLE),
	}
	TokenCmpEq = Token{
		ID:  TokenIDCmp,
		Imm: uint8(TypeEq),
	}
	TokenCmpGE = Token{
		ID:  TokenIDCmp,
		Imm: uint8(TypeGE),
	}
	TokenCmpGt = Token{
		ID:  TokenIDCmp,
		Imm: uint8(TypeGt),
	}
	TokenCmpNeq = Token{
		ID:  TokenIDCmp,
		Imm: uint8(TypeNeq),
	}
)

func tokenizeForProgram(str string) []string {
	runes := []rune(str)

	strTokens := make([]string, 0)
	currentToken := make([]rune, 0)
	currentTokenIsString := false

	finishToken := func() {
		if len(currentToken) == 0 {
			return
		}

		strTokens = append(strTokens, string(currentToken))
		currentToken = make([]rune, 0)
		currentTokenIsString = false
	}

	appendRune := func(r rune) bool {
		if !currentTokenIsString && unicode.IsSpace(r) {
			return true
		}

		if len(currentToken) == 0 && r == '"' {
			currentTokenIsString = true
		}

		currentToken = append(currentToken, r)

		if currentTokenIsString && r == '"' && len(currentToken) != 1 {
			return true
		}

		return false
	}

	for _, v := range runes {
		finishing := appendRune(v)
		if finishing {
			finishToken()
		}
	}
	finishToken()

	return strTokens
}

func StringToProgram(str string) ([]Token, error) {
	strTokens := tokenizeForProgram(str)

	basicTokens := map[string]Token{
		"DROP":  TokenDROP,
		"DUP":   TokenDUP,
		"ROL":   TokenROL,
		"FIELD": TokenFieldResolve,
		"AND":   TokenAND,
		"OR":    TokenOR,
		"XOR":   TokenXOR,
		"NOT":   TokenNOT,
		"<":     TokenCmpLt,
		"<=":    TokenCmpLE,
		"==":    TokenCmpEq,
		">=":    TokenCmpGE,
		">":     TokenCmpGt,
		"!=":    TokenCmpNeq,
	}

	tokenGenerators := []func(string) (Token, bool){
		func(s string) (Token, bool) {
			if s[len(s)-1] == 'u' {
				if u, err := strconv.ParseUint(s[:len(s)-1], 10, 64); err == nil {
					return Token{ID: TokenIDPush, Imm: u}, true
				} else {
					return Token{}, false
				}
			} else {
				if i, err := strconv.ParseInt(s, 10, 64); err == nil {
					return Token{ID: TokenIDPush, Imm: i}, true
				} else {
					return Token{}, false
				}
			}
		},
		func(s string) (Token, bool) {
			if !strings.HasPrefix(s, "\"") || !strings.HasSuffix(s, "\"") {
				return Token{}, false
			}

			str := strings.TrimPrefix(strings.TrimSuffix(s, "\""), "\"")
			return Token{ID: TokenIDPush, Imm: str}, true
		},
	}

	tokens := make([]Token, len(strTokens))
	for k, v := range strTokens {
		gotToken := false
		token := Token{}

		if bt, ok := basicTokens[v]; ok {
			token = bt
			gotToken = true
		} else {
			for _, gen := range tokenGenerators {
				var genToken Token
				if genToken, ok = gen(v); ok {
					token = genToken
					gotToken = true
					break
				}
			}
		}

		if !gotToken {
			return tokens, errors.New(fmt.Sprint("unknown token: ", v))
		}

		tokens[k] = token
	}

	return tokens, nil
}

func StringFromProgram(program []Token) (string, error) {
	builder := strings.Builder{}

	for _, v := range program {
		switch v.ID {
		case TokenIDPush:
			switch reflect.TypeOf(v.Imm).Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				builder.WriteString(fmt.Sprint(v.Imm))
				break

			case reflect.String:
				builder.WriteString(v.Imm.(string))
				break

			default:
				return "", errors.New("bad argument for a PUSH token")
			}
			break
		case TokenIDDrop:
			builder.WriteString("DROP")
			break
		case TokenIDDup:
			builder.WriteString("DUP")
			break
		case TokenIDRoll:
			builder.WriteString("ROL")
			break
		case TokenIDField:
			builder.WriteString("FIELD")
			if v.Imm != nil {
				builder.WriteRune('_')

				switch v2 := v.Imm.(type) {
				case uint64:
					builder.WriteString(fmt.Sprint(v2))
				case string:
					builder.WriteRune('"')
					builder.WriteString(v2)
					builder.WriteRune('"')
				default:
					return "", errors.New("bad argument for a FIELD token")
				}
			}
			break
		case TokenIDCmp:
			if comparisonType, ok := v.Imm.(uint8); ok {
				if cTypeStr, ok := comparisonToStringMap[CType(comparisonType)]; ok {
					builder.WriteString(cTypeStr)
				} else {
					builder.WriteString(fmt.Sprintf("CMP_%d", comparisonType))
				}
			} else {
				return "", errors.New("bad argument for a CMP token")
			}
			break

		case TokenIDBinaryLogic:
			if opType, ok := v.Imm.(uint8); ok {
				switch util.Relation(opType) {
				case util.RelAnd:
					builder.WriteString("AND")
					break
				case util.RelOr:
					builder.WriteString("OR")
					break
				case util.RelXor:
					builder.WriteString("XOR")
					break
				default:
					builder.WriteString(fmt.Sprintf("LB_%d", opType))
					break
				}
			} else {
				return "", errors.New("bad argument for a LB token")
			}
			break

		case TokenIDUnaryLogic:
			if opType, ok := v.Imm.(uint8); ok {
				if util.Relation(opType) == util.RelNotP {
					builder.WriteString("NOT")
				} else {
					builder.WriteString(fmt.Sprintf("LU_%d", opType))
				}
			}
			break
		}

		builder.WriteRune(' ')
	}

	return builder.String(), nil
}

type Comparer struct {
	Fields map[CField]interface{}

	program  []Token
	instrPtr int

	stack    [stackSize]interface{}
	stackPtr int
}

func NewComparer(program []Token) *Comparer {
	c := &Comparer{
		Fields: make(map[CField]interface{}),

		program:  program,
		instrPtr: 0,

		stack:    [stackSize]interface{}{},
		stackPtr: 0,
	}

	return c
}

func (c *Comparer) push(v interface{}) (err error) {
	if c.stackPtr == stackSize {
		err = ErrorStackOverflow
		return
	}

	c.stack[c.stackPtr] = v
	c.stackPtr++

	return
}

func (c *Comparer) pop() (v interface{}, err error) {
	if c.stackPtr == 0 {
		return nil, ErrorStackUnderflow
	}

	c.stackPtr--
	v = c.stack[c.stackPtr]

	return
}

func (c *Comparer) SingleStep() (err error) {
	getUint := func(v interface{}) (uint64, error) {
		if v == nil {
			return 0, errors.New("the passed value is nil")
		}

		val := reflect.ValueOf(v)
		field := uint64(0)

		switch val.Type().Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			field = uint64(val.Int())
			break
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			field = val.Uint()
			break
		default:
			return 0, ErrorBadTokenImmediate
		}

		return field, nil
	}

	popBool := func() (bool, error) {
		if v, err := c.pop(); err != nil {
			return false, err
		} else {
			if b, ok := v.(bool); !ok {
				return false, ErrorBadStackContent
			} else {
				return b, nil
			}
		}
	}

	pushLogic := func(v interface{}, p, q bool) error {
		var rel uint64
		if rel, err = getUint(v); err != nil {
			return err
		}

		if err = c.push(util.TTableEval(p, q, util.Relation(rel))); err != nil {
			return err
		}

		return nil
	}

	v := c.program[c.instrPtr]

	switch v.ID {
	case TokenIDPush:
		if err = c.push(v.Imm); err != nil {
			return
		}
		break

	case TokenIDDrop:
		if _, err = c.pop(); err != nil {
			return
		}
		break

	case TokenIDDup:
		var vi interface{}
		if vi, err = c.pop(); err != nil {
			return
		}

		if err = c.push(vi); err != nil {
			return
		}
		if err = c.push(vi); err != nil {
			return
		}

		break

	case TokenIDRoll:
		var over, top interface{}

		if top, err = c.pop(); err != nil {
			return
		}
		if over, err = c.pop(); err != nil {
			return
		}

		_ = c.push(top)
		_ = c.push(over)

		break

	case TokenIDField:
		var field uint64

		if field, err = getUint(v.Imm); err != nil {
			var poppedField interface{}
			if poppedField, err = c.pop(); err != nil {
				return
			}

			switch resolution := poppedField.(type) {
			case string:
				return ErrorUnknownField
			case uint64:
				field = resolution
				break
			default:
				return ErrorBadFiledType
			}
		}

		if f, ok := c.Fields[CField(field)]; !ok {
			return ErrorUnknownField
		} else {
			if err = c.push(f); err != nil {
				return
			}
		}

		break

	case TokenIDCmp:
		var cmpV uint64
		if cmpV, err = getUint(v.Imm); err != nil {
			return
		}
		cmp := CType(cmpV)

		var lhs, rhs interface{}
		if lhs, err = c.pop(); err != nil {
			return
		}
		if rhs, err = c.pop(); err != nil {
			return
		}

		sRes := util.Spaceship(lhs, rhs)

		if sRes == 2 {
			err = ErrorIncomparable
			return
		} else {
			if err = c.push(util.MatCond(cmp == TypeLt, sRes < 0) &&
				util.MatCond(cmp == TypeLE, sRes <= 0) &&
				util.MatCond(cmp == TypeEq, sRes == 0) &&
				util.MatCond(cmp == TypeGE, sRes >= 0) &&
				util.MatCond(cmp == TypeGt, sRes > 0) &&
				util.MatCond(cmp == TypeNeq, sRes != 0)); err != nil {
				return
			}
		}

	case TokenIDBinaryLogic:
		var lhs, rhs bool

		if lhs, err = popBool(); err != nil {
			return
		}
		if rhs, err = popBool(); err != nil {
			return
		}

		if err = pushLogic(v.Imm, lhs, rhs); err != nil {
			return
		}

		break

	case TokenIDUnaryLogic:
		var lhs bool

		if lhs, err = popBool(); err != nil {
			return
		}

		if err = pushLogic(v.Imm, lhs, false); err != nil {
			return
		}

		break

	default:
		err = ErrorBadTokenID
		return
	}

	return
}

func (c *Comparer) Run() (result bool, err error) {
	for c.instrPtr < len(c.program) {
		if err = c.SingleStep(); err != nil {
			return
		}
		c.instrPtr++
	}

	if c.instrPtr == len(c.program) {
		popped, err := c.pop()

		if err != nil {
			return false, err
		}

		if b, ok := popped.(bool); ok {
			return b, nil
		} else {
			return false, errors.New("bad last value left on stack")
		}
	}

	return false, errors.New("FIXME")
}
