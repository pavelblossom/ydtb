package bot

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"

	"ydtb/config"
)

type (
	Bot struct {
		api     *tgbotapi.BotAPI
		updates tgbotapi.UpdatesChannel
		conf    *config.Config
		videos  map[int][]Video // list videos for each chat
	}

	Callback struct {
		URL    string `json:"url"`
		ChatID int64  `json:"chat_id"`
	}

	Video struct {
		Title string
		ID    string
	}

	Message struct {
		chatID int64
		text   string
	}
)

func (b *Bot) sendButtonMessage(update *tgbotapi.Update, videos []Video) error {
	var buttonMessageRows [][]tgbotapi.InlineKeyboardButton

	for number := range videos {
		cb, err := json.Marshal(&Callback{
			URL:    videos[number].ID,
			ChatID: update.Message.Chat.ID},
		)
		if err != nil {
			return fmt.Errorf("Json marshal callback error: %w", err)
		}

		inlineKeyboardText := html.UnescapeString(videos[number].Title)
		inlineKeyBoardData := string(cb)

		buttonMessageRow := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(inlineKeyboardText, inlineKeyBoardData),
		)

		buttonMessageRows = append(buttonMessageRows, buttonMessageRow)
	}

	buttonMessage := tgbotapi.NewInlineKeyboardMarkup(buttonMessageRows...)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		fmt.Sprintf("Results on search:\n*%s*", update.Message.Text))
	msg.ReplyMarkup = buttonMessage
	msg.ParseMode = "Markdown"

	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("Error during button message sending: %w", err)
	}

	return nil
}

func (b *Bot) sendTextMessage(message Message) error {
	if message.chatID > 0 {
		reply := tgbotapi.NewMessage(message.chatID, message.text)
		reply.ParseMode = "Markdown"
		reply.DisableWebPagePreview = true
		_, err := b.api.Send(reply)
		if err != nil {
			return fmt.Errorf("Error during text message sending : %w", message, err)
		}
	} else {
		return nil
	}

	return nil
}

// New is a main layer with bot instance
func New(conf *config.Config) (*Bot, error) {
	client := &http.Client{}

	if conf.UseProxy {
		proxyURL, err := url.Parse(conf.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("parse proxy URL error: %w", err)
		}

		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	}

	botAPI, err := tgbotapi.NewBotAPIWithClient(conf.TelegramToken, client)
	if err != nil {
		return nil, fmt.Errorf("Bot start error: %w", err)
	}

	logrus.Println("Authorization is done for acc: ", botAPI.Self.UserName)

	updateConf := tgbotapi.NewUpdate(0)
	updateConf.Timeout = 60

	updatesChan, err := botAPI.GetUpdatesChan(updateConf)
	if err != nil {
		return nil, fmt.Errorf("Udpate channel error: %w", err)
	}

	return &Bot{api: botAPI, updates: updatesChan, conf: conf}, err
}

func (b *Bot) Start() {
	b.videos = make(map[int][]Video)
	var wg sync.WaitGroup
	errChan := make(chan error, b.conf.Concurrency)

	for update := range b.updates {
		update := update
		wg.Add(1)

		go func(errChan chan<- error) {
			defer wg.Done()
			var err error

			if update.Message == nil && update.CallbackQuery == nil {
				return
			}

			if update.Message != nil && !update.Message.IsCommand() {
				logrus.Printf("[MESSAGE] USERID: %d USERNAME: %s MESSAGE: %s\n",
					update.Message.From.ID,
					update.Message.From.UserName,
					update.Message.Text)

				err = b.implementMessage(&update)
				if err != nil {
					errChan <- fmt.Errorf("Error on start implementMessage: %w USERID: %d",
						err, update.Message.From.ID)

					err = b.sendTextMessage(Message{
						chatID: update.Message.Chat.ID,
						text:   messageImplementErr})
					if err != nil {
						logrus.Error(err)
					}
					return
				}
				errChan <- nil
				return
			}

			if update.Message != nil && update.Message.IsCommand() {
				logrus.Printf("[COMMAND] USERID: %d USERNAME: %s MESSAGE: %s\n",
					update.Message.From.ID,
					update.Message.From.UserName,
					update.Message.Text)
				err = b.implementCommands(&update)
				if err != nil {
					errChan <- fmt.Errorf("Error during implement command: %w USERID: %d",
						err, update.Message.From.ID)

					err = b.sendTextMessage(Message{
						chatID: update.Message.Chat.ID,
						text:   commandsImplementErr})
					if err != nil {
						logrus.Error(err)
					}
					return
				}
				errChan <- nil
				return
			}

			if update.Message == nil && update.CallbackQuery != nil {
				logrus.Printf("[CALLBACK] USERID: %d USERNAME: %s CALLBACKDATA: %s\n",
					update.CallbackQuery.From.ID,
					update.CallbackQuery.From.UserName,
					update.CallbackQuery.Data)
				err = b.implementCallback(&update)
				if err != nil {
					errChan <- fmt.Errorf("Error during implement callback: %w USERID: %d",
						err, update.CallbackQuery.Message.From.ID)

					err = b.sendTextMessage(Message{
						chatID: update.CallbackQuery.Message.Chat.ID,
						text:   callbackImplementErr})
					if err != nil {
						logrus.Error(err)
					}
					return
				}
				errChan <- nil
				return
			}
		}(errChan)
		wg.Wait()

		err := <-errChan
		if err != nil {
			logrus.Error(err)
		}
	}
}
