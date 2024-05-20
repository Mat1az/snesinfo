package main

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Snes struct {
	romName   string
	romLayout string
	romType   string
	romSize   int
	sram      int
	country   string
	license   string
	version   string
	checkC    string
	check     string
}

func (s *Snes) String() string {
	return fmt.Sprintf("%s,%s,%s,%d,%d,%s,%s,%s,%s,%s", s.romName, s.romLayout, s.romType, s.romSize, s.sram, s.country, s.license, s.version, s.checkC, s.check)
}

func decode(x []byte) Snes {
	snes := Snes{}
	romName := make([]byte, 21)
	var romLayout byte
	var romType byte
	var romSize byte
	var sramSize byte
	var country byte
	var license byte
	var version byte
	checkComp := make([]byte, 2)
	checksum := make([]byte, 2)
	copy(romName, x[0:21])
	romLayout = x[21]
	romType = x[22]
	romSize = x[23]
	sramSize = x[24]
	country = x[25]
	license = x[26]
	version = x[27]
	checkComp = x[28:30]
	checksum = x[30:32]
	//ROM Name
	snes.romName = string(romName)
	//ROM Layout
	bitString := fmt.Sprintf("%08b", romLayout)
	a, _ := strconv.ParseInt(string(bitString[3]), 2, 8)
	b, _ := strconv.ParseInt(string(bitString[5]), 2, 8)
	c, _ := strconv.ParseInt(string(bitString[6]), 2, 8)
	d, _ := strconv.ParseInt(string(bitString[7]), 2, 8)
	var str strings.Builder
	//FIXME is ExLo/ExHi && Lo/Hi complementary? do nested if so
	if a == 0 {
		str.WriteString("SlowROM/")
	} else {
		str.WriteString("FastROM/")
	}
	if b == 1 {
		str.WriteString("ExHiROM")
	}
	if c == 1 {
		str.WriteString("ExLoROM")
	}
	if d == 0 {
		str.WriteString("LoROM")
	} else {
		str.WriteString("HiROM")
	}

	snes.romLayout = str.String()
	//ROM Type
	str.Reset()
	s := strings.Split(fmt.Sprintf("%02X", romType), "")
	hasCo := true
	if s[0] == "0" && s[1] == "0" {
		hasCo = false
		str.WriteString("ROM")
	}
	if s[0] == "0" && s[1] == "1" {
		hasCo = false
		str.WriteString("ROM + RAM")
	}
	if s[0] == "0" && s[1] == "2" {
		hasCo = false
		str.WriteString("ROM + RAM + SRAM")
	}
	if hasCo {
		switch s[1] {
		case "3":
			str.WriteString("ROM + ")
		case "4":
			str.WriteString("ROM + RAM + ")
		case "5":
			str.WriteString("ROM + RAM + SRAM + ")
		case "6":
			str.WriteString("ROM + SRAM + ")
		}
		switch s[0] {
		case "0": //DSP
			str.WriteString("DSP")
		case "1": //GSU (SuperFX)
			str.WriteString("GSU (SuperFX)")
		case "2": //OBC1
			str.WriteString("OBC1")
		case "3": //SA-1
			str.WriteString("SA-1")
		case "4": //S-DD1
			str.WriteString("S-DD1")
		case "5": //S-RTC
			str.WriteString("S-RTC")
		case "E": //Other
			str.WriteString("Other")
		case "F": //Custom
			str.WriteString("Custom")
		}
	}
	snes.romType = str.String()
	//ROM Size
	//FIXME Testing & coding needed for: Non Power-of-2 ROM Size.
	snes.romSize = int(math.Pow(2, float64(romSize)) / 125)
	//SRAM Size
	if sramSize != 0 {
		snes.sram = int(16 * math.Pow(2, float64(sramSize-1)))
	}
	/*
		switch int(sramSize) {
		case 0: // no sram
		case 1: // 16kb
		case 2: // 32kb
		case 3: // 64kb
		case 4: // 128kb
		case 5: // 256kb
		}*/
	//Country Code
	countryCode := fmt.Sprintf("%02X", country)
	switch countryCode {
	case "00":
		snes.country = "Japan"
	case "01":
		snes.country = "USA"
	case "02":
		snes.country = "Europe"
	case "03":
		snes.country = "Swenden"
	case "04":
		snes.country = "Finland"
	case "05":
		snes.country = "Denmark"
	case "06":
		snes.country = "France"
	case "07":
		snes.country = "Netherlands"
	case "08":
		snes.country = "Spain"
	case "09":
		snes.country = "Germany"
	case "10":
		snes.country = "Italy"
	case "11":
		snes.country = "China"
	case "12":
		snes.country = "Indonesia"
	case "13":
		snes.country = "Korea"
	default:
		snes.country = countryCode
	}
	//License Code
	switch license {
	case 1, 51:
		snes.license = "Nintendo"
	case 8:
		snes.license = "Capcom"
	case 175:
		snes.license = "Namco"
	case 164:
		snes.license = "Konami"
	case 200:
		snes.license = "Koei"
	default:
		snes.license = fmt.Sprintf("%02X", license)
	}
	//Version
	//TODO decode decimal version pending
	if version == 0 {
		snes.version = "1.0"
	} else {
		snes.version = fmt.Sprintf("%02X", version)
	}
	//Checksum
	snes.checkC = fmt.Sprintf("%02X", checkComp)
	snes.check = fmt.Sprintf("%02X", checksum)
	return snes
}

