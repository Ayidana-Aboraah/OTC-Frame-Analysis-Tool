package Display

import (
	"encoding/binary"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

type IOpair struct {
	in  string
	out string
}
type Result struct {
	file IOpair
	err  error
}

func LFTB[in Output, out Output](tasks []IOpair) {
	workers := runtime.NumCPU()
	jobs := make(chan IOpair, len(tasks))
	results := make(chan Result, len(tasks))
	var wg sync.WaitGroup

	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				err := ContinousFrameSetToBin[in, out](job.in, job.out)
				if err != nil {
					results <- Result{file: job, err: err}
				}
			}
		}()
	}

	for _, task := range tasks {
		jobs <- task
	}
	close(jobs)

	go func() {
		wg.Wait()
		close(results)
	}()

	// TODO: List info from results upon completion
}

func ContinousFrameSetToBin[in Output, outT Output](path, outPath string) error {
	out, err := os.Create(outPath)
	if err != nil {
		return (err)
	}

	dir, err := os.ReadDir(path)
	if err != nil {
		return (err)
	}

	header := [2]uint16{}

	for i, entry := range dir {
		var extension string
		var a in

		switch any(a).(type) {
		case float32:
			extension = "base"
		case uint16:
			extension = "int"
		case Pair:
			extension = "rle"
		}

		if entry.IsDir() || !strings.Contains(entry.Name(), extension) {
			continue
		}

		cf, size := LoadThermal(path + "/" + entry.Name())

		if i == 0 {
			metadata := LoadMetaData(cf)
			header[0] = uint16(metadata.Width)
			header[1] = uint16(metadata.Height)

			var buf [4]byte
			binary.LittleEndian.PutUint16(buf[:2], header[0])
			binary.LittleEndian.PutUint16(buf[2:], header[1])
			_, err = out.Write(buf[:])
			if err != nil {
				return (err)
			}
		}

		b := FileToValue[in](cf, size)
		o := ValueToValue[in, outT](&b) // TODO: Adjust how o handles the in & out type being the same
		out.Write(ValueToBytes(&o))
	}
	return nil
}

func LoadFrameSetFiles(startFrame int, endFrame int, extension, path string) (files []*os.File, sizes []int64) {
	files = make([]*os.File, endFrame-startFrame)
	sizes = make([]int64, endFrame-startFrame)
	current_file := 0

	dir, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range dir {

		if entry.IsDir() || !strings.Contains(entry.Name(), extension) {
			continue
		}

		pieces := strings.Split(entry.Name(), "_")
		frameIdx, err := strconv.ParseInt(pieces[1], 10, 64)
		if err != nil {
			log.Fatalln(err)
		}

		if frameIdx < int64(startFrame) || frameIdx > int64(endFrame) {
			continue
		}

		files[current_file], sizes[current_file] = LoadThermal(path + "/" + entry.Name())
		current_file++
	}

	return
}

func FilesToValue[V Output](files []*os.File, sizes []int64) (set [][]V) {
	set = make([][]V, len(files))
	for i := range set {
		set[i] = FileToValue[V](files[i], sizes[i]-HeaderOffset)
	}
	return
}
