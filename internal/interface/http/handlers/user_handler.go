package handlers

import (
	"clean_architecture/internal/interface/dto"
	"clean_architecture/internal/interface/http/common"
	"clean_architecture/internal/usecase/user"
	"github.com/gin-gonic/gin"
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

func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, "request payload is invalid", err.Error())
		return
	}

	userEntity, err := h.register.Execute(c.Request.Context(), user.RegisterUserInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		common.Error(c, err)
		return
	}

	common.Created(c, dto.ToUserResponse(userEntity))
}

func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, "request payload is invalid", err.Error())
		return
	}

	result, err := h.login.Execute(c.Request.Context(), user.LoginUserInput{
		Email:      req.Email,
		Password:   req.Password,
		DeviceName: req.DeviceName,
		UserAgent:  c.Request.UserAgent(),
		IPAddress:  c.ClientIP(),
	})
	if err != nil {
		common.Error(c, err)
		return
	}

	common.OK(c, dto.AuthResponse{
		User:         dto.ToUserResponse(result.User),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})
}

func (h *UserHandler) Refresh(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, "request payload is invalid", err.Error())
		return
	}

	result, err := h.refresh.Execute(c.Request.Context(), user.RefreshSessionInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		common.Error(c, err)
		return
	}

	common.OK(c, dto.AuthResponse{
		User:         dto.ToUserResponse(result.User),
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	})
}

func (h *UserHandler) Logout(c *gin.Context) {
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, "request payload is invalid", err.Error())
		return
	}
	if err := h.logout.Execute(c.Request.Context(), user.LogoutSessionInput{RefreshToken: req.RefreshToken}); err != nil {
		common.Error(c, err)
		return
	}
	common.NoContent(c)
}

func (h *UserHandler) LogoutAll(c *gin.Context) {
	if err := h.logoutAll.Execute(c.Request.Context(), user.LogoutAllSessionsInput{UserID: common.UserID(c)}); err != nil {
		common.Error(c, err)
		return
	}
	common.NoContent(c)
}

func (h *UserHandler) ListSessions(c *gin.Context) {
	items, err := h.listSessions.Execute(c.Request.Context(), user.ListSessionsInput{UserID: common.UserID(c)})
	if err != nil {
		common.Error(c, err)
		return
	}
	response := make([]dto.SessionResponse, 0, len(items))
	for _, item := range items {
		response = append(response, dto.ToSessionResponse(item))
	}
	common.OK(c, response)
}

func (h *UserHandler) RevokeSession(c *gin.Context) {
	var req dto.RevokeSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ValidationError(c, "request payload is invalid", err.Error())
		return
	}
	if err := h.revokeSession.Execute(c.Request.Context(), user.RevokeSessionInput{
		UserID:    common.UserID(c),
		SessionID: req.SessionID,
	}); err != nil {
		common.Error(c, err)
		return
	}
	common.NoContent(c)
}
