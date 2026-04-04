package entities

import "time"

type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
)

type Post struct {
	ID          string
	AuthorID    string
	Title       string
	Content     string
	Status      PostStatus
	LikesCount  int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt *time.Time
}

func NewPost(id, authorID, title, content string, now time.Time) (*Post, error) {
	post := &Post{
		ID:        id,
		AuthorID:  authorID,
		Title:     title,
		Content:   content,
		Status:    PostStatusDraft,
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC(),
	}

	if err := post.Validate(); err != nil {
		return nil, err
	}

	return post, nil
}

func (p *Post) Validate() error {
	if p.Title == "" {
		return ErrEmptyTitle
	}
	if p.Content == "" {
		return ErrEmptyContent
	}
	if p.AuthorID == "" || p.ID == "" {
		return ErrInvalidInput
	}

	return nil
}

func (p *Post) Publish(now time.Time) error {
	if p.Status == PostStatusPublished {
		return ErrPostAlreadyPublic
	}

	p.Status = PostStatusPublished
	p.UpdatedAt = now.UTC()
	p.PublishedAt = &p.UpdatedAt

	return nil
}

func (p *Post) UpdateContent(title, content string, now time.Time) error {
	p.Title = title
	p.Content = content
	p.UpdatedAt = now.UTC()
	return p.Validate()
}

func (p *Post) Like() {
	p.LikesCount++
}

func (p *Post) CanBeEditedBy(userID string, role UserRole) bool {
	return p.AuthorID == userID || role == RoleAdmin
}
