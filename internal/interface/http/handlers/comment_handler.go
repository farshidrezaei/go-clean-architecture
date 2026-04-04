package handlers

import (
	"strconv"

	"clean_architecture/internal/interface/dto"
	"clean_architecture/internal/interface/http/common"
	"clean_architecture/internal/usecase/comment"
	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	add    *comment.AddComment
	update *comment.UpdateComment
	delete *comment.DeleteComment
	list   *comment.ListComments
}

func NewCommentHandler(add *comment.AddComment, update *comment.UpdateComment, deleteUC *comment.DeleteComment, list *comment.ListComments) *CommentHandler {
	return &CommentHandler{add: add, update: update, delete: deleteUC, list: list}
}

func (h *CommentHandler) Add(c *gin.Context) {
	var req dto.AddCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, "request payload is invalid", err.Error())
		return
	}

	commentEntity, err := h.add.Execute(c.Request.Context(), comment.AddCommentInput{
		PostID:  c.Param("postID"),
		ActorID: common.UserID(c),
		Body:    req.Body,
	})
	if err != nil {
		common.Error(c, err)
		return
	}

	common.Created(c, dto.ToCommentResponse(commentEntity))
}

func (h *CommentHandler) Update(c *gin.Context) {
	var req dto.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, "request payload is invalid", err.Error())
		return
	}

	commentEntity, err := h.update.Execute(c.Request.Context(), comment.UpdateCommentInput{
		CommentID: c.Param("commentID"),
		ActorID:   common.UserID(c),
		ActorRole: common.UserRole(c),
		Body:      req.Body,
	})
	if err != nil {
		common.Error(c, err)
		return
	}

	common.OK(c, dto.ToCommentResponse(commentEntity))
}

func (h *CommentHandler) Delete(c *gin.Context) {
	if err := h.delete.Execute(c.Request.Context(), comment.DeleteCommentInput{
		CommentID: c.Param("commentID"),
		ActorID:   common.UserID(c),
		ActorRole: common.UserRole(c),
	}); err != nil {
		common.Error(c, err)
		return
	}

	common.NoContent(c)
}

func (h *CommentHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	result, err := h.list.Execute(c.Request.Context(), comment.ListCommentsInput{
		PostID: c.Param("postID"),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		common.Error(c, err)
		return
	}

	items := make([]dto.CommentResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, dto.ToCommentResponse(item))
	}

	common.OK(c, dto.PaginatedResponse[dto.CommentResponse]{
		Items: items,
		Meta:  dto.PageMeta{Page: result.Page, Limit: result.Limit, Total: result.Total},
	})
}
