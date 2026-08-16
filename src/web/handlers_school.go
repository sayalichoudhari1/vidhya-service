package web

import (
	"net/http"

	"github.com/gorilla/mux"

	"vidhya-service/src/app"
	"vidhya-service/src/services/school"
	"vidhya-service/src/util"
)

func handleCreateSchool(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req school.CreateRequest
		if err := util.DecodeJSON(r, &req); err != nil {
			util.WriteError(w, r, err)
			return
		}
		created, err := a.SchoolService.Create(r.Context(), req)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteCreated(w, created)
	}
}

func handleListSchools(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := util.ParsePagination(r)
		schools, err := a.SchoolService.List(r.Context(), page.Limit, page.Offset)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, schools)
	}
}

func handleGetSchool(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["school_id"]
		s, err := a.SchoolService.Get(r.Context(), id)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, s)
	}
}

func handleUpdateSchool(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["school_id"]
		var req school.UpdateRequest
		if err := util.DecodeJSON(r, &req); err != nil {
			util.WriteError(w, r, err)
			return
		}
		updated, err := a.SchoolService.Update(r.Context(), id, req)
		if err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteOK(w, updated)
	}
}

func handleDeleteSchool(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["school_id"]
		if err := a.SchoolService.Delete(r.Context(), id); err != nil {
			util.WriteError(w, r, err)
			return
		}
		util.WriteNoContent(w)
	}
}
