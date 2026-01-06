package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/namkangwaan/docvault-go/internal/models"
	"github.com/namkangwaan/docvault-go/internal/repository"
	"github.com/namkangwaan/docvault-go/internal/services"
)

type DocumentHandler struct {
	documentService *services.DocumentService
}

func NewDocumentHandler(documentService *services.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		documentService: documentService,
	}
}

// ListDocuments returns paginated documents (public endpoint)
func (h *DocumentHandler) ListDocuments(c *fiber.Ctx) error {
	// Parse query parameters
	limit, _ := strconv.Atoi(c.Query("limit", "30"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	filter := &repository.DocumentFilter{
		SearchQuery: c.Query("search"),
		OrderNumber: c.Query("order_number"),
		Limit:       limit,
		Offset:      offset,
	}

	// Category filter
	if categoryID := c.Query("category_id"); categoryID != "" {
		id, err := strconv.ParseInt(categoryID, 10, 64)
		if err == nil {
			filter.CategoryID = &id
		}
	}

	// Date range filter
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		t, err := time.Parse("2006-01-02", dateFrom)
		if err == nil {
			filter.EffectiveDateFrom = &t
		}
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		t, err := time.Parse("2006-01-02", dateTo)
		if err == nil {
			filter.EffectiveDateTo = &t
		}
	}

	// Only show published documents for public endpoint
	published := true
	filter.IsPublished = &published

	documents, total, err := h.documentService.List(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list documents",
		})
	}

	return c.JSON(fiber.Map{
		"documents": documents,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// GetDocument returns a single document with files (public endpoint)
func (h *DocumentHandler) GetDocument(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid document id",
		})
	}

	document, err := h.documentService.GetByID(c.Context(), id, true)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "document not found",
		})
	}

	// Increment view count
	go h.documentService.IncrementViewCount(c.Context(), id)

	return c.JSON(document)
}

// CreateDocument creates a new document (admin endpoint)
func (h *DocumentHandler) CreateDocument(c *fiber.Ctx) error {
	var doc models.Document
	if err := c.BodyParser(&doc); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.documentService.Create(c.Context(), &doc); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create document",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(doc)
}

// UpdateDocument updates a document (admin endpoint)
func (h *DocumentHandler) UpdateDocument(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid document id",
		})
	}

	var doc models.Document
	if err := c.BodyParser(&doc); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	doc.ID = id

	if err := h.documentService.Update(c.Context(), &doc); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update document",
		})
	}

	return c.JSON(doc)
}

// DeleteDocument deletes a document (admin endpoint)
func (h *DocumentHandler) DeleteDocument(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid document id",
		})
	}

	if err := h.documentService.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete document",
		})
	}

	return c.JSON(fiber.Map{
		"message": "document deleted successfully",
	})
}
