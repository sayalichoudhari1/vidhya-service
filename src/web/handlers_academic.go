package web

import (
	"net/http"

	"github.com/gorilla/mux"

	"vidhya-service/src/app"
	"vidhya-service/src/services/academic"
	"vidhya-service/src/util"
)

// ------------------------------------------------------------------ classes ---

func handleCreateClass(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		var req academic.CreateClassRequest
		if err := util.DecodeJSON(r, &req); err != nil {
			util.WriteError(w, r, err)
			return
		}
		created, err := a.AcademicService.CreateClass(r.Context(), schoolID, req)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteCreated(w, created)
	}
}

func handleListClasses(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		page := util.ParsePagination(r)
		classes, err := a.AcademicService.ListClasses(r.Context(), schoolID, page.Limit, page.Offset)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, classes)
	}
}

func handleGetClass(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		class, err := a.AcademicService.GetClass(r.Context(), schoolID, mux.Vars(r)["class_id"])
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, class)
	}
}

func handleUpdateClass(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		var req academic.UpdateClassRequest
		if err := util.DecodeJSON(r, &req); err != nil {
			util.WriteError(w, r, err)
			return
		}
		updated, err := a.AcademicService.UpdateClass(r.Context(), schoolID, mux.Vars(r)["class_id"], req)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, updated)
	}
}

func handleDeleteClass(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		if err := a.AcademicService.DeleteClass(r.Context(), schoolID, mux.Vars(r)["class_id"]); err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteNoContent(w)
	}
}

// ----------------------------------------------------------------- subjects ---

func handleCreateSubject(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		var req academic.CreateSubjectRequest
		if err := util.DecodeJSON(r, &req); err != nil {
			util.WriteError(w, r, err)
			return
		}
		created, err := a.AcademicService.CreateSubject(r.Context(), schoolID, req)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteCreated(w, created)
	}
}

func handleListSubjects(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		page := util.ParsePagination(r)
		subjects, err := a.AcademicService.ListSubjects(r.Context(), schoolID, page.Limit, page.Offset)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, subjects)
	}
}

func handleUpdateSubject(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		var req academic.UpdateSubjectRequest
		if err := util.DecodeJSON(r, &req); err != nil {
			util.WriteError(w, r, err)
			return
		}
		updated, err := a.AcademicService.UpdateSubject(r.Context(), schoolID, mux.Vars(r)["subject_id"], req)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, updated)
	}
}

func handleDeleteSubject(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		if err := a.AcademicService.DeleteSubject(r.Context(), schoolID, mux.Vars(r)["subject_id"]); err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteNoContent(w)
	}
}

// ----------------------------------------------------------- class-subjects ---

func handleAssignSubjectToClass(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := resolveSchoolID(r); err != nil {
			util.WriteError(w, r, err)
			return
		}
		var req academic.AssignSubjectRequest
		if err := util.DecodeJSON(r, &req); err != nil {
			util.WriteError(w, r, err)
			return
		}
		created, err := a.AcademicService.AssignSubjectToClass(r.Context(), mux.Vars(r)["class_id"], req)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteCreated(w, created)
	}
}

func handleListClassSubjects(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := resolveSchoolID(r); err != nil {
			util.WriteError(w, r, err)
			return
		}
		rows, err := a.AcademicService.ListSubjectsForClass(r.Context(), mux.Vars(r)["class_id"])
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, rows)
	}
}

func handleRemoveClassSubject(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := resolveSchoolID(r); err != nil {
			util.WriteError(w, r, err)
			return
		}
		if err := a.AcademicService.RemoveSubjectFromClass(r.Context(), mux.Vars(r)["class_subject_id"]); err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteNoContent(w)
	}
}
