package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version     string
	tempo       float32
	instruments []instrument
}

type instrument struct {
	num     uint32
	name    string
	measure []step
}

type step []byte

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {

	p := &Pattern{}

	f, err := os.Open(path)
	if err != nil {
		return p, err
	}

	headerBin := make([]byte, 13)
	_, err = f.Read(headerBin)
	if err != nil {
		return p, err
	}
	_, err = parseHeader(headerBin)
	if err != nil {
		return p, err
	}

	numBytesSlice := make([]byte, 1)

	_, err = f.Read(numBytesSlice)
	if err != nil {
		return p, err
	}

	numBytesRemaining := uint64(numBytesSlice[0])

	versionBin := make([]byte, 32)
	numBytesRemaining = numBytesRemaining - 32

	_, err = f.Read(versionBin)
	if err != nil {
		return p, err
	}

	p.version = string(bytes.Trim(versionBin, "\x00"))

	tempoBin := make([]byte, 4)
	numBytesRemaining = numBytesRemaining - 4

	_, err = f.Read(tempoBin)
	if err != nil {
		return p, err
	}

	buf := bytes.NewReader(tempoBin)
	binary.Read(buf, binary.LittleEndian, &p.tempo)

	for numBytesRemaining > 0 {

		n, i, err := readInstrument(f)
		if err != nil {
			return p, err
		}

		p.instruments = append(p.instruments, *i)
		numBytesRemaining = numBytesRemaining - uint64(n)
	}

	return p, nil
}

func (p *Pattern) String() string {

	version := fmt.Sprintf("Saved with HW Version: %v\n", p.version)
	tempo := fmt.Sprintf("Tempo: %v\n", p.tempo)

	instruments := ""

	for _, instrument := range p.instruments {
		var line = fmt.Sprintf("(%d) %s\t|", instrument.num, instrument.name)

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

func readInstrument(f *os.File) (uint16, *instrument, error) {

	inst := &instrument{}
	numBytesRead := uint16(0)

	numBin := make([]byte, 4)
	numBytesRead += 4

	_, err := f.Read(numBin)
	if err != nil {
		return numBytesRead, inst, err
	}

	buf := bytes.NewReader(numBin)
	var instrNum uint32
	binary.Read(buf, binary.LittleEndian, &instrNum)

	inst.num = instrNum

	nameLengthBin := make([]byte, 1)
	numBytesRead++

	_, err = f.Read(nameLengthBin)
	if err != nil {
		return numBytesRead, inst, err
	}

	nameLength := nameLengthBin[0]

	nameBin := make([]byte, nameLength)
	numBytesRead += uint16(nameLength)

	_, err = f.Read(nameBin)
	if err != nil {
		return numBytesRead, inst, err
	}

	inst.name = string(nameBin)

	for i := 0; i < 4; i++ {

		stepBin := make([]byte, 4)
		numBytesRead += uint16(4)

		_, err := f.Read(stepBin)
		if err != nil {
			return numBytesRead, inst, err
		}

		inst.measure = append(inst.measure, stepBin)
	}

	return numBytesRead, inst, nil
}

func parseHeader(h []byte) (string, error) {

	headerBin := bytes.Trim(h, "\x00")

	if string(headerBin) != "SPLICE" {
		return "", errors.New("invalid header")
	}

	return string(headerBin), nil
}
