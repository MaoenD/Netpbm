package Netpbm

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type PGM struct {
	data          [][]uint8
	width, height int
	magicNumber   string
	max           uint8
}

func ReadPGM(filename string) (*PGM, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	//open the file, return error if failed to open and secure close after the end of the function
	reader := bufio.NewReader(file)

	// Read magic number
	magicNumber, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading magic number: %v", err)
	}
	magicNumber = strings.TrimSpace(magicNumber) // trim the magic number from the whitespaces
	if magicNumber != "P2" && magicNumber != "P5" {
		return nil, fmt.Errorf("invalid magic number: %s", magicNumber)
	}
	// A lot of flag checking during the code since it was quite hard to find the error at the beginning if the test phase
	// Read dimensions
	dimensions, err := reader.ReadString('\n') //
	if err != nil {
		return nil, fmt.Errorf("error reading dimensions: %v", err)
	}
	var width, height int                                                        // declare variables width and height
	_, err = fmt.Sscanf(strings.TrimSpace(dimensions), "%d %d", &width, &height) // trim the dimensions from the whitespaces
	if err != nil {                                                              // check if there is an error
		return nil, fmt.Errorf("invalid dimensions: %v", err)
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid dimensions: width and height must be positive")
	} // Check if the dimensions are positive in case you wanted to test a negative number

	// Read max value
	maxValue, err := reader.ReadString('\n') //
	if err != nil {
		return nil, fmt.Errorf("error reading max value: %v", err)
	}
	maxValue = strings.TrimSpace(maxValue)
	var max2 int
	_, err = fmt.Sscanf(maxValue, "%d", &max2)
	if err != nil {
		return nil, fmt.Errorf("invalid max value: %v", err)
	} // Check if the max value is valid

	data := make([][]uint8, height)
	expectedBytesPerPixel := 1 // Allocate a 2D slice for image data. Each element in the slice represents a row of pixels and define the expected number of bytes per pixel since for grayscale images, typically, it's 1 byte per pixel.

	if magicNumber == "P2" {
		// Read P2 format in ASCII format
		for y := 0; y < height; y++ {
			line, err := reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("error reading data at row %d: %v", y, err)
			}
			fields := strings.Fields(line)  // Split the line into individual fields, each field represents one pixel's value.
			rowData := make([]uint8, width) // Allocate a slice to hold pixel values for the current row.
			for x, field := range fields {  // Iterate through each pixel in the line
				if x >= width {
					return nil, fmt.Errorf("index out of range at row %d", y)
				}
				var pixelValue uint8
				_, err := fmt.Sscanf(field, "%d", &pixelValue)
				if err != nil {
					return nil, fmt.Errorf("error parsing pixel value at row %d, column %d: %v", y, x, err)
				}
				rowData[x] = pixelValue // Store the pixel value in the row slice
			}
			data[y] = rowData // Assign the row data to the corresponding row in the image data.
		}
	} else if magicNumber == "P5" {
		// Read P5 format in binary format
		for y := 0; y < height; y++ {
			row := make([]byte, width*expectedBytesPerPixel) // Allocate a slice to hold pixel values for the current row.
			n, err := reader.Read(row)
			if err != nil {
				if err == io.EOF {
					return nil, fmt.Errorf("unexpected end of file at row %d", y)
				}
				return nil, fmt.Errorf("error reading pixel data at row %d: %v", y, err)
			}
			if n < width*expectedBytesPerPixel {
				return nil, fmt.Errorf("unexpected end of file at row %d, expected %d bytes, got %d", y, width*expectedBytesPerPixel, n)
			} // flag for the same reason as before

			rowData := make([]uint8, width) // Allocate a slice to store the pixel values for the current row.
			for x := 0; x < width; x++ {
				pixelValue := uint8(row[x*expectedBytesPerPixel])
				rowData[x] = pixelValue
			} // Convert the raw byte data to pixel values and store them in rowData. accesses the byte data for the pixel then  Store the converted pixel value in the 'rowData' slice at position x.
			data[y] = rowData // Assign the rowData slice to the corresponding row in the 'data' slice.
		}
	}

	// Return the PGM struct
	return &PGM{data, width, height, magicNumber, uint8(max2)}, nil
}

