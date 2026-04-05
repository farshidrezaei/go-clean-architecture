package handlers

import (
	"strconv"

	"clean_architecture/internal/interface/dto"
	"clean_architecture/internal/interface/http/common"
	"clean_architecture/internal/interface/http/port"
	"clean_architecture/internal/usecase/post"
)

type PostHandler struct {
	create *post.CreatePost
	update *post.UpdatePost
	delete *post.DeletePost
	like   *post.LikePost
	get    *post.GetPost
	list   *post.ListPosts
}

func NewPostHandler(create *post.CreatePost, update *post.UpdatePost, deleteUC *post.DeletePost, like *post.LikePost, get *post.GetPost, list *post.ListPosts) *PostHandler {
	return &PostHandler{create: create, update: update, delete: deleteUC, like: like, get: get, list: list}
}

func (h *PostHandler) Create(ctx port.Context) {
	var req dto.CreatePostRequest
	if err := ctx.BindJSON(&req); err != nil {
		common.ValidationError(ctx, "request payload is invalid", err.Error())
		return
	}

	postEntity, err := h.create.Execute(ctx.Context(), post.CreatePostInput{
		AuthorID:   common.UserID(ctx),
		Title:      req.Title,
		Content:    req.Content,
		PublishNow: req.PublishNow,
	})
	if err != nil {
		common.Error(ctx, err)
		return
	}

	common.Created(ctx, dto.ToPostResponse(postEntity))
}

func (h *PostHandler) Update(ctx port.Context) {
	var req dto.UpdatePostRequest
	if err := ctx.BindJSON(&req); err != nil {
		common.ValidationError(ctx, "request payload is invalid", err.Error())
		return
	}

	postEntity, err := h.update.Execute(ctx.Context(), post.UpdatePostInput{
		PostID:    ctx.Param("postID"),
		ActorID:   common.UserID(ctx),
		ActorRole: common.UserRole(ctx),
		Title:     req.Title,
		Content:   req.Content,
	})
	if err != nil {
		common.Error(ctx, err)
		return
	}

	common.OK(ctx, dto.ToPostResponse(postEntity))
}

func (h *PostHandler) Delete(ctx port.Context) {
	if err := h.delete.Execute(ctx.Context(), post.DeletePostInput{
		PostID:    ctx.Param("postID"),
		ActorID:   common.UserID(ctx),
		ActorRole: common.UserRole(ctx),
	}); err != nil {
		common.Error(ctx, err)
		return
	}

	common.NoContent(ctx)
}

func (h *PostHandler) Like(ctx port.Context) {
	if err := h.like.Execute(ctx.Context(), post.LikePostInput{
		PostID:  ctx.Param("postID"),
		ActorID: common.UserID(ctx),
	}); err != nil {
		common.Error(ctx, err)
		return
	}

	common.NoContent(ctx)
}

func (h *PostHandler) Get(ctx port.Context) {
	postEntity, err := h.get.Execute(ctx.Context(), post.GetPostInput{PostID: ctx.Param("postID")})
	if err != nil {
		common.Error(ctx, err)
		return
	}

	common.OK(ctx, dto.ToPostResponse(postEntity))
}

func (h *PostHandler) List(ctx port.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	result, err := h.list.Execute(ctx.Context(), post.ListPostsInput{Page: page, Limit: limit})
	if err != nil {
		common.Error(ctx, err)
		return
	}

	items := make([]dto.PostResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, dto.ToPostResponse(item))
	}

	common.OK(ctx, dto.PaginatedResponse[dto.PostResponse]{
		Items: items,
		Meta:  dto.PageMeta{Page: result.Page, Limit: result.Limit, Total: result.Total},
	})
}
