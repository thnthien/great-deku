package telegram

import (
	"fmt"
	"unsafe"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

func getString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

type Config struct {
	Bot     *tgbotapi.BotAPI
	Token   string
	ChatID  int64
	Enabler zap.AtomicLevel
}

type Writer struct {
	*tgbotapi.BotAPI
	chatID int64
}

func New(token string, chatID int64) (*Writer, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &Writer{bot, chatID}, nil
}

func NewWithBot(bot *tgbotapi.BotAPI, chatID int64) *Writer {
	return &Writer{bot, chatID}
}

func (w *Writer) Write(b []byte) (int, error) {
	sentMsgs := make([]*tgbotapi.Message, 0)
	n := 0
	for i := 0; i < len(b); i += 4096 {
		length := i + 4096
		if length > len(b) {
			length = len(b)
		}
		message := tgbotapi.NewMessage(w.chatID, getString(b[i:length]))
		message.ParseMode = "Html"
		sentMsg, err := w.Send(message)
		if err != nil {
			return 0, fmt.Errorf("cannot sent message: %s", err)
		}
		sentMsgs = append(sentMsgs, &sentMsg)
		n = length
	}
	return n, nil
}
