package functions

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func AllAgents() []string {
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

func IsPersonInDB(agent string) bool {
	all := AllAgents()
	for _, one := range all {
		if one == agent {
			return true
		}
	}
	return false
}

func CurrentAgent() string {
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	stat, _ := db.Query("SELECT user_name, entry_time FROM users")
	var name, t string
	var names, ts []string
	for stat.Next() {
		stat.Scan(&name, &t)
		names = append(names, name)
		ts = append(ts, t)
	}
	if len(names) != 0 {
		latest := names[0]
		lt := FromStrToTime(ts[0])
		for i := 0; i < len(names); i++ {
			test := FromStrToTime(ts[i])
			if test.Compare(lt) == 1 {
				lt = FromStrToTime(ts[i])
				latest = names[i]
			}
		}
		return latest
	}
	return ""
}

func CurrentTask(name string) (string, string, int) {
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	stat, _ := db.Query("SELECT view FROM tasks WHERE user=? AND done=?", name, false)
	var res string
	for stat.Next() {
		stat.Scan(&res)
	}

	stat, _ = db.Query("SELECT view, result, time_ended FROM tasks WHERE user=? AND done=?", name, true)
	var views []string
	var results []int
	var times []string
	for stat.Next() {
		var vi, ti string
		var re int
		stat.Scan(&vi, re, ti)
		views = append(views, vi)
		results = append(results, re)
		times = append(times, ti)
	}
	if len(results) != 0 {
		lv := views[0]
		lr := results[0]
		lt := times[0]
		for i, t := range times {
			one := FromStrToTime(t)
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

func CheckTask(task string) bool {
	last_sym := ""
	for i, s := range task {
		oper := strings.Contains("+-*/", string(s))
		if oper && (last_sym == "oper" || last_sym == "" || i == len(task)-1) {
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

func IsFinished(username string) bool {
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	com, _ := db.Query("SELECT done FROM tasks WHERE id_task=(SELECT last_task FROM users WHERE user_name=?)", username)
	var res bool
	for com.Next() {
		com.Scan(&res)
	}
	return res
}

func HasTask(username string) bool {
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	stat, _ := db.Query("SELECT last_task FROM users WHERE user_name=?", username)
	var a int
	for stat.Next() {
		stat.Scan(&a)
	}
	return a != 0
}

func WriteNewTask(task, name string) {
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	now := CreateTime(time.Now()).StringTime()
	com, _ := db.Prepare("INSERT INTO tasks(view, time_started, done, user) VALUES(?, ?, ?, ?)")
	com.Exec(task, now, false, name)
	com, _ = db.Prepare("UPDATE users SET entry_time=?, last_task=(SELECT id_task FROM tasks WHERE user=? AND done=?) WHERE user_name=?")
	com.Exec(now, name, false, name)
}

func WriteNewUser(name string) {
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	com, _ := db.Prepare("INSERT INTO users(user_name, entry_time) VALUES(?, ?)")
	com.Exec(name, CreateTime(time.Now()).StringTime())
}

type MyTime struct {
	t        time.Time
	year     int
	month    int
	day      int
	hours    int
	minutes  int
	seconds  int
	nseconds int
}

func CreateTime(t time.Time) *MyTime {
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

func (m *MyTime) StringTime() string {
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

type TasksInformation struct {
	Info []string
}

func GetTasks() *TasksInformation {
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	stat, _ := db.Query("SELECT done, view, time_started, time_ended FROM tasks")
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
		res = append(res, views[i] + "    " + strconv.FormatBool(done[i]) + "    " + starts[i] + "    " + ends[i])
	}
	return &TasksInformation{Info: res}
}

type Operations struct {
	Plus int
	Minus int
	Multi int
	Division int
}

func GetOperations() *Operations {
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	stat, _ := db.Query("SELECT time FROM operation_time")
	var time_ int
	var times []int
	for stat.Next() {
		stat.Scan(&time_)
		times = append(times, time_)
	}
	return &Operations{Plus: times[0], Minus: times[1], Multi: times[2], Division: times[3]}
}

func UpdateOperations(pl, mi, mu, di int) {
	db, _ := sql.Open("sqlite3", "./dbs/main_db.db")
	defer db.Close()
	com, _ := db.Prepare("UPDATE operation_time SET time=? WHERE symbol=?")
	com.Exec(pl, "+")
	com, _ = db.Prepare("UPDATE operation_time SET time=? WHERE symbol=?")
	com.Exec(mi, "-")
	com, _ = db.Prepare("UPDATE operation_time SET time=? WHERE symbol=?")
	com.Exec(mu, "*")
	com, _ = db.Prepare("UPDATE operation_time SET time=? WHERE symbol=?")
	com.Exec(di, "/")
}
