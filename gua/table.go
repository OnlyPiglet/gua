package gua

import "os"

const MaxSymbol = 512

var TableBuffer = make([]Symbol, MaxSymbol)
var LuaTable = TableBuffer

func init() {
	TableBuffer[0] = Symbol{
		Name:   "type",
		Object: NewCFunctionObject(LuaType),
	}
	TableBuffer[1] = Symbol{
		Name:   "tonumber",
		Object: NewCFunctionObject(LuaObj2Number),
	}
	TableBuffer[2] = Symbol{
		Name:   "next",
		Object: NewCFunctionObject(LuaNext),
	}
	TableBuffer[3] = Symbol{
		Name:   "nextvar",
		Object: NewCFunctionObject(LuaNextVar),
	}
	TableBuffer[4] = Symbol{
		Name:   "print",
		Object: NewCFunctionObject(LuaPrint),
	}
	TableBuffer[5] = Symbol{
		Name:   "dofile",
		Object: NewCFunctionObject(LuaInternalDoFile),
	}
	TableBuffer[6] = Symbol{
		Name:   "dostring",
		Object: NewCFunctionObject(LuaInternalDoString),
	}
}

var LuaNTable Word = 7

type List struct {
	S    *Symbol
	Next *List
}

/* Variables to controll garbage collection */
var (
	LuaBlock   Word = 10 /* to check when garbage collector will be called */
	LuaNentity Word = 0  /* counter of new entities (strings and arrays) */
)

func LuaMarkObject(o Object) {
	if o.Tag() == STRING {
		s, _ := o.(StringObject)
		s.Mark()
	}
	if o.Tag() == ARRAY {
		a := o.(ArrayObject)
		LuaHashMark(a.Value())
	}
}

func LuaFindSymbol(s string) int {
	var l *List
	var p *List

	for l != nil {
		if s == l.S.Name {
			if p != nil {
				p.Next = l.Next
				l.Next = sea
			}
		}

		p = l
		l = l.Next
	}

}

func LuaNextVar() {
	index := 0
	o := LuaGetParam(1)
	if o == nil {
		os.Stderr.WriteString("too few arguments to function `nextvar\n`")
		return
	}
	if LuaGetParam(2) != nil {
		os.Stderr.WriteString("too many arguments to function `nextvar\n`")
		return
	}
	if o.Tag() == NIl {
		index = 0
	} else if o.Tag() != STRING {
		os.Stderr.WriteString("incorrect argument to function `nextvar\n`")
		return
	} else {
		ostr := o.(StringObject)
		stop := LuaNTable.ToInt()
		for ; index < stop; index++ {
			if LuaTable[index].Name == ostr.Value() {
				break
			}

		}
		if index == LuaNTable.ToInt() {
			os.Stderr.WriteString("name not found in function `nextvar`\n")
			return
		}
		index++
		for index < LuaNTable.ToInt() && LuaTable[index].Object.Tag() == NIl {
			index++
		}

		if index == LuaNTable.ToInt() {
			LuaPushNil()
			LuaPushNil()
			return
		}
	}
	{
		name := NewStringObject(LuaTable[index].Name)
		if LuaPushObject(name) != 0 {
			return
		}
		if LuaPushObject(LuaTable[index].Object) != 0 {
			return
		}
	}
}

// LuaInternalDoFile TODO
func LuaInternalDoFile() {
	obj := LuaGetParam(1)
	if LuaIsString(obj) && LuaGetParam(2) != nil {
		LuaPushNumber(1)
	} else {
		LuaPushNil()
	}
}

// LuaInternalDoString TODO
func LuaInternalDoString() {
	obj := LuaGetParam(1)
	if LuaIsString(obj) && LuaGetParam(2) != nil {
		LuaPushNumber(1)
	} else {
		LuaPushNil()
	}

}

func LuaGetString(obj Object) string {
	if obj == nil || obj.Tag() == NIl {
		return ""
	}
	//

}

func LuaTravStack(fn func(obj Object)) {

	for i := StackTopIndex - 1; i >= StackBaseIndex; i-- {
		fn(Stack[i])
	}

}
