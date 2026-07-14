package main

import (
	"log"
	"net/http"
	"os"

	planeapi "github.com/andrelair-platform/minicloud-plane/internal/api"
	natspub "github.com/andrelair-platform/minicloud-plane/internal/nats"
	"github.com/andrelair-platform/minicloud-plane/internal/plane"
	"github.com/andrelair-platform/minicloud-plane/internal/webhook"
)

func main() {
	planeURL := mustEnv("PLANE_URL")        // https://plane.devandre.sbs
	planeToken := mustEnv("PLANE_TOKEN")    // API token from Plane god-mode
	workspace := mustEnv("PLANE_WORKSPACE") // workspace slug
	natsURL := getEnv("NATS_URL", "nats://nats.nats.svc.cluster.local:4222")
	webhookSecret := getEnv("PLANE_WEBHOOK_SECRET", "")
	port := getEnv("PORT", "8080")

	planeClient := plane.NewClient(planeURL, planeToken, workspace)

	publisher, err := natspub.NewPublisher(natsURL, workspace)
	if err != nil {
		log.Fatalf("NATS connect failed: %v", err)
	}
	defer publisher.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"minicloud-plane"}`))
	})

	mux.Handle("/webhook", webhook.NewHandler(webhookSecret, publisher))
	mux.Handle("/api/", planeapi.NewHandler(planeClient))

	log.Printf("minicloud-plane listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
