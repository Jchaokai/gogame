package apt

import (
	"fmt"
	"gogame/noise"
	"math"
	"math/rand"
)

//+ / * - sin cos tan SimpleNoise const...
/*
	leaf node (0 children)
	single node (sin,cos)
	double node (+,-)
*/
type Node interface {
	Eval(x, y float32) float32
	AddRandom(node Node)
	NodeCount() (nodeCount,nilCount int)
}


type LeafNode struct {
}

func (leaf *LeafNode) AddRandom(node Node){
	//panic("已经无法添加节点")
	fmt.Println("leaf node add random node")
}
func (leaf *LeafNode) NodeCount() (nodeCount,nilCount int){
	return 1,0
}

type SingleNode struct {
	Child Node
}
func (single *SingleNode) AddRandom(node Node) {
	if single.Child == nil {
		single.Child = node
	}else{
		single.Child.AddRandom(node)
	}
}

func (single *SingleNode) NodeCount() (nodeCount,nilCount int) {
	if single.Child == nil {
		return 1,1
	}else{
		childNodeCount,childNilCount := single.Child.NodeCount()
		return 1+childNodeCount,childNilCount
	}
}

type DoubleNode struct {
	LeftChild Node
	RightNode Node
}

func (double *DoubleNode) AddRandom(node Node) {
	r := rand.Intn(2)
	if r == 0 {
		if double.LeftChild == nil {
			double.LeftChild = node
		}else{
			double.LeftChild.AddRandom(node)
		}
	}else{
		if double.RightNode == nil {
			double.RightNode =node
		}else{
			double.RightNode.AddRandom(node)
		}
	}
}
func (double *DoubleNode) NodeCount() (nodeCount,nilCount int){
	var leftCount,leftNilCount,rightCount,rightNilCount int
	if double.LeftChild == nil {
		leftCount = 0
		leftNilCount =1
	}else{
		leftCount,leftNilCount  = double.LeftChild.NodeCount()
	}
	if double.RightNode == nil {
		rightCount =0
		rightNilCount = 1
	}else{
		rightCount,rightNilCount = double.RightNode.NodeCount()
	}
	return 1+leftCount+rightCount,leftNilCount+rightNilCount
}
type OpNoise struct {
	DoubleNode
}

func (op *OpNoise) Eval(x, y float32) float32 {
	return 80*noise.Snoise2(op.LeftChild.Eval(x, y), op.RightNode.Eval(x, y)) - 2.0
}

type OpPlus struct {
	DoubleNode
}

func (op *OpPlus) Eval(x, y float32) float32 {
	return op.LeftChild.Eval(x, y) + op.RightNode.Eval(x, y)
}

type OpMinus struct {
	DoubleNode
}

func (op *OpMinus) Eval(x, y float32) float32 {
	return op.LeftChild.Eval(x, y) - op.RightNode.Eval(x, y)
}

type OpMult struct {
	DoubleNode
}

func (op *OpMult) Eval(x, y float32) float32 {
	return op.LeftChild.Eval(x, y) * op.RightNode.Eval(x, y)
}

type OpDiv struct {
	DoubleNode
}

func (op *OpDiv) Eval(x, y float32) float32 {
	return op.LeftChild.Eval(x, y) / op.RightNode.Eval(x, y)
}

type OpAtan2 struct {
	DoubleNode
}

func (op *OpAtan2) Eval(x, y float32) float32 {
	return float32(math.Atan2(float64(y), float64(x)))
}

type OpAtan struct {
	SingleNode
}

func (op *OpAtan) Eval(x, y float32) float32 {
	return float32(math.Atan(float64(op.Child.Eval(x, y))))
}

type OpSin struct {
	SingleNode
}

func (op *OpSin) Eval(x, y float32) float32 {
	return float32(math.Sin(float64(op.Child.Eval(x, y))))
}

type OpCos struct {
	SingleNode
}

func (op *OpCos) Eval(x, y float32) float32 {
	return float32(math.Cos(float64(op.Child.Eval(x, y))))
}

type OpX struct {
	LeafNode
}

func (op *OpX) Eval(x, y float32) float32 {
	return x
}

type OpY struct {
	LeafNode
}

func (op *OpY) Eval(x, y float32) float32 {
	return y
}

type OpConst struct {
	LeafNode
	value float32
}

func (op *OpConst) Eval(x, y float32) float32 {
	return op.value
}


func GetRandomNode() Node{
	r := rand.Intn(9)
	switch r {
	case 0:
		return &OpPlus{}
	case 1:
		return &OpMinus{}
	case 2:
		return &OpMult{}
	case 3:
		return &OpDiv{}
	case 4 :
		return &OpAtan2{}
	case 5:
		return &OpAtan{}
	case 6:
		return &OpCos{}
	case 7:
		return &OpSin{}
	case 8:
		return &OpNoise{}
	}
	panic("get random node err")
}

func GetRandomLeaf() Node {
	r := rand.Intn(3)
	switch r {
	case 0:
		return &OpX{}
	case 1:
		return &OpY{}
	case 2:
		return &OpConst{LeafNode{}, rand.Float32()*2 - 1}
	}
	panic("get random leaf err")
}
