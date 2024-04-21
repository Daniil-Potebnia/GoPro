package server

import (
	"encoding/json"
	"first_project/agent"
	"first_project/functions"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"
)

func DoArithmetic(w http.ResponseWriter, r *http.Request) { // Отвечает за отрисовку главной странницы
	agent.Start()
	temp, _ := template.ParseFiles("templates/do_arithmetic.html")
	token, err := r.Cookie("jwt_token")
	var current_user, current_task, prev string
	var res float64
	if err == nil && token.Value != "" {
		current_user := functions.GetToken(token.Value)
		current_task, prev, res := functions.CurrentTask(current_user)
		temp.Execute(w, agent.CreateAgent(current_user, current_task, prev, res))
	} else {
		temp.Execute(w, agent.CreateAgent(current_user, current_task, prev, res))
	}
}

func SaveData(w http.ResponseWriter, r *http.Request) { // Обработка данных с главной странницы (агента, примеры)
	status := http.StatusOK
	if r.Method == http.MethodPost {
		token, _ := r.Cookie("jwt_token")
		current_user := functions.GetToken(token.Value)
		current_task, _, _ := functions.CurrentTask(current_user)
		new_task := r.FormValue("task")
		if new_task != "" && current_task == "" {
			if !functions.CheckTask(new_task) {
				status = http.StatusBadRequest
			} else {
				functions.WriteNewTask(new_task, current_user)
			}
		}
	}
	temp, _ := template.ParseFiles("templates/redirect.html")
	temp.Execute(w, status)
}

func ShowData(w http.ResponseWriter, r *http.Request) { // Показывает список выражений
	token, _ := r.Cookie("jwt_token")
	tasks := functions.GetTasks(functions.GetToken(token.Value))
	temp, _ := template.ParseFiles("templates/show_data.html")
	temp.Execute(w, tasks)
}

func OperationTime(w http.ResponseWriter, r *http.Request) { // Странница с возможностью регулировать время выполнения операций
	token, _ := r.Cookie("jwt_token")
	ops := functions.GetOperations(functions.GetToken(token.Value))
	temp, _ := template.ParseFiles("templates/operation_time.html")
	temp.Execute(w, ops)
}

func UpdateOperations(w http.ResponseWriter, r *http.Request) { // Обрабатывает данные со странницы операций
	status := http.StatusOK
	if r.Method == http.MethodPost {
		plus, _ := strconv.Atoi(r.FormValue("plus"))
		minus, _ := strconv.Atoi(r.FormValue("minus"))
		multi, _ := strconv.Atoi(r.FormValue("multi"))
		division, _ := strconv.Atoi(r.FormValue("division"))
		token, _ := r.Cookie("jwt_token")
		functions.UpdateOperations(plus, minus, multi, division, functions.GetToken(token.Value))
	}
	temp, _ := template.ParseFiles("templates/redirect.html")
	temp.Execute(w, status)
}

func FormRegister(w http.ResponseWriter, r *http.Request) {
	temp, _ := template.ParseFiles("templates/reg_log.html")
	temp.Execute(w, "регистрации")
}

