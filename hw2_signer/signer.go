package main

import (
	"sync"
	"strings"
	"strconv"
	"sort"
)

type orderedMsg struct {
	data string
	order int
}

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
	crcCh := make(chan orderedMsg)
	defer close(crcCh)

	dataStr := ""
	cnt := 0
	for data := range in {
		dataStr = strconv.Itoa(data.(int))
		md5 := DataSignerMd5(dataStr)
		go calcFirstStep(crcCh, dataStr, md5, cnt)
		cnt++
	}

	crcResults := make([]string, cnt, 100)

	for i:=0; i<cnt; i++ {
		result := <- crcCh
		crcResults[result.order] = result.data
	}

	out <-crcResults
}

func calcFirstStep(out chan orderedMsg, data string, md5 string, order int) {
	ch1 := make(chan orderedMsg)
	ch2 := make(chan orderedMsg)

	go calcCRC(data, ch1, 0)
	go calcCRC(md5, ch2, 0)

	crc := <-ch1
	close(ch1)

	crcmd := <-ch2
	close(ch2)

	out <- orderedMsg{data: strings.Join([]string{crc.data, crcmd.data}, "~"), order: order}
}

func calcCRC(data string, out chan<- orderedMsg, order int) {
	out <- orderedMsg{data: DataSignerCrc32(data), order: order}
}

func MultiHash(in, out chan interface{}) {
	incomeArr := <-in

	ch := make(chan orderedMsg)
	cnt := 0
	for _, data := range incomeArr.([]string) {
		go calcSecondStep(data, ch, cnt)
		cnt++
	}

	results := make([]string, cnt, 100)

	for i:=0; i<cnt; i++ {
		res := <-ch
		results[res.order] = res.data
	}

	out <- results
}

func calcSecondStep(data string, out chan orderedMsg, order int) {
	NUM := 6
	ch := make(chan orderedMsg)
	results := make([]string, NUM)
	for i:=0; i<NUM; i++ {
		go calcCRC(strconv.Itoa(i) + data, ch, i)
	}

	for i:=0; i<NUM; i++ {
		res := <-ch
		results[res.order] = res.data
	}

	out <- orderedMsg{data: strings.Join(results, ""), order: order}
}

func CombineResults(in, out chan interface{}) {
	results := <-in
	sort.Strings(results.([]string))
	out <- strings.Join(results.([]string), "_")
}