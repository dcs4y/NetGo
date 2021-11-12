package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"strconv"
)

func main() {
	b, err := ioutil.ReadFile("banner.txt")
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println(string(b))
	r := bufio.NewReader(bytes.NewReader(b))
	lines := []string{}
	i := 0
	for {
		i++
		l, _, err := r.ReadLine()
		if err != nil {
			fmt.Println(err)
			break
		} else {
			if i <= 20 {
				lines = append(lines, fillString(string(l)))
			} else {
				fmt.Println(lines[i%20] + string(l))
			}
		}
	}
	fmt.Println("ok")
}

func fillString(s string) string {
	count := 120
	if len(s) < count {
		return s + fmt.Sprintf("%"+strconv.Itoa(count-len(s))+"s", "")
	}
	return s
}
