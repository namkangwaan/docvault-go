package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/namkangwaan/docvault-go/internal/services"
	"github.com/namkangwaan/docvault-go/internal/storage"
)

type FileHandler struct {
	fileService     *services.FileService
	documentService *services.DocumentService
	storage         *storage.NFSStorage
}

func NewFileHandler(fileService *services.FileService, documentService *services.DocumentService, storage *storage.NFSStorage) *FileHandler {
	return &FileHandler{
		fileService:     fileService,
		documentService: documentService,
		storage:         storage,
	}
}

// UploadFile handles single file upload with validation (admin endpoint)
func (h *FileHandler) UploadFile(c *fiber.Ctx) error {
	// Get document ID from form
	documentID, err := strconv.ParseInt(c.FormValue("document_id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid document_id",
		})
	}

	// Get file from form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no file uploaded",
		})
	}

	// Open file
	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to open uploaded file",
		})
	}
	defer file.Close()

	// Upload file with validation
	uploadedFile, report, err := h.fileService.UploadFile(c.Context(), file, fileHeader.Filename, documentID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":           err.Error(),
			"security_report": report,
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"file":            uploadedFile,
		"security_report": report,
	})
}

// InitChunkedUpload initializes a chunked upload session (admin endpoint)
func (h *FileHandler) InitChunkedUpload(c *fiber.Ctx) error {
	type InitRequest struct {
		FileName   string `json:"file_name"`
		TotalSize  int64  `json:"total_size"`
		ChunkSize  int64  `json:"chunk_size"`
		DocumentID int64  `json:"document_id"`
	}

	var req InitRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// Generate session ID
	sessionID := uuid.New().String()

	// Initialize chunked upload
	chunker := h.storage.GetChunker()
	session, err := chunker.InitSession(sessionID, req.FileName, req.TotalSize, req.ChunkSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to initialize upload session",
		})
	}

	return c.JSON(fiber.Map{
		"session_id":   session.ID,
		"total_chunks": session.TotalChunks,
	})
}

// UploadChunk handles individual chunk upload (admin endpoint)
func (h *FileHandler) UploadChunk(c *fiber.Ctx) error {
	sessionID := c.FormValue("session_id")
	chunkIndex, err := strconv.Atoi(c.FormValue("chunk_index"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid chunk_index",
		})
	}

	// Get chunk data
	fileHeader, err := c.FormFile("chunk")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no chunk uploaded",
		})
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to open chunk",
		})
	}
	defer file.Close()

	// Read chunk data
	chunkData := make([]byte, fileHeader.Size)
	if _, err := file.Read(chunkData); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to read chunk",
		})
	}

	// Save chunk
	chunker := h.storage.GetChunker()
	if err := chunker.SaveChunk(sessionID, chunkIndex, chunkData); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to save chunk",
		})
	}

	// Get progress
	received, total, _ := chunker.Progress(sessionID)

	return c.JSON(fiber.Map{
		"message":  "chunk uploaded successfully",
		"progress": fmt.Sprintf("%d/%d", received, total),
	})
}

// FinalizeChunkedUpload finalizes the chunked upload (admin endpoint)
func (h *FileHandler) FinalizeChunkedUpload(c *fiber.Ctx) error {
	type FinalizeRequest struct {
		SessionID  string `json:"session_id"`
		DocumentID int64  `json:"document_id"`
	}

	var req FinalizeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	chunker := h.storage.GetChunker()

	// Check if complete
	complete, err := chunker.IsComplete(req.SessionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if !complete {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "not all chunks received",
		})
	}

	// Get session info
	session, err := chunker.GetSession(req.SessionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Generate final file path in temp directory
	finalPath := filepath.Join(os.TempDir(), "docvault-finalized", session.FileName)
	if err := os.MkdirAll(filepath.Dir(finalPath), 0700); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create finalized directory",
		})
	}

	// Finalize upload
	if err := chunker.Finalize(req.SessionID, finalPath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to finalize upload",
		})
	}

	// Clean up session
	defer chunker.CleanupSession(req.SessionID)

	return c.JSON(fiber.Map{
		"message": "upload completed successfully",
		"file":    session.FileName,
	})
}

// GetUploadStatus returns the status of a chunked upload (admin endpoint)
func (h *FileHandler) GetUploadStatus(c *fiber.Ctx) error {
	sessionID := c.Params("session_id")

	chunker := h.storage.GetChunker()
	received, total, err := chunker.Progress(sessionID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "session not found",
		})
	}

	complete, _ := chunker.IsComplete(sessionID)

	return c.JSON(fiber.Map{
		"session_id":      sessionID,
		"chunks_received": received,
		"total_chunks":    total,
		"complete":        complete,
		"progress":        float64(received) / float64(total) * 100,
	})
}

// DownloadFile handles file download with Range support (public endpoint)
func (h *FileHandler) DownloadFile(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid file id",
		})
	}

	// Get file reader
	reader, file, err := h.fileService.GetFileReader(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "file not found",
		})
	}
	defer reader.Close()

	// Get range handler
	rangeHandler := h.storage.GetRangeHandler()

	// Parse range header
	rangeHeader := c.Get("Range")
	rangeReq, err := rangeHandler.ParseRange(rangeHeader, file.FileSize)
	if err != nil && rangeHeader != "" {
		return c.Status(fiber.StatusRequestedRangeNotSatisfiable).JSON(fiber.Map{
			"error": "invalid range",
		})
	}

	// Set response headers
	c.Set("Content-Type", file.MimeType)
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file.OriginalName))
	c.Set("Accept-Ranges", "bytes")

	if rangeReq != nil {
		c.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", rangeReq.Start, rangeReq.End, file.FileSize))
		c.Set("Content-Length", strconv.FormatInt(rangeReq.End-rangeReq.Start+1, 10))
		c.Status(fiber.StatusPartialContent)
	} else {
		c.Set("Content-Length", strconv.FormatInt(file.FileSize, 10))
	}

	// Stream file
	_, err = rangeHandler.WriteRange(c.Response().BodyWriter(), reader, rangeReq)
	if err != nil {
		return err
	}

	// Increment download count on document
	if file.DocumentID > 0 {
		go h.documentService.IncrementDownloadCount(c.Context(), file.DocumentID)
	}

	return nil
}

// PreviewFile serves file for inline preview (public endpoint)
func (h *FileHandler) PreviewFile(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid file id",
		})
	}

	// Get file reader
	reader, file, err := h.fileService.GetFileReader(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "file not found",
		})
	}
	defer reader.Close()

	// Set response headers for inline display
	c.Set("Content-Type", file.MimeType)
	c.Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, file.OriginalName))
	c.Set("Content-Length", strconv.FormatInt(file.FileSize, 10))

	// Stream file
	return c.SendStream(reader)
}

// DeleteFile deletes a file (admin endpoint)
func (h *FileHandler) DeleteFile(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid file id",
		})
	}

	if err := h.fileService.DeleteFile(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete file",
		})
	}

	return c.JSON(fiber.Map{
		"message": "file deleted successfully",
	})
}
