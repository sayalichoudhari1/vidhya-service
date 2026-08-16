package web

import (
	"net/http"

	"github.com/gorilla/mux"

	"vidhya-service/src/app"
	"vidhya-service/src/model"
	"vidhya-service/src/requestid"
)

// staff is the set of roles that can manage a school's academic/RBAC data
// (everything except reading their own records).
var staff = []model.Role{model.RoleAdmin, model.RolePrincipal}

// staffAndTeachers additionally includes teachers, who can take attendance
// and enter marks but not manage classes/subjects/enrollment.
var staffAndTeachers = []model.Role{model.RoleAdmin, model.RolePrincipal, model.RoleTeacher}

// anyRole includes every role - used for read endpoints where the handler
// itself narrows access further (e.g. a student can only read their own
// record; see ensureSelfOrParentOrStaff).
var anyRole = []model.Role{model.RoleAdmin, model.RolePrincipal, model.RoleTeacher, model.RoleStudent, model.RoleParent}

// NewRouter builds the full HTTP routing table for Vidhya Service.
func NewRouter(a *app.App) *mux.Router { //nolint:funlen // route table is intentionally linear and easy to scan
	router := mux.NewRouter().StrictSlash(true)
	router.Use(requestid.Middleware)
	router.Use(CORSMiddleware(a.Config.Server.AllowedOrigins))

	api := router.PathPrefix(a.Config.Server.URLPrefix).Subrouter()

	// ---------------------------------------------------------- public ---
	api.Handle("/health", handleHealth(a)).Methods(http.MethodGet)
	api.HandleFunc("/version", handleVersion).Methods(http.MethodGet)
	api.Handle("/v1/auth/login", handleLogin(a)).Methods(http.MethodPost)

	authed := AuthMiddleware(a.JWTIssuer)

	route := func(path string, handler http.Handler, method string, roles ...model.Role) {
		api.Handle(path, Chain(handler, authed, RequireRoles(roles...))).Methods(method)
	}

	// -------------------------------------------------------------- me ---
	route("/v1/auth/me", handleMe(a), http.MethodGet, anyRole...)

	// ---------------------------------------------------------- schools ---
	route("/v1/schools", handleCreateSchool(a), http.MethodPost, model.RoleAdmin)
	route("/v1/schools", handleListSchools(a), http.MethodGet, model.RoleAdmin)
	route("/v1/schools/{school_id}", handleGetSchool(a), http.MethodGet, anyRole...)
	route("/v1/schools/{school_id}", handleUpdateSchool(a), http.MethodPut, model.RoleAdmin, model.RolePrincipal)
	route("/v1/schools/{school_id}", handleDeleteSchool(a), http.MethodDelete, model.RoleAdmin)

	// ----------------------------------------------------------- classes ---
	route("/v1/schools/{school_id}/classes", handleCreateClass(a), http.MethodPost, staff...)
	route("/v1/schools/{school_id}/classes", handleListClasses(a), http.MethodGet, staffAndTeachers...)
	route("/v1/schools/{school_id}/classes/{class_id}", handleGetClass(a), http.MethodGet, anyRole...)
	route("/v1/schools/{school_id}/classes/{class_id}", handleUpdateClass(a), http.MethodPut, staff...)
	route("/v1/schools/{school_id}/classes/{class_id}", handleDeleteClass(a), http.MethodDelete, staff...)

	// ---------------------------------------------------------- subjects ---
	route("/v1/schools/{school_id}/subjects", handleCreateSubject(a), http.MethodPost, staff...)
	route("/v1/schools/{school_id}/subjects", handleListSubjects(a), http.MethodGet, staffAndTeachers...)
	route("/v1/schools/{school_id}/subjects/{subject_id}", handleUpdateSubject(a), http.MethodPut, staff...)
	route("/v1/schools/{school_id}/subjects/{subject_id}", handleDeleteSubject(a), http.MethodDelete, staff...)

	// ---------------------------------------------------- class-subjects ---
	route("/v1/schools/{school_id}/classes/{class_id}/subjects", handleAssignSubjectToClass(a), http.MethodPost, staff...)
	route("/v1/schools/{school_id}/classes/{class_id}/subjects", handleListClassSubjects(a), http.MethodGet, staffAndTeachers...)
	route("/v1/schools/{school_id}/classes/{class_id}/subjects/{class_subject_id}", handleRemoveClassSubject(a), http.MethodDelete, staff...)

	// ---------------------------------------------------------- teachers ---
	route("/v1/schools/{school_id}/teachers", handleCreateTeacher(a), http.MethodPost, staff...)
	route("/v1/schools/{school_id}/teachers", handleListTeachers(a), http.MethodGet, staff...)
	route("/v1/schools/{school_id}/teachers/{teacher_id}", handleGetTeacher(a), http.MethodGet, staff...)
	route("/v1/schools/{school_id}/teachers/{teacher_id}", handleDeleteTeacher(a), http.MethodDelete, staff...)

	// ---------------------------------------------------------- students ---
	route("/v1/schools/{school_id}/students", handleCreateStudent(a), http.MethodPost, staff...)
	route("/v1/schools/{school_id}/students", handleListStudents(a), http.MethodGet, staffAndTeachers...)
	route("/v1/schools/{school_id}/students/{student_id}", handleGetStudent(a), http.MethodGet, anyRole...)
	route("/v1/schools/{school_id}/students/{student_id}", handleUpdateStudent(a), http.MethodPut, staff...)
	route("/v1/schools/{school_id}/students/{student_id}", handleDeleteStudent(a), http.MethodDelete, staff...)

	// -------------------------------------------------------- attendance ---
	route("/v1/schools/{school_id}/attendance", handleMarkAttendance(a), http.MethodPost, staffAndTeachers...)
	route("/v1/schools/{school_id}/attendance/bulk", handleMarkAttendanceBulk(a), http.MethodPost, staffAndTeachers...)
	route("/v1/schools/{school_id}/attendance/{attendance_id}", handleDeleteAttendance(a), http.MethodDelete, staffAndTeachers...)
	route("/v1/schools/{school_id}/classes/{class_id}/attendance", handleGetClassAttendance(a), http.MethodGet, staffAndTeachers...)
	route("/v1/schools/{school_id}/students/{student_id}/attendance", handleGetStudentAttendance(a), http.MethodGet, anyRole...)
	route("/v1/schools/{school_id}/students/{student_id}/attendance/summary", handleGetStudentAttendanceSummary(a), http.MethodGet, anyRole...)

	// -------------------------------------------------------------- exams ---
	route("/v1/schools/{school_id}/exams", handleCreateExam(a), http.MethodPost, staffAndTeachers...)
	route("/v1/schools/{school_id}/classes/{class_id}/exams", handleListClassExams(a), http.MethodGet, anyRole...)
	route("/v1/schools/{school_id}/exams/{exam_id}", handleGetExam(a), http.MethodGet, staffAndTeachers...)
	route("/v1/schools/{school_id}/exams/{exam_id}", handleDeleteExam(a), http.MethodDelete, staffAndTeachers...)
	route("/v1/schools/{school_id}/exams/{exam_id}/results", handleRecordExamResult(a), http.MethodPost, staffAndTeachers...)
	route("/v1/schools/{school_id}/exams/{exam_id}/results/bulk", handleRecordExamResultsBulk(a), http.MethodPost, staffAndTeachers...)
	route("/v1/schools/{school_id}/exams/{exam_id}/results", handleListExamResults(a), http.MethodGet, staffAndTeachers...)
	route("/v1/schools/{school_id}/students/{student_id}/marksheet", handleGetMarksheet(a), http.MethodGet, anyRole...)

	return router
}
