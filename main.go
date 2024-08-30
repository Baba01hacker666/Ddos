package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	botAPI       *tgbotapi.BotAPI
	startTime    time.Time
	statusMutex  sync.Mutex
	lastPingTime time.Time
)

func init() {
	startTime = time.Now()
	lastPingTime = startTime
}

func stressTest(url string, concurrency int, totalRequests int) string {
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < totalRequests/concurrency; j++ {
				resp, err := http.Get(url)
				if err != nil {
					fmt.Println("Error:", err)
					continue
				}
				resp.Body.Close()
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)
	return fmt.Sprintf("Completed %d requests in %v", totalRequests, duration)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	statusMutex.Lock()
	defer statusMutex.Unlock()

	uptime := time.Since(startTime)
	lastPing := time.Since(lastPingTime)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Bot Status</title>
    <style>
        body { font-family: Arial, sans-serif; }
        .container { width: 80%%; margin: auto; padding: 20px; }
        h1 { color: #333; }
        p { font-size: 18px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Bot Status</h1>
        <p><strong>Status:</strong> up</p>
        <p><strong>Uptime:</strong> %s</p>
        <p><strong>Last Ping:</strong> %s</p>
        <p><strong>Latency:</strong> %s</p>
    </div>
</body>
</html>`, uptime, lastPing, latency())
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	statusMutex.Lock()
	lastPingTime = time.Now()
	statusMutex.Unlock()
	fmt.Fprintln(w, "Ping received")
}

func latency() string {
	// Placeholder for latency calculation
	return "N/A"
}

func main() {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		fmt.Println("Error: TELEGRAM_BOT_TOKEN environment variable is not set.")
		return
	}

	var err error
	botAPI, err = tgbotapi.NewBotAPI(botToken)
	if err != nil {
		fmt.Println("Error creating bot:", err)
		return
	}

	fmt.Printf("Authorized on account %s\n", botAPI.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := botAPI.GetUpdatesChan(u)

	go func() {
		http.HandleFunc("/", indexHandler)
		http.HandleFunc("/ping", pingHandler)
		port := "8080"
		fmt.Printf("Starting HTTP server on :%s\n", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			fmt.Println("Error starting HTTP server:", err)
		}
	}()

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Send the URL to stress test.")
				botAPI.Send(msg)
			case "test":
				url := update.Message.CommandArguments()
				if url == "" {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Please provide a URL.")
					botAPI.Send(msg)
					continue
				}

				result := stressTest(url, 100, 9999) // Adjust concurrency and totalRequests as needed
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, result)
				botAPI.Send(msg)
			default:
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command. Use /start to get instructions.")
				botAPI.Send(msg)
			}
		}
	}
}
