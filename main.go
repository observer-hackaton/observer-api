package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/garyburd/redigo/redis"
	"log"
	"math/rand"
	"time"
	"net/http"
)

type Monitor struct {
	ID      string `json:"id"`
	Monitor struct {
		Check struct {
			Arguments string `json:"arguments"`
			Interval  int    `json:"interval"`
			Type      string `json:"type"`
		} `json:"check"`
		Notifier struct {
			Arguments string `json:"arguments"`
			Type      string `json:"type"`
		} `json:"notifier"`
	} `json:"monitor"`
}


var pool *redis.Pool

func main() {

	pool = redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", "127.0.0.1:6379")
		if err != nil {
			return nil, err
		}
		return c, err
	},
		10)
	defer pool.Close()

	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/monitor", getMonitors),
		rest.Get("/monitor/#monitor", getMonitorId),
		rest.Post("/monitor", postMonitor),
		rest.Delete("/monitor/#monitor", deleteMonitorId),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(":8080", api.MakeHandler()))
}

func getMonitors(w rest.ResponseWriter, req *rest.Request) {
	client := pool.Get()
	defer client.Close()
	response, _ := redis.Strings(client.Do("KEYS", "monitor-*"))
	w.WriteJson(&response)
}

func getMonitorId(w rest.ResponseWriter, req *rest.Request) {
	client := pool.Get()
	defer client.Close()
	response, _ := redis.String(client.Do("GET", "monitor-" + req.PathParam("monitor")))
	if response == "" {
		w.WriteHeader(404)
		w.WriteJson("monitor not found")
	} else {
		w.WriteJson(&response)
	}
}

func postMonitor(w rest.ResponseWriter, req *rest.Request) {
	var monitor Monitor
	err := req.DecodeJsonPayload(&monitor)
	if monitor.ID == "" {
		monitor.ID = RandomString(12)
		client := pool.Get()
		defer client.Close()

		monitorStr, err := json.Marshal(monitor)
    if err != nil {
        fmt.Println(err)
        return
    }

		client.Do("SET", "monitor-" + monitor.ID, monitorStr)
	} else {
		w.WriteHeader(400)
		w.WriteJson("monitor ID should be empty when creating, use PATCH for modification")
	}
	if err != nil {
		w.WriteHeader(400)
		w.WriteJson("invalid monitor payload")
	} else {
		w.WriteJson("monitor created")
	}
}

func deleteMonitorId(w rest.ResponseWriter, req *rest.Request) {
	client := pool.Get()
	defer client.Close()
	client.Do("DEL", "monitor-" + req.PathParam("monitor"))
	w.WriteJson("removed.")
}


func RandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
