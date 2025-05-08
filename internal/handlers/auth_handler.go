package handlers

import (
	"database/sql"
	"errors"
	"net"
	"net/http"
	"time"

	"template/internal/repository"
	"template/internal/services"
	"template/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:80/auth/google/callback",
		ClientID:     "",
		ClientSecret: "",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
	oauthStateString = "random"
)

func (uh *UserHandler) GetRegister(w http.ResponseWriter, r *http.Request) {}
func (uh *UserHandler) GetLogin(w http.ResponseWriter, r *http.Request)    {}
func (uh *UserHandler) PostRegister(w http.ResponseWriter, r *http.Request) {
	email, pwd, repeat := utils.CleanString(r.FormValue("email")), r.FormValue("password"), r.FormValue("repeat")

	if repeat != pwd || !utils.IsValidEmail(email) || len(pwd) < 8 || len(pwd) > 72 {
		unprocessable(w)
		return
	}

	// Password handled as sql.NullString for consistency with repo/service
	passwordHash := sql.NullString{}
	if pwd != "" {
		passwordHash = sql.NullString{String: pwd, Valid: true}
	}

	u := repository.User{
		Email:        email,
		PasswordHash: passwordHash,
	}

	created, err := uh.US.Create(r.Context(), u)
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyExist) {
			conflict(w, err.Error())
			return
		}
		internal(w, err)
		return
	}
	cookieHash, err := uh.SS.CreateSession(r.Context(), created.ID, net.IP(utils.GetIPAddressBytes(r)), r.UserAgent())
	if err != nil {
		internal(w, err)
		return
	}

	// Set secure session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    cookieHash,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(24 * time.Hour.Seconds()),
	})

	http.Redirect(w, r, "/app", http.StatusSeeOther)
}
func (uh *UserHandler) PostLogin(w http.ResponseWriter, r *http.Request)  {}
func (uh *UserHandler) PostLogout(w http.ResponseWriter, r *http.Request) {}

func (uh *UserHandler) HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r,
		googleOauthConfig.AuthCodeURL(oauthStateString, oauth2.AccessTypeOffline),
		http.StatusTemporaryRedirect)
}

func (uh *UserHandler) LoginWithGoogle(w http.ResponseWriter, r *http.Request) {}
