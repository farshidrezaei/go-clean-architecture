package handlers

import (
	"strconv"

	"clean_architecture/internal/interface/dto"
	"clean_architecture/internal/interface/http/common"
	"clean_architecture/internal/interface/http/port"
	"clean_architecture/internal/usecase/comment"
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

func (h *CommentHandler) Add(ctx port.Context) {
	var req dto.AddCommentRequest
	if err := ctx.BindJSON(&req); err != nil {
		common.ValidationError(ctx, "request payload is invalid", err.Error())
		return
	}

	commentEntity, err := h.add.Execute(ctx.Context(), comment.AddCommentInput{
		PostID:  ctx.Param("postID"),
		ActorID: common.UserID(ctx),
		Body:    req.Body,
	})
	if err != nil {
		common.Error(ctx, err)
		return
	}

	common.Created(ctx, dto.ToCommentResponse(commentEntity))
}

func (h *CommentHandler) Update(ctx port.Context) {
	var req dto.UpdateCommentRequest
	if err := ctx.BindJSON(&req); err != nil {
		common.ValidationError(ctx, "request payload is invalid", err.Error())
		return
	}

	commentEntity, err := h.update.Execute(ctx.Context(), comment.UpdateCommentInput{
		CommentID: ctx.Param("commentID"),
		ActorID:   common.UserID(ctx),
		ActorRole: common.UserRole(ctx),
		Body:      req.Body,
	})
	if err != nil {
		common.Error(ctx, err)
		return
	}

	common.OK(ctx, dto.ToCommentResponse(commentEntity))
}

func (h *CommentHandler) Delete(ctx port.Context) {
	if err := h.delete.Execute(ctx.Context(), comment.DeleteCommentInput{
		CommentID: ctx.Param("commentID"),
		ActorID:   common.UserID(ctx),
		ActorRole: common.UserRole(ctx),
	}); err != nil {
		common.Error(ctx, err)
		return
	}

	common.NoContent(ctx)
}

func (h *CommentHandler) List(ctx port.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	result, err := h.list.Execute(ctx.Context(), comment.ListCommentsInput{
		PostID: ctx.Param("postID"),
		Page:   page,
		Limit:  limit,
	})
	if err != nil {
		common.Error(ctx, err)
		return
	}

	items := make([]dto.CommentResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, dto.ToCommentResponse(item))
	}

	common.OK(ctx, dto.PaginatedResponse[dto.CommentResponse]{
		Items: items,
		Meta:  dto.PageMeta{Page: result.Page, Limit: result.Limit, Total: result.Total},
	})
}
