package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/m-ariany/telegram-gpt-bot/internal/config"
	"github.com/m-ariany/telegram-gpt-bot/internal/limiter"
	"github.com/m-ariany/telegram-gpt-bot/internal/retry"

	"github.com/go-redis/redis_rate/v10"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	gpt "github.com/m-ariany/gpt-chat-client"
	"github.com/redis/go-redis/v9"
)

const (
	instructionMsg = `Adopt the persona of a Senior Software Engineer with a specialization in Golang. 
	Address the user's inquiries with precision, adhering strictly to best practices in software engineering and Go. 
	Should you incorporate Golang code snippets in your replies, ensure they are properly enclosed within triple backticks for clarity. 
	All responses should be provided in Farsi language.
	`
)

func main() {

	cnf := config.LoadConfigOrPanic()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	redisClient := redis.NewClient(&redis.Options{
		Addr:       cnf.Redis.Address,
		Password:   cnf.Redis.Password,
		ClientName: "golang",
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		panic(err)
	}

	redisLimiter := redis_rate.NewLimiter(redisClient)
	messageRateLimiter := limiter.NewLimiter(redisLimiter, "user-message").PerDay(cnf.Telegram.MessageRateLimit)

	var temperature float32 = 0.0
	gptClient, err := gpt.NewClient(gpt.ClientConfig{
		ApiUrl:      "https://api.gilas.io/v1",
		ApiKey:      cnf.Gilas.ApiKey,
		ApiTimeout:  time.Minute * 2,
		Model:       "gpt-3.5-turbo",
		Temperature: &temperature,
	})
	if err != nil {
		panic(err)
	}

	bot, err := tgbotapi.NewBotAPI(cnf.Telegram.ApiKey)
	if err != nil {
		panic(err)
	}
	//bot.Debug = true

	// Create a new UpdateConfig struct with an offset of 0. Offsets are used
	// to make sure Telegram knows we've handled previous values and we don't
	// need them repeated.
	updateConfig := tgbotapi.NewUpdate(0)

	// Tell Telegram we should wait up to 30 seconds on each request for an
	// update. This way we can get information just as quickly as making many
	// frequent requests without having to send nearly as many.
	updateConfig.Timeout = 30

	// Start listening to the incomming messages
	updates := bot.GetUpdatesChan(updateConfig)

	for {
		select {
		case <-sigs:
			os.Exit(1)
		case update, ok := <-updates:
			if !ok { // channel is closed
				os.Exit(1)
			}

			go func() {

				// We only want to look at messages that mention the bot.
				if update.Message == nil || !strings.Contains(update.Message.Text, fmt.Sprintf("@%s", cnf.Telegram.BotName)) {
					return
				}

				// Only answer messages from a specific group
				if update.Message.Chat == nil || update.Message.Chat.ID != cnf.Telegram.GroupId {
					return
				}

				// Remove the bot name from the message
				msgText := strings.ReplaceAll(update.Message.Text, fmt.Sprintf("@%s", cnf.Telegram.BotName), "")

				// Skip empty messages
				if len(strings.TrimSpace(msgText)) == 0 {
					return
				}

				// Generally, every message should have a From
				if update.Message.From == nil || update.Message.From.ID == 0 {
					return
				}

				userId := strconv.FormatInt(update.Message.From.ID, 10)
				result, err := messageRateLimiter.Allow(ctx, userId) // Apply rate limit to the user
				if err != nil {
					return
				}

				chReply := make(chan string, 2) // maximum 2 messages will be written on this channel
				defer close(chReply)

				go func(ch <-chan string) {
					retryHandler := retry.NewRetryHandler(time.Second, time.Millisecond*500, 5)
					for replyMsg := range ch {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, replyMsg)
						msg.ReplyToMessageID = update.Message.MessageID
						retryHandler.Do(func() error {
							if _, err := bot.Send(msg); err != nil {
								log.Printf("failed to send message %+v", err)
								return err
							}
							return nil
						})
					}
				}(chReply)

				replyMessage := ""
				if result.Allowed == 0 {
					replyMessage = fmt.Sprintf("من فقط به %d تا سوال هر نفر در روز جواب میدم 🙈 لطفا بقیه کمک کنن!", cnf.Telegram.MessageRateLimit)
					chReply <- replyMessage
				} else {
					replyMessage = "یکم صبر کن الان جوابت رو میدم."
					chReply <- replyMessage

					c := gptClient.Clone()
					c.Instruct(instructionMsg)
					replyMessage, err = c.Prompt(ctx, msgText)
					if err != nil {
						replyMessage = "نشد که به Gilas.io .وصل شم😢 اگه بازم این اتفاق افتاد به میلاد خبر بدین."
						log.Printf("failed to prompt the model %+v", err)
					}
					chReply <- replyMessage
				}
			}()
		}
	}
}
