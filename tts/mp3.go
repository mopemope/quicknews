package tts

import (
	"fmt"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/dmulholl/mp3lib"
)

func MergeMP3(outpath string, inpaths []string) error {
	var totalFrames uint32
	var totalBytes uint32
	var totalFiles int
	var firstBitRate int

	// If the list of input files includes the output file we'll end up in an infinite loop.
	for _, filepath := range inpaths {
		if filepath == outpath {
			return errors.New("the list of input files includes the output file.")
		}
	}

	// Create the output file.
	outfile, err := os.Create(outpath)
	if err != nil {
		return errors.Wrap(err, "creating output file")
	}

	// Loop over the input files and append their MP3 frames to the output file.
	for _, inpath := range inpaths {

		infile, err := os.Open(inpath)
		if err != nil {
			return errors.Wrap(err, "opening input file")
		}

		isFirstFrame := true

		for {
			// Read the next frame from the input file.
			frame := mp3lib.NextFrame(infile)
			if frame == nil {
				break
			}

			// Skip the first frame if it's a VBR header.
			if isFirstFrame {
				isFirstFrame = false
				if mp3lib.IsXingHeader(frame) || mp3lib.IsVbriHeader(frame) {
					continue
				}
			}

			// If we detect more than one bitrate we'll need to add a VBR header to the output file.
			if firstBitRate == 0 {
				firstBitRate = frame.BitRate
			}

			// Write the frame to the output file.
			_, err := outfile.Write(frame.RawBytes)
			if err != nil {
				return errors.Wrap(err, "writing to output file")
			}

			totalFrames += 1
			totalBytes += uint32(len(frame.RawBytes))
		}

		if err := infile.Close(); err != nil {
			// Consider logging the error or returning it if critical
			fmt.Println("Error closing input file:", err)
		}
		totalFiles += 1
	}

	if err := outfile.Close(); err != nil {
		return errors.Wrap(err, "closing output file")
	}

	return nil
}
