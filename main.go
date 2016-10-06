package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/caarlos0/alelogo"
	"github.com/garyburd/redigo/redis"
	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	pool := &redis.Pool{
		MaxIdle:   1,
		MaxActive: 10,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.DialURL(os.Getenv("REDIS_URL"))
			if err != nil {
				log.Panic("Could not connect to redis. Cause: " + err.Error())
				return nil, err
			}
			return conn, err
		},
	}
	defer pool.Close()

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Panic(err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// without a port binded, heroku complains and eventually kills the process.
	go serve()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	for update := range updates {
		update := update
		if update.Message == nil {
			continue
		}
		log.Println("Message from:", *update.Message.From)
		if update.Message.Command() == "login" {
			go login(pool, bot, update)
			continue
		}
		if update.Message.Command() == "balance" {
			go balance(pool, bot, update)
			continue
		}
		log.Println("Unknown command", update.Message.Text)
		bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Os únicos comandos suportados são /login e /balance",
		))
	}
}

func balance(pool *redis.Pool, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	id := string(update.Message.From.ID)
	cpf, _ := redis.String(pool.Get().Do("GET", id+".cpf"))
	pwd, _ := redis.String(pool.Get().Do("GET", id+".pwd"))
	if cpf == "" || pwd == "" {
		bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Por favor, faça login novamente...",
		))
		return
	}
	client, err := alelogo.New(cpf, pwd)
	if err != nil {
		log.Println(err.Error())
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
		return
	}
	cards, err := client.Balance()
	if err != nil {
		log.Println(err.Error())
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
		return
	}
	for _, card := range cards {
		bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Saldo do cartao "+strings.TrimSpace(card.Title)+" é "+card.Balance,
		))
	}
}

func login(pool *redis.Pool, bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	parts := strings.Split(
		strings.TrimSpace(update.Message.CommandArguments()), " ",
	)
	if len(parts) != 2 {
		bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"Para fazer login, diga\n\n/login CPF Senha",
		))
		return
	}
	cpf, pwd := parts[0], parts[1]
	_, err := alelogo.New(cpf, pwd)
	if err != nil {
		log.Println(err.Error())
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
		return
	}
	id := string(update.Message.From.ID)
	_, err = pool.Get().Do("SET", id+".cpf", cpf)
	if err != nil {
		log.Println(err.Error())
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
		return
	}
	_, err = pool.Get().Do("SET", id+".pwd", pwd)
	if err != nil {
		log.Println(err.Error())
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
		return
	}
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Sucesso!"))
}

func serve() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
