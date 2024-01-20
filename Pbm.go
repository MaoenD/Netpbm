package Netpbm

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type PBM struct {
	data          [][]bool
	width, height int
	magicNumber   string
}

// ReadPBM reads a PBM image from a file and returns a struct that represents the image.
func ReadPBM(filename string) (*PBM, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	var pbm PBM

	// Read the magic number
	line, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading magic number: %w", err)
	}
	pbm.magicNumber = strings.TrimSpace(line)

	// Skip comments and read width and height
	for {
		line, err = reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading dimensions: %w", err)
		}
		if strings.HasPrefix(line, "#") {
			continue // Skip comments
		}
		parts := strings.Fields(line)
		if len(parts) == 2 {
			pbm.width, err = strconv.Atoi(parts[0])
			if err != nil {
				return nil, err
			}
			pbm.height, err = strconv.Atoi(parts[1])
			if err != nil {
				return nil, err
			}
			break
		}
	}

	// Prepare the data slice
	pbm.data = make([][]bool, pbm.height)
	for i := range pbm.data {
		pbm.data[i] = make([]bool, pbm.width)
	}

	// Decode the image data based on the magic number
	switch pbm.magicNumber {
	case "P1":
		// Decode P1 (ASCII)
		for y := 0; y < pbm.height; y++ {
			for x := 0; x < pbm.width; x++ {
				var ch rune
				for {
					ch, _, err = reader.ReadRune()
					if err != nil {
						return nil, err
					}
					if ch == '0' || ch == '1' {
						pbm.data[y][x] = ch == '1'
						break
					}
				}
			}
		}
	case "P4":
		for y := 0; y < pbm.height; y++ {
			for x := 0; x < pbm.width; x += 8 {
				byteVal, err := reader.ReadByte()
				if err != nil {
					if err == io.EOF && y == pbm.height-1 && x >= pbm.width-8 {
						break
					}
					return nil, err
				}
				for bit := 0; bit < 8; bit++ {
					if x+bit < pbm.width { // Check for padding bits at the end of the row
						pbm.data[y][x+bit] = byteVal&(1<<(7-bit)) != 0
					}
				}
			}
		}
	default:
		return nil, fmt.Errorf("unsupported magic number: %s", pbm.magicNumber)
	}

	return &pbm, nil
}
func (pbm *PBM) Size() (int, int) {
	return pbm.width, pbm.height
}
func (pbm *PBM) At(x, y int) bool {
	if x >= 0 && x < pbm.width && y >= 0 && y < pbm.height {
		return pbm.data[y][x]
	}
	return false
}

func (pbm *PBM) Set(x, y int, value bool) {
	if x >= 0 && x < pbm.width && y >= 0 && y < pbm.height {
		pbm.data[y][x] = value
	}
}

func (pbm *PBM) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	_, err = writer.WriteString(pbm.magicNumber + "\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(writer, "%d %d\n", pbm.width, pbm.height)
	if err != nil {
		return err
	}
	if pbm.magicNumber == "P1" {
		for _, row := range pbm.data {
			for _, pixel := range row {
				if pixel {
					_, err = writer.WriteString("1 ")
				} else {
					_, err = writer.WriteString("0 ")
				}
				if err != nil {
					return err
				}
			}
			_, err = writer.WriteString("\n")
			if err != nil {
				return err
			}
		}
	} else if pbm.magicNumber == "P4" {
		// Handle P4 format
		for y := 0; y < pbm.height; y++ {
			var row []byte
			for x := 0; x < pbm.width; x++ {
				if x%8 == 0 {
					row = append(row, 0) // Append a new byte for every 8 pixels
				}
				if pbm.data[y][x] {
					byteIndex := x / 8
					bitIndex := uint(x % 8)
					row[byteIndex] |= 1 << (7 - bitIndex)
				}
			}
			if _, err := writer.Write(row); err != nil {
				return err
			}
		}
	}

	return nil
}

func (pbm *PBM) Invert() {
	for y := range pbm.data {
		for x := range pbm.data[y] {
			pbm.data[y][x] = !pbm.data[y][x]
		}
	}
}

func (pbm *PBM) Flip() {
	for y := range pbm.data {
		for x := 0; x < pbm.width/2; x++ {
			pbm.data[y][x], pbm.data[y][pbm.width-x-1] = pbm.data[y][pbm.width-x-1], pbm.data[y][x]
		}
	}
}

func (pbm *PBM) Flop() {
	for y := 0; y < pbm.height/2; y++ {
		pbm.data[y], pbm.data[pbm.height-y-1] = pbm.data[pbm.height-y-1], pbm.data[y]
	}
}

func (pbm *PBM) SetMagicNumber(magicNumber string) {
	pbm.magicNumber = magicNumber
}
