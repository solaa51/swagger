package middleware

import (
	"expvar"
	"github.com/solaa51/swagger/cFunc"
	"net/http"
)

//允许跨域访问中间件

type CrossDomain struct {
}

// 记录访问量
var lastVisits = expvar.NewInt("last_day_visits")
var todayVisits = expvar.NewInt("visits")
var today = cFunc.Date("Y-m-d", 0)

func (C *CrossDomain) Handle(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, DELETE, POST, GET, PATCH, PUT")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Api-Key, X-Requested-With, Content-Type, Accept, Authorization, Token")

	td := cFunc.Date("Y-m-d", 0)
	if td != today {
		lastVisits.Set(todayVisits.Value())
		todayVisits.Set(0)
		today = td
	}
	todayVisits.Add(1)

	return true
}
