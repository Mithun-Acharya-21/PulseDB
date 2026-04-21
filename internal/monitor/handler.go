package monitor

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

type CreateMonitorRequest struct {
	Name      string `json:"name"       binding:"required"`
	URL       string `json:"url"        binding:"required,url"`
	IntervalS int    `json:"interval_s" binding:"required,min=10,max=86400"`
}

type UpdateMonitorRequest struct {
	Name      *string `json:"name"`
	URL       *string `json:"url"`
	IntervalS *int    `json:"interval_s"`
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	m := &Monitor{Name: req.Name, URL: req.URL, IntervalS: req.IntervalS}
	if err := h.repo.Create(c.Request.Context(), m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create"})
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *Handler) GetByID(c *gin.Context) {
	m, err := h.repo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get"})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *Handler) List(c *gin.Context) {
	monitors, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list"})
		return
	}
	if monitors == nil {
		monitors = []*Monitor{}
	}
	c.JSON(http.StatusOK, monitors)
}

func (h *Handler) Update(c *gin.Context) {
	m, err := h.repo.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get"})
		return
	}
	var req UpdateMonitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name != nil      { m.Name = *req.Name }
	if req.URL != nil       { m.URL = *req.URL }
	if req.IntervalS != nil { m.IntervalS = *req.IntervalS }
	if err := h.repo.Update(c.Request.Context(), m); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
		return
	}
	c.JSON(http.StatusOK, m)
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.repo.Delete(c.Request.Context(), c.Param("id")); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) ListChecks(c *gin.Context) {
	id := c.Param("id")
	if _, err := h.repo.GetByID(c.Request.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed"})
		return
	}
	checks, err := h.repo.ListChecks(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list checks"})
		return
	}
	if checks == nil {
		checks = []*Check{}
	}
	c.JSON(http.StatusOK, checks)
}