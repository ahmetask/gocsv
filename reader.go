package gocsv

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/ahmetask/worker"
	"io"
	"math"
	"os"
	"strings"
	"sync"
)

type Reader interface {
	Read() chan OptionalRowData
	Close() error
}

type reader struct {
	fp *os.File
	ReaderConfig
	converter converter
	pool      *worker.Pool
}

type ReaderConfig struct {
	FilePath        string
	ChunkSize       int    //Default
	LineBuffer      int64  //Default
	Separator       string //Default " "
	StringSeparator string //Default ","
	ProducerBuffer  int    //Default 100
	producerChannel chan OptionalRowData
	WorkerCount     int
	Format          interface{}
	ConvertFunction ConvertField
}

func (r *ReaderConfig) Validate() error {
	if r.FilePath == "" {
		return errors.New("filePatch config value must be given")
	}

	if r.ProducerBuffer == 0 {
		r.ProducerBuffer = 100
	}

	if r.Separator == "" {
		r.Separator = " "
	}

	if r.LineBuffer == 0 {
		r.LineBuffer = 250 * 1024
	}

	if r.ChunkSize == 0 {
		r.ChunkSize = 100
	}

	if r.WorkerCount == 0 {
		r.WorkerCount = 8
	}
	return nil
}

func NewReader(config ReaderConfig) (Reader, error) {
	err := config.Validate()
	if err != nil {
		return nil, err
	}
	pool := worker.NewWorkerPool(config.WorkerCount, 100)
	pool.Start()

	c, err := newConverter(config.StringSeparator, config.Format, config.ConvertFunction)
	if err != nil {
		return nil, err
	}
	return &reader{
		ReaderConfig: config,
		converter:    c,
		pool:         pool,
	}, nil
}

func (r *reader) open() error {
	file, err := os.Open(r.FilePath)

	if err != nil {
		fmt.Println("cannot read the file", err)
		return err
	}
	r.fp = file
	return nil
}

func (r *reader) Read() chan OptionalRowData {
	ch := make(chan OptionalRowData, r.ProducerBuffer)
	readErr := r.open()
	if readErr != nil {
		close(ch)
		r.pool.Stop()
		return ch
	}

	r.producerChannel = ch

	go r.process()

	return ch
}

func (r *reader) Close() error {
	return r.fp.Close()
}

func (r *reader) process() {

	linesPool := sync.Pool{New: func() interface{} {
		lines := make([]byte, r.ReaderConfig.LineBuffer)
		return lines
	}}

	stringPool := sync.Pool{New: func() interface{} {
		lines := ""
		return lines
	}}

	br := bufio.NewReader(r.fp)

	wg := &sync.WaitGroup{}
	for {
		buf := linesPool.Get().([]byte)

		n, err := br.Read(buf)
		buf = buf[:n]

		if n == 0 {
			if err != nil {
				wg.Wait()
				close(r.producerChannel)
				break
			}
			if err == io.EOF {
				wg.Wait()
				close(r.producerChannel)
				break
			}
		}

		nextUntilNewline, err := br.ReadBytes('\n')

		if err != io.EOF {
			buf = append(buf, nextUntilNewline...)
		}

		r.processChunk(buf, &linesPool, &stringPool, wg)
	}
}

type ChunkJob struct {
	chunk      []byte
	linesPool  *sync.Pool
	stringPool *sync.Pool
	reader     *reader
	wg         *sync.WaitGroup
}

func (j *ChunkJob) Do() {
	var wg2 sync.WaitGroup

	lines := j.stringPool.Get().(string)
	lines = string(j.chunk)

	j.linesPool.Put(j.chunk)

	lineSlice := strings.Split(lines, "\n")

	j.stringPool.Put(lines)

	chunkSize := j.reader.ReaderConfig.ChunkSize
	n := len(lineSlice)
	noOfThread := n / chunkSize

	if n%chunkSize != 0 {
		noOfThread++
	}

	for i := 0; i < (noOfThread); i++ {
		wg2.Add(1)
		go func(s int, e int) {
			defer wg2.Done()
			for i2 := s; i2 < e; i2++ {
				text := lineSlice[i2]
				if len(text) == 0 {
					continue
				}

				values := strings.Split(text, j.reader.Separator)
				d, err := j.reader.converter.convert(values)
				j.reader.producerChannel <- &Row{V: d, Error: err}
			}

		}(i*chunkSize, int(math.Min(float64((i+1)*chunkSize), float64(len(lineSlice)))))
	}

	wg2.Wait()
	lineSlice = nil
	j.wg.Done()
}

func (r *reader) processChunk(chunk []byte, linesPool *sync.Pool, stringPool *sync.Pool, wg *sync.WaitGroup) {
	wg.Add(1)
	r.pool.Submit(&ChunkJob{
		chunk:      chunk,
		linesPool:  linesPool,
		stringPool: stringPool,
		reader:     r,
		wg:         wg,
	})
}
