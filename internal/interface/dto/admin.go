package dto

type UpdateUserRoleRequest struct {
	Role string `json:"role" binding:"required"`
}
