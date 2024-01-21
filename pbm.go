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
	//open the file return error if failed to open and secure close after the end of the function

	lecture := bufio.NewReader(file)
	var pbm PBM

	line, err := lecture.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading magic number: %w", err)
	}
	pbm.magicNumber = strings.TrimSpace(line)
	//  Read the magic number, trim and store the magic number

	for {
		line, err = lecture.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading dimensions: %w", err)
		}
		if strings.HasPrefix(line, "#") {
			continue
		} // Skip comments and read width and height aka dimensions
		parts := strings.Fields(line)
		if len(parts) == 2 {
			pbm.width, err = strconv.Atoi(parts[0])
			if err != nil {
				return nil, err
			}
			pbm.height, err = strconv.Atoi(parts[1])
			if err != nil {
				return nil, err
			} // Convert width and height from string to int.
			break
		}
	}

	pbm.data = make([][]bool, pbm.height)
	for i := range pbm.data {
		pbm.data[i] = make([]bool, pbm.width)
	}
	// Init the data slice based on the read dimensions.

	switch pbm.magicNumber { // Decode the image data according to the magic number.
	case "P1":
		for y := 0; y < pbm.height; y++ {
			for x := 0; x < pbm.width; x++ {
				var ch rune
				for {
					ch, _, err = lecture.ReadRune()
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
		// Handle P1 (ASCII) format.read a character. If it is a 0 or 1, store it in the data slice as pixel
	case "P4":
		for y := 0; y < pbm.height; y++ { // Read the image data row by row handling padding bits at the end of the row
			for x := 0; x < pbm.width; x += 8 {
				byteVal, err := lecture.ReadByte()
				if err != nil {
					if err == io.EOF && y == pbm.height-1 && x >= pbm.width-8 {
						break // Ignore EOF error if we are at the end of the file and the last byte is a padding byte
					}
					return nil, err // Return an error if we are not at the end of the file
				}
				for bit := 0; bit < 8; bit++ {
					if x+bit < pbm.width { // Check for padding bits at the end of the row
						pbm.data[y][x+bit] = byteVal&(1<<(7-bit)) != 0
					}
				}
			}
		}
	default: // Return an error message if the magic number is not supported
		return nil, fmt.Errorf("unsupported magic number: %s", pbm.magicNumber)
	}

	return &pbm, nil
}
func (pbm *PBM) Size() (int, int) {
	return pbm.width, pbm.height
} // Size returns the width and height of the image.

func (pbm *PBM) At(x, y int) bool {
	if x >= 0 && x < pbm.width && y >= 0 && y < pbm.height {
		return pbm.data[y][x]
	}
	return false // Check if the pixel is in bounds if in bound it returns the pixel value if not it returns false
}

func (pbm *PBM) Set(x, y int, value bool) {
	if x >= 0 && x < pbm.width && y >= 0 && y < pbm.height {
		pbm.data[y][x] = value
	}
} // Check if the pixel is in bounds if it's good it sets the pixel value if not it does nothing

func (pbm *PBM) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	//open the file return error if failed to open and secure close after the end of the function
	writer := bufio.NewWriter(file)
	defer writer.Flush() // Flush the writer at the end of the function

	_, err = writer.WriteString(pbm.magicNumber + "\n")
	if err != nil {
		return err
	} // Write the magic number into the file

	_, err = fmt.Fprintf(writer, "%d %d\n", pbm.width, pbm.height)
	if err != nil {
		return err
	} // Write the dimensions into the file
	if pbm.magicNumber == "P1" { // Handle P1 format (ASCII) format
		for _, row := range pbm.data { // Write the image data row by row
			for _, pixel := range row {
				if pixel {
					_, err = writer.WriteString("1 ")
				} else {
					_, err = writer.WriteString("0 ")
				} //  if it's a pixel write 1 if not write 0
				if err != nil {
					return err
				} //if there is an error return it
			}
			_, err = writer.WriteString("\n")
			if err != nil {
				return err
			} // Write a new line at the end of each row
		}
	} else if pbm.magicNumber == "P4" { // Handle P4 format (binary) format
		for y := 0; y < pbm.height; y++ { // Write the image data row by row
			var row []byte // Create a slice of bytes to store the row data
			for x := 0; x < pbm.width; x++ {
				if x%8 == 0 { // Check if we need to append a new byte to the slice
					row = append(row, 0) // Append a new byte for every 8 pixels
				}
				if pbm.data[y][x] { // Set the bit in the byte if the pixel is set
					byteIndex := x / 8                    // Calculate the index of the byte in the slice
					bitIndex := uint(x % 8)               // Calculate the index of the bit in the byte
					row[byteIndex] |= 1 << (7 - bitIndex) // Set the bit in the byte
				}
			}
			if _, err := writer.Write(row); err != nil { //
				return err
			}
		}
	}

	return nil // Return nil if no error occurs
}

func (pbm *PBM) Invert() {
	for y := range pbm.data {
		for x := range pbm.data[y] {
			pbm.data[y][x] = !pbm.data[y][x] // Invert the pixel value
		}
	}
}

// Flip the image vertically
func (pbm *PBM) Flip() {
	for y := range pbm.data {
		for x := 0; x < pbm.width/2; x++ { // the loop will run until variable "y" reaches half of the height of the PBM image.
			pbm.data[y][x], pbm.data[y][pbm.width-x-1] = pbm.data[y][pbm.width-x-1], pbm.data[y][x] //Inside each iteration of the loop, it swaps two rows in the pixel data array stored in variable "data".
		}
	}
}

func (pbm *PBM) Flop() {
	for y := 0; y < pbm.height/2; y++ {
		pbm.data[y], pbm.data[pbm.height-y-1] = pbm.data[pbm.height-y-1], pbm.data[y]
	} //For each row, it swaps its position with another row. This row has an equal distance from both ends of the image (pbm.height/2 - y - 1). Every iteration of this loop, two rows will be swapped: one from top half and one from bottom half.

}

func (pbm *PBM) SetMagicNumber(magicNumber string) {
	pbm.magicNumber = magicNumber // Set the magic number of the PBM image. The magic number is stored in the variable "magicNumber". The function takes a string as an argument and sets the variable to the value of the argument.
}
