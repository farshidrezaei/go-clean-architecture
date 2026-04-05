package handlers

import (
	"strconv"

	"clean_architecture/internal/interface/dto"
	"clean_architecture/internal/interface/http/common"
	"clean_architecture/internal/interface/http/port"
	adminuc "clean_architecture/internal/usecase/admin"
)

type AdminHandler struct {
	listUsers   *adminuc.ListUsers
	promoteUser *adminuc.PromoteUser
}

func NewAdminHandler(listUsers *adminuc.ListUsers, promoteUser *adminuc.PromoteUser) *AdminHandler {
	return &AdminHandler{listUsers: listUsers, promoteUser: promoteUser}
}

func (h *AdminHandler) ListUsers(ctx port.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	result, err := h.listUsers.Execute(ctx.Context(), adminuc.ListUsersInput{
		ActorRole: common.UserRole(ctx),
		Page:      page,
		Limit:     limit,
	})
	if err != nil {
		common.Error(ctx, err)
		return
	}

	items := make([]dto.UserResponse, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, dto.ToUserResponse(item))
	}
	common.OK(ctx, dto.PaginatedResponse[dto.UserResponse]{
		Items: items,
		Meta:  dto.PageMeta{Page: result.Page, Limit: result.Limit, Total: result.Total},
	})
}

func (h *AdminHandler) UpdateUserRole(ctx port.Context) {
	var req dto.UpdateUserRoleRequest
	if err := ctx.BindJSON(&req); err != nil {
		common.ValidationError(ctx, "request payload is invalid", err.Error())
		return
	}

	userEntity, err := h.promoteUser.Execute(ctx.Context(), adminuc.PromoteUserInput{
		ActorRole: common.UserRole(ctx),
		UserID:    ctx.Param("userID"),
		Role:      req.Role,
	})
	if err != nil {
		common.Error(ctx, err)
		return
	}
	common.OK(ctx, dto.ToUserResponse(userEntity))
}
