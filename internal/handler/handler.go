package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"org-structure-api/internal/response"
	"org-structure-api/internal/service"
)

type Handler struct {
	service service.DepartmentService
}

func NewHandler(service service.DepartmentService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Routes() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)

		normalizedPath := strings.TrimRight(r.URL.Path, "/")
		if normalizedPath == "" {
			normalizedPath = "/"
		}

		switch {
		case r.Method == http.MethodPost && normalizedPath == "/departments":
			h.createDepartment(w, r)
		case r.Method == http.MethodGet && strings.HasPrefix(normalizedPath, "/departments/"):
			h.getDepartment(w, r)
		case r.Method == http.MethodPatch && strings.HasPrefix(normalizedPath, "/departments/"):
			h.updateDepartment(w, r)
		case r.Method == http.MethodDelete && strings.HasPrefix(normalizedPath, "/departments/"):
			h.deleteDepartment(w, r)
		case r.Method == http.MethodPost && strings.HasSuffix(normalizedPath, "/employees"):
			h.createEmployee(w, r)
		default:
			response.Error(w, http.StatusNotFound, "route not found")
		}
	})
}

type createDepartmentRequest struct {
	Name     string `json:"name"`
	ParentID *int   `json:"parent_id"`
}

func (h *Handler) createDepartment(w http.ResponseWriter, r *http.Request) {
	var req createDepartmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid json body")
		return
	}

	department, err := h.service.CreateDepartment(r.Context(), service.CreateDepartmentInput{
		Name:     req.Name,
		ParentID: req.ParentID,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, department)
}

type createEmployeeRequest struct {
	FullName string  `json:"full_name"`
	Position string  `json:"position"`
	HiredAt  *string `json:"hired_at"`
}

func (h *Handler) createEmployee(w http.ResponseWriter, r *http.Request) {
	departmentID, ok := parseDepartmentIDFromEmployeePath(r.URL.Path)
	if !ok {
		response.Error(w, http.StatusNotFound, "route not found")
		return
	}

	var req createEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid json body")
		return
	}

	hiredAt, err := parseOptionalDate(req.HiredAt)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "hired_at must be in YYYY-MM-DD format")
		return
	}

	employee, err := h.service.CreateEmployee(r.Context(), departmentID, service.CreateEmployeeInput{
		FullName: req.FullName,
		Position: req.Position,
		HiredAt:  hiredAt,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, employee)
}

func (h *Handler) getDepartment(w http.ResponseWriter, r *http.Request) {
	departmentID, ok := parseDepartmentIDFromDepartmentPath(r.URL.Path)
	if !ok {
		response.Error(w, http.StatusNotFound, "route not found")
		return
	}

	depth := 1
	if rawDepth := r.URL.Query().Get("depth"); rawDepth != "" {
		parsedDepth, err := strconv.Atoi(rawDepth)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "depth must be an integer")
			return
		}
		if parsedDepth < 0 || parsedDepth > 5 {
			response.Error(w, http.StatusBadRequest, "depth must be between 0 and 5")
			return
		}
		depth = parsedDepth
	}

	includeEmployees := true
	if rawInclude := r.URL.Query().Get("include_employees"); rawInclude != "" {
		parsedInclude, err := strconv.ParseBool(rawInclude)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "include_employees must be a boolean")
			return
		}
		includeEmployees = parsedInclude
	}

	node, err := h.service.GetDepartmentTree(r.Context(), departmentID, depth, includeEmployees)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, node)
}

func (h *Handler) updateDepartment(w http.ResponseWriter, r *http.Request) {
	departmentID, ok := parseDepartmentIDFromDepartmentPath(r.URL.Path)
	if !ok {
		response.Error(w, http.StatusNotFound, "route not found")
		return
	}

	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid json body")
		return
	}

	input := service.UpdateDepartmentInput{}

	if nameRaw, exists := raw["name"]; exists {
		input.NameSet = true

		var name string
		if err := json.Unmarshal(nameRaw, &name); err != nil {
			response.Error(w, http.StatusBadRequest, "name must be a string")
			return
		}
		input.Name = name
	}

	if parentRaw, exists := raw["parent_id"]; exists {
		input.ParentSet = true
		if string(parentRaw) != "null" {
			var parentID int
			if err := json.Unmarshal(parentRaw, &parentID); err != nil {
				response.Error(w, http.StatusBadRequest, "parent_id must be an integer or null")
				return
			}
			input.ParentID = &parentID
		}
	}

	department, err := h.service.UpdateDepartment(r.Context(), departmentID, input)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, department)
}

func (h *Handler) deleteDepartment(w http.ResponseWriter, r *http.Request) {
	departmentID, ok := parseDepartmentIDFromDepartmentPath(r.URL.Path)
	if !ok {
		response.Error(w, http.StatusNotFound, "route not found")
		return
	}

	var reassignTo *int
	if rawReassign := r.URL.Query().Get("reassign_to_department_id"); rawReassign != "" {
		value, err := strconv.Atoi(rawReassign)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "reassign_to_department_id must be an integer")
			return
		}
		reassignTo = &value
	}

	err := h.service.DeleteDepartment(r.Context(), departmentID, service.DeleteDepartmentInput{
		Mode:                   r.URL.Query().Get("mode"),
		ReassignToDepartmentID: reassignTo,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		response.Error(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrConflict):
		response.Error(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrCycleDetected):
		response.Error(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrValidation):
		response.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrSelfParent),
		errors.Is(err, service.ErrBadRequest),
		errors.Is(err, service.ErrMissingReassign),
		errors.Is(err, service.ErrInvalidDeleteMode):
		response.Error(w, http.StatusBadRequest, err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func parseDepartmentIDFromDepartmentPath(path string) (int, bool) {
	trimmed := strings.Trim(path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 || parts[0] != "departments" {
		return 0, false
	}

	value, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, false
	}

	return value, true
}

func parseDepartmentIDFromEmployeePath(path string) (int, bool) {
	trimmed := strings.Trim(path, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) != 3 || parts[0] != "departments" || parts[2] != "employees" {
		return 0, false
	}

	value, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, false
	}

	return value, true
}

func parseOptionalDate(raw *string) (*time.Time, error) {
	if raw == nil {
		return nil, nil
	}
	if strings.TrimSpace(*raw) == "" {
		return nil, nil
	}

	parsed, err := time.Parse("2006-01-02", *raw)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}
