package atomicStack

import (
	"sync/atomic"
	"unsafe"
)

// NewAtomicStack 初始化一个栈 并将栈顶设置为空
func NewAtomicStack[V any]() *AtomicStack[V] {
	return &AtomicStack[V]{head: unsafe.Pointer(&nodeData[V]{})}
}

// AtomicStack 栈顶指针
type AtomicStack[V any] struct {
	head unsafe.Pointer
}

// nodeData 数据类型及下一个指针
type nodeData[V any] struct {
	val  V
	next unsafe.Pointer
}

// Push 入栈
func (s *AtomicStack[V]) Push(val V) {
	for {
		// 获取栈顶
		top := (*nodeData[V])(atomic.LoadPointer(&s.head))
		// 创建新节点
		newNode := &nodeData[V]{val: val, next: unsafe.Pointer(top)}
		if atomic.CompareAndSwapPointer(&s.head, unsafe.Pointer(top), unsafe.Pointer(newNode)) {
			return
		}
	}
}

// Pop 出栈
func (s *AtomicStack[V]) Pop() *V {
	for {
		// 获取栈顶
		top := (*nodeData[V])(atomic.LoadPointer(&s.head))
		if top == nil { //如果栈顶为空
			return nil
		}
		// 获取栈顶的下一个节点
		next := (*nodeData[V])(atomic.LoadPointer(&top.next))
		if atomic.CompareAndSwapPointer(&s.head, unsafe.Pointer(top), unsafe.Pointer(next)) {
			return &top.val
		}
	}
}
