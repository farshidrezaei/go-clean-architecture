package entities

import "time"

type Comment struct {
	ID        string
	PostID    string
	AuthorID  string
	Body      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewComment(id, postID, authorID, body string, now time.Time) (*Comment, error) {
	comment := &Comment{
		ID:        id,
		PostID:    postID,
		AuthorID:  authorID,
		Body:      body,
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC(),
	}

	if err := comment.Validate(); err != nil {
		return nil, err
	}

	return comment, nil
}

func (c *Comment) Validate() error {
	if c.ID == "" || c.PostID == "" || c.AuthorID == "" {
		return ErrInvalidInput
	}
	if c.Body == "" {
		return ErrEmptyCommentBody
	}

	return nil
}

func (c *Comment) UpdateBody(body string, now time.Time) error {
	c.Body = body
	c.UpdatedAt = now.UTC()
	return c.Validate()
}

func (c *Comment) CanBeManagedBy(userID string, role UserRole) bool {
	return c.AuthorID == userID || role == RoleAdmin
}
