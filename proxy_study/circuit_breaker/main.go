package main

import (
	"errors"
	"github.com/afex/hystrix-go/hystrix"
	"log"
	"net/http"
	"time"
)

func main() {
	h := hystrix.NewStreamHandler()
	h.Start()
	go http.ListenAndServe(":1234", h)

	hystrix.ConfigureCommand("aaa", hystrix.CommandConfig{
		Timeout:                1000,
		MaxConcurrentRequests:  1,
		SleepWindow:            5000,
		RequestVolumeThreshold: 1,
		ErrorPercentThreshold:  1,
	})

	for i := 0; i < 1000; i++ {
		err := hystrix.Do("aaa", func() error {
			if i == 0 {
				return errors.New("service error")
			}
			log.Println("do service")
			return nil
		}, nil)
		if err != nil {
			log.Println("hystrix err:", err.Error())
			time.Sleep(time.Second * 1)
			log.Println("sleep 1 second")
		}
	}
	time.Sleep(time.Second * 100)


}
