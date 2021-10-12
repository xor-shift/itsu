package vm

/*
data types:
bool, number, string
*/

const (
	KindNumber = 0
	KindBool   = 1
	KindString = 2
)

const (
	comparisonTypeLt  = 0
	comparisonTypeLE  = 1
	comparisonTypeEq  = 2
	comparisonTypeGE  = 3
	comparisonTypeGt  = 4
	comparisonTypeNeq = 5
)

const (
	OpNCONST   = 0x00 //8 byte argument
	OpNCONST_0 = 0x01
	OpNCONST_1 = 0x02
	OpNCONST_2 = 0x03

	OpBCONST_0 = 0x04
	OpBCONST_1 = 0x05

	OpNLOAD   = 0x06 //4 byte argument
	OpBLOAD   = 0x07 //4 byte argument
	OpSTRLOAD = 0x08 //4 byte argument
)

const (
	OpSDUP  = 0x10 //n -- n n
	OpSDROP = 0x11 //n --
	OpSSWAP = 0x12 //n1 n2 -- n2 n1
	OpSOVER = 0x13 //n1 n2 -- n1 n2 n1
	OpSROT  = 0x14 //n1 n2 n3 -- n2 n3 n1
)

const (
	OpNCMP   = 0x20 //n1 n2 -- n (negative -> n1<n2, zero -> n1==n2, positive -> n1>n2)
	OpSCMP   = 0x21 //s1 s2 -- strcmp(s1, s2)
	OpLT     = 0x22 //n -- b (see NCMP)
	OpLE     = 0x23
	OpEQ     = 0x24
	OpGE     = 0x25
	OpGT     = 0x26
	OpNE     = 0x27
	OpLAND   = 0x28 //b1 b2 -- (b1&&b2)
	OpLOR    = 0x29
	OpLXOR   = 0x2A
	OpLTTBLB = 0x2B //1 byte argument, least significant 4 bits are used
	OpLNOT   = 0x2C //b -- !b
	OpLTTBLU = 0x2D //same argument as OpLTTBLB
)

const (
	OpNADD   = 0x30
	OpNSUB   = 0x31
	OpNMUL   = 0x32
	OpNDIV   = 0x33
	OpNFMOD  = 0x34
	OpNPOW   = 0x35
	OpNSQRT  = 0x36
	OpNTRUNC = 0x37
	OpNFLOOR = 0x38
	OpNCEIL  = 0x39
	OpNSHL   = 0x3A
	OpNSHR   = 0x3B
)

const (
	OpHLT = 0xE0
)

type OpcodeProperties struct {
	ArgSize        int
	Name           string
	IsConstLoading bool
}

var (
	BadOpcodeProps = OpcodeProperties{
		ArgSize:        0,
		Name:           "[invalid opcode]",
		IsConstLoading: false,
	}
)

func GetOpcodeProperties(opcode byte) OpcodeProperties {
	m := map[byte]OpcodeProperties{
		OpNCONST:   {8, "NCONST", false},
		OpNCONST_0: {0, "NCONST_0", false},
		OpNCONST_1: {0, "NCONST_1", false},
		OpNCONST_2: {0, "NCONST_2", false},
		OpBCONST_0: {0, "BCONST_0", false},
		OpBCONST_1: {0, "BCONST_1", false},
		OpNLOAD:    {4, "NLOAD", true},
		OpBLOAD:    {4, "BLOAD", true},
		OpSTRLOAD:  {4, "STRLOAD", true},
		OpSDUP:     {0, "SDUP", false},
		OpSDROP:    {0, "SDROP", false},
		OpSSWAP:    {0, "SSWAP", false},
		OpSOVER:    {0, "SOVER", false},
		OpSROT:     {0, "SROT", false},
		OpNCMP:     {0, "NCMP", false},
		OpSCMP:     {0, "SCMP", false},
		OpLT:       {0, "LT", false},
		OpLE:       {0, "LE", false},
		OpEQ:       {0, "EQ", false},
		OpGE:       {0, "GE", false},
		OpGT:       {0, "GT", false},
		OpNE:       {0, "NE", false},
		OpLAND:     {0, "LAND", false},
		OpLOR:      {0, "LOR", false},
		OpLXOR:     {0, "LXOR", false},
		OpLTTBLB:   {0, "LTTBLB", false},
		OpLNOT:     {0, "LNOT", false},
		OpLTTBLU:   {0, "LTTBLU", false},
		OpHLT:      {0, "HLT", false},
	}

	if v, ok := m[opcode]; ok {
		return v
	} else {
		return BadOpcodeProps
	}
}

func (p OpcodeProperties) Bad() bool { return p == BadOpcodeProps }
