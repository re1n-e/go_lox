package main

const (
	OP_CONSTANT = iota
	OP_NIL
	OP_TRUE
	OP_FALSE
	OP_POP
	OP_DEFINE_GLOBAL
	OP_EQUAL
	OP_GREATER
	OP_LESS
	OP_ADD
	OP_SUBTRACT
	OP_MULTIPLY
	OP_DIVIDE
	OP_NOT
	OP_NEGATE
	OP_PRINT
	OP_RETURN
)

// ValueType represents the type of a value in the VM
type ValueType int

const (
	VAL_BOOL ValueType = iota
	VAL_NIL
	VAL_NUMBER
	VAL_STRING
	VAL_OBJ
)

// Value represents any value that can be stored in the VM
type Value struct {
	Type ValueType
	Bool bool
	Num  float64
	String string
	obj  Obj
}

func OBJ_TYPE(value Value) ObjType {
	return AsObj(value).Type
}

func BoolVal(b bool) Value {
	return Value{Type: VAL_BOOL, Bool: b}
}

func NilVal() Value {
	return Value{Type: VAL_NIL}
}

func NumberVal(n float64) Value {
	return Value{Type: VAL_NUMBER, Num: n}
}

func StringVal(s string) Value {
	return Value{Type: VAL_STRING, String: s}
}

func ObjVal(object Obj) Value {
	return Value{Type: VAL_OBJ, obj: object}
}

func IsBool(value Value) bool {
	return value.Type == VAL_BOOL
}

func IsNil(value Value) bool {
	return value.Type == VAL_NIL
}

func IsNumber(value Value) bool {
	return value.Type == VAL_NUMBER
}

func IsString(value Value) bool {
	return value.Type == VAL_STRING
}

func IsObj(value Value) bool {
	return value.Type == VAL_OBJ
}

func AsBool(value Value) bool {
	return value.Bool
}

func AsNumber(value Value) float64 {
	return value.Num
}

func AsString(value Value) string {
	return value.String
}

func AsObj(value Value) Obj {
	return value.obj
}

func valuesEqual(a Value, b Value) bool {
	if a.Type != b.Type {
		return false
	}
	switch a.Type {
	case VAL_BOOL:
		return AsBool(a) == AsBool(b)
	case VAL_NIL:
		return true
	case VAL_NUMBER:
		return AsNumber(a) == AsNumber(b)
	default:
		return false
	}
}
