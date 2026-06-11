package Display

import (
	"encoding/binary"
	"io"
	"log"
	"math"
	"os"
)

type Pair struct {
	Idx    uint16
	Length uint32
}

type Output interface {
	float32 | uint16 | Pair
}

type MetaData struct {
	Width            uint32
	Height           uint32
	FrameIdx         uint32
	HardwareFrameIdx uint32
	Timestamp        uint64
}

const (
	WindowSizeFloat  int64 = 4
	WindowSizeUint16       = 2
	WindowSizeRLE          = 6
	WindowSizeMap          = 10
	WindowSizeMapRLE       = 4
)

const HeaderOffset int64 = 24 // Should be about 16? or more since I think there's a little more data than expected added on

func OutputData(file string, out []byte) {
	f, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	f.Write(out)
}

func LoadThermal(path string) (*os.File, int64) {
	// Load the data file
	file, err := os.OpenFile(path, os.O_RDONLY, os.ModeAppend)
	if err != nil {
		log.Fatal(err)
	}

	// Convert the byte array into an array of float 32s

	info, err := file.Stat()
	if err != nil {
		log.Fatal(info)
	}

	return file, info.Size()
}

func LoadMetaData(file *os.File) MetaData {
	data := [HeaderOffset]byte{} // Doing this to try to allocate this on the stack instead of the heap
	_, err := file.ReadAt(data[0:], 0)
	if err != nil {
		log.Fatalln(err)
	}

	// TODO: Reduce the size of each type
	return MetaData{
		Width:            binary.LittleEndian.Uint32((data)[:4]),
		Height:           binary.LittleEndian.Uint32((data)[4:8]),
		FrameIdx:         binary.LittleEndian.Uint32((data)[8:12]),
		HardwareFrameIdx: binary.LittleEndian.Uint32((data)[12:16]),
		Timestamp:        binary.LittleEndian.Uint64((data)[16:24]),
	}
}

func FileToValue[V Output](file *os.File, size int64) (data []V) {
	defer file.Close()

	var windowSize int64

	switch any(data).(type) {
	case []uint16:
		windowSize = WindowSizeUint16
	case []float32:
		windowSize = WindowSizeFloat
	case []Pair:
		windowSize = WindowSizeRLE
	}

	window := make([]byte, int(windowSize))

	data = make([]V, int(size/windowSize))

	for i := range data {
		_, err := file.ReadAt(window, int64(i)*windowSize)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		switch d := any(&data).(type) {
		case *[]float32:
			(*d)[i] = math.Float32frombits(binary.LittleEndian.Uint32(window))
		case *[]uint16:
			(*d)[i] = binary.LittleEndian.Uint16(window[:2])
		case *[]Pair:
			(*d)[i] = Pair{
				Idx:    binary.LittleEndian.Uint16(window[:2]),
				Length: binary.LittleEndian.Uint32(window[2:6]),
			}
		}
	}
	return
}

// // TODO: Make sure size is adjusted according to if the file is formatted or not
// func FileToFloat(file *os.File, size int64) (data []float32) {
// 	defer file.Close()

// 	windowSize := 4
// 	window := make([]byte, windowSize)
// 	data = make([]float32, (size / int64(windowSize)))

// 	for i := range data {
// 		_, err := file.ReadAt(window, int64(i*windowSize))
// 		if err != nil {
// 			if err == io.EOF {
// 				break
// 			}
// 			log.Fatal(err)
// 		}
// 		data[i] = math.Float32frombits(binary.LittleEndian.Uint32(window))
// 	}
// 	return
// }

// func FileToInt(file *os.File, size int64) (data []uint16) {
// 	defer file.Close()

// 	windowSize := 2
// 	window := make([]byte, windowSize)
// 	data = make([]uint16, (size / int64(windowSize)))

// 	for i := range data {
// 		_, err := file.ReadAt(window, int64(i*windowSize))
// 		if err != nil {
// 			if err == io.EOF {
// 				break
// 			}
// 			log.Fatal(err)
// 		}
// 		data[i] = binary.LittleEndian.Uint16(window[:2])
// 	}
// 	return
// }

// func FileToRLE(file *os.File, size int64) (data []Pair) {
// 	defer file.Close()

// 	windowSize := 6
// 	window := make([]byte, windowSize)
// 	data = make([]Pair, int(size/int64(windowSize)))

// 	for i := range data {
// 		_, err := file.ReadAt(window, int64(i*windowSize))
// 		if err != nil {
// 			if err == io.EOF {
// 				break
// 			}
// 			log.Fatal(err)
// 		}

// 		data[i] = Pair{
// 			Idx:    binary.LittleEndian.Uint16(window[:2]),
// 			Length: binary.LittleEndian.Uint32(window[2:6]),
// 		}
// 	}
// 	return
// }

