package web

import (
	"net/http"

	"github.com/gorilla/mux"

	"vidhya-service/src/app"
	"vidhya-service/src/apperrors"
	"vidhya-service/src/services/attendance"
	"vidhya-service/src/util"
)

// handleMarkAttendance handles POST /v1/schools/{school_id}/attendance - the
// core "fill attendance" API for a single student.
func handleMarkAttendance(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		claims := ClaimsFromContext(r.Context())
		var req attendance.MarkRequest
		if err := util.DecodeJSON(r, &req); err != nil {
			util.WriteError(w, r, err)
			return
		}
		record, err := a.AttendanceService.Mark(r.Context(), schoolID, claims.UserID, req)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteCreated(w, record)
	}
}

// handleMarkAttendanceBulk handles POST /v1/schools/{school_id}/attendance/bulk
// - lets a teacher take a whole class's attendance in one call.
func handleMarkAttendanceBulk(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		claims := ClaimsFromContext(r.Context())
		var req attendance.BulkMarkRequest
		if err := util.DecodeJSON(r, &req); err != nil {
			util.WriteError(w, r, err)
			return
		}
		records, err := a.AttendanceService.MarkBulk(r.Context(), schoolID, claims.UserID, req)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteCreated(w, records)
	}
}

// handleGetStudentAttendance handles GET
// /v1/schools/{school_id}/students/{student_id}/attendance?from=&to=
func handleGetStudentAttendance(a *app.App) http.HandlerFunc {
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

		q := r.URL.Query()
		rows, err := a.AttendanceService.ListForStudent(r.Context(), studentID, q.Get("from"), q.Get("to"))
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, rows)
	}
}

// handleGetStudentAttendanceSummary handles GET
// /v1/schools/{school_id}/students/{student_id}/attendance/summary?from=&to=
func handleGetStudentAttendanceSummary(a *app.App) http.HandlerFunc {
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

		q := r.URL.Query()
		summary, err := a.AttendanceService.Summary(r.Context(), studentID, q.Get("from"), q.Get("to"))
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, summary)
	}
}

// handleGetClassAttendance handles GET
// /v1/schools/{school_id}/classes/{class_id}/attendance?date=YYYY-MM-DD
func handleGetClassAttendance(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := resolveSchoolID(r); err != nil {
			util.WriteError(w, r, err)
			return
		}
		date := r.URL.Query().Get("date")
		if date == "" {
			util.WriteError(w, r, apperrors.WithMessage(apperrors.ErrInvalidPayload, "date query parameter is required (YYYY-MM-DD)"))
			return
		}
		rows, err := a.AttendanceService.ListForClassOnDate(r.Context(), mux.Vars(r)["class_id"], date)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, rows)
	}
}

func handleDeleteAttendance(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := resolveSchoolID(r); err != nil {
			util.WriteError(w, r, err)
			return
		}
		if err := a.AttendanceService.Delete(r.Context(), mux.Vars(r)["attendance_id"]); err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteNoContent(w)
	}
}
