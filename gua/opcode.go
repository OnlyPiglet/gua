package gua

import (
	"fmt"
	"os"
	"strconv"
)

const MaxStack = 256

type Real float64

var Stack = make([]Object, MaxStack)

func init() {
	for i, _ := range Stack {
		Stack[i] = BaseObject{
			tag:   Mark,
			value: nil,
		}
	}
}

var (
	StackTopIndex  = 1
	StackBaseIndex = 1
)

type Symbol struct {
	Name   string
	Object Object
}

type Word uint16

func (w Word) ToInt() int {
	return int(w)
}

type Byte uint

type CodeWordM struct {
	c1 byte
	c2 byte
}

type CodeWord struct {
	M CodeWordM
	W Word
}

type CodeFloat struct {
}

type CFunction func()

type Object interface {
	Tag() ObjectTypeTag
	SetTag(ObjectTypeTag)
	SetValue(val any)
}

func ObjectEqual(a Object, b Object) bool {
	if a == b {
		return true
	}
	if a.Tag() != b.Tag() {
		return false
	}
	tag := a.Tag()
	switch tag {
	case NUMBER:
		return a.(NumberObject).Value() == b.(NumberObject).Value()
	case STRING:
		return a.(StringObject).Value() == b.(StringObject).Value()
	case NIl:
		return true
	case CFUNCTION:
		return a.(CFunctionObject).Value() == b.(CFunctionObject).Value()
	case USERDATA:
		return a.(UserDataObject).Value() == b.(UserDataObject).Value()
		//TODO other tag
	}
	return false
}

type BaseObject struct {
	tag   ObjectTypeTag
	value any
}

func (bo BaseObject) Tag() ObjectTypeTag {
	return bo.tag
}

func (bo BaseObject) SetTag(tag ObjectTypeTag) {
	bo.tag = tag
}

func (bo BaseObject) SetValue(val any) {
	bo.value = val
}

type MarkObject struct {
	BaseObject
}

type NilObject struct {
	BaseObject
}

type NumberObject struct {
	BaseObject
}

type ByteObject struct {
	BaseObject
}

func (o ByteObject) Value() Byte {
	return o.BaseObject.value.(Byte)
}

func (o NumberObject) Value() float64 {
	return o.BaseObject.value.(float64)
}

type MarkableObject interface {
	Mark()
	IsMarked() bool
}

type StringObject struct {
	mark bool
	BaseObject
}

func (o StringObject) Mark() {
	o.mark = true
}

func (o StringObject) IsMarked() bool {
	return o.mark
}

func (o StringObject) Value() string {
	return o.BaseObject.value.(string)
}

type ArrayObject struct {
	Mark bool
	BaseObject
}

func (o ArrayObject) Value() *Hash {
	return o.BaseObject.value.(*Hash)
}

type CFunctionObject struct {
	BaseObject
}

func NewCFunctionObject(fn CFunction) *CFunctionObject {
	return &CFunctionObject{
		BaseObject{
			tag:   CFUNCTION,
			value: fn,
		},
	}
}

func NewNumberObject(n float64) NumberObject {
	return NumberObject{
		BaseObject: BaseObject{
			tag:   NUMBER,
			value: n,
		},
	}
}

func NewStringObject(s string) StringObject {
	return StringObject{
		mark: false,
		BaseObject: BaseObject{
			tag:   STRING,
			value: s,
		},
	}
}

func (o CFunctionObject) Value() *CFunction {
	return o.BaseObject.value.(*CFunction)
}

type UserDataObject struct {
	BaseObject
}

func (o UserDataObject) Value() interface{} {
	return o.BaseObject.value
}

// ObjectTypeTag mark which tag of object is
type ObjectTypeTag uint

const (
	Mark ObjectTypeTag = iota
	NIl
	NUMBER
	STRING
	ARRAY
	FUNCTION
	CFUNCTION
	USERDATA
)

func (o ObjectTypeTag) String() string {
	switch o {
	case NUMBER:
		return "number"
	case STRING:
		return "string"
	case ARRAY:
		return "table"
	case FUNCTION:

		return "function"
	case CFUNCTION:

		return "cfunction"
	case USERDATA:

		return "userdata"
	case NIl:

		return "nil"
	case Mark:

		return "mark"
	}
	return "unknown"
}

func stackFlow() int {
	if (StackTopIndex - 0) > (MaxStack - 1) {
		os.Stderr.WriteString("stack overflow")
	}
	return 1
}

// LuaPushNil Push a nil object
func LuaPushNil() int {
	if stackFlow() == 1 {
		return 1
	}
	Stack[StackTopIndex].SetTag(NIl)
	return 0
}

