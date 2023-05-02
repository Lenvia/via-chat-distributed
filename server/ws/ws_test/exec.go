package main

import (
	"log"
	"os/exec"
	"strconv"
	"sync"
)

var wg = sync.WaitGroup{}
var ch = make(chan int, 20)

func main() {

	for i := 500; i <= 600; i++ {
		wg.Add(1)
		go execCommand(i)
	}

	wg.Wait()

	log.Println("okkkkkkkkkkkkkkkkkk")
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
	//cmd := exec.Command("./mock_ws_client_coon.exe", strI)
	cmd := exec.Command("go",
		"run",
		"/Users/yy/GithubProjects/via-web/ws/ws_test/mock_ws_client_coon.go",
		strI)

	err := cmd.Start()

	if err != nil {
		log.Println(err)
	}

	//time.Sleep(time.Second * 1)
	<-ch
	wg.Done()
}
