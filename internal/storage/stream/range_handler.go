package stream

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Range represents an HTTP Range request
type Range struct {
	Start int64
	End   int64
}

// RangeHandler handles HTTP Range requests for file downloads
type RangeHandler struct {
	bufferPool *BufferPool
}

// NewRangeHandler creates a new range handler
func NewRangeHandler(bufferPool *BufferPool) *RangeHandler {
	return &RangeHandler{
		bufferPool: bufferPool,
	}
}

// ParseRange parses the Range header from an HTTP request
func (rh *RangeHandler) ParseRange(rangeHeader string, fileSize int64) (*Range, error) {
	if rangeHeader == "" {
		return nil, nil
	}

	// Range header format: "bytes=start-end"
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return nil, fmt.Errorf("invalid range header format")
	}

	rangeStr := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangeStr, "-")

	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range format")
	}

	var start, end int64
	var err error

	if parts[0] != "" {
		start, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid start range: %w", err)
		}
	}

	if parts[1] != "" {
		end, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid end range: %w", err)
		}
	} else {
		end = fileSize - 1
	}

	// Validate range
	if start < 0 || end >= fileSize || start > end {
		return nil, fmt.Errorf("invalid range values")
	}

	return &Range{
		Start: start,
		End:   end,
	}, nil
}

// SetRangeHeaders sets appropriate headers for range response
func (rh *RangeHandler) SetRangeHeaders(w http.ResponseWriter, r *Range, fileSize int64, contentType string) {
	if r != nil {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", r.Start, r.End, fileSize))
		w.Header().Set("Content-Length", strconv.FormatInt(r.End-r.Start+1, 10))
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.Header().Set("Content-Length", strconv.FormatInt(fileSize, 10))
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Accept-Ranges", "bytes")
}

// WriteRange writes a range of data from reader to writer
func (rh *RangeHandler) WriteRange(writer io.Writer, reader io.ReadSeeker, r *Range) (int64, error) {
	if r == nil {
		buffer := rh.bufferPool.Get()
		defer rh.bufferPool.Put(buffer)
		return io.CopyBuffer(writer, reader, *buffer)
	}

	// Seek to start position
	if _, err := reader.Seek(r.Start, io.SeekStart); err != nil {
		return 0, fmt.Errorf("failed to seek: %w", err)
	}

	// Create a limited reader for the range
	limitedReader := io.LimitReader(reader, r.End-r.Start+1)

	buffer := rh.bufferPool.Get()
	defer rh.bufferPool.Put(buffer)

	return io.CopyBuffer(writer, limitedReader, *buffer)
}
