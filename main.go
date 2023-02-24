package main

import (
	"fmt"
	"time"
)

// ЗАДАНИЕ:
// * сделать из плохого кода хороший;
// * важно сохранить логику появления ошибочных тасков;
// * сделать правильную мультипоточность обработки заданий.
// Обновленный код отправить через merge-request.

// приложение эмулирует получение и обработку тасков, пытается и получать и обрабатывать в многопоточном режиме
// В конце должно выводить успешные таски и ошибки выполнены остальных тасков

// A Ttype represents a meaninglessness of our life
type Ttype struct {
	id         int
	cT         string // время создания
	fT         string // время выполнения
	taskRESULT []byte
}

func taskGenerator(ch chan Ttype) {
    for {
        ct := time.Now().Format(time.RFC3339)
        if time.Now().Nanosecond()%2 > 0 { // вот такое условие появления ошибочных тасков
            ct = "Some error occured"
        }
        ch <- Ttype{cT: ct, id: int(time.Now().UnixNano())} // передаем таск на выполнение
    }
}

func main() {
    superChan := make(chan Ttype, 10)
    doneChan, errorChan := make(chan Ttype, 10), make(chan error, 10)
	closeChan := make(chan struct{})

    defer func() {
        close(superChan)
        close(doneChan)
        close(errorChan)
        close(closeChan)
    }()

    go taskGenerator(superChan)

    doneTasks := map[int]Ttype{}
	errors := []error{}
    timelock := time.After(time.Second * 3)

	go func() {
		for {
			select {
			case data := <-errorChan:
				errors = append(errors, data)

			case data := <-doneChan:
				doneTasks[data.id] = data

			case data := <-superChan:
				tt, _ := time.Parse(time.RFC3339, data.cT)
				if tt.After(time.Now().Add(-20 * time.Second)) {
					data.taskRESULT = []byte("task has been successed")
					doneChan <- data
				} else {
					data.taskRESULT = []byte("something went wrong")
					errorChan <- fmt.Errorf("Task id %d time %s, error %s", data.id, data.cT, data.taskRESULT)
				}
                data.fT = time.Now().Format(time.RFC3339Nano)
				time.Sleep(time.Millisecond * 150)

			case <-timelock:
				closeChan <- struct{}{}
			}
		}
	}()

	<-closeChan

	println("Errors:")
	for error := range errors {
		println(error)
	}

	println("Done tasks:")
	for task := range doneTasks {
		println(task)
	}
}