// LuaPushNumber Push a nil object
func LuaPushNumber(n Real) int {
	if stackFlow() == 1 {
		return 1
	}
	Stack[StackTopIndex].SetTag(NUMBER)
	Stack[StackTopIndex] = NewNumberObject(float64(n))
	StackTopIndex++
	return 0
}

func LuaPushObject(o Object) int {
	if stackFlow() == 1 {
		return 1
	}
	Stack[StackTopIndex] = o
	StackTopIndex = StackTopIndex + 1
	return 0
}

func LuaPushString(s string) int {
	if stackFlow() == 1 {
		return 1
	}
	Stack[StackTopIndex].SetTag(STRING)
	Stack[StackTopIndex].SetValue(s)
	StackTopIndex = StackTopIndex + 1
	return 0
}

func LuaPushCFunction(fn LuaCFunction) int {
	if stackFlow() == 1 {
		return 1
	}
	Stack[StackTopIndex].SetTag(CFUNCTION)
	Stack[StackTopIndex].(BaseObject).SetValue(fn)
	StackTopIndex = StackTopIndex + 1
	return 0
}

func LuaGetParam(number int) Object {
	if number <= 0 || number > StackTopIndex-StackBaseIndex {
		return nil
	}
	return Stack[StackBaseIndex+number-1]
}

func LuaIsNil(o Object) bool {
	return o != nil && o.Tag() == NIl
}

func LuaIsNumber(o Object) bool {
	return o != nil && o.Tag() == NUMBER
}

func LuaIsString(o Object) bool {
	return o != nil && o.Tag() == STRING
}

func LuaIsTable(o Object) bool {
	return o != nil && o.Tag() == ARRAY
}

func LuaIsCFunction(o Object) bool {
	return o != nil && o.Tag() == CFUNCTION
}

func LuaIsUserData(o Object) bool {
	return o != nil && o.Tag() == USERDATA
}

func LuaType() {
	o := LuaGetParam(1)
	LuaPushString(o.Tag().String())
}

func LuaObj2Number() {
	o := LuaGetParam(1)
	LuaPushObject(luaConvToNumber(o))
}

func luaConvToNumber(obj Object) Object {
	var cvt Object

	if obj.Tag() == NUMBER {
		ob := obj.(NumberObject)
		cvt = NumberObject{BaseObject{
			tag:   NUMBER,
			value: ob.Value(),
		}}
		return cvt
	}

	if obj.Tag() == STRING {
		ob := obj.(StringObject)
		cvt = NumberObject{BaseObject{
			tag:   NUMBER,
			value: strconv.ParseFloat(ob.Value(), 10),
		}}
		return cvt
	}

	return nil
}

func LuaPrint() {
	i := 1
	var obj Object
	for obj = LuaGetParam(i); obj != nil; {
		if LuaIsNumber(obj) {
			fmt.Println("%v", LuaGetNumber(obj))
		}
		if LuaIsString(obj) {
			fmt.Println("%s", obj.(StringObject).Value())
		}
		if LuaIsCFunction(obj) {
			fmt.Println("cfunction: %p\n", LuaGetCFunction(obj))
		}

		if LuaIsUserData(obj) {
			fmt.Println("userdata: %p", LuaGetUserData(obj))
		}
		if LuaIsTable(obj) {
			fmt.Println("table: %p", obj)
		}
		if LuaIsNil(obj) {
			fmt.Println("nil")
		} else {
			fmt.Println("invalid value to print")
		}

		i = i + 1
	}

}

func LuaGetNumber(obj Object) Real {
	if obj == nil || obj.Tag() == NIl {
		return 0.0
	}

	newObj, e := LuaToNumber(obj)
	if e == 1 {
		return 0.0
	}

	return Real(newObj.(NumberObject).Value())
}

func LuaToNumber(obj Object) (Object, int) {
	if obj.Tag() != STRING {
		os.Stderr.WriteString("unexpected type at conversion to number")
		return obj, 1
	}
	//ostr :=
	n, _ := strconv.ParseFloat(obj.(StringObject).Value(), 10)
	obj = Object(NewNumberObject(n))
	return obj, 0
}

func LuaGetCFunction(obj Object) *CFunction {
	if obj == nil || obj.Tag() != CFUNCTION {
		return nil
	}
	return obj.(CFunctionObject).Value()
}

func LuaGetUserData(obj Object) interface{} {
	if obj == nil || obj.Tag() != USERDATA {
		return nil
	}
	return obj.(UserDataObject).Value()
}
