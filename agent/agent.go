package agent

import (
	"database/sql"
	"first_project/functions"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Agent struct {
	Name string
	Task string
	Prev string
	Res  int
}

func CreateAgent(name, task, prev string, res int) *Agent {
	return &Agent{Name: name, Task: task, Prev: prev, Res: res}
}

type MainWorker struct {
	Workers []Worker
}

func CreateMain() *MainWorker {
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	stat, _ := db.Query("SELECT id_task, view FROM tasks WHERE done=?", false)
	var id int
	var view string
	var ids []int
	var views []string
	for stat.Next() {
		stat.Scan(&id, &view)
		ids = append(ids, id)
		views = append(views, view)
	}
	var workers []Worker
	for i := 0; i < len(ids); i++ {
		workers = append(workers, *CreateWorker(ids[i], views[i]))
	}
	return &MainWorker{Workers: workers}
}

func (m *MainWorker) Start() {
	for _, w := range m.Workers {
		w.Parse()
		w.Solving()
	}
}

type Worker struct {
	Id      int
	Task    string
	Numbers chan string
	Res     chan int
	Maxgoru int
	Wg      *sync.WaitGroup
	Result  int
	Done    bool
}

func CreateWorker(id int, task string) *Worker {
	return &Worker{Id: id, Task: task, Numbers: make(chan string), Res: make(chan int), Maxgoru: 3, Wg: &sync.WaitGroup{}, Result: 0, Done: false}
}

func (w *Worker) ChangeNumOfOps(n int) {
	if n > 1 {
		w.Maxgoru = n
	}
}

func (w *Worker) Parse() {
	var n string
	for i := 0; i < len(w.Task); i++ {
		if strings.Contains("0123456789*/", string(w.Task[i])) {
			n += string(w.Task[i])
			if i == len(w.Task)-1 {
				go func() {
					w.Numbers <- n
				}()
				n = ""
			}
		} else if strings.Contains("+-", string(w.Task[i])) {
			go func() {
				w.Numbers <- n
			}()
			n = ""
			if w.Task == "-" {
				n += "-"
			}
		}
	}
}

func (w *Worker) Solving() {
	o := functions.GetOperations()
	for i := 0; i < w.Maxgoru; i++ {
		w.Wg.Add(1)
		go func() {
			defer w.Wg.Done()
			defer close(w.Numbers)
			defer close(w.Res)
			for {
				s := <-w.Numbers
				if strings.Contains(s, "*") || strings.Contains(s, "/") {
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
							first = false
						} else {
							first = true
							op = string(n)
						}
					}
					n1, _ := strconv.Atoi(num1)
					n2, _ := strconv.Atoi(num2)
					if op == "*" {
						w.Res <- n1 * n2
					} else if op == "/" {
						w.Res <- n1 / n2
					}
				} else {
					n, _ := strconv.Atoi(s)
					w.Res <- n
				}
			}
		}()
	}
	fmt.Println(0)
	w.Wg.Wait()
	w.Wg.Add(1)
	go func() {
		defer w.Wg.Done()
		defer close(w.Res)
		for k := range w.Res {
			if k > 0 {
				time.Sleep(time.Second * time.Duration(o.Plus))
				w.Result += k
			} else {
				time.Sleep(time.Second * time.Duration(o.Minus))
				w.Result += k
			}
		}
		w.Done = true
	}()
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	com, _ := db.Prepare("UPDATE tasks SET done=?, result=?, time_ended=? WHERE id_task=?")
	com.Exec(true, w.Result, functions.CreateTime(time.Now()).StringTime(), w.Id)
}
