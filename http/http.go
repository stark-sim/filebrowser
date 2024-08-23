package http

import (
	"io/fs"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
)

type modifyRequest struct {
	What  string   `json:"what"`  // Answer to: what data type?
	Which []string `json:"which"` // Answer to: which fields?
}

func NewHandler(
	imgSvc ImgService,
	fileCache FileCache,
	store *storage.Storage,
	server *settings.Server,
	assetsFs fs.FS,
) (http.Handler, error) {
	server.Clean()

	r := mux.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Security-Policy", `default-src 'self'; style-src 'unsafe-inline';`)
			next.ServeHTTP(w, r)
		})
	})
	index, static := getStaticHandlers(store, server, assetsFs)

	// NOTE: This fixes the issue where it would redirect if people did not put a
	// trailing slash in the end. I hate this decision since this allows some awful
	// URLs https://www.gorillatoolkit.org/pkg/mux#Router.SkipClean
	r = r.SkipClean(true)

	monkey := func(fn handleFunc, prefix string) http.Handler {
		return handle(fn, prefix, store, server)
	}

	r.HandleFunc("/health", healthHandler)
	r.PathPrefix("/static").Handler(static)
	r.NotFoundHandler = index

	api := r.PathPrefix("/api").Subrouter()

	api.Handle("/login", monkey(loginHandler, ""))
	api.Handle("/signup", monkey(signupHandler, ""))
	api.Handle("/renew", monkey(renewHandler, ""))

	users := api.PathPrefix("/users").Subrouter()
	users.Handle("", monkey(usersGetHandler, "")).Methods("GET")
	users.Handle("", monkey(userPostHandler, "")).Methods("POST")
	users.Handle("/{id:[0-9]+}", monkey(userPutHandler, "")).Methods("PUT")
	users.Handle("/{id:[0-9]+}", monkey(userGetHandler, "")).Methods("GET")
	users.Handle("/{id:[0-9]+}", monkey(userDeleteHandler, "")).Methods("DELETE")

	api.PathPrefix("/resources").Handler(monkey(resourceGetHandler, "/api/resources")).Methods("GET")
	api.PathPrefix("/resources").Handler(monkey(resourceDeleteHandler(fileCache), "/api/resources")).Methods("DELETE")
	api.PathPrefix("/resources").Handler(monkey(resourcePostHandler(fileCache), "/api/resources")).Methods("POST")
	api.PathPrefix("/resources").Handler(monkey(resourcePutHandler, "/api/resources")).Methods("PUT")
	api.PathPrefix("/resources").Handler(monkey(resourcePatchHandler(fileCache), "/api/resources")).Methods("PATCH")

	api.PathPrefix("/tus").Handler(monkey(tusPostHandler(), "/api/tus")).Methods("POST")
	api.PathPrefix("/tus").Handler(monkey(tusHeadHandler(), "/api/tus")).Methods("HEAD", "GET")
	api.PathPrefix("/tus").Handler(monkey(tusPatchHandler(), "/api/tus")).Methods("PATCH")

	api.PathPrefix("/usage").Handler(monkey(diskUsage, "/api/usage")).Methods("GET")

	api.Path("/shares").Handler(monkey(shareListHandler, "/api/shares")).Methods("GET")
	api.PathPrefix("/share").Handler(monkey(shareGetsHandler, "/api/share")).Methods("GET")
	api.PathPrefix("/share").Handler(monkey(sharePostHandler, "/api/share")).Methods("POST")
	api.PathPrefix("/share").Handler(monkey(shareDeleteHandler, "/api/share")).Methods("DELETE")

	api.Handle("/settings", monkey(settingsGetHandler, "")).Methods("GET")
	api.Handle("/settings", monkey(settingsPutHandler, "")).Methods("PUT")

	api.PathPrefix("/raw").Handler(monkey(rawHandler, "/api/raw")).Methods("GET")
	api.PathPrefix("/preview/{size}/{path:.*}").
		Handler(monkey(previewHandler(imgSvc, fileCache, server.EnableThumbnails, server.ResizePreview), "/api/preview")).Methods("GET")
	api.PathPrefix("/command").Handler(monkey(commandsHandler, "/api/command")).Methods("GET")
	api.PathPrefix("/search").Handler(monkey(searchHandler, "/api/search")).Methods("GET")

	public := api.PathPrefix("/public").Subrouter()
	public.PathPrefix("/dl").Handler(monkey(publicDlHandler, "/api/public/dl/")).Methods("GET")
	public.PathPrefix("/share").Handler(monkey(publicShareHandler, "/api/public/share/")).Methods("GET")

	// 百度网盘相关接口
	bd := api.PathPrefix("/bd").Subrouter()
	bd.PathPrefix("/login").Handler(monkey(bdLogin, "/api/bd/login/")).Methods("POST")
	bd.PathPrefix("/user-info").Handler(monkey(bdUserInfo, "/api/bd/user-info/")).Methods("POST")
	bd.PathPrefix("/dir").Handler(monkey(bdShowDirInfo, "/api/bd/dir/")).Methods("POST")
	bd.PathPrefix("/download").Handler(monkey(bdDownLoad, "/api/bd/download/")).Methods("POST")
	bd.PathPrefix("/progress").Handler(monkey(bdDownloadProgress, "/api/bd/progress/")).Methods("POST")
	bd.PathPrefix("/progress").Handler(monkey(bdDeleteDownloadProgress, "/api/bd/progress/")).Methods("DELETE")
	bd.PathPrefix("/access-token/refresh").Handler(monkey(bdRefreshAccessToken, "/api/bd/access-token/refresh")).Methods("PUT")
	bd.PathPrefix("/access-token").Handler(monkey(bdGetAccessToken, "/api/bd/access-token/")).Methods("GET")

	// 端脑云存储相关接口
	cephalonDisk := api.PathPrefix("/cd").Subrouter()
	cephalonDisk.PathPrefix("/download").Handler(monkey(cephalonDiskDownload, "/api/cd/download")).Methods("POST")
	cephalonDisk.PathPrefix("/dir").Handler(monkey(cephalonDirInfo, "/api/cd/dir")).Methods("GET")
	cephalonDisk.PathPrefix("/user-space").Handler(monkey(cephalonUserSpace, "/api/cd/user-space")).Methods("GET")
	cephalonDisk.PathPrefix("/download/size").Handler(monkey(cephalonDiskDownloadProgress, "/api/cd/download/size")).Methods("GET")
	return stripPrefix(server.BaseURL, r), nil
}
