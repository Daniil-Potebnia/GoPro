package agent

import (
	"database/sql"
	"first_project/functions"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func Start() { // Запускает работу workers 
	main := CreateMain()
	main.Start()
}

type Agent struct { // Помогает в передачи данных на главнную странницу
	Name string
	Task string
	Prev string
	Res  float64
}

func CreateAgent(name, task, prev string, res float64) *Agent { // Конструктор агента
	return &Agent{Name: name, Task: task, Prev: prev, Res: res}
}

type MonitWorker interface { // Интерфейс мониторинга
	Start()
}

type MainWorker struct { // Мониторит работу рабочих
	Workers []Worker
}

func CreateMain() *MainWorker { // Конструктор монитора
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	stat, _ := db.Query("SELECT id_task, view, user FROM tasks WHERE done=?", false)
	var id int
	var view string
	var user string
	var ids []int
	var views []string
	var users []string
	for stat.Next() {
		stat.Scan(&id, &view, &user)
		ids = append(ids, id)
		views = append(views, view)
		users = append(users, user)
	}
	var workers []Worker
	for i := 0; i < len(ids); i++ {
		workers = append(workers, *CreateWorker(ids[i], views[i], users[i]))
	}
	return &MainWorker{Workers: workers}
}

func (m *MainWorker) Start() { // Начало работы и добавления рабочих(парсинг, начало решения)
	for _, w := range m.Workers {
		if !w.Started {
			w.Parse()
			w.Solving()
			w.Started = true
		}
	}
}

type Worker struct { // Структура рабочих, каждый отвечает за 1 задачу 1 пользователя
	Id      int
	User    string
	Task    string
	Numbers chan string
	Res     []float64
	Result  float64
	MaxGor  int
	Done    bool
	Started bool
	Mu      *sync.Mutex
	Length  int
}

func CreateWorker(id int, task, user string) *Worker { // Конструктор рабочего
	return &Worker{Id: id, User: user, Task: task, Numbers: make(chan string), Res: []float64{}, Result: 0, Done: false, Started: false, Mu: &sync.Mutex{}, MaxGor: 5}
}

func (w *Worker) Parse() { // Парсит примеры
	go func() {
		defer close(w.Numbers)
		var n string
		for i := 0; i < len(w.Task); i++ {
			if strings.Contains("0123456789*/", string(w.Task[i])) { // Добавляет в канал рабочего все слагаемые примера (5, -8, 4*7,  8/2, - всё это считается слагаемым) и считает их количество 
				n += string(w.Task[i])
				if i == len(w.Task)-1 {
					w.Numbers <- n
					w.Mu.Lock()
					w.Length++
					w.Mu.Unlock()
					n = ""
				}
			} else if strings.Contains("+-", string(w.Task[i])) {
				w.Numbers <- n
				w.Mu.Lock()
				w.Length++
				w.Mu.Unlock()
				n = ""
				if string(w.Task[i]) == "-" {
					n += "-"
				}
			}
		}
	}()
}

func (w *Worker) Solving() { // Решает пример
	o := functions.GetOperations(w.User)
	for i := 0; i < w.MaxGor; i++ { // Паралельно может быть столько решающих горутин, сколько слагаемых в примере
		go func() {
			for s := range w.Numbers {
				if strings.Contains(s, "*") || strings.Contains(s, "/") { // Если в слагаемом есть * или /, то оно решается, если нет - добавляется в слайс чисел
					num1 := ""
					op := ""
					first := false
					num2 := ""
					for _, n := range s {
						if strings.Contains("0123456789-", string(n)) {
							if first {
								num2 += string(n)
							} else {
								num1 += string(n)
							}
						} else if first {
							n1, _ := strconv.Atoi(num1)
							n2, _ := strconv.Atoi(num2)
							if op == "*" {
								time.Sleep(time.Second * time.Duration(o.Multi))
								num1 = strconv.Itoa(n1 * n2)
							} else if op == "/" {
								time.Sleep(time.Second * time.Duration(o.Division))
								num1 = strconv.Itoa(n1 / n2)
							}
							num2 = ""
							op = string(n)
						} else {
							first = true
							op = string(n)
						}
					}
					n1, _ := strconv.Atoi(num1)
					n2, _ := strconv.Atoi(num2)
					w.Mu.Lock()
					if op == "*" {
						time.Sleep(time.Second * time.Duration(o.Multi))
						w.Res = append(w.Res, float64(n1*n2))
					} else if op == "/" {
						time.Sleep(time.Second * time.Duration(o.Division))
						w.Res = append(w.Res, float64(n1)/float64(n2))
					}
					w.Mu.Unlock()
				} else {
					n, _ := strconv.Atoi(s)
					w.Mu.Lock()
					w.Res = append(w.Res, float64(n))
					w.Mu.Unlock()
				}
			}
			w.Mu.Lock()
			if w.Length == len(w.Res) && !w.Done { // Если все слагаемые упрощены, то начинается подсчёт окончательного ответа
				for _, k := range w.Res {
					if k > 0 {
						time.Sleep(time.Second * time.Duration(o.Plus))
						w.Result += k
					} else {
						time.Sleep(time.Second * time.Duration(o.Minus))
						w.Result += k
					}
				}
				w.Done = true
				db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
				defer db.Close()
				com, _ := db.Prepare("UPDATE tasks SET done=?, result=?, time_ended=? WHERE id_task=?") // Сохранение в базу данных
				com.Exec(true, w.Result, functions.CreateTime(time.Now()).StringTime(), w.Id)
			}
			w.Mu.Unlock()
		}()
	}
}
