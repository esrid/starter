package handlers

import (
	"template/internal/services"
)

type UserHandler struct {
	US *services.UserService
	SS *services.SessionService
}
