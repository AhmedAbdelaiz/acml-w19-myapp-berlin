package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
)

var (
	dbHost     = envOrDefault("MYAPP_DATABASE_HOST", "localhost")
	dbPort     = envOrDefault("MYAPP_DATABASE_PORT", "5432")
	dbUser     = envOrDefault("MYAPP_DATABASE_USER", "root")
	dbPassword = envOrDefault("MYAPP_DATABASE_PASSWORD", "secret")
	dbName     = envOrDefault("MYAPP_DATABASE_NAME", "myapp")

	cacheHost = envOrDefault("MYAPP_CACHE_HOST", "localhost")
	cachePort = envOrDefault("MYAPP_CACHE_PORT", "6379")

	webHost = envOrDefault("MYAPP_WEB_HOST", "")
	webPort = envOrDefault("MYAPP_WEB_PORT", "8080")

	db    *sql.DB
	cache *redis.Client
)

func envOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT * FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	fmt.Fprintln(w, "ID | Name")
	fmt.Fprintln(w, "---+--------")
	for rows.Next() {
		var (
			id   int
			name string
		)

		rows.Scan(&id, &name)

		fmt.Fprintf(w, "%2d | %s\n", id, name)
	}
}

func myCachedHandler(w http.ResponseWriter, r *http.Request) {
	n, err := cache.Get("n").Result()

	if err == redis.Nil {
		n = strconv.Itoa(rand.Intn(100))
		cache.Set("n", n, 5*time.Second)
	}

	fmt.Fprintf(w, "n = %s\n", n)
}

func main() {
	dbInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	var err error
	db, err = sql.Open("postgres", dbInfo)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	cache = redis.NewClient(&redis.Options{
		Addr: cacheHost + ":" + cachePort,
	})
	if _, err := cache.Ping().Result(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", myHandler)
	http.HandleFunc("/cache", myCachedHandler)
	log.Print("Listening on " + webHost + ":" + webPort + "...")
	http.ListenAndServe(webHost+":"+webPort, nil)
}
