package main

import (
	"log"
	"strconv"
	"sync"
	"via-chat-distributed/ws/ws_test/test"
)

var wg = sync.WaitGroup{}
var ch = make(chan int, 20)

func main() {

	for i := 0; i <= 100; i++ {
		wg.Add(1)
		go execCommand(i)
	}

	wg.Wait()
}

func execCommand(i int) {
	defer func() {
		//捕获read抛出的panic
		if err := recover(); err != nil {
			log.Println("execCommand", err)
		}
	}()

	ch <- i
	strI := strconv.Itoa(i)

	test.StartFunc(strI)

	//time.Sleep(time.Second * 1)
	<-ch
	wg.Done()
}
