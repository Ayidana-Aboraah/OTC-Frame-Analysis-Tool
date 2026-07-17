package main

import (
	"fmt"
	"log"
	"os"
	"super/Display"
)

const fp string = "data/frame_00000000_base.bin"
const ttt string = "data/bframe_00000000_base.bin"

func main() {
	file, size := Display.LoadThermal(ttt)
	f := Display.FileToValue[float32](file, size-Display.HeaderOffset)
	Display.TemperaturesToBMP(f, 640, 480, "./out.bmp")
	// rle := Display.ValueToValue[float32, Display.Pair](&f)
	// stats := Display.FrameStatsRLE(rle)
	// fmt.Println((stats))
	// CheckAppAndCurrentRLE()

	// path := "C:/Users/ayidana.aboraah/Documents/Stuff2/Thermal Frame Analysis/data"
	// Display.ContinousFrameSetToBin[float32](path+"/Session_20260610_152352/frames", path+"/x.out")

}

type CompressionRate struct {
	BaseSize       int
	BaseIntSize    int
	BaseRLESize    int
	CompressionInt float32
	CompressionRLE float32
}

func GetGeneralCompressionRate(path string) {
	var rates []CompressionRate
	// Scan through all formatted binary files
	dir, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}

		//	With each binary Find the base size, IntCompression Size, RLE Compression Size
		file, size := Display.LoadThermal(path + "/" + entry.Name())

		f32 := Display.FileToValue[float32](file, size-Display.HeaderOffset)
		u16 := Display.ValueToValue[float32, uint16](&f32)
		rle := Display.ValueToValue[uint16, Display.Pair](&u16)

		baseSize := len(f32) * Display.WindowSizeFloat
		intSize := len(u16) * Display.WindowSizeUint16
		rleSize := len(rle) * Display.WindowSizeRLE

		compressionInt := (float32(intSize) / float32(baseSize)) * 100
		compressionRLE := (float32(rleSize) / float32(baseSize)) * 100
		rate := CompressionRate{BaseSize: baseSize, BaseIntSize: intSize, BaseRLESize: rleSize, CompressionInt: compressionInt, CompressionRLE: compressionRLE}
		fmt.Println(rate)
		rates = append(rates, rate)
	}

	var compressionAvg CompressionRate
	for _, rate := range rates {
		compressionAvg.BaseSize += rate.BaseSize
		compressionAvg.BaseIntSize += rate.BaseIntSize
		compressionAvg.BaseRLESize += rate.BaseRLESize
		compressionAvg.CompressionInt += rate.CompressionInt
		compressionAvg.CompressionRLE += rate.CompressionRLE
	}
	compressionAvg.BaseSize /= len(rates)
	compressionAvg.BaseIntSize /= len(rates)
	compressionAvg.BaseRLESize /= len(rates)
	compressionAvg.CompressionInt /= float32(len(rates))
	compressionAvg.CompressionRLE /= float32(len(rates))

	fmt.Println("Avg Rates: ", compressionAvg)
}

func CheckAppAndCurrentRLE() {
	file, size := Display.LoadThermal("data/Session_20260610_152352/frames/frame_00000000_int.bin")
	i := Display.FileToValue[uint16](file, size-Display.HeaderOffset)
	rle := Display.ValueToValue[uint16, Display.Pair](&i)
	Display.OutputData("RLE.thermal", Display.ValueToBytes[Display.Pair](&rle))

	// constructedFile, cfs := Display.LoadThermal("data/Session_20260610_152352/frames/frame_00000008_rle.bin")
	// appRLE := Display.FileToValue[Display.Pair](constructedFile, cfs-Display.HeaderOffset)

	check_int := Display.ValueToValue[Display.Pair, uint16](&rle)

	if len(i) != len(check_int) {
		fmt.Println("Unequal Lenghts, Original: ", len(i), ", New: ", len(check_int))
	}

	// Display.TemperaturesIntToBMP(i, 640, 480, "out/bitGo.bmp")
	// Display.TemperaturesIntToBMP(Display.ValueToValue[Display.Pair, uint16](&appRLE), 640, 480, "out/bitC#.bmp")

	// Display.OutputData("RLE 2.thermal", Display.RLEToBytes(&ReconstructedRLE))
	// if len(appRLE) != len(rle) {
	// 	fmt.Println("Unequal Reconstruction Length | Our: ", len(rle), "Your: ", len(appRLE))
	// }

	// for i := range rle {
	// 	if appRLE[i].Idx != rle[i].Idx {
	// 		fmt.Println("Unequal Reconstruction ELement Idx | Our: ", appRLE[i].Idx, "Your: ", rle[i].Idx)
	// 	}
	// 	if uint16(appRLE[i].Length) != uint16(rle[i].Length) {
	// 		fmt.Println("Unequal Reconstruction ELement Length | Our: ", appRLE[i].Length, "Your: ", rle[i].Length)
	// 	}
	// }

	// fmt.Println(rle)
}
