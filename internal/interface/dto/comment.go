package dto

import (
	"time"

	"clean_architecture/internal/domain/entities"
)

type AddCommentRequest struct {
	Body string `json:"body" binding:"required"`
}

type UpdateCommentRequest struct {
	Body string `json:"body" binding:"required"`
}

type CommentResponse struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	AuthorID  string    `json:"author_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToCommentResponse(comment *entities.Comment) CommentResponse {
	return CommentResponse{
		ID:        comment.ID,
		PostID:    comment.PostID,
		AuthorID:  comment.AuthorID,
		Body:      comment.Body,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
	}
}
