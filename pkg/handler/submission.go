package handler

import (
	"strconv"

	"github.com/VDEN5/t-bmstu/pkg/database"
	"github.com/gin-gonic/gin"
)

func (h *Handler) getSumbissionCode(c *gin.Context) {
	stringSubmissionId := c.Param("id")
	submissionId, err := strconv.Atoi(stringSubmissionId)
	if err != nil {
		c.String(404, "There are no such submission")
		return
	}
	code, err := database.GetSubmissionCode(submissionId)
	if err != nil {
		c.String(404, "There are no such submission")
		return
	}
	c.String(200, code)
}