// Size returns the width and height of the image.
func (pgm *PGM) Size() (int, int) {
	return pgm.width, pgm.height
} // return the width and height of the image

// At returns the value of the pixel at (x, y).
func (pgm *PGM) At(x, y int) uint8 {
	return pgm.data[y][x]
} // return the value of the pixel at (x, y)

// Set sets the value of the pixel at (x, y).
func (pgm *PGM) Set(x, y int, value uint8) {
	pgm.data[y][x] = value
} // set the value of the pixel at (x, y)

// Save saves the PGM image to a file and returns an error if there was a problem.
func (pgm *PGM) Save(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close() // Create or overwrite a file with the specified filename, return an error if file creation fails then secure that the file is closed when the function exits

	writer := bufio.NewWriter(file)
	_, err = fmt.Fprintln(writer, pgm.magicNumber)
	if err != nil {
		return fmt.Errorf("error writing magic number: %v", err)
	} // Write the magic number to the file and handle any errors.

	_, err = fmt.Fprintf(writer, "%d %d\n", pgm.width, pgm.height)
	if err != nil {
		return fmt.Errorf("error writing dimensions: %v", err)
	} // Write image dimensions (width and height) to the file.

	_, err = fmt.Fprintln(writer, pgm.max)
	if err != nil {
		return fmt.Errorf("error writing max value: %v", err)
	} // Write the maximum gray value (max) to the file.

	if pgm.magicNumber == "P2" {
		err = savePGM(writer, pgm, false) // Pass false for isBinary
	} else if pgm.magicNumber == "P5" {
		err = savePGM(writer, pgm, true) // Pass true for isBinary
	}

	if err != nil {
		return err
	} // Check for errors during the saving of image data and return any errors encountered.

	return writer.Flush() // Flush the buffered writer to ensure all data is written to the file.
}

func savePGM(file *bufio.Writer, pgm *PGM, isBinary bool) error {
	for y := 0; y < pgm.height; y++ {
		for x := 0; x < pgm.width; x++ {
			// Write the pixel value
			if isBinary {
				err := file.WriteByte(byte(pgm.data[y][x]))
				if err != nil {
					return fmt.Errorf("error writing binary pixel data at row %d, column %d: %v", y, x, err)
				}
			} else {
				_, err := fmt.Fprint(file, pgm.data[y][x])
				if err != nil {
					return fmt.Errorf("error writing pixel data at row %d, column %d: %v", y, x, err)
				}
			}

			// Add a space after each pixel, except the last one in a row
			if x < pgm.width-1 && !isBinary {
				_, err := fmt.Fprint(file, " ")
				if err != nil {
					return fmt.Errorf("error writing space after pixel at row %d, column %d: %v", y, x, err)
				}
			}
		}

		// Add a newline after each row
		if !isBinary {
			_, err := fmt.Fprintln(file)
			if err != nil {
				return fmt.Errorf("error writing newline after row %d: %v", y, err)
			}
		}
	}
	return nil
}

// Invert inverts the colors of the PGM image.
func (pgm *PGM) Invert() {
	for i := 0; i < pgm.height; i++ {
		for j := 0; j < pgm.width; j++ {
			pgm.data[i][j] = pgm.max - pgm.data[i][j] // Invert the pixel value,this is done by subtracting the pixel value from the maximum possible value. if near max it goes light  so inverting it turns it very black
		}
	}
}

