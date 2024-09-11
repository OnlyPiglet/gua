package main

import (
	"fmt"
)

const GAPCODE = 50
const FIELDS_PER_FLUSH = 10

type OpCode byte

const (
	PUSHNIL OpCode = iota
	PUSH0
	PUSH1
	PUSH2
	PUSHBYTE
	PUSHWORD
	PUSHFLOAT
	PUSHSTRING
	PUSHLOCAL0
	PUSHLOCAL1
	PUSHLOCAL2
	PUSHLOCAL3
	PUSHLOCAL4
	PUSHLOCAL5
	PUSHLOCAL6
	PUSHLOCAL7
	PUSHLOCAL8
	PUSHLOCAL9
	PUSHLOCAL
	PUSHGLOBAL
	PUSHINDEXED
	PUSHMARK
	PUSHOBJECT
	STORELOCAL0
	STORELOCAL1
	STORELOCAL2
	STORELOCAL3
	STORELOCAL4
	STORELOCAL5
	STORELOCAL6
	STORELOCAL7
	STORELOCAL8
	STORELOCAL9
	STORELOCAL
	STOREGLOBAL
	STOREINDEXED0
	STOREINDEXED
	STORERECORD
	ADJUST
	CREATEARRAY
	EQOP
	LTOP
	LEOP
	ADDOP
	SUBOP
	MULTOP
	DIVOP
	CONCOP
	MINUSOP
	NOTOP
	ONTJMP
	ONFJMP
	JMP
	UPJMP
	IFFJMP
	IFFUPJMP
	POP
	CALLFUNC
	RETCODE
	HALT
	SETFUNCTION
	SETLINE
	RESET
)

type CodeWord struct {
	w uint16
	m struct {
		c1 byte
		c2 byte
	}
}

type CodeFloat struct {
	f float32
	m struct {
		c1 byte
		c2 byte
		c3 byte
		c4 byte
	}
}

type YYSTYPE struct {
	vInt   int
	vLong  int64
	vFloat float32
	pChar  *string
	vWord  uint16
	pByte  *byte
}

const (
	WRONGTOKEN = 257
	NIL        = 258
	IF         = 259
	THEN       = 260
	ELSE       = 261
	ELSEIF     = 262
	WHILE      = 263
	DO         = 264
	REPEAT     = 265
	UNTIL      = 266
	END        = 267
	RETURN     = 268
	LOCAL      = 269
	NUMBER     = 270
	FUNCTION   = 271
	STRING     = 272
	NAME       = 273
	DEBUG      = 274
	AND        = 275
	OR         = 276
	NE         = 277
	LE         = 278
	GE         = 279
	CONC       = 280
	UNARY      = 281
	NOT        = 282
)

type Byte byte
type Word uint16

var maxcode Word
var maxmain Word
var maxcurr Word
var code []Byte
var initcode []Byte
var basepc []Byte
var maincode Word
var pc Word

const MAXVAR = 32

var varbuffer = make([]int64, MAXVAR)
var nvarbuffer int

var localvar = make([]Word, 100) // Assuming STACKGAP is a large number. Adjust as needed.
var nlocalvar int

const MAXFIELDS = FIELDS_PER_FLUSH * 2

var fields = make([]Word, MAXFIELDS)
var nfields int
var ntemp int
var err bool

func code_byte(c Byte) {
	if pc > maxcurr-2 {
		maxcurr += GAPCODE
		newBasepc := make([]Byte, maxcurr)
		copy(newBasepc, basepc)
		basepc = newBasepc
		if basepc == nil {
			panic("not enough memory")
			err = true
		}
	}
	basepc[pc] = c
	pc++
}

func code_word(n Word) {
	codeWord := CodeWord{w: n}
	code_byte(codeWord.m.c1)
	code_byte(codeWord.m.c2)
}

func code_float(n float32) {
	codeFloat := CodeFloat{f: n}
	code_byte(codeFloat.m.c1)
	code_byte(codeFloat.m.c2)
	code_byte(codeFloat.m.c3)
	code_byte(codeFloat.m.c4)
}

func code_word_at(p *Byte, n Word) {
	codeWord := CodeWord{w: n}
	*p = codeWord.m.c1
	p = p[1:]
	*p = codeWord.m.c2
}

func push_field(name Word) {
	if nfields < MAXFIELDS-1 {
		fields[nfields] = name
		nfields++
	} else {
		panic("too many fields in a constructor")
		err = true
	}
}

