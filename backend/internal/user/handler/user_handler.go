package handler

import (
	"net/http"

	"github.com/leebrouse/ems/backend/common/genopenapi/user"
	usrModel "github.com/leebrouse/ems/backend/user/model"
	"github.com/leebrouse/ems/backend/user/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Ensure UserHandler implements user.ServerInterface
var _ user.ServerInterface = (*UserHandler)(nil)

func (h *UserHandler) ListRoles(c *gin.Context) {
	roles, err := h.svc.ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var res []user.Role
	for _, r := range roles {
		name := r.Name
		desc := r.Description
		res = append(res, user.Role{
			Name:        &name,
			Description: &desc,
		})
	}
	c.JSON(http.StatusOK, res)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var body user.CreateUserJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	roles := []string{}
	if body.Roles != nil {
		roles = *body.Roles
	}

	u, err := h.svc.CreateUser(c.Request.Context(), body.Username, body.Password, roles)
	if err != nil {
		if err == service.ErrUsernameExists {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, h.toUserResponse(u))
}

func (h *UserHandler) DeleteUser(c *gin.Context, id int32) {
	if err := h.svc.DeleteUser(c.Request.Context(), int64(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *UserHandler) GetUser(c *gin.Context, id int32) {
	u, err := h.svc.GetUser(c.Request.Context(), int64(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, h.toUserResponse(u))
}

func (h *UserHandler) ListUsers(c *gin.Context, params user.ListUsersParams) {
	page := 1
	if params.Page != nil {
		page = int(*params.Page)
	}
	size := 20
	if params.Size != nil {
		size = int(*params.Size)
	}

	users, total, err := h.svc.ListUsers(c.Request.Context(), page, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var res []user.User
	for _, u := range users {
		res = append(res, h.toUserResponse(&u))
	}

	c.JSON(http.StatusOK, gin.H{
		"users": res,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

func (h *UserHandler) UpdateUser(c *gin.Context, id int32) {
	var body user.UpdateUserJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.svc.UpdateUser(c.Request.Context(), int64(id), body.Password, body.Roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, h.toUserResponse(u))
}

func (h *UserHandler) UpdateUserRoles(c *gin.Context, id int32) {
	var body user.UpdateUserRolesJSONBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.svc.UpdateUser(c.Request.Context(), int64(id), nil, &body.Roles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, h.toUserResponse(u))
}

func (h *UserHandler) toUserResponse(u *usrModel.User) user.User {
	id := int32(u.ID)
	username := u.Username
	roles := []string{}
	for _, r := range u.Roles {
		roles = append(roles, r.Name)
	}
	return user.User{
		Id:       &id,
		Username: &username,
		Roles:    &roles,
	}
}
