package Display

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"reflect"
)

type Pair struct {
	Idx    uint16
	Length uint32
}

type Output interface {
	float32 | uint16 | Pair | uint32
}

type Decoder map[uint32]Pair

type Encoder map[Pair]uint32

type MetaData struct {
	Width            uint32
	Height           uint32
	FrameIdx         uint32
	HardwareFrameIdx uint32
	Timestamp        uint64
}

const (
	WindowSizeFloat  = 4
	WindowSizeUint16 = 2
	WindowSizeRLE    = 6
	WindowSizeMapRLE = 4
	WindowSizeMap    = 10
)

const HeaderOffset int64 = 24 // Should be about 16? or more since I think there's a little more data than expected added on

var RLEMap map[Pair]uint32

func MapFromDecoder(decoder *map[uint32]Pair) {
	RLEMap = make(map[Pair]uint32, len(*decoder))
	for k, v := range *decoder {
		RLEMap[v] = k
	}
}

func MapToDecoder() (decoder map[uint32]Pair) {
	decoder = make(map[uint32]Pair, len(RLEMap))
	for k, v := range RLEMap {
		decoder[v] = k
	}
	return
}

func OutputData(file string, out []byte) {
	f, err := os.Create(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	f.Write(out)
}

func LoadThermal(path string) (*os.File, int64) {
	file, err := os.OpenFile(path, os.O_RDONLY, os.ModeAppend)
	if err != nil {
		log.Fatal(err)
	}

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

func FileToValue[V Output | Decoder](file *os.File, size int64) (data []V) {
	defer file.Close()

	var windowSize int

	switch any(data).(type) {
	case []uint16:
		windowSize = WindowSizeUint16
	case []float32:
		windowSize = WindowSizeFloat
	case []Pair:
		windowSize = WindowSizeRLE
	case []uint32:
		windowSize = WindowSizeMapRLE
	case []Decoder:
		windowSize = WindowSizeMap
	}

	window := make([]byte, windowSize)

	data = make([]V, int(size)/windowSize)

	for i := range data {
		_, err := file.ReadAt(window, int64(i*windowSize))
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
			(*d)[i] = binary.BigEndian.Uint16(window[:2])
		case *[]Pair:
			(*d)[i] = Pair{
				Idx:    binary.LittleEndian.Uint16(window[:2]),
				Length: binary.LittleEndian.Uint32(window[2:6]),
			}
		case *[]uint32:
			(*d)[i] = binary.LittleEndian.Uint32(window[:4])
		case *[]Decoder:
			(*d)[0][binary.LittleEndian.Uint32(window[:4])] = Pair{
				Idx:    binary.LittleEndian.Uint16(window[4:6]),
				Length: binary.LittleEndian.Uint32(window[6:10]),
			}
		}
	}
	return
}

func ByteToValue[V Output | Decoder](data *[]byte) (out []V) {
	var windowSize int

	switch any(data).(type) {
	case []uint16:
		windowSize = WindowSizeUint16
	case []float32:
		windowSize = WindowSizeFloat
	case []Pair:
		windowSize = WindowSizeRLE
	case []uint32:
		windowSize = WindowSizeMapRLE
	case []Decoder:
		windowSize = WindowSizeMap
	}

	out = make([]V, len(*data)/windowSize)
	for i := range out {
		c := i * 6
		switch o := any(&out).(type) {
		case *[]float32:
			(*o)[i] = math.Float32frombits(binary.LittleEndian.Uint32((*data)[c : c+4]))
		case *[]uint16:
			(*o)[i] = binary.LittleEndian.Uint16((*data)[c : c+2])
		case *[]Pair:
			(*o)[i] = Pair{
				Idx:    binary.LittleEndian.Uint16((*data)[c : c+2]),
				Length: binary.LittleEndian.Uint32((*data)[c+2 : c+4]),
			}
		case *[]uint32:
			(*o)[i] = binary.LittleEndian.Uint32((*data)[c : c+4])
		case *[]Decoder:
			(*o)[0][binary.LittleEndian.Uint32((*data)[c:c+4])] = Pair{
				Idx:    binary.LittleEndian.Uint16((*data)[c+4 : c+6]),
				Length: binary.LittleEndian.Uint32((*data)[c+6 : c+10]),
			}
		}
	}
	return
}

func ValueToBytes[V Output | map[uint32]Pair](value *[]V) (out []byte) {
	var windowSize int

	switch any((*value)).(type) {
	case []uint16:
		windowSize = WindowSizeUint16
	case []float32:
		windowSize = WindowSizeFloat
	case []Pair:
		windowSize = WindowSizeRLE
	case []uint32:
		windowSize = WindowSizeMapRLE
	case []map[Pair]uint32:
		windowSize = WindowSizeMap
	}

	out = make([]byte, (len(*value) * windowSize))
	c := 0
	for i, v := range *value {
		c = i * windowSize
		switch x := any((v)).(type) {
		case float32:
			binary.LittleEndian.PutUint32(out[c:c+4], math.Float32bits(x))
		case uint16:
			binary.LittleEndian.PutUint16(out[c:c+2], x)
		case Pair:
			binary.LittleEndian.PutUint16(out[c:c+2], x.Idx)
			binary.LittleEndian.PutUint32(out[c+2:c+6], x.Length)
		case uint32:
			binary.LittleEndian.PutUint32(out[c:c+4], x)
		case map[uint32]Pair:
			i := 0
			for idx, pair := range x {
				c := i * 10
				binary.LittleEndian.PutUint32(out[c:c+4], idx)
				binary.LittleEndian.PutUint16(out[c+4:c+6], pair.Idx)
				binary.LittleEndian.PutUint32(out[c+6:c+10], pair.Length)
			}
		}
	}
	return
}

func ValueToValue[In Output, Out Output](value *[]In) (out []Out) {
	inputType := reflect.TypeOf(*value)
	outputType := reflect.TypeOf(out)

	a := !(inputType == reflect.TypeFor[[]Pair]() && outputType == reflect.TypeFor[[]uint32]())
	b := !(inputType == reflect.TypeFor[[]uint32]() && outputType == reflect.TypeFor[[]Pair]())

	// TODO: Evaluate the different cases in which we would not want to pre-allocate our output type
	if inputType != outputType && a && b {
		out = make([]Out, len(*value))
	}

	// Translation Functions

	FtoI := func(in *[]float32, out *[]uint16) {
		for i, v := range *in {
			(*out)[i] = uint16(float32(v * 10))
		}
	}
	ItoF := func(in *[]uint16, out *[]float32) {
		for i, v := range *in {
			(*out)[i] = float32(v) * .1
		}
	}
	ItoRLE := func(in *[]uint16, out *[]Pair) {
		(*out)[0] = Pair{Idx: (*in)[0], Length: 0}
		current_pair := uint16(0)

		for _, val := range *in {
			if (*out)[current_pair].Idx == val { // When we see old values, we add to the length of the current pair
				(*out)[current_pair].Length++
			} else { // When we see new values, we create a new pair with a length of 1
				current_pair++
				(*out)[current_pair] = Pair{Idx: val, Length: 1}
			}
		}
		fmt.Println("Current Pair Idx: ", current_pair)
		*out = (*out)[:current_pair+1]
	}
	RLEtoI := func(in *[]Pair, out *[]uint16) {
		total := 0
		for _, v := range *in {
			for range v.Length {
				total += int(v.Length)
				(*out) = append(*out, v.Idx)
			}
		}
		fmt.Println("Total RLE vals: ", total)
	}
	RLEtoMapRLE := func(in *[]Pair, out *[]uint32) {
		for i, v := range *in {
			x, found := (RLEMap)[v]
			if !found {
				x = uint32(len(RLEMap))
				(RLEMap)[v] = x
			}
			(*out)[i] = x
		}
	}

	MapRLEtoRLE := func(in *[]uint32, out *[]Pair) {
		decoder := MapToDecoder()
		for i, v := range *in {
			x := (decoder)[v]
			(*out)[i] = x
		}
	}

	switch in := any(value).(type) {
	case *[]float32:
		switch o := any(&out).(type) {
		case *[]float32:
			(*o) = *in
		case *[]uint16:
			FtoI(in, o)
		case *[]Pair:
			i := make([]uint16, len(*in))
			FtoI(in, &i)
			ItoRLE(&i, o)
		case *[]uint32: // TODO: Use multiple translation functions to get here
			i := make([]uint16, len(*in))
			rle := make([]Pair, len(i))
			FtoI(in, &i)
			ItoRLE(&i, &rle)
			RLEtoMapRLE(&rle, o)
		}

	case *[]uint16:
		switch o := any(&out).(type) {
		case *[]float32:
			ItoF(in, o)
		case *[]uint16:
			(*o) = *in
		case *[]Pair:
			ItoRLE(in, o)
		case *[]uint32:
			rle := make([]Pair, len(*in))
			ItoRLE(in, &rle)
			RLEtoMapRLE(&rle, o)
		}

	case *[]Pair:
		switch o := any(&out).(type) {
		case *[]float32:
			i := []uint16{}
			RLEtoI(in, &i)
			ItoF(&i, o)
		case *[]uint16:
			RLEtoI(in, o)
		case *[]Pair:
			(*o) = *in
		case *[]uint32:
			RLEtoMapRLE(in, o)
		}

	case *[]uint32:
		switch o := any(&out).(type) {
		case *[]float32:
			rle := make([]Pair, len(*in))
			i := []uint16{}

			MapRLEtoRLE(in, &rle)
			RLEtoI(&rle, &i)
			ItoF(&i, o)

		case *[]uint16:
			rle := make([]Pair, len(*in))
			MapRLEtoRLE(in, &rle)
			RLEtoI(&rle, o)

		case *[]Pair:
			MapRLEtoRLE(in, o)

		case *[]uint32:
			(*o) = *in
		}
	}
	return
}
