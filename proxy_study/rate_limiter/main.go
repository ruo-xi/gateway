package main

import (
	"context"
	"golang.org/x/time/rate"
	"log"
	"time"
)

func main() {
	l := rate.NewLimiter(2, 3)
	log.Println(l.Limit(), l.Burst())
	for i := 0; i < 100; i++ {
		log.Println("before Wait")
		c, _ := context.WithTimeout(context.Background(), time.Millisecond*1)
		if err := l.Wait(c); err != nil {
			log.Println("limiter wait er: ", err)
		}
		log.Println("after wait")

		r := l.Reserve()
		log.Println("reverse Deelay", r.Delay())

		a := l.Allow()
		log.Println("Allow", a)
	}
}
