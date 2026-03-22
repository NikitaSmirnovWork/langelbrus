package handlers

import (
	"net/http"
	"wordbot/database"
	"wordbot/models"

	"github.com/gin-gonic/gin"
)

// GET /api/modules/:moduleId/topics
func GetTopics(c *gin.Context) {
	userID   := c.GetString("user_id")
	moduleID := c.Param("moduleId")

	// Verify module belongs to user
	var count int
	database.DB.QueryRow(`SELECT COUNT(*) FROM modules WHERE id = ? AND user_id = ?`, moduleID, userID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		return
	}

	rows, err := database.DB.Query(
		`SELECT id, module_id, user_id, name, created_at FROM topics WHERE module_id = ? AND user_id = ? ORDER BY created_at ASC`,
		moduleID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch topics"})
		return
	}
	defer rows.Close()

	topics := []models.Topic{}
	for rows.Next() {
		var t models.Topic
		if err := rows.Scan(&t.ID, &t.ModuleID, &t.UserID, &t.Name, &t.CreatedAt); err != nil {
			continue
		}
		topics = append(topics, t)
	}

	c.JSON(http.StatusOK, topics)
}

// POST /api/modules/:moduleId/topics
func CreateTopic(c *gin.Context) {
	userID   := c.GetString("user_id")
	moduleID := c.Param("moduleId")

	var req models.CreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	// Verify module belongs to user
	var count int
	database.DB.QueryRow(`SELECT COUNT(*) FROM modules WHERE id = ? AND user_id = ?`, moduleID, userID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Module not found"})
		return
	}

	id := generateID()
	_, err := database.DB.Exec(
		`INSERT INTO topics (id, module_id, user_id, name) VALUES (?, ?, ?, ?)`,
		id, moduleID, userID, req.Name,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create topic"})
		return
	}

	c.JSON(http.StatusCreated, models.Topic{ID: id, ModuleID: moduleID, UserID: userID, Name: req.Name})
}

// PUT /api/modules/:moduleId/topics/:id
func UpdateTopic(c *gin.Context) {
	userID  := c.GetString("user_id")
	topicID := c.Param("id")

	var req models.CreateTopicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	res, err := database.DB.Exec(
		`UPDATE topics SET name = ? WHERE id = ? AND user_id = ?`,
		req.Name, topicID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update topic"})
		return
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": topicID, "name": req.Name})
}

// DELETE /api/modules/:moduleId/topics/:id
func DeleteTopic(c *gin.Context) {
	userID  := c.GetString("user_id")
	topicID := c.Param("id")

	database.DB.Exec("PRAGMA foreign_keys = ON")
	res, err := database.DB.Exec(
		`DELETE FROM topics WHERE id = ? AND user_id = ?`,
		topicID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete topic"})
		return
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": topicID})
}
