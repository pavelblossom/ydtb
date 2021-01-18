package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/sirupsen/logrus"
)

// implementCallback implements callbacks (user menu selection from search result)
func (b *Bot) implementCallback(update *tgbotapi.Update) error {
	callback := Callback{}
	err := json.Unmarshal([]byte(update.CallbackQuery.Data), &callback)
	if err != nil {
		return err
	}

	go func() {
		err = b.sendTextMessage(Message{
			chatID: callback.ChatID,
			text:   downloadStarted})
		if err != nil {
			logrus.Error(err)
		}
	}()

	workDir, err := os.Getwd()
	if err != nil {
		return err
	}

	downloadsDir := filepath.Join(workDir, b.conf.DownloadsDir)
	chatIDDir := filepath.Join(downloadsDir, fmt.Sprint(update.CallbackQuery.Message.Chat.ID))

	filename, err := downloadAudioByLink(callback.URL, chatIDDir)
	if err != nil {
		return err
	}

	fullFilePath := filepath.Join(chatIDDir, filename)
	defer func() {
		err = removeContents(chatIDDir)
		if err != nil {
			logrus.Errorf("remove file %s after upload error: %v", fullFilePath, err)
		}
	}()

	_, err = b.api.Send(tgbotapi.NewAudioUpload(callback.ChatID, fullFilePath))

	return err
}

// implementMessage implements message updates (e.g. to initiate search or download by link)
func (b *Bot) implementMessage(update *tgbotapi.Update) error {
	regAnyURL, err := regexp.Compile(`(http|ftp|https)://([\w-]+(?:(?:\.[\w-]+)+))([\w.,@?^=%&:/~+#-]*[\w@?^=%&/~+#-])?`)
	if err != nil {
		return fmt.Errorf("regexp compile (find URL in message) err: %w", err)
	}

	switch regAnyURL.MatchString(update.Message.Text) {
	case true:
		var videoURL = regAnyURL.FindString(update.Message.Text)
		regYtURL, err := regexp.Compile(`(http(s|):|)\/\/(www\.|)yout(.*?)\/(embed\/|watch.*?v=|)([a-z_A-Z0-9\-]{11})`)
		if err != nil {
			return fmt.Errorf("regexp compile (find youtube link in message) err: %w", err)
		}

		if regYtURL.MatchString(videoURL) {
			videoURL = regYtURL.FindString(videoURL)
		}

		go func() {
			err = b.sendTextMessage(Message{
				chatID: update.Message.Chat.ID,
				text:   downloadStarted})
			if err != nil {
				logrus.Error(err)
			}
		}()

		workDir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("retrieve work dir error: %w", err)
		}

		chatIDDir := filepath.Join(workDir, b.conf.DownloadsDir, fmt.Sprint(update.Message.Chat.ID))

		filename, err := downloadAudioByLink(videoURL, chatIDDir)
		if err != nil {
			return fmt.Errorf("download audio by link error: %w", err)
		}

		fullFilePath := filepath.Join(chatIDDir, filename)
		defer func() {
			err = removeContents(chatIDDir)
			if err != nil {
				logrus.Errorf("remove file %s after upload error: %v", fullFilePath, err)
			}
		}()

		_, err = os.Stat(fullFilePath)
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found after download: %w", err)
		}

		_, err = b.api.Send(tgbotapi.NewAudioUpload(update.Message.Chat.ID, fullFilePath))
		if err != nil {
			return fmt.Errorf("send file to user error: %w", err)
		}
	case false:
		videos, err := getVideosList(update.Message.Text, b.conf.YoutubeToken)
		if err != nil {
			return fmt.Errorf("retrieve video list error: %w", err)
		}

		err = b.sendButtonMessage(update, videos)
		if err != nil {
			return fmt.Errorf("send button message to user error: %w", err)
		}
	}

	return err
}

// implementCommands implements slash-commands (e.g. /start)
func (b *Bot) implementCommands(update *tgbotapi.Update) error {
	switch update.Message.Command() {
	case "start":
		err := b.sendTextMessage(Message{
			chatID: update.Message.Chat.ID,
			text:   greetingMessage})
		if err != nil {
			return err
		}
	case "help":
		err := b.sendTextMessage(Message{
			chatID: update.Message.Chat.ID,
			text:   helpMessage})
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("got non-exist command from user: %d", update.Message.From.ID)
	}

	return nil
}

func removeContents(dir string) error {
	fi, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range fi {
		err = os.RemoveAll(filepath.Join(dir, f.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}
