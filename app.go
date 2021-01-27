// Web application daemon
package main

import (
	// some std libz
	"fmt"
	"os"
	"strconv"

	// eh well its a http server so ofc I need a HTTP server library
	"net/http"

	// pull in gorm
	"gorm.io/gorm"

	// only allow postgresql because im a masochist
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"

	// gorm prometheus exporter
	gormPrometheus "gorm.io/plugin/prometheus"

	// logger because I am not crazy
	log "github.com/sirupsen/logrus"

	// prometheus because im hardcore
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	"github.com/slok/go-http-metrics/middleware"
	"github.com/slok/go-http-metrics/middleware/std"
)

// Person a sql table xd
type Person struct {
	gorm.Model
	UID  uint
	Name string
	Age  int
}

var (
	db *gorm.DB
)

// NewUser makes a new User from a POST request
func NewUser(w http.ResponseWriter, r *http.Request) {
	// only allow http POST requests
	if r.Method == "POST" {
		// parse through the form
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		// Parse age and name from request form
		Name := r.FormValue("USERNAME")
		Age := r.FormValue("AGE")
		// check if we actually got a value
		if Name == "" || Age == "" {
			http.Error(w, "DID NOT RECEIVE A FIELD", http.StatusBadRequest)
			return
		}
		// err convert string to int
		AgeInt, err := strconv.Atoi(Age)
		if err != nil {
			http.Error(w, "Invalid Age Input", http.StatusBadRequest)
			return
		}
		//add to database
		fmt.Fprintf(w, "Successfully Received Process and will be added to database xd")
		db.Create(&Person{Name: Name, Age: AgeInt})
		log.Info("Added to database with Name " + Name + " and Age " + strconv.Itoa(AgeInt))
	} else {
		http.Error(w, "INVALID REQUEST METHOD", http.StatusBadRequest)
	}
}

func init() {
	// initialize logger
	log.SetOutput(os.Stdout)
	// connect to database
	dsn := os.Getenv("DB_DSN")
	var err error
	if dsn != "" {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatal("Database failed to connect with the DSN: " + dsn)
		}
	} else {
		db, err = gorm.Open(sqlite.Open("db.db"), &gorm.Config{})
		if err != nil {
			log.Fatal(err.Error())
		}
	}
	// register gorm promtheus plugin
	db.Use(gormPrometheus.New(gormPrometheus.Config{
		DBName:          "userdb",
		RefreshInterval: 5,
		StartServer:     true,
		HTTPServerPort:  9091,
	}))
	// execute the migration xd
	db.AutoMigrate(&Person{})
}

func main() {
	// prometheus middleware
	mdlw := middleware.New(middleware.Config{
		Recorder: metrics.NewRecorder(metrics.Config{}),
	})
	// register functions and serve requests
	mux := http.NewServeMux()
	mux.HandleFunc("/user/new", NewUser)
	h := std.Handler("", mdlw, mux)
	log.Info("start to Listen for prometheus data")
	go func() {
		// listen for prometheus data
		log.Fatal(http.ListenAndServe(":9090", promhttp.Handler()).Error())
	}()
	// listen for new requests
	log.Info("starting to listen for new requests at HTTP 0.0.0.0:6443")
	log.Fatal(http.ListenAndServe(":6443", h).Error())
}
