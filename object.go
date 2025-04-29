package main

type ObjType int

const (
)

type Obj struct {
	Type ObjType
}

func IsObjType(value Value, Type ObjType) bool {
	return IsObj(value) && AsObj(value).Type == Type
}
