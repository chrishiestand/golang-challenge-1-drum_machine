package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	instruments []Instrument
	version     string
	tempo       float32
}

// Instrument is the high level representation of the
// instrument section of a drum machine pattern
type Instrument struct {
	measure []Step
	num     uint32
	name    string
}

// Step is the representation of a step within a musical measure
// in the drum machine pattern
type Step []byte

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (Pattern, error) {

	var p Pattern

	f, err := os.Open(path)
	if err != nil {
		return p, err
	}

	headerBin := make([]byte, 13)
	if _, err = f.Read(headerBin); err != nil {
		return p, err
	}

	if _, err = parseHeader(headerBin); err != nil {
		return p, err
	}

	numBytesSlice := make([]byte, 1)

	if _, err = f.Read(numBytesSlice); err != nil {
		return p, err
	}

	numBytesRemaining := uint64(numBytesSlice[0])

	remainingBytes := make([]byte, numBytesRemaining)

	if _, err := io.ReadFull(f, remainingBytes); err != nil {
		return p, err
	}

	versionBin, remainingBytes := remainingBytes[0:32], remainingBytes[32:]

	p.version = string(bytes.Trim(versionBin, "\x00"))

	tempoBin, remainingBytes := remainingBytes[0:4], remainingBytes[4:]
	buf := bytes.NewReader(tempoBin)
	binary.Read(buf, binary.LittleEndian, &p.tempo)

	p.instruments = readInstruments(remainingBytes)

	if err := f.Close(); err != nil {
		return p, err
	}
	return p, nil
}

// String converts a drum machine pattern into a string
func (p Pattern) String() string {

	version := fmt.Sprintf("Saved with HW Version: %v\n", p.version)
	tempo := fmt.Sprintf("Tempo: %v\n", p.tempo)

	instruments := ""

	for _, instrument := range p.instruments {
		line := fmt.Sprintf("(%d) %s\t|", instrument.num, instrument.name)

		for _, measure := range instrument.measure {

			for _, beat := range measure {
				if beat == 0x01 {
					line += "x"
				} else {
					line += "-"
				}
			}

			line += "|"
		}
		line += "\n"
		instruments += line
	}

	return version + tempo + instruments
}

func readInstruments(remainingBytes []byte) []Instrument {

	instruments := make([]Instrument, 0)

	for len(remainingBytes) > 0 {

		i, rb := readInstrument(remainingBytes)
		remainingBytes = rb
		instruments = append(instruments, i)
	}
	return instruments
}

func readInstrument(remainingBytes []byte) (Instrument, []byte) {

	var inst Instrument

	numBin, remainingBytes := remainingBytes[0:4], remainingBytes[4:]

	buf := bytes.NewReader(numBin)
	binary.Read(buf, binary.LittleEndian, &inst.num)

	nameLengthBin, remainingBytes := remainingBytes[0:1], remainingBytes[1:]

	nameLength := nameLengthBin[0]

	nameBin, remainingBytes := remainingBytes[0:nameLength], remainingBytes[nameLength:]

	inst.name = string(nameBin)

	for i := 0; i < 4; i++ {

		stepBin, rb := remainingBytes[0:4], remainingBytes[4:]
		remainingBytes = rb

		inst.measure = append(inst.measure, stepBin)
	}

	return inst, remainingBytes
}

func parseHeader(h []byte) (string, error) {

	headerBin := bytes.Trim(h, "\x00")

	if string(headerBin) != "SPLICE" {
		return "", errors.New("invalid header")
	}

	return string(headerBin), nil
}
