package handlers

import (
	"net/http"
	"wordbot/database"
	"wordbot/models"

	"github.com/gin-gonic/gin"
)

// GET /api/modules
func GetModules(c *gin.Context) {
	userID := c.GetString("user_id")

	rows, err := database.DB.Query(
		`SELECT id, user_id, name, created_at FROM modules WHERE user_id = ? ORDER BY created_at ASC`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch modules"})
		return
	}
	defer rows.Close()

	modules := []models.Module{}
	for rows.Next() {
		var m models.Module
		if err := rows.Scan(&m.ID, &m.UserID, &m.Name, &m.CreatedAt); err != nil {
			continue
		}
		// Count topics
		var count int
		database.DB.QueryRow(`SELECT COUNT(*) FROM topics WHERE module_id = ?`, m.ID).Scan(&count)
		m.Topics = make([]models.Topic, 0)
		_ = count
		modules = append(modules, m)
	}

	c.JSON(http.StatusOK, modules)
}

// POST /api/modules
func CreateModule(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	id := generateID()
	_, err := database.DB.Exec(
		`INSERT INTO modules (id, user_id, name) VALUES (?, ?, ?)`,
		id, userID, req.Name,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create module"})
		return
	}

	c.JSON(http.StatusCreated, models.Module{ID: id, UserID: userID, Name: req.Name})
}

// PUT /api/modules/:id
func UpdateModule(c *gin.Context) {
	userID := c.GetString("user_id")
	moduleID := c.Param("id")

	var req models.CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	res, err := database.DB.Exec(
		`UPDATE modules SET name = ? WHERE id = ? AND user_id = ?`,
		req.Name, moduleID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update module"})
		return
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": moduleID, "name": req.Name})
}

// DELETE /api/modules/:id
func DeleteModule(c *gin.Context) {
	userID := c.GetString("user_id")
	moduleID := c.Param("id")

	database.DB.Exec("PRAGMA foreign_keys = ON")
	res, err := database.DB.Exec(
		`DELETE FROM modules WHERE id = ? AND user_id = ?`,
		moduleID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete module"})
		return
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": moduleID})
}
