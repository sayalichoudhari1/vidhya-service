package web

import (
	"net/http"

	"github.com/gorilla/mux"

	"vidhya-service/src/app"
	"vidhya-service/src/services/exam"
	"vidhya-service/src/util"
)

func handleCreateExam(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		claims := ClaimsFromContext(r.Context())
		var req exam.CreateExamRequest
		if err := util.DecodeJSON(r, &req); err != nil {
			util.WriteError(w, r, err)
			return
		}
		created, err := a.ExamService.CreateExam(r.Context(), schoolID, claims.UserID, req)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteCreated(w, created)
	}
}

func handleListClassExams(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := resolveSchoolID(r); err != nil {
			util.WriteError(w, r, err)
			return
		}
		page := util.ParsePagination(r)
		exams, err := a.ExamService.ListExamsByClass(r.Context(), mux.Vars(r)["class_id"], page.Limit, page.Offset)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, exams)
	}
}

func handleGetExam(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		e, err := a.ExamService.GetExam(r.Context(), schoolID, mux.Vars(r)["exam_id"])
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, e)
	}
}

func handleDeleteExam(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		if err := a.ExamService.DeleteExam(r.Context(), schoolID, mux.Vars(r)["exam_id"]); err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteNoContent(w)
	}
}

// handleRecordExamResult handles POST
// /v1/schools/{school_id}/exams/{exam_id}/results - records one student's
// marks for that exam.
func handleRecordExamResult(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		claims := ClaimsFromContext(r.Context())
		var req exam.RecordMarksRequest
		if err := util.DecodeJSON(r, &req); err != nil {
			util.WriteError(w, r, err)
			return
		}
		result, err := a.ExamService.RecordMarks(r.Context(), schoolID, mux.Vars(r)["exam_id"], claims.UserID, req)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteCreated(w, result)
	}
}

// handleRecordExamResultsBulk handles POST
// /v1/schools/{school_id}/exams/{exam_id}/results/bulk
func handleRecordExamResultsBulk(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		claims := ClaimsFromContext(r.Context())
		var req exam.BulkRecordMarksRequest
		if err := util.DecodeJSON(r, &req); err != nil {
			util.WriteError(w, r, err)
			return
		}
		results, err := a.ExamService.RecordMarksBulk(r.Context(), schoolID, mux.Vars(r)["exam_id"], claims.UserID, req)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteCreated(w, results)
	}
}

func handleListExamResults(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		results, err := a.ExamService.ListResultsForExam(r.Context(), schoolID, mux.Vars(r)["exam_id"])
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, results)
	}
}

// handleGetMarksheet handles GET
// /v1/schools/{school_id}/students/{student_id}/marksheet - the core
// "student marksheet" API.
func handleGetMarksheet(a *app.App) http.HandlerFunc {
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

		marksheet, err := a.ExamService.Marksheet(r.Context(), studentID)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, marksheet)
	}
}
