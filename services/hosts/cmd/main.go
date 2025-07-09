package main

import "fmt"

var nums = []int{0, 2, 4, 6, 8, 10}
var target = 6

func find(arr []int, k int) []int {
    items := make(map[int]int)
	for i, v := range arr {
		if j, ok := items[k-v]; ok {
			return []int{i, j}
		}
		items[v] = i
	}


    return nil
}

func main() {
    fmt.Println(find(nums, target))
}