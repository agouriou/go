package main

import (
	"fmt"
	"github.com/golang/tour/tree"
)

// Walk walks the tree t sending all values
// from the tree to the channel ch.
func Walk(t *tree.Tree, ch chan int) {
	ch <- t.Value
	if t.Left != nil {
		Walk(t.Left, ch)
	}
	if t.Right != nil {
		Walk(t.Right, ch)
	}
}

var elemMap = make(map[int]int)

func consume(ch chan int){
	elem, opened := <-ch
	for opened {
		fmt.Printf("%v, %v\n", elem, opened)
		_, elemInMap := elemMap[elem]
		fmt.Printf("elemInMap %v\n", elemInMap)
		if elemInMap {
			delete(elemMap, elem)
		} else {
			elemMap[elem] = 1		
		}
		elem, opened = <-ch
		fmt.Printf("%v, %v\n", elem, opened)
	}
}

// Same determines whether the trees
// t1 and t2 contain the same values.
func Same(t1, t2 *tree.Tree) bool {
	
	ch := make(chan int)
	go consume(ch)
	
	Walk(t1, ch)
	Walk(t2, ch)
	

	for k, _ := range elemMap {
		fmt.Printf("map %v\n", k)
	}
	return len(elemMap) == 0
}

func main() {
	var t *tree.Tree = tree.New(1)
	fmt.Println(Same(t, t))
	//fmt.Println(Same(tree.New(1), tree.New(1)))
	//fmt.Println(Same(tree.New(2), tree.New(1)))
}

