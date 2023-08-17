package main

import (
	"context"
	"fmt"
	"log"
	"sync"
)

const writeLimit = 10

func main() {
	ctx := context.Background()
	// multiplier := 2
	poolSize := 2
	fmt.Println(poolSize)

	list := createList()
	chunks := chunkSlice(list, writeLimit)
	fmt.Println(chunks)

	ch := make(chan []int, poolSize)

	results := make(chan bool, poolSize)
	defer close(results)

	ctxCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	go sendChunks(ctxCancel, ch, chunks)

	// jobs:
	// 	for {
	// 		val, ok := <-ch
	// 		if !ok {
	// 			break jobs
	// 		}
	// 		fmt.Println(val)
	// 	}

	var wg sync.WaitGroup
	wg.Add(poolSize)

	for i := 0; i < poolSize; i++ {
		go func(ch <-chan []int, i int) {
			write(ch, i)
			wg.Done()
		}(ch, i)
	}
	wg.Wait()
}

func write(ch <-chan []int, i int) {
	defer func() {
		log.Println("shutting down goroutine ", i)
	}()

job:
	for {
		val, ok := <-ch
		if !ok {
			break job
		}
		fmt.Println(val)
	}
}

func createList() []int {
	var list []int
	for i := 0; i < 100; i++ {
		list = append(list, i)
	}
	return list
}

func chunkSlice(list []int, limit int) [][]int {
	var chunks [][]int
	for i := 0; i < len(list); i += limit {
		end := i + limit

		if end > len(list) {
			end = len(list)
		}
		chunks = append(chunks, list[i:end])
	}
	fmt.Println(chunks)
	return chunks
}

func sendChunks(ctx context.Context, ch chan<- []int, chunks [][]int) {
	defer close(ch)
	numChunks := len(chunks)
	log.Printf("got %d chunks to send\n", numChunks)

	// check if context is done or if is possible to send chunk to channel
	for i := 0; i < numChunks; i++ {
		select {
		case <-ctx.Done():
			log.Println("killing sendChunks!!! context with cancel is done.")
			return
		case ch <- chunks[i]:
			log.Println("sent chunck: ", i+1)
		}
	}
	log.Println("all chuncks were sent. closing channel.")
}
