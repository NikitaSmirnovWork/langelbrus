package handlers

import (
	"net/http"
	"wordbot/database"
	"wordbot/models"

	"github.com/gin-gonic/gin"
)

// GET /api/topics/:topicId/words
func GetWords(c *gin.Context) {
	userID  := c.GetString("user_id")
	topicID := c.Param("topicId")

	// Verify topic belongs to user
	var count int
	database.DB.QueryRow(`SELECT COUNT(*) FROM topics WHERE id = ? AND user_id = ?`, topicID, userID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}

	rows, err := database.DB.Query(
		`SELECT id, topic_id, user_id, word, pos, description, example, created_at 
		 FROM words WHERE topic_id = ? AND user_id = ? ORDER BY created_at ASC`,
		topicID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch words"})
		return
	}
	defer rows.Close()

	words := []models.Word{}
	for rows.Next() {
		var w models.Word
		if err := rows.Scan(&w.ID, &w.TopicID, &w.UserID, &w.Word, &w.POS, &w.Description, &w.Example, &w.CreatedAt); err != nil {
			continue
		}
		words = append(words, w)
	}

	c.JSON(http.StatusOK, words)
}

// POST /api/topics/:topicId/words
func CreateWord(c *gin.Context) {
	userID  := c.GetString("user_id")
	topicID := c.Param("topicId")

	var req models.CreateWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Word is required"})
		return
	}

	// Verify topic belongs to user
	var count int
	database.DB.QueryRow(`SELECT COUNT(*) FROM topics WHERE id = ? AND user_id = ?`, topicID, userID).Scan(&count)
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Topic not found"})
		return
	}

	if req.POS == "" { req.POS = "other" }

	id := generateID()
	_, err := database.DB.Exec(
		`INSERT INTO words (id, topic_id, user_id, word, pos, description, example) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, topicID, userID, req.Word, req.POS, req.Description, req.Example,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create word"})
		return
	}

	c.JSON(http.StatusCreated, models.Word{
		ID: id, TopicID: topicID, UserID: userID,
		Word: req.Word, POS: req.POS, Description: req.Description, Example: req.Example,
	})
}

// PUT /api/topics/:topicId/words/:id
func UpdateWord(c *gin.Context) {
	userID := c.GetString("user_id")
	wordID := c.Param("id")

	var req models.CreateWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Word is required"})
		return
	}

	if req.POS == "" { req.POS = "other" }

	res, err := database.DB.Exec(
		`UPDATE words SET word = ?, pos = ?, description = ?, example = ? WHERE id = ? AND user_id = ?`,
		req.Word, req.POS, req.Description, req.Example, wordID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update word"})
		return
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Word not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": wordID, "word": req.Word, "pos": req.POS, "description": req.Description, "example": req.Example})
}

// DELETE /api/topics/:topicId/words/:id
func DeleteWord(c *gin.Context) {
	userID := c.GetString("user_id")
	wordID := c.Param("id")

	res, err := database.DB.Exec(
		`DELETE FROM words WHERE id = ? AND user_id = ?`,
		wordID, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete word"})
		return
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Word not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": wordID})
}
