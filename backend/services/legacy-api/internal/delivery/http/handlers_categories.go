package httpapi

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/moon-eye/velune/services/legacy-api/internal/usecase"
	errs "github.com/moon-eye/velune/shared/errors"
)

type categoryCreateReq struct {
	Name     string     `json:"name" validate:"required,min=1,max=200"`
	ParentID *uuid.UUID `json:"parentId"`
}

type categoryUpdateReq struct {
	Name     string     `json:"name" validate:"required,min=1,max=200"`
	ParentID *uuid.UUID `json:"parentId"`
}

func (s *Server) createCategory(w http.ResponseWriter, r *http.Request) {
	uid, err := mustUserID(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	var req categoryCreateReq
	if err := decodeJSON(r, &req); err != nil {
		WriteError(w, err)
		return
	}
	if err := validateStruct(s, &req); err != nil {
		WriteError(w, err)
		return
	}
	c, err := s.Categories.Create(r.Context(), uid, usecase.CreateCategoryInput{
		Name:     req.Name,
		ParentID: req.ParentID,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, c)
}

func (s *Server) listCategories(w http.ResponseWriter, r *http.Request) {
	uid, err := mustUserID(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	page, limit := parsePageLimit(r)
	list, total, err := s.Categories.List(r.Context(), uid, page, limit)
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"items": list, "total": total, "page": page, "limit": limit})
}

func (s *Server) updateCategory(w http.ResponseWriter, r *http.Request) {
	uid, err := mustUserID(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		WriteError(w, err)
		return
	}
	var req categoryUpdateReq
	if err := decodeJSON(r, &req); err != nil {
		WriteError(w, err)
		return
	}
	if err := validateStruct(s, &req); err != nil {
		WriteError(w, err)
		return
	}
	v, ok := parseInt64Query(r, "version")
	if !ok {
		WriteError(w, errs.New("VALIDATION_ERROR", "version query is required", http.StatusBadRequest))
		return
	}
	c, err := s.Categories.Update(r.Context(), uid, id, v, usecase.UpdateCategoryInput{
		Name:     req.Name,
		ParentID: req.ParentID,
	})
	if err != nil {
		WriteError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, c)
}

func (s *Server) deleteCategory(w http.ResponseWriter, r *http.Request) {
	uid, err := mustUserID(r)
	if err != nil {
		WriteError(w, err)
		return
	}
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		WriteError(w, err)
		return
	}
	v, ok := parseInt64Query(r, "version")
	if !ok {
		WriteError(w, errs.New("VALIDATION_ERROR", "version query is required", http.StatusBadRequest))
		return
	}
	if err := s.Categories.Delete(r.Context(), uid, id, v); err != nil {
		WriteError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
