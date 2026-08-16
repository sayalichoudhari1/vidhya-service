package web

import (
	"net/http"

	"github.com/gorilla/mux"

	"vidhya-service/src/app"
	"vidhya-service/src/apperrors"
	"vidhya-service/src/model"
	"vidhya-service/src/services/student"
	"vidhya-service/src/util"
)

// handleCreateStudent handles POST /v1/schools/{school_id}/students - the
// core "create student record" API.
func handleCreateStudent(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		var req student.CreateRequest
		if err := util.DecodeJSON(r, &req); err != nil {
			util.WriteError(w, r, err)
			return
		}
		created, err := a.StudentService.Create(r.Context(), schoolID, req)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteCreated(w, created)
	}
}

// handleListStudents handles GET /v1/schools/{school_id}/students, with an
// optional ?classId= filter.
func handleListStudents(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		var classID *string
		if v := r.URL.Query().Get("classId"); v != "" {
			classID = &v
		}
		page := util.ParsePagination(r)
		students, err := a.StudentService.List(r.Context(), schoolID, classID, page.Limit, page.Offset)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, students)
	}
}

// handleGetStudent handles GET /v1/schools/{school_id}/students/{student_id}.
// A STUDENT/PARENT may only fetch their own (or their child's) record.
func handleGetStudent(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := resolveSchoolID(r); err != nil {
			util.WriteError(w, r, err)
			return
		}
		studentID := mux.Vars(r)["student_id"]
		if err := ensureSelfOrParentOrStaff(r, studentID); err != nil {
			util.WriteError(w, r, err)
			return
		}

		s, err := a.StudentService.Get(r.Context(), studentID)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, s)
	}
}

func handleUpdateStudent(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := resolveSchoolID(r); err != nil {
			util.WriteError(w, r, err)
			return
		}
		var req student.UpdateRequest
		if err := util.DecodeJSON(r, &req); err != nil {
			util.WriteError(w, r, err)
			return
		}
		updated, err := a.StudentService.Update(r.Context(), mux.Vars(r)["student_id"], req)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, updated)
	}
}

func handleDeleteStudent(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := resolveSchoolID(r); err != nil {
			util.WriteError(w, r, err)
			return
		}
		if err := a.StudentService.Delete(r.Context(), mux.Vars(r)["student_id"]); err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteNoContent(w)
	}
}

// ensureSelfOrParentOrStaff restricts a STUDENT to their own record. Parent
// scoping to a specific child is a phase-2 enhancement (requires a
// parent<->student link check); for now PARENT is treated like staff-read
// and can be tightened once that link is modeled.
func ensureSelfOrParentOrStaff(r *http.Request, studentID string) error {
	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		return apperrors.ErrUnauthorized
	}
	if claims.Role == model.RoleStudent && claims.UserID != studentID {
		return apperrors.WithMessage(apperrors.ErrForbidden, "students may only view their own record")
	}
	return nil
}
