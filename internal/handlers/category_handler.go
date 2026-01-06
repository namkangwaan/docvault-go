package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/namkangwaan/docvault-go/internal/models"
	"github.com/namkangwaan/docvault-go/internal/services"
)

type CategoryHandler struct {
	categoryService *services.CategoryService
}

func NewCategoryHandler(categoryService *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
	}
}

// ListCategories returns all categories (public endpoint)
func (h *CategoryHandler) ListCategories(c *fiber.Ctx) error {
	includeInactive := c.Query("include_inactive") == "true"
	
	categories, err := h.categoryService.List(c.Context(), includeInactive)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list categories",
		})
	}

	return c.JSON(fiber.Map{
		"categories": categories,
	})
}

// GetCategory returns a single category
func (h *CategoryHandler) GetCategory(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid category id",
		})
	}

	category, err := h.categoryService.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "category not found",
		})
	}

	return c.JSON(category)
}

// CreateCategory creates a new category (admin endpoint)
func (h *CategoryHandler) CreateCategory(c *fiber.Ctx) error {
	var category models.Category
	if err := c.BodyParser(&category); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if err := h.categoryService.Create(c.Context(), &category); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create category",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(category)
}

// UpdateCategory updates a category (admin endpoint)
func (h *CategoryHandler) UpdateCategory(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid category id",
		})
	}

	var category models.Category
	if err := c.BodyParser(&category); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	category.ID = id

	if err := h.categoryService.Update(c.Context(), &category); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to update category",
		})
	}

	return c.JSON(category)
}

// DeleteCategory deletes a category (admin endpoint)
func (h *CategoryHandler) DeleteCategory(c *fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid category id",
		})
	}

	if err := h.categoryService.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to delete category",
		})
	}

	return c.JSON(fiber.Map{
		"message": "category deleted successfully",
	})
}