func flush_record(n int) {
	if n == 0 {
		return
	}
	code_byte(STORERECORD)
	code_byte(Byte(n))
	for i := n - 1; i >= 0; i-- {
		code_word(fields[i])
	}
	ntemp -= n
}

func flush_list(m, n int) {
	if n == 0 {
		return
	}
	if m == 0 {
		code_byte(STORELIST0)
	} else {
		code_byte(STORELIST)
		code_byte(Byte(m))
	}
	code_byte(Byte(n))
	ntemp -= n
}

func incr_ntemp() {
	if ntemp+nlocalvar+MAXVAR+1 < 1000 { // Assuming STACKGAP is a large number. Adjust as needed.
		ntemp++
	} else {
		panic("stack overflow")
		err = true
	}
}

func add_nlocalvar(n int) {
	if ntemp+nlocalvar+MAXVAR+n < 1000 { // Assuming STACKGAP is a large number. Adjust as needed.
		nlocalvar += n
	} else {
		panic("too many local variables or expression too complicate")
		err = true
	}
}

func incr_nvarbuffer() {
	if nvarbuffer < MAXVAR-1 {
		nvarbuffer++
	} else {
		panic("variable buffer overflow")
		err = true
	}
}

func code_number(f float32) {
	var i Word = Word(f)
	if f == float32(i) {
		if i <= 2 {
			code_byte(PUSH0 + Byte(i))
		} else if i <= 255 {
			code_byte(PUSHBYTE)
			code_byte(Byte(i))
		} else {
			code_byte(PUSHWORD)
			code_word(i)
		}
	} else {
		code_byte(PUSHFLOAT)
		code_float(f)
	}
	incr_ntemp()
}

func lua_localname(n Word) int {
	for i := nlocalvar - 1; i >= 0; i-- {
		if n == localvar[i] {
			return i
		}
	}
	return -1
}

func lua_pushvar(number int64) {
	if number > 0 {
		code_byte(PUSHGLOBAL)
		code_word(Word(number - 1))
		incr_ntemp()
	} else if number < 0 {
		number = -number - 1
		if number < 10 {
			code_byte(PUSHLOCAL0 + Byte(number))
		} else {
			code_byte(PUSHLOCAL)
			code_byte(Byte(number))
		}
		incr_ntemp()
	} else {
		code_byte(PUSHINDEXED)
		ntemp--
	}
}

func lua_codeadjust(n int) {
	code_byte(ADJUST)
	code_byte(Byte(n + nlocalvar))
}

func lua_codestore(i int) {
	if varbuffer[i] > 0 {
		code_byte(STOREGLOBAL)
		code_word(Word(varbuffer[i] - 1))
	} else if varbuffer[i] < 0 {
		number := -varbuffer[i] - 1
		if number < 10 {
			code_byte(STORELOCAL0 + Byte(number))
		} else {
			code_byte(STORELOCAL)
			code_byte(Byte(number))
		}
	} else {
		j := i + 1
		upper := 0
		param := 0
		for ; j < nvarbuffer; j++ {
			if varbuffer[j] == 0 {
				upper++
			}
		}
		param = upper*2 + i
		if param == 0 {
			code_byte(STOREINDEXED0)
		} else {
			code_byte(STOREINDEXED)
			code_byte(Byte(param))
		}
	}
}

func yyerror(s string) {
	msg := fmt.Sprintf("%s near \"%s\" at line %d in file \"%s\"", s, "lua_lasttext()", "lua_linenumber()", "lua_filename()")
	panic(msg)
	err = true
}

func yywrap() bool {
	return true
}

func lua_parse() int {
	init := make([]Byte, GAPCODE)
	initcode = init
	maincode = 0
	maxmain = GAPCODE
	if init == nil {
		panic("not enough memory")
		return 1
	}
	err = false
	if yyparse() || err {
		return 1
	}
	initcode[maincode] = HALT
	init = initcode
	// if LISTING {
	//    PrintCode(init, init+maincode)
	// }
	if lua_execute(init) {
		return 1
	}
	return 0
}

// func PrintCode(code, end []Byte) {
//    // Implementation omitted for brevity.
// }

func lua_execute(code []Byte) bool {
	return false
}

// Assume these functions are implemented elsewhere or replaced with appropriate logic.
func lua_findsymbol(s *string) Word {
	return 0
}

func lua_findconstant(s *string) Word {
	return 0
}

func yyparse() bool {
	return false
}