func FormLogin(w http.ResponseWriter, r *http.Request) {
	temp, _ := template.ParseFiles("templates/reg_log.html")
	temp.Execute(w, "входа")
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		login := r.FormValue("login")
		password := r.FormValue("password")
		if !functions.IsPersonInDB(login) && login != "" && password != "" {
			functions.WriteNewUser(login, password)
			token := functions.CreateToken(login)
			http.SetCookie(w, &http.Cookie{Name: "jwt_token", Value: token, Expires: time.Now().Add(24 * 7 * time.Hour)})
			temp, _ := template.ParseFiles("templates/redirect.html")
			temp.Execute(w, http.StatusOK)
		} else {
			temp, _ := template.ParseFiles("templates/redirect.html")
			temp.Execute(w, http.StatusConflict)
		}
	} else {
		temp, _ := template.ParseFiles("templates/redirect.html")
		temp.Execute(w, http.StatusBadRequest)
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		login := r.FormValue("login")
		password := r.FormValue("password")
		if functions.IsPersonInDB(login) && functions.ComparePasswords(login, password) {
			token := functions.CreateToken(login)
			http.SetCookie(w, &http.Cookie{Name: "jwt_token", Value: token, Expires: time.Now().Add(24 * 7 * time.Hour)})
			temp, _ := template.ParseFiles("templates/redirect.html")
			temp.Execute(w, http.StatusOK)
		} else {
			temp, _ := template.ParseFiles("templates/redirect.html")
			temp.Execute(w, http.StatusConflict)
		}
	} else {
		temp, _ := template.ParseFiles("templates/redirect.html")
		temp.Execute(w, http.StatusBadRequest)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		http.SetCookie(w, &http.Cookie{Name: "jwt_token", Value: "", Expires: time.Now().Add(24 * 7 * time.Hour)})
		temp, _ := template.ParseFiles("templates/redirect.html")
		temp.Execute(w, http.StatusOK)
	} else {
		temp, _ := template.ParseFiles("templates/redirect.html")
		temp.Execute(w, http.StatusBadRequest)
	}
}

type RegJs struct {
	Login    string
	Password string
}

func RegisterApi(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var data RegJs
		byte_data, _ := io.ReadAll(r.Body)
		json.Unmarshal(byte_data, &data)
		if !functions.IsPersonInDB(data.Login) && data.Login != "" && data.Password != "" {
			functions.WriteNewUser(data.Login, data.Password)
			token := functions.CreateToken(data.Login)
			http.SetCookie(w, &http.Cookie{Name: "jwt_token", Value: token, Expires: time.Now().Add(24 * 7 * time.Hour)})
			ans, _ := json.Marshal(map[string]int{"result": http.StatusOK})
			w.Write(ans)
		} else {
			ans, _ := json.Marshal(map[string]int{"result": http.StatusBadRequest})
			w.Write(ans)
		}
	}
}

func LoginApi(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var data RegJs
		byte_data, _ := io.ReadAll(r.Body)
		json.Unmarshal(byte_data, &data)
		if functions.IsPersonInDB(data.Login) && functions.ComparePasswords(data.Login, data.Password) {
			token := functions.CreateToken(data.Login)
			http.SetCookie(w, &http.Cookie{Name: "jwt_token", Value: token, Expires: time.Now().Add(24 * 7 * time.Hour)})
			ans := []byte(fmt.Sprintf(`{"result": %d, "token": "%s"}`, http.StatusOK, token))
			w.Write(ans)
		} else {
			ans := []byte(fmt.Sprintf(`{"result": %d, "token": %s}`, http.StatusBadRequest, `""`))
			w.Write(ans)
		}
	}
}

func StartServer() {
	agent.Start()
	http.HandleFunc("/", http.HandlerFunc(DoArithmetic))
	http.HandleFunc("/save_data", http.HandlerFunc(SaveData))
	http.HandleFunc("/show_data", http.HandlerFunc(ShowData))
	http.HandleFunc("/operation_time", http.HandlerFunc(OperationTime))
	http.HandleFunc("/update_operations", http.HandlerFunc(UpdateOperations))
	http.HandleFunc("/register_form", http.HandlerFunc(FormRegister))
	http.HandleFunc("/login_form", http.HandlerFunc(FormLogin))
	http.HandleFunc("/register", http.HandlerFunc(Register))
	http.HandleFunc("/login", http.HandlerFunc(Login))
	http.HandleFunc("/logout", http.HandlerFunc(Logout))

	http.HandleFunc("/api/v1/register", http.HandlerFunc(RegisterApi))
	http.HandleFunc("/api/v1/login", http.HandlerFunc(LoginApi))

	http.ListenAndServe(":8080", nil)
}
