package dto

import (
	"time"

	"clean_architecture/internal/domain/entities"
)

type CreatePostRequest struct {
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content" binding:"required"`
	PublishNow bool   `json:"publish_now"`
}

type UpdatePostRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type PostResponse struct {
	ID          string     `json:"id"`
	AuthorID    string     `json:"author_id"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Status      string     `json:"status"`
	LikesCount  int        `json:"likes_count"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
}

func ToPostResponse(post *entities.Post) PostResponse {
	return PostResponse{
		ID:          post.ID,
		AuthorID:    post.AuthorID,
		Title:       post.Title,
		Content:     post.Content,
		Status:      string(post.Status),
		LikesCount:  post.LikesCount,
		CreatedAt:   post.CreatedAt,
		UpdatedAt:   post.UpdatedAt,
		PublishedAt: post.PublishedAt,
	}
}
