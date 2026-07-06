package Display

import (
	"encoding/binary"
	"log"
	"os"
	"strconv"
	"strings"
)

func ContinousFrameSetToBin[v Output](path, outPath string) {
	out, err := os.Create(outPath)
	if err != nil {
		panic(err)
	}

	dir, err := os.ReadDir(path)
	if err != nil {
		panic(err)
	}

	header := [2]uint16{}

	for i, entry := range dir {
		var extension string
		var a v

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
				panic(err)
			}
		}

		b := FileToValue[v](cf, size)
		out.Write(ValueToBytes(&b))
	}
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
