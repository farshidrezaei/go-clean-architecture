package handlers

import (
	"clean_architecture/internal/interface/dto"
	"clean_architecture/internal/interface/http/common"
	"clean_architecture/internal/interface/http/port"
	"clean_architecture/internal/usecase/user"
)

type UserHandler struct {
	register      *user.RegisterUser
	login         *user.LoginUser
	refresh       *user.RefreshSession
	logout        *user.LogoutSession
	logoutAll     *user.LogoutAllSessions
	listSessions  *user.ListSessions
	revokeSession *user.RevokeSession
}

func NewUserHandler(register *user.RegisterUser, login *user.LoginUser, refresh *user.RefreshSession, logout *user.LogoutSession, logoutAll *user.LogoutAllSessions, listSessions *user.ListSessions, revokeSession *user.RevokeSession) *UserHandler {
	return &UserHandler{register: register, login: login, refresh: refresh, logout: logout, logoutAll: logoutAll, listSessions: listSessions, revokeSession: revokeSession}
}

func (h *UserHandler) Register(ctx port.Context) {
	var req dto.RegisterUserRequest
	if err := ctx.BindJSON(&req); err != nil {
		common.ValidationError(ctx, "request payload is invalid", err.Error())
		return
	}

	userEntity, err := h.register.Execute(ctx.Context(), user.RegisterUserInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		common.Error(ctx, err)
		return
	}

	common.Created(ctx, dto.ToUserResponse(userEntity))
}

func (h *UserHandler) Login(ctx port.Context) {
	var req dto.LoginUserRequest
	if err := ctx.BindJSON(&req); err != nil {
		common.ValidationError(ctx, "request payload is invalid", err.Error())
		return
	}

	result, err := h.login.Execute(ctx.Context(), user.LoginUserInput{
		Email:      req.Email,
		Password:   req.Password,
		DeviceName: req.DeviceName,
		UserAgent:  ctx.Request().UserAgent(),
		IPAddress:  ctx.ClientIP(),
	})
	if err != nil {
		common.Error(ctx, err)
		return
	}

	common.OK(ctx, dto.AuthResponse{
		User:         dto.ToUserResponse(result.User),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})
}

func (h *UserHandler) Refresh(ctx port.Context) {
	var req dto.RefreshTokenRequest
	if err := ctx.BindJSON(&req); err != nil {
		common.ValidationError(ctx, "request payload is invalid", err.Error())
		return
	}

	result, err := h.refresh.Execute(ctx.Context(), user.RefreshSessionInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		common.Error(ctx, err)
		return
	}

	common.OK(ctx, dto.AuthResponse{
		User:         dto.ToUserResponse(result.User),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})
}

func (h *UserHandler) Logout(ctx port.Context) {
	var req dto.RefreshTokenRequest
	if err := ctx.BindJSON(&req); err != nil {
		common.ValidationError(ctx, "request payload is invalid", err.Error())
		return
	}
	if err := h.logout.Execute(ctx.Context(), user.LogoutSessionInput{RefreshToken: req.RefreshToken}); err != nil {
		common.Error(ctx, err)
		return
	}
	common.NoContent(ctx)
}

func (h *UserHandler) LogoutAll(ctx port.Context) {
	if err := h.logoutAll.Execute(ctx.Context(), user.LogoutAllSessionsInput{UserID: common.UserID(ctx)}); err != nil {
		common.Error(ctx, err)
		return
	}
	common.NoContent(ctx)
}

func (h *UserHandler) ListSessions(ctx port.Context) {
	items, err := h.listSessions.Execute(ctx.Context(), user.ListSessionsInput{UserID: common.UserID(ctx)})
	if err != nil {
		common.Error(ctx, err)
		return
	}
	response := make([]dto.SessionResponse, 0, len(items))
	for _, item := range items {
		response = append(response, dto.ToSessionResponse(item))
	}
	common.OK(ctx, response)
}

func (h *UserHandler) RevokeSession(ctx port.Context) {
	var req dto.RevokeSessionRequest
	if err := ctx.BindJSON(&req); err != nil {
		common.ValidationError(ctx, "request payload is invalid", err.Error())
		return
	}
	if err := h.revokeSession.Execute(ctx.Context(), user.RevokeSessionInput{
		UserID:    common.UserID(ctx),
		SessionID: req.SessionID,
	}); err != nil {
		common.Error(ctx, err)
		return
	}
	common.NoContent(ctx)
}
