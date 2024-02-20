package server

import (
	"first_project/agent"
	"first_project/functions"
	"html/template"
	"net/http"
	"strconv"
)

func DoArithmetic(w http.ResponseWriter, r *http.Request) { // Отвечает за отрисовку главной странницы
	agent.Start()
	temp, _ := template.ParseFiles("templates/do_arithmetic.html")
	current_user := functions.CurrentAgent()
	current_task, prev, res := functions.CurrentTask(current_user)
	temp.Execute(w, agent.CreateAgent(current_user, current_task, prev, res))
}

func SaveData(w http.ResponseWriter, r *http.Request) { // Обработка данных с главной странницы (агента, примеры)
	status := http.StatusOK
	if r.Method == http.MethodPost {
		current_user := functions.CurrentAgent()
		current_task, _, _ := functions.CurrentTask(current_user)
		new_task := r.FormValue("task")
		new_agent := r.FormValue("new_user")
		if new_task != "" && current_task == "" || new_agent != "" {
			if (!functions.CheckTask(new_task) || new_task == "") && new_agent == "" {
				status = http.StatusBadRequest
			} else if current_user == "" && new_agent == "" {
				status = http.StatusBadRequest
			} else {
				if new_agent == "" || new_agent == current_user {
					if functions.IsFinished(current_user) || !functions.HasTask(current_user) {
						functions.WriteNewTask(new_task, current_user)
					} else {
						status = http.StatusConflict
					}
				} else {
					if !functions.IsPersonInDB(new_agent) {
						functions.WriteNewUser(new_agent)
					}
					if new_task != "" && functions.CheckTask(new_task) {
						functions.WriteNewTask(new_task, new_agent)
					} else if new_task == "" && new_agent != "" {
						functions.SwapUser(new_agent)
					}
				}
			}
		}
	}
	temp, _ := template.ParseFiles("templates/redirect.html")
	temp.Execute(w, status)
}

func ShowData(w http.ResponseWriter, r *http.Request) { // Показывает список выражений
	tasks := functions.GetTasks()
	temp, _ := template.ParseFiles("templates/show_data.html")
	temp.Execute(w, tasks)
}

func OperationTime(w http.ResponseWriter, r *http.Request) { // Странница с возможностью регулировать время выполнения операций
	ops := functions.GetOperations()
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
		functions.UpdateOperations(plus, minus, multi, division)
	}
	temp, _ := template.ParseFiles("templates/redirect.html")
	temp.Execute(w, status)
}

func StartServer() {
	agent.Start()
	http.HandleFunc("/", http.HandlerFunc(DoArithmetic))
	http.HandleFunc("/save_data", http.HandlerFunc(SaveData))
	http.HandleFunc("/show_data", http.HandlerFunc(ShowData))
	http.HandleFunc("/operation_time", http.HandlerFunc(OperationTime))
	http.HandleFunc("/update_operations", http.HandlerFunc(UpdateOperations))
	http.ListenAndServe(":8080", nil)
}
