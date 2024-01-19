package Netpbm

import (
	"bufio"
	"fmt"
	"os"
)

// PGM represents a Portable GrayMap image.
type PGM struct {
	data          [][]uint8
	width, height int
	magicNumber   string
	max           uint8
}

// ReadPGM reads a PGM image from a file and returns a struct that represents the image.
func ReadPGM(filename string) (*PGM, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	var magicNumber string
	if scanner.Scan() {
		magicNumber = scanner.Text()
	} else {
		return nil, fmt.Errorf("unable to read magic number")
	}

	if magicNumber != "P2" && magicNumber != "P5" {
		return nil, fmt.Errorf("unsupported PGM format: %s", magicNumber)
	}

	var width, height int
	if scanner.Scan() {
		_, err := fmt.Sscanf(scanner.Text(), "%d", &width)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unable to read width")
	}

	if scanner.Scan() {
		_, err := fmt.Sscanf(scanner.Text(), "%d", &height)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unable to read height")
	}

	var maxVal uint8
	if scanner.Scan() {
		_, err := fmt.Sscanf(scanner.Text(), "%d", &maxVal)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unable to read max value")
	}

	data := make([][]uint8, height)
	for i := 0; i < height; i++ {
		data[i] = make([]uint8, width)
		for j := 0; j < width; j++ {
			if scanner.Scan() {
				var value uint8
				_, err := fmt.Sscanf(scanner.Text(), "%d", &value)
				if err != nil {
					return nil, err
				}
				data[i][j] = value
			} else {
				return nil, fmt.Errorf("unable to read pixel data")
			}
		}
	}

	return &PGM{
		data:        data,
		width:       width,
		height:      height,
		magicNumber: magicNumber,
		max:         maxVal,
	}, nil
}

// Size returns the width and height of the image.
func (pgm *PGM) Size() (int, int) {
	return pgm.width, pgm.height
}

// At returns the value of the pixel at (x, y).
func (pgm *PGM) At(x, y int) uint8 {
	return pgm.data[y][x]
}

// Set sets the value of the pixel at (x, y).
func (pgm *PGM) Set(x, y int, value uint8) {
	pgm.data[y][x] = value
}

// Save saves the PGM image to a file and returns an error if there was a problem.
func (pgm *PGM) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	writer := bufio.NewWriter(file)
	defer func(writer *bufio.Writer) {
		err := writer.Flush()
		if err != nil {

		}
	}(writer)

	_, err = writer.WriteString(pgm.magicNumber + "\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(writer, "%d %d\n", pgm.width, pgm.height)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(writer, "%d\n", pgm.max)
	if err != nil {
		return err
	}

	for _, row := range pgm.data {
		for _, pixel := range row {
			_, err = fmt.Fprintf(writer, "%d ", pixel)
			if err != nil {
				return err
			}
		}
		_, err = writer.WriteString("\n")
		if err != nil {
			return err
		}
	}

	return nil
}

// Invert inverts the colors of the PGM image.
func (pgm *PGM) Invert() {
	for i := 0; i < pgm.height; i++ {
		for j := 0; j < pgm.width; j++ {
			pgm.data[i][j] = pgm.max - pgm.data[i][j]
		}
	}
}

// Flip flips the PGM image horizontally.
func (pgm *PGM) Flip() {
	for i := 0; i < pgm.height; i++ {
		for j := 0; j < pgm.width/2; j++ {
			pgm.data[i][j], pgm.data[i][pgm.width-j-1] = pgm.data[i][pgm.width-j-1], pgm.data[i][j]
		}
	}
}

// Flop flops the PGM image vertically.
func (pgm *PGM) Flop() {
	for i := 0; i < pgm.height/2; i++ {
		pgm.data[i], pgm.data[pgm.height-i-1] = pgm.data[pgm.height-i-1], pgm.data[i]
	}
}

// SetMagicNumber sets the magic number of the PGM image.
func (pgm *PGM) SetMagicNumber(magicNumber string) {
	pgm.magicNumber = magicNumber
}

// SetMaxValue sets the max value of the PGM image.
func (pgm *PGM) SetMaxValue(maxValue uint8) {
	pgm.max = maxValue
}

// Rotate90CW rotates the PGM image 90° clockwise.
func (pgm *PGM) Rotate90CW() {
	// Create a new PGM image with swapped width and height
	newData := make([][]uint8, pgm.width)
	for i := 0; i < pgm.width; i++ {
		newData[i] = make([]uint8, pgm.height)
	}

	for i := 0; i < pgm.height; i++ {
		for j := 0; j < pgm.width; j++ {
			newData[j][pgm.height-i-1] = pgm.data[i][j]
		}
	}

	pgm.data = newData
	pgm.width, pgm.height = pgm.height, pgm.width
}

// ToPBM converts the PGM image to PBM.
func (pgm *PGM) ToPBM() *PBM {
	data := make([][]bool, pgm.height)
	for i := 0; i < pgm.height; i++ {
		data[i] = make([]bool, pgm.width)
		for j := 0; j < pgm.width; j++ {
			data[i][j] = pgm.data[i][j] > pgm.max/2
			// The pixel is considered true if its value is greater than half of the max value.
		}
	}

	return &PBM{
		data:        data,
		width:       pgm.width,
		height:      pgm.height,
		magicNumber: "P1",
	}
}
