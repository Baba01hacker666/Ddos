package main

import (
    "fmt"
    "net/http"
    "os"
    "sync"
    "time"

    "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

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

func main() {
    botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
    if botToken == "" {
        fmt.Println("Error: TELEGRAM_BOT_TOKEN environment variable is not set.")
        return
    }

    bot, err := tgbotapi.NewBotAPI(botToken)
    if err != nil {
        fmt.Println("Error creating bot:", err)
        return
    }

    fmt.Printf("Authorized on account %s\n", bot.Self.UserName)

    u := tgbotapi.NewUpdate(0)
    u.Timeout = 60

    updates := bot.GetUpdatesChan(u)

    for update := range updates {
        if update.Message == nil {
            continue
        }

        if update.Message.IsCommand() {
            switch update.Message.Command() {
            case "start":
                msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Send the URL to stress test.")
                bot.Send(msg)
            case "test":
                url := update.Message.CommandArguments()
                if url == "" {
                    msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Please provide a URL.")
                    bot.Send(msg)
                    continue
                }

                result := stressTest(url, 100, 1000) // Adjust concurrency and totalRequests as needed
                msg := tgbotapi.NewMessage(update.Message.Chat.ID, result)
                bot.Send(msg)
            default:
                msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command. Use /start to get instructions.")
                bot.Send(msg)
            }
        }
    }
}

