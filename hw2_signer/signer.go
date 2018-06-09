package main

import (
	"sync"
	"strings"
	"strconv"
	"sort"
)

func ExecutePipeline(args ...job) {
	wg := &sync.WaitGroup{}
	firstCh := make(chan interface{})
	prevCh := firstCh
	for _, jobFunc := range args {
		nextCh := make(chan interface{})
		wg.Add(1)
		go PipeFunc(wg, jobFunc, prevCh, nextCh)
		prevCh = nextCh
	}
	close(firstCh)
	wg.Wait()
}

func PipeFunc(wg *sync.WaitGroup, f job, in, out chan interface{}) {
	f(in, out)
	close(out)
	wg.Done()
}

func SingleHash(in, out chan interface{}) {
	data := <-in
	md5 := DataSignerMd5(data.(string))

	ch1 := make(chan string)
	ch2 := make(chan string)

	go calcCRC(data, ch1)
	go calcCRC(md5, ch2)

	crc := <-ch1
	close(ch1)

	crcmd := <-ch2
	close(ch2)

	out <- strings.Join([]string{crc, crcmd}, "~")
}

func calcCRC(data interface{}, out <-chan string) {
	out <- DataSignerCrc32(data.(string))
}

func MultiHash(in, out chan interface{}) {
	NUM := 6
	data := <-in
	channels := make([]chan string, NUM)
	results := make([]string, NUM)
	for i:=0; i<NUM; i++ {
		ch := make(chan string)
		channels[i] = ch
		go calcCRC(strconv.Itoa(i) + data.(string), ch)
	}

	for i, ch := range channels {
		results[i] = <-ch
	}

	out <- strings.Join(results, "")
}

func CombineResults(in, out chan interface{}) {
	IN_LIMIT := 100
	results := make([]string, IN_LIMIT)
	for data := range in {
		results = append(results, data.(string))
	}
	sort.Strings(results)
	out <- strings.Join(results, "_")
}