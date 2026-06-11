package Display

import (
	"log"
	"os"
	"strconv"
	"strings"
)

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
