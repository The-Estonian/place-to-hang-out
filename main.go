package main

import (
	// "crypto/tls"
	"fmt"
	"net/http"

	"forum/database"
	"forum/urlHandlers"
	"forum/validateData"

	_ "github.com/mattn/go-sqlite3"
	// "golang.org/x/crypto/acme/autocert"
)

func main() {
	database.Engine()
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Commented out code is for valid Domains
	// certManager := autocert.Manager{
	// 	Prompt:     autocert.AcceptTOS,
	// 	HostPolicy: autocert.HostWhitelist("Domain name here!"),
	// }
	// server.TLSConfig = &tls.Config{
	// 	MinVersion:               tls.VersionTLS12,
	// 	PreferServerCipherSuites: true,
	// 	InsecureSkipVerify:       true,
	// 	// GetCertificate:           certManager.GetCertificate,
	// }

	staticFiles := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", staticFiles))
	mux.HandleFunc("/register", urlHandlers.HandleRegister)
	mux.HandleFunc("/login", urlHandlers.HandleLogin)
	mux.HandleFunc("/logout", urlHandlers.HandleLogout)
	mux.HandleFunc("/", urlHandlers.HandleForum)
	mux.HandleFunc("/post", urlHandlers.HandlePost)
	mux.HandleFunc("/postcontent", urlHandlers.HandlePostContent)
	mux.HandleFunc("/profile", urlHandlers.HandleProfile)
	mux.HandleFunc("/googleAuth", urlHandlers.HandleGoogleAuth)
	mux.HandleFunc("/githubAuth", urlHandlers.HandleGithubAuth)

	fmt.Println("Server hosted at: https://localhost:" + "8080")
	fmt.Println("To Kill Server press Ctrl+C")

	err := server.ListenAndServe()
	validateData.CheckErr(err)
}
