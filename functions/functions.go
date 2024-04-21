package functions

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/mattn/go-sqlite3"
)

const hmacSampleSecret = "super_secret_signature"

func AllAgents() []string { // Возвращает всех пользователей из базы данных
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	stat, _ := db.Query("SELECT user_name FROM users")
	var one string
	var all []string
	for stat.Next() {
		stat.Scan(&one)
		all = append(all, one)
	}
	return all
}

func IsPersonInDB(agent string) bool { // Проверка на наличие в базе данных
	all := AllAgents()
	for _, one := range all {
		if one == agent {
			return true
		}
	}
	return false
}

func CurrentTask(name string) (string, string, float64) { // Возвращает текущую задачу текущего агента
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	stat, _ := db.Query("SELECT view FROM tasks WHERE user=? AND done=?", name, false)
	var res string
	for stat.Next() {
		stat.Scan(&res)
	}

	stat, _ = db.Query("SELECT view, result, time_ended FROM tasks WHERE user=? AND done=?", name, true)
	var views []string
	var results []float64
	var times []string
	for stat.Next() {
		var vi, ti string
		var re float64
		stat.Scan(&vi, &re, &ti)
		views = append(views, vi)
		results = append(results, re)
		times = append(times, ti)
	}
	if len(results) != 0 {
		lv := views[0]
		lr := results[0]
		lt := times[0]
		for i := 0; i < len(results); i++ {
			one := FromStrToTime(times[i])
			two := FromStrToTime(lt)
			if one.Compare(two) == 1 {
				lv = views[i]
				lr = results[i]
				lt = times[i]
			}
		}
		return res, lv, lr
	}
	return res, "", 0
}

func CheckTask(task string) bool { // Проверка на правильность выражения
	last_sym := ""
	for i, s := range task {
		oper := strings.Contains("+-*/", string(s))
		if oper && (last_sym == "oper" || last_sym == "" || i == len(task)-1) || !strings.Contains("0123456789-+/* ", string(s)) {
			return false
		}
		if oper {
			last_sym = "oper"
		} else {
			last_sym = "num"
		}
	}
	return true
}

func IsFinished(username string) bool { // Проверка закончено ли прошлое задание данного агента
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	com, _ := db.Query("SELECT done FROM tasks WHERE id_task=(SELECT last_task FROM users WHERE user_name=?)", username)
	var res bool
	for com.Next() {
		com.Scan(&res)
	}
	return res
}

func HasTask(username string) bool { // Проверка есть ли задание у данного пользователя
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	stat, _ := db.Query("SELECT last_task FROM users WHERE user_name=?", username)
	var a int
	for stat.Next() {
		stat.Scan(&a)
	}
	return a != 0
}

func WriteNewTask(task, name string) { // Обновляет текущее задание у агента
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	now := CreateTime(time.Now()).StringTime()
	com, _ := db.Prepare("INSERT INTO tasks(view, time_started, done, user) VALUES(?, ?, ?, ?)")
	com.Exec(task, now, false, name)
	com, _ = db.Prepare("UPDATE users SET last_task=(SELECT id_task FROM tasks WHERE user=? AND done=?) WHERE user_name=?")
	com.Exec(name, false, name)
}

func WriteNewUser(name, password string) { // Добавляет нового агента в дб
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	com, _ := db.Prepare("INSERT INTO users(user_name, password) VALUES(?, ?)")
	com.Exec(name, password)
	new_com, _ := db.Prepare("INSERT INTO operation_time(plus, minus, multi, divis, user) VALUES(?, ?, ?, ?, ?)")
	new_com.Exec(1, 1, 1, 1, name)
}

func ComparePasswords(user, password string) bool { // Проверяет пароль
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	stat, _ := db.Query("SELECT password FROM users WHERE user_name=?", user)
	var real_pass string
	for stat.Next() {
		stat.Scan(&real_pass)
	}
	return real_pass == password
}

type MyTime struct { // Структура для удобной обработки времени
	t        time.Time
	year     int
	month    int
	day      int
	hours    int
	minutes  int
	seconds  int
	nseconds int
}

func CreateTime(t time.Time) *MyTime { // Создание "Удобного времени"
	year, month_, day := t.Date()
	hours := t.Hour()
	minutes := t.Minute()
	seconds := t.Second()
	nseconds := t.Nanosecond()
	var month int
	for i, m := range strings.Split("January, February, March, April, May, June, July, August, September, October, November, December", ", ") {
		if m == month_.String() {
			month = i + 1
			break
		}
	}
	return &MyTime{t: t, year: year, month: month, day: day, hours: hours, minutes: minutes, seconds: seconds, nseconds: nseconds}
}

