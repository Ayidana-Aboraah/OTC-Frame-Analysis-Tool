package test

import (
	"super/Display"
	"testing"
)

func TestSuper(t *testing.T) {
	data := Display.LoadFormatedThermal("C:/Users/ayidana.aboraah/Documents/Stuff2/Thermal Frame Analysis/data/frame_00000012.bin")

	// Try out some of the compression methods mentioned
	intData := Display.ConvertToU16(&data)
	// // Run through each Conversion approach and print the binary to files
	mp, rle := Display.MapRLE(&intData)

	revertedIntData := Display.RevertMapRLE(mp, rle)

	for i, v := range intData {
		if i == len(revertedIntData) {
			break
		}
		if revertedIntData[i] != v {
			t.Errorf("Data Corrupted at: %d\nOriginal: %d, New: %d", i, v, revertedIntData[i])
		}
	}
}