func FileToMap(file *os.File, size int64) (data map[uint32]Pair) {
	defer file.Close()

	windowSize := 10
	window := make([]byte, windowSize)
	data = make(map[uint32]Pair, (size / int64(windowSize)))

	for i := range len(data) {
		_, err := file.ReadAt(window, int64(i*windowSize))
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		data[binary.LittleEndian.Uint32(window[:4])] = Pair{
			Idx:    binary.LittleEndian.Uint16(window[4:6]),
			Length: binary.LittleEndian.Uint32(window[6:10]),
		}
	}
	return
}

func FileToMapRLE(file *os.File, size int64, mp map[uint32]Pair) (data []Pair) {
	defer file.Close()

	windowSize := 10
	window := make([]byte, windowSize)
	data = make([]Pair, (size / int64(windowSize)))

	for i := range data {
		_, err := file.ReadAt(window, int64(i*windowSize))
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		data[binary.LittleEndian.Uint32(window[:4])] = Pair{
			Idx:    binary.LittleEndian.Uint16(window[4:6]),
			Length: binary.LittleEndian.Uint32(window[6:10]),
		}
	}
	return
}

// Note: You have to manually allocate out
func RLEToMapAndRLE(rle *[]Pair, out *map[Pair]uint32) (mapRLE []uint32) {
	mapRLE = make([]uint32, len(*rle))

	for i, v := range *rle {
		x, found := (*out)[v]
		if !found {
			x = uint32(len(*out))
			(*out)[v] = x
		}

		mapRLE[i] = x
	}
	return
}

func MapRLEToBytes(rle *[]uint32) (rleBytes []byte) {
	rleBytes = make([]byte, len(*rle)*4)

	for i, v := range *rle {
		c := i * 4
		binary.LittleEndian.PutUint32(rleBytes[c:c+4], v)
	}

	return
}

func MapToBytes(mp *map[Pair]uint32) (mapBytes []byte) {
	mapBytes = make([]byte, len(*mp)*10)
	i := 0
	for pair, idx := range *mp {
		c := i * 10
		binary.LittleEndian.PutUint32(mapBytes[c:c+4], idx)
		binary.LittleEndian.PutUint16(mapBytes[c+4:c+6], pair.Idx)
		binary.LittleEndian.PutUint32(mapBytes[c+6:c+10], pair.Length)
	}
	return
}

func FloatToInt(data *[]float32) (out []uint16) {
	out = make([]uint16, len(*data))
	for i, v := range *data {
		out[i] = uint16(float32(v * 10))
		// fmt.Println(out[i])
	}
	return out
}

func IntToFloat(data *[]uint16) (out []float32) {
	out = make([]float32, len(*data))
	for i, v := range *data {
		out[i] = float32(v) * .1
		// fmt.Println(out[i])
	}
	return out
}

func IntToBytes(data *[]uint16) (out []byte) {
	out = make([]byte, len(*data)*2)
	for i, v := range *data {
		binary.LittleEndian.PutUint16(out[i:i+2], v)
	}
	return out
}

func FloatToRLE(data *[]float32) []Pair {
	i := FloatToInt(data)
	return IntToRLE(&i)
}

func IntToRLE(data *[]uint16) (out []Pair) {
	out = make([]Pair, len(*data))

	out[0] = Pair{Idx: (*data)[0], Length: 0}
	current_pair := uint16(0)

	// Loop through data
	for _, val := range *data {
		if out[current_pair].Idx == val { // When we see old values, we add to the length of the current pair
			out[current_pair].Length++
		} else { // When we see new values, we create a new pair with a length of 1
			current_pair++
			out[current_pair] = Pair{Idx: val, Length: 1}
		}
	}
	out = out[:current_pair+1]

	return
}

func RLEToBytes(pairs *[]Pair) (out []byte) {
	out = make([]byte, (len(*pairs) * 6))
	c := 0
	for i, v := range *pairs {
		c = i * 6
		binary.LittleEndian.PutUint16(out[c:c+2], v.Idx)
		binary.LittleEndian.PutUint32(out[c+2:c+6], v.Length)
	}
	return
}

// When passing in Data, you should only pass the data section
func BytesToRLE(data []byte) (rle []Pair) {
	rle = make([]Pair, len(data)/6)
	for i := range rle {
		c := i * 6
		rle[i] = Pair{
			Idx:    binary.LittleEndian.Uint16(data[c : c+2]),
			Length: binary.LittleEndian.Uint32(data[c+2 : c+4]),
		}
	}
	return
}

func RLEToInt(rle *[]Pair) (out []uint16) {
	for _, v := range *rle {
		for range v.Length {
			out = append(out, v.Idx)
		}
	}
	return
}