func (m *MyTime) StringTime() string { // Преобразует в строку время
	return strconv.Itoa(m.year) + "-" + strconv.Itoa(m.month) + "-" + strconv.Itoa(m.day) + " " + strconv.Itoa(m.hours) + ":" + strconv.Itoa(m.minutes) + ":" + strconv.Itoa(m.seconds) + "." + strconv.Itoa(m.nseconds)
}

func FromStrToTime(date string) time.Time {
	data := strings.Split(date, " ")
	ymd := strings.Split(data[0], "-")
	hms := strings.Split(data[1], ":")
	sn := strings.Split(hms[2], ".")
	year, _ := strconv.Atoi(ymd[0])
	month, _ := strconv.Atoi(ymd[1])
	day, _ := strconv.Atoi(ymd[2])
	hour, _ := strconv.Atoi(hms[0])
	mintes, _ := strconv.Atoi(hms[1])
	seconds, _ := strconv.Atoi(sn[0])
	nseconds, _ := strconv.Atoi(sn[1])
	return time.Date(year, time.Month(month), day, hour, mintes, seconds, nseconds, time.UTC)
}

type TasksInformation struct { // Хранит задания, помогает отображать данные на сайте
	Info []string
}

func GetTasks(name string) *TasksInformation { // Получает все задания агента, решённые и нерешённые
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	stat, _ := db.Query("SELECT done, view, time_started, time_ended FROM tasks WHERE user=?", name)
	var d bool
	var view, start, end string
	var done []bool
	var views, starts, ends []string
	for stat.Next() {
		stat.Scan(&d, &view, &start, &end)
		done = append(done, d)
		views = append(views, view)
		starts = append(starts, start)
		ends = append(ends, end)
	}
	res := []string{"Выражение        Выполнено        Дата начала        Дата оканчания"}
	for i := 0; i < len(done); i++ {
		res = append(res, views[i]+"    "+strconv.FormatBool(done[i])+"    "+starts[i]+"    "+ends[i])
	}
	return &TasksInformation{Info: res}
}

type Operations struct { // Хранит время выполнения операций
	Plus     int
	Minus    int
	Multi    int
	Division int
}

func GetOperations(user string) *Operations { // Получает из дб время выполнения операций
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	stat, _ := db.Query("SELECT plus, minus, multi, divis FROM operation_time WHERE user=?", user)
	var plus, minus, multi, div int
	for stat.Next() {
		stat.Scan(&plus, &minus, &multi, &div)
	}
	return &Operations{Plus: plus, Minus: minus, Multi: multi, Division: div}
}

func UpdateOperations(pl, mi, mu, di int, user string) { // Записывает новое время в дб
	if pl != 0 || mi != 0 || mu != 0 || di != 0 {
		db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
		defer db.Close()
		sent := "UPDATE operation_time SET"
		counter := 0
		if pl != 0 {
			sent += " plus=" + strconv.Itoa(pl)
			counter++
		}
		if mi != 0 {
			if counter != 0 {
				sent += ","
			}
			sent += " minus=" + strconv.Itoa(mi)
			counter++
		}
		if mu != 0 {
			if counter != 0 {
				sent += ","
			}
			sent += " multi=" + strconv.Itoa(mu)
			counter++
		}
		if di != 0 {
			if counter != 0 {
				sent += ","
			}
			sent += " divis=" + strconv.Itoa(di)
			counter++
		}
		sent += " WHERE user=?"
		com, _ := db.Prepare(sent)
		com.Exec(user)
	}
}

func GetToken(token string) string {
	tokenFromString, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			panic(fmt.Errorf("unexpected signing method: %v", token.Header["alg"]))
		}

		return []byte(hmacSampleSecret), nil
	})
	if claims, ok := tokenFromString.Claims.(jwt.MapClaims); ok {
		return fmt.Sprint(claims["name"]) 
	}
	return ""
}

func CreateToken(login string) string {
	now := time.Now()
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": login,
		"nbf":  now.Add(time.Minute).Unix(),
		"exp":  now.Add(24 * 7 * time.Hour).Unix(),
		"iat":  now.Unix(),
	}).SignedString([]byte(hmacSampleSecret))
	return token
}
