package gua

import (
	"fmt"
	"os"
)

var (
	ListHead *ArrayList = nil
)

type ArrayList struct {
	Array *Hash
	Next  *ArrayList
}

type Node struct {
	Ref  Object
	Val  Object
	Next *Node
}

type Hash struct {
	Mark  uint
	Nhash int
	List  []*Node
}

// hashCreate 初始化hash
func hashCreate(nhash int) *Hash {
	return &Hash{
		Nhash: nhash,
		Mark:  0,
		List:  make([]*Node, nhash),
	}
}

func LuaCreateArray(nhash int) *Hash {
	newArray := new(ArrayList)
	newArray.Array = hashCreate(nhash)
	if LuaNentity == LuaBlock {
		//TODO lua_pack()
	}
	LuaNentity = LuaNentity + 1
	newArray.Next = ListHead
	ListHead = newArray
	return newArray.Array
}

func LuaHashMark(h *Hash) {
	if h.Mark == 0 {
		h.Mark = 1
		for i := range h.Nhash {
			if h.List[i] != nil {
				LuaMarkObject(h.List[i].Ref)
				LuaMarkObject(h.List[i].Val)
			}
		}
	}
}

func freelist(n *Node) {
	for n != nil {
		next := n.Next
		n = nil
		n = next
	}
}

func hashDelete(h *Hash) {
	for i := range h.Nhash {
		freelist(h.List[i])
	}
	h.List = nil
	h = nil
}

func LuaHashCollector() {
	curr := ListHead
	var prev *ArrayList

	for curr != nil {
		next := curr.Next
		if curr.Array.Mark != 1 {
			if prev == nil {
				ListHead = next
			} else {
				prev.Next = next
			}
			hashDelete(curr.Array)
			curr = nil
		} else {
			curr.Array.Mark = 0
			prev = curr
		}
		curr = next
	}
}

func head(t *Hash, ref Object) int {
	if ref.Tag() == NUMBER {
		return int(ref.(NumberObject).Value()) % t.Nhash
	}
	if ref.Tag() == STRING {
		v := ref.(StringObject).Value()
		h := 0
		for i := 0; i < len(v); i++ {
			h = h << 8
			vi := int(v[i] + 0)
			h = h + vi
			h = h % t.Nhash
		}
		return h
	}
	os.Stderr.WriteString(fmt.Sprintf("Invalid %s Type For Head ", ref.Tag()))
	return -1
}

func present(t *Hash, ref Object, h int) *Node {
	var n *Node
	if ref.Tag() == NUMBER {
		n = t.List[h]
		for n != nil {

			if n.Ref.Tag() == NUMBER && ref.(NumberObject).Value() == n.Ref.(NumberObject).Value() {
				return n
			}

			n = n.Next

		}
	}
	if ref.Tag() == STRING {

		n = t.List[h]
		for n != nil {

			if ref.Tag() == STRING && ref.(StringObject).Value() == n.Ref.(StringObject).Value() {
				return n
			}

			n = n.Next
		}

	}
	return n
}

func LuaHashDefine(t *Hash, ref Object) Object {
	h := head(t, ref)
	if h == -1 {
		return nil
	}

	n := present(t, ref, h)

	if n == nil {
		n = new(Node)
		n.Ref = ref

		n.Val.SetTag(NIl)

		n.Next = t.List[h]
		t.List[h] = n

	}
	return n.Val
}

func firstNode(a *Hash, h int) {
	if h < a.Nhash {

		var i int = 0

		for i = h; i < a.Nhash; i++ {
			if a.List[i] != nil {

				if a.List[i].Val.Tag() != NIl {
					LuaPushObject(a.List[i].Ref)
					LuaPushObject(a.List[i].Val)
					return
				} else {

					next := a.List[i].Next
					for next != nil && next.Val.Tag() == NIl {
						next = next.Next
					}
					if next != nil {
						LuaPushObject(next.Ref)
						LuaPushObject(next.Val)
						return
					}
				}
			}
		}
	}
	LuaPushNil()
	LuaPushNil()
}

func LuaNext() {
	var a *Hash
	o := LuaGetParam(1)
	r := LuaGetParam(2)
	if o == nil || r == nil {
		os.Stderr.WriteString("too few arguments to function `next'")
		return
	}
	if LuaGetParam(3) != nil {
		os.Stderr.WriteString("too many arguments to function `next'")
		return
	}
	if o.Tag() != ARRAY {
		os.Stderr.WriteString("first argument of function `next' is not a table")
		return
	}
	if r.Tag() == NIl {
		firstNode(a, 0)
		return
	} else {
		h := head(a, r)
		if h >= 0 {
			n := a.List[h]
			for n != nil {
				if ObjectEqual(n.Ref, r) {
					if n.Next == nil {
						firstNode(a, h+1)
						return
					} else if n.Next.Val.Tag() != NIl {
						LuaPushObject(n.Next.Ref)
						LuaPushObject(n.Next.Val)
						return
					} else {
						next := n.Next.Next
						for next != nil && next.Val.Tag() == NIl {
							next = next.Next
						}
						if next == nil {
							firstNode(a, h+1)
							return
						} else {
							LuaPushObject(next.Ref)
							LuaPushObject(next.Val)
						}
						return
					}
				}
				n = n.Next
			}
			if n == nil {
				os.Stderr.WriteString("error in function 'next': reference not found")
			}
		}
	}
}