/*
//0 0x7FC0 LoROM
//1 0xFFC0 HiROM
//2 0x81C0 LoROM + SMC Header
//3 0x101C0 HiROM + SMC Header
*/
func doChecksum(f *os.File) int {
	s, _ := f.Stat()
	hasSMC := s.Size()%1024 != 0
	checksum := make([]byte, 4)
	var result []byte
	result = make([]byte, 2)
	var toTest []int64
	if hasSMC {
		//0x81C0 LoROM + SMC Header
		//0x101C0 HiROM + SMC Header
		//fmt.Println("has smc header")
		toTest = []int64{0x81C0, 0x101C0}
	} else {
		//0x7FC0 LoROM
		//0xFFC0 HiROM
		//fmt.Println("smc header-less")
		toTest = []int64{0x7FC0, 0xFFC0}
	}
	//FIXME Some few roms/hackroms have both LoROM/HiROM valid, extra check methods needed
	for i, offset := range toTest {
		f.Seek(offset+28, 0)
		f.Read(checksum)
		subtle.XORBytes(result, checksum[0:2], checksum[2:4])
		/* debug
		test := subtle.XORBytes(result, checksum[0:2], checksum[2:4])
		fmt.Println("offset: ", offset)
		fmt.Println("offset applied: ", offset+28)
		fmt.Println("offsets: ", fmt.Sprintf("%02X", offset))
		fmt.Println("offsets applied: ", fmt.Sprintf("%02X", offset+28))
		fmt.Println("check0: ", checksum[0:2])
		fmt.Println("check1: ", checksum[2:4])
		fmt.Println("check1s: ", fmt.Sprintf("%02X", checksum[0:2]))
		fmt.Println("check1s: ", fmt.Sprintf("%02X", checksum[2:4]))
		fmt.Println("test is: ", test)
		fmt.Println("result is: ", result)
		fmt.Println("result is: ", fmt.Sprintf("%02X", result))
		*/
		if bytes.Equal(result, []byte{0xFF, 0xFF}) {
			if hasSMC {
				i++
				i++
			}
			return i
		}
	}
	return -1 //Invalid as not found in ^for
}

func getSnes(f *os.File) string {
	rom := make([]byte, 32)
	var offset int64
	switch doChecksum(f) {
	case 0:
		//0x7FC0 LoROM
		offset = 0x7FC0
		//fmt.Println("found case 0")
	case 1:
		//0xFFC0 HiROM
		offset = 0xFFC0
		//fmt.Println("found case 1")
	case 2:
		//0x81C0 LoROM + SMC Header
		offset = 0x81C0
		//fmt.Println("found case 2")
	case 3:
		//0x101C0 HiROM + SMC Header
		offset = 0x101C0
		//fmt.Println("found case 3")
	default: //Invalid ROM
		//fmt.Println("found invalid")
		return "Invalid ROM"
	}
	f.Seek(offset, 0)
	f.Read(rom)
	snes := decode(rom)
	return snes.String()
}

func main() {
	// frontend
	if len(os.Args) < 2 {
		//TODO options pending: json/xml, delimiters, etc.
		fmt.Println("Usage: snesinfo [options] <ROM image filename>")
		os.Exit(1)
	}
	files, _ := filepath.Glob(os.Args[1])
	for _, file := range files {
		f, _ := os.Open(file)
		fmt.Println(getSnes(f))
	}
}
