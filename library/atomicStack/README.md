栈操作
    先入后出

### 未经实际使用

```go
package main
func main() {
	stk := NewAtomicStack[int]()
	go stk.Push(1)
	go stk.Push(2)
	go stk.Push(3)
	go stk.Push(4)

	time.Sleep(time.Second * 1)

	fmt.Println(*stk.Pop())
	fmt.Println(*stk.Pop())
	fmt.Println(*stk.Pop())
	fmt.Println(*stk.Pop())
	fmt.Println(*stk.Pop())
}
```