// Flip flips the PGM image horizontally.
func (pgm *PGM) Flip() {
	for i := 0; i < pgm.height; i++ { // Loop over the first half of the columns in the row, only going up to half the width ensures that each pixel is swapped only once.
		for j := 0; j < pgm.width/2; j++ {
			pgm.data[i][j], pgm.data[i][pgm.width-j-1] = pgm.data[i][pgm.width-j-1], pgm.data[i][j]
		} // Swap the pixel at position j with its counterpart on the other side of the row. pgm.data[i][j] is the pixel on the left side of the row, and pgm.data[i][pgm.width-j-1] is the corresponding pixel on the right side. The '-1' is necessary because arrays begins at 0 in go
	}
}

// Flop flops the PGM image vertically.
func (pgm *PGM) Flop() {
	for i := 0; i < pgm.height/2; i++ {
		pgm.data[i], pgm.data[pgm.height-i-1] = pgm.data[pgm.height-i-1], pgm.data[i] // Exchange the current row (pgm.data[i]) with its vertically mirrored counterpart. The counterpart row is identified by 'pgm.height-i-1', which effectively calculates the mirrored row index from the bottom of the image.
	}
}

// SetMagicNumber sets the magic number of the PGM image.
func (pgm *PGM) SetMagicNumber(magicNumber string) {
	pgm.magicNumber = magicNumber // Set the magic number of the PGM image. The magic number is stored in the variable "magicNumber". The function takes a string as an argument and sets the variable to the value of the argument.
}

// SetMaxValue sets the max value of the PGM image.
func (pgm *PGM) SetMaxValue(maxValue uint8) {
	if maxValue <= 0 {
		panic("Invalid maximum value")
	} // Check if the maximum value is valid if equal or less than 0 it will panic

	scaleFactor := float64(maxValue) / float64(pgm.max) // Calculate the scale factor to adjust pixel values. This is done by dividing the new maximum value by the current maximum value. The scaling ensures that the image's relative luminance levels are maintained even after changing the maximum grayscale value.
	for i := 0; i < pgm.height; i++ {
		for j := 0; j < pgm.width; j++ {
			pixelValue := uint8(float64(pgm.data[i][j]) * scaleFactor)
			pgm.data[i][j] = pixelValue
		} // Scale the pixel's grayscale value and convert it back to uint8; the scaling adjusts each pixel's brightness to the new range.
	}

	pgm.max = maxValue // Update the maximum grayscale value of the image to the new value.
}

// Rotate90CW rotates the PGM image 90° clockwise.
func (pgm *PGM) Rotate90CW() {
	// Create a new PGM image with swapped width and height
	newData := make([][]uint8, pgm.width)
	for i := 0; i < pgm.width; i++ {
		newData[i] = make([]uint8, pgm.height)
	} // Iterate through the original image data and populate the new rotated image

	for i := 0; i < pgm.height; i++ {
		for j := 0; j < pgm.width; j++ {
			newData[j][pgm.height-i-1] = pgm.data[i][j]
		}
	} // Rotate the pixel values by 90 degrees clockwise, the pixel at (i, j) in the original image becomes the pixel at (j, height-i-1) in the rotated image

	pgm.data = newData
	pgm.width, pgm.height = pgm.height, pgm.width
} // Update the PGM struct to use the new rotated data and update the width and height accordingly

// ToPBM converts the PGM image to PBM.
func (pgm *PGM) ToPBM() *PBM {
	pbm := &PBM{
		data:        make([][]bool, pgm.height),
		width:       pgm.width,
		height:      pgm.height,
		magicNumber: "P1",
	}
	for y := 0; y < pgm.height; y++ {
		pbm.data[y] = make([]bool, pgm.width)
		for x := 0; x < pgm.width; x++ {
			pbm.data[y][x] = pgm.data[y][x] < uint8(pgm.max/2)
		} // Convert grayscale pixel values to binary in PBM format ,pixels with values less than half of the maximum value become 'true' (1), otherwise 'false' (0)
	}
	return pbm
}
