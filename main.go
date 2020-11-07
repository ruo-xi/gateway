package main

import (
	"fmt"
	"runtime"
)

type people struct {
	name string
	age  int
}

func (p *people) changeName() {
	p.name = "2"
}

type me struct {
	people
}

func (m me) changeName() {
	m.name = "3"
	fmt.Println(m)
}

func main() {
	p := people{
		name: "1",
		age:  1,
	}
	m := me{people: p}
	fmt.Printf("%p \n", m)
	m.changeName()
	fmt.Println(m.people)
}

func sayHello() {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			fmt.Printf("tcp: pnic serving \n %s, \n %s", err, buf)
		}
	}()
	panic("nihao")

	//http.ListenAndServe()
}
