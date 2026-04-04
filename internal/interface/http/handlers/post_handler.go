package handlers

import (
	"strconv"

	"clean_architecture/internal/interface/dto"
	"clean_architecture/internal/interface/http/common"
	"clean_architecture/internal/usecase/post"
	"github.com/gin-gonic/gin"
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

func (h *PostHandler) Create(c *gin.Context) {
	var req dto.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, "request payload is invalid", err.Error())
		return
	}

	postEntity, err := h.create.Execute(c.Request.Context(), post.CreatePostInput{
		AuthorID:   common.UserID(c),
		Title:      req.Title,
		Content:    req.Content,
		PublishNow: req.PublishNow,
	})
	if err != nil {
		common.Error(c, err)
		return
	}

	common.Created(c, dto.ToPostResponse(postEntity))
}

func (h *PostHandler) Update(c *gin.Context) {
	var req dto.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, "request payload is invalid", err.Error())
		return
	}

	postEntity, err := h.update.Execute(c.Request.Context(), post.UpdatePostInput{
		PostID:    c.Param("postID"),
		ActorID:   common.UserID(c),
		ActorRole: common.UserRole(c),
		Title:     req.Title,
		Content:   req.Content,
	})
	if err != nil {
		common.Error(c, err)
		return
	}

	common.OK(c, dto.ToPostResponse(postEntity))
}

func (h *PostHandler) Delete(c *gin.Context) {
	if err := h.delete.Execute(c.Request.Context(), post.DeletePostInput{
		PostID:    c.Param("postID"),
		ActorID:   common.UserID(c),
		ActorRole: common.UserRole(c),
	}); err != nil {
		common.Error(c, err)
		return
	}

	common.NoContent(c)
}

func (h *PostHandler) Like(c *gin.Context) {
	if err := h.like.Execute(c.Request.Context(), post.LikePostInput{
		PostID:  c.Param("postID"),
		ActorID: common.UserID(c),
	}); err != nil {
		common.Error(c, err)
		return
	}

	common.NoContent(c)
}

func (h *PostHandler) Get(c *gin.Context) {
	postEntity, err := h.get.Execute(c.Request.Context(), post.GetPostInput{PostID: c.Param("postID")})
	if err != nil {
		common.Error(c, err)
		return
	}

	common.OK(c, dto.ToPostResponse(postEntity))
}

func (h *PostHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	result, err := h.list.Execute(c.Request.Context(), post.ListPostsInput{Page: page, Limit: limit})
	if err != nil {
		common.Error(c, err)
		return
	}

	items := make([]dto.PostResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, dto.ToPostResponse(item))
	}

	common.OK(c, dto.PaginatedResponse[dto.PostResponse]{
		Items: items,
		Meta:  dto.PageMeta{Page: result.Page, Limit: result.Limit, Total: result.Total},
	})
}
