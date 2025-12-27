package v1

import (
	"net/http"
	"strconv"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (r *V1) createUser(ctx *fiber.Ctx) error {
	var body entity.User

	if err := ctx.BodyParser(&body); err != nil {
		r.l.Errorw("restapi - v1 - createUser", zap.Error(err))
		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	err := r.u.Create(ctx.UserContext(), body)
	if err != nil {
		r.l.Errorw("restapi - v1 - createUser", zap.Error(err))
		return errorResponse(ctx, http.StatusInternalServerError, "service problems")
	}

	return ctx.SendStatus(http.StatusCreated)
}

func (r *V1) getUser(ctx *fiber.Ctx) error {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid user id")
	}

	user, err := r.u.GetByID(ctx.UserContext(), id)
	if err != nil {
		r.l.Errorw("restapi - v1 - getUser", zap.Error(err))
		return errorResponse(ctx, http.StatusNotFound, "user not found")
	}

	return ctx.Status(http.StatusOK).JSON(user)
}

func (r *V1) updateUser(ctx *fiber.Ctx) error {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid user id")
	}

	var body entity.User
	if err := ctx.BodyParser(&body); err != nil {
		r.l.Errorw("restapi - v1 - updateUser", zap.Error(err))
		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}
	body.ID = id

	err = r.u.Update(ctx.UserContext(), body)
	if err != nil {
		r.l.Errorw("restapi - v1 - updateUser", zap.Error(err))
		return errorResponse(ctx, http.StatusInternalServerError, "service problems")
	}

	return ctx.SendStatus(http.StatusOK)
}

func (r *V1) deleteUser(ctx *fiber.Ctx) error {
	id, err := strconv.ParseInt(ctx.Params("id"), 10, 64)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, "invalid user id")
	}

	err = r.u.Delete(ctx.UserContext(), id)
	if err != nil {
		r.l.Errorw("restapi - v1 - deleteUser", zap.Error(err))
		return errorResponse(ctx, http.StatusInternalServerError, "service problems")
	}

	return ctx.SendStatus(http.StatusNoContent)
}
