package handlers

import (
	"strconv"

	"clean_architecture/internal/interface/dto"
	"clean_architecture/internal/interface/http/common"
	adminuc "clean_architecture/internal/usecase/admin"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	listUsers   *adminuc.ListUsers
	promoteUser *adminuc.PromoteUser
}

func NewAdminHandler(listUsers *adminuc.ListUsers, promoteUser *adminuc.PromoteUser) *AdminHandler {
	return &AdminHandler{listUsers: listUsers, promoteUser: promoteUser}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	result, err := h.listUsers.Execute(c.Request.Context(), adminuc.ListUsersInput{
		ActorRole: common.UserRole(c),
		Page:      page,
		Limit:     limit,
	})
	if err != nil {
		common.Error(c, err)
		return
	}

	items := make([]dto.UserResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, dto.ToUserResponse(item))
	}
	common.OK(c, dto.PaginatedResponse[dto.UserResponse]{
		Items: items,
		Meta:  dto.PageMeta{Page: result.Page, Limit: result.Limit, Total: result.Total},
	})
}

func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	var req dto.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, "request payload is invalid", err.Error())
		return
	}

	userEntity, err := h.promoteUser.Execute(c.Request.Context(), adminuc.PromoteUserInput{
		ActorRole: common.UserRole(c),
		UserID:    c.Param("userID"),
		Role:      req.Role,
	})
	if err != nil {
		common.Error(c, err)
		return
	}
	common.OK(c, dto.ToUserResponse(userEntity))
}
