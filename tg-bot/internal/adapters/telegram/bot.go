package telegram

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.uber.org/zap"
)

type Bot struct {
	api     *tgbotapi.BotAPI
	usecase *service.BotUseCase
	logger  *zap.Logger
}

func NewBot(token string, usecase *service.BotUseCase, logger *zap.Logger) (*Bot, error) {
	if logger == nil {
		logger = zap.NewNop()
	}

	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		api:     api,
		usecase: usecase,
		logger:  logger,
	}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := b.api.GetUpdatesChan(updateConfig)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-updates:
			if update.Message != nil {
				if err := b.handleMessage(ctx, update.Message); err != nil {
					b.logger.Error("failed to handle telegram message", zap.Error(err))
				}
				continue
			}

			if update.CallbackQuery != nil {
				if err := b.handleCallback(ctx, update.CallbackQuery); err != nil {
					b.logger.Error("failed to handle telegram callback", zap.Error(err))
				}
				continue
			}
		}
	}
}

func (b *Bot) handleMessage(ctx context.Context, message *tgbotapi.Message) error {
	if message.From == nil {
		return nil
	}

	telegramUserID := message.From.ID
	chatID := message.Chat.ID
	text := message.Text

	switch text {
	case "/start":
		return b.sendMainMenu(chatID)
	case "Добавить подписку":
		response := b.usecase.StartAddSubscription(ctx, telegramUserID)
		return b.sendText(chatID, response)
	default:
		result, err := b.usecase.HandleProductURL(ctx, telegramUserID, text)
		if err != nil {
			return b.sendText(chatID, "Не понял сообщение. Нажми /start или выбери действие.")
		}

		return b.sendSizeChoice(chatID, result)
	}
}

func (b *Bot) sendMainMenu(chatID int64) error {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Добавить подписку"),
		),
	)

	keyboard.ResizeKeyboard = true

	msg := tgbotapi.NewMessage(chatID, "Выбери действие")
	msg.ReplyMarkup = keyboard

	if _, err := b.api.Send(msg); err != nil {
		return fmt.Errorf("send main menu: %w", err)
	}

	return nil
}

func (b *Bot) sendText(chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)

	if _, err := b.api.Send(msg); err != nil {
		return fmt.Errorf("send telegram message: %w", err)
	}

	return nil
}

func (b *Bot) sendSizeChoice(chatID int64, result *service.SizeChoiceResult) error {
	rows := make([][]tgbotapi.InlineKeyboardButton, 0, len(result.Sizes))

	for _, size := range result.Sizes {
		buttonText := fmt.Sprintf("%s — %d руб.", size.SizeName, size.PriceMinor/100)
		callbackData := fmt.Sprintf("size:%d", size.OptionID)

		row := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData),
		)

		rows = append(rows, row)
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	msg := tgbotapi.NewMessage(chatID, result.Text)
	msg.ReplyMarkup = keyboard

	if _, err := b.api.Send(msg); err != nil {
		return fmt.Errorf("send telegram size choice: %w", err)
	}

	return nil
}

func (b *Bot) handleCallback(ctx context.Context, callback *tgbotapi.CallbackQuery) error {
	if callback.Message == nil {
		return nil
	}

	telegramUserID := callback.From.ID
	chatID := callback.Message.Chat.ID

	const prefix = "size:"

	if !strings.HasPrefix(callback.Data, prefix) {
		return b.answerCallback(callback.ID, "Неизвестное действие")
	}

	optionID, err := strconv.ParseInt(strings.TrimPrefix(callback.Data, prefix), 10, 64)
	if err != nil {
		return b.answerCallback(callback.ID, "Не смог прочитать размер")
	}

	response, err := b.usecase.SelectSize(ctx, telegramUserID, optionID)
	if err != nil {
		b.logger.Error(
			"failed to select size",
			zap.Int64("telegram_user_id", telegramUserID),
			zap.Int64("option_id", optionID),
			zap.Error(err),
		)

		return b.answerCallback(callback.ID, "Не смог выбрать размер")
	}

	if err := b.answerCallback(callback.ID, "Размер выбран"); err != nil {
		return err
	}

	return b.sendText(chatID, response)
}

func (b *Bot) answerCallback(callbackID string, text string) error {
	callback := tgbotapi.NewCallback(callbackID, text)

	if _, err := b.api.Request(callback); err != nil {
		return fmt.Errorf("answer telegram callback: %w", err)
	}

	return nil
}
