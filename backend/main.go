package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"subhub-backend/database"
	"subhub-backend/handlers"
	
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db := database.ConnectionDB()
	defer func() {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	database.RunMigrations(db)

	http.HandleFunc("/api/auth/register",
		handlers.AuthRegisterHandler(db))
	http.HandleFunc("/api/auth/login",
		handlers.AuthLoginHandler(db))
	http.HandleFunc("/api/subscriptions",
		handlers.SubscriptionsHandler(db))
	http.HandleFunc("/api/subscriptions/",
		handlers.SubscriptionByIDHandler(db))
	http.HandleFunc("/api/analytics",
		handlers.AnalyticsHandler(db))
	http.HandleFunc("/api/profile",
		handlers.ProfileHandler(db))
	http.HandleFunc("/api/groups",
		handlers.GroupsHandler(db))
	http.HandleFunc("/api/groups/",
		handlers.GroupByIDHandler(db))
	http.HandleFunc("/api/groups/join",
		handlers.JoinGroupHandler(db))

	registerFrontendRoutes()

	fmt.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func registerFrontendRoutes() {
	frontendDir, err := resolveFrontendDir()
	if err != nil {
		log.Printf("frontend directory is unavailable: %v", err)
		return
	}

	http.Handle("/assets/", http.StripPrefix("/assets/",
		http.FileServer(http.Dir(filepath.Join(frontendDir, "assets")))))

	http.Handle("/pages/", http.StripPrefix("/pages/",
		http.FileServer(http.Dir(filepath.Join(frontendDir, "pages")))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, filepath.Join(frontendDir, "pages", "index.html"))
			return
		}

		pageName := strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")
		if pageName == "" || strings.Contains(pageName, "/") ||
			!strings.HasSuffix(pageName, ".html") {
			http.NotFound(w, r)
			return
		}

		pagePath := filepath.Join(frontendDir, "pages", pageName)
		if info, err := os.Stat(pagePath); err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}

		http.ServeFile(w, r, pagePath)
	})
}

func resolveFrontendDir() (string, error) {
	candidates := []string{
		"frontend",
		filepath.Join("..", "frontend"),
	}

	for _, candidate := range candidates {
		frontendDir := filepath.Clean(candidate)
		info, err := os.Stat(frontendDir)
		if err == nil && info.IsDir() {
			return frontendDir, nil
		}
	}

	return "", fmt.Errorf("looked in %q and %q",
		candidates[0], candidates[1])
}
