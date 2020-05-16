package apt

import "math"

//+ / * - sin cos tan SimpleNoise const...
/*
	leaf node (0 children)
	single node (sin,cos)
	double node (+,-)
*/
type Node interface {
	Eval(x, y float32) float32
}
type LeafNode struct {
}

type SingleNode struct {
	Child Node
}

type DoubleNode struct {
	LeftChild Node
	RightNode Node
}

type OpPlus struct {
	DoubleNode
}

func (op *OpPlus) Eval(x, y float32) float32 {
	return op.LeftChild.Eval(x, y) + op.RightNode.Eval(x, y)
}

type OpSin struct {
	SingleNode
}

func (op *OpSin) Eval(x, y float32) float32 {
	return float32(math.Sin(float64(op.Child.Eval(x, y))))
}

type OpX LeafNode

func (op *OpX) Eval(x, y float32) float32 {
	return x
}

type OpY LeafNode

func (op *OpY) Eval(x, y float32) float32 {
	return y
}
