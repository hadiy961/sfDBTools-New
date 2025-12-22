package backup

import (
	"bufio"
	"fmt"
	"os"

	"sfDBTools/pkg/consts"
)

// createBufferedOutputFile membuat output file dengan buffered writer
func (s *Service) createBufferedOutputFile(outputPath string) (*os.File, *bufio.Writer, error) {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal membuat file output: %w", err)
	}
	bufWriter := bufio.NewWriterSize(outputFile, consts.BackupWriterBufferSize)
	return outputFile, bufWriter, nil
}
