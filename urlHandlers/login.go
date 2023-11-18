package urlHandlers

import (
	"fmt"
	"html/template"
	"net/http"

	"forum/cleanData"
	"forum/dbconnections"
	"forum/validateData"

	"github.com/google/uuid"
)

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	template, err := template.ParseFiles("./templates/login.html")
	if err != nil {
		http.Error(w, "Template not found!"+err.Error(), http.StatusInternalServerError)
	}
	if r.Method != http.MethodPost {
		template.Execute(w, nil)
		return
	}

	formDataUsername := r.FormValue("username")
	formDataPassword := r.FormValue("password")

	errorLog := []string{}
	dataValid := true
	validatePassword, _ := validateData.ValidatePassword(formDataPassword, formDataPassword)
	if !validateData.ValidateName(formDataUsername) {
		errorLog = append(errorLog, "Username should be minimum 3 letters!")
		dataValid = false
	}
	if !validatePassword {
		errorLog = append(errorLog, "Password should be at least 6 letters long!")
		dataValid = false
	}
	if !dataValid {
		executeErr := template.Execute(w, errorLog)
		if executeErr != nil {
			fmt.Println("Template error: ", executeErr)
		}
		return
	}

	username := cleanData.CleanName(formDataUsername)
	if dbconnections.LoginUser(username, formDataPassword) {
		fmt.Println("USER LOGGED IN WITH RIGHT CREDENTIALS!")
		id := uuid.New()
		fmt.Println("User Cookie to be saved to DB and forward to Browser DB: ", id.String())
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	executeErr := template.Execute(w, "Username or Password incorrect")
	if executeErr != nil {
		fmt.Println("Template error: ", executeErr)
	}
}