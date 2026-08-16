package web

import (
	"net/http"

	"github.com/gorilla/mux"

	"vidhya-service/src/app"
	"vidhya-service/src/services/teacher"
	"vidhya-service/src/util"
)

func handleCreateTeacher(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		var req teacher.CreateRequest
		if err := util.DecodeJSON(r, &req); err != nil {
			util.WriteError(w, r, err)
			return
		}
		created, err := a.TeacherService.Create(r.Context(), schoolID, req)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteCreated(w, created)
	}
}

func handleListTeachers(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schoolID, err := resolveSchoolID(r)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		page := util.ParsePagination(r)
		teachers, err := a.TeacherService.List(r.Context(), schoolID, page.Limit, page.Offset)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, teachers)
	}
}

func handleGetTeacher(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := resolveSchoolID(r); err != nil {
			util.WriteError(w, r, err)
			return
		}
		t, err := a.TeacherService.Get(r.Context(), mux.Vars(r)["teacher_id"])
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, t)
	}
}

func handleDeleteTeacher(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := resolveSchoolID(r); err != nil {
			util.WriteError(w, r, err)
			return
		}
		if err := a.TeacherService.Delete(r.Context(), mux.Vars(r)["teacher_id"]); err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteNoContent(w)
	}
}
