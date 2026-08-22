package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/models"
	"github.com/Alexxx-Hug/price-catcher-monorepo/tg-bot/internal/models/events"
	"github.com/google/uuid"
)

type BotUseCase struct {
	parser               ProductParser
	producer             UserActionProducer
	userStates           map[int64]models.UserState
	pending              map[int64]models.PendingSubscription
	subscriptionProvider SubscriptionProvider
}

type SizeChoiceResult struct {
	Text  string
	Sizes []models.ProductSize
}

type DeleteSubscriptionResult struct {
	Text          string
	Subscriptions []models.Subscription
}

func NewBotUseCase(parser ProductParser, producer UserActionProducer, subscriptionProvider SubscriptionProvider) *BotUseCase {
	return &BotUseCase{
		parser:               parser,
		producer:             producer,
		userStates:           make(map[int64]models.UserState),
		pending:              make(map[int64]models.PendingSubscription),
		subscriptionProvider: subscriptionProvider,
	}
}

// return message to user: "Пришли ссылку на товар"
func (u *BotUseCase) StartAddSubscription(ctx context.Context, telegramUserID int64) string {
	u.userStates[telegramUserID] = models.StateWaitingProductURL
	return "Жду ссылку на товар!"
}

// send request to monitor service and return text with all size of product
func (u *BotUseCase) HandleProductURL(ctx context.Context, telegramUserID int64, url string) (*SizeChoiceResult, error) {
	state := u.userStates[telegramUserID]
	if state != models.StateWaitingProductURL {
		return nil, fmt.Errorf("unexpected user state: %s", state)
	}

	product, err := u.parser.ParseProduct(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse product: %w", err)
	}

	if product == nil {
		return nil, fmt.Errorf("product is nil")
	}

	if len(product.Sizes) == 0 {
		return nil, fmt.Errorf("product has no sizes")
	}

	u.pending[telegramUserID] = models.PendingSubscription{
		Product: *product,
	}

	sizes := make([]models.ProductSize, 0, len(product.Sizes))
	for _, size := range product.Sizes {
		sizes = append(sizes, size)
	}

	sizeChoiceResult := SizeChoiceResult{
		Text:  "Выбери нужный размер",
		Sizes: sizes,
	}

	u.userStates[telegramUserID] = models.StateWaitingSizeChoiceForAddSubscription

	return &sizeChoiceResult, nil
}

// when user choose the size service will return him a message: "принял в обработку"
func (u *BotUseCase) SelectSize(ctx context.Context, telegramUserID int64, optionID int64) (string, error) {
	state := u.userStates[telegramUserID]
	if state != models.StateWaitingSizeChoiceForAddSubscription {
		return "", fmt.Errorf("unexpected user state")
	}

	pending, exist := u.pending[telegramUserID]
	if !exist {
		return "", fmt.Errorf("pending subscription not found")
	}

	var selectedSize *models.ProductSize
	for _, size := range pending.Product.Sizes {
		if size.OptionID == optionID {
			selectedSize = &size
		}
	}

	if selectedSize == nil {
		return "", fmt.Errorf("product size with option_id=%d not found", optionID)
	}

	payload := events.AddSubscriptionPayload{
		Product:     pending.Product,
		ProductSize: *selectedSize,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal add subscription payload: %w", err)
	}

	event := events.UserActionEvent{
		ActionID:       uuid.NewString(),
		TelegramUserID: telegramUserID,
		Type:           events.UserActionAddSubscription,
		Payload:        payloadBytes,
		CreatedAt:      time.Now(),
	}

	if err := u.producer.SendUserAction(ctx, event); err != nil {
		return "", fmt.Errorf("failed to send user action event: %w", err)
	}

	delete(u.pending, telegramUserID)
	u.userStates[telegramUserID] = models.StateIdle

	return fmt.Sprintf("Принял, добавляю подписку на \"%s\", размер: %s", pending.Product.Name, selectedSize.SizeName), nil
}

func (u *BotUseCase) ListUserSubscription(ctx context.Context, telegramUserID int64) (string, error) {
	subscriptions, err := u.subscriptionProvider.ListUserSubscriptions(ctx, telegramUserID)
	if err != nil {
		return "", fmt.Errorf("failed to get user subscriptions: %w", err)
	}

	if len(subscriptions) == 0 {
		return "У тебя пока нет подписок", nil
	}

	var builder strings.Builder

	builder.WriteString("Ваши подписки:\n")

	for i, sub := range subscriptions {
		builder.WriteString(fmt.Sprintf(
			"\n%d. %s\nБренд: %s\nРазмер: %s\nСтоимость: %d\nURL: %s\n",
			i+1, sub.ProductName, sub.Brand, sub.SizeName, sub.PriceMinor/100, sub.URL,
		))
	}

	return builder.String(), nil
}

func (u *BotUseCase) DeleteUserSubscription(ctx context.Context, telegramUserID int64) (*DeleteSubscriptionResult, error) {
	subscriptions, err := u.subscriptionProvider.ListUserSubscriptions(ctx, telegramUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user subscriptions: %w", err)
	}

	if len(subscriptions) == 0 {
		return &DeleteSubscriptionResult{
			Text: "У тебя пока нет подписок",
		}, nil
	}

	u.userStates[telegramUserID] = models.StateWaitingSubscriptionChoiceForDelete

	return &DeleteSubscriptionResult{
		Text:          "Выбери подписку для удаления:",
		Subscriptions: subscriptions,
	}, nil
}

func (u *BotUseCase) SelectSubscriptionForDelete(ctx context.Context, telegramUserID, subscriptionID int64) (string, error) {
	state := u.userStates[telegramUserID]
	if state != models.StateWaitingSubscriptionChoiceForDelete {
		return "", fmt.Errorf("unexpected user state")
	}

	payload := events.DeleteSubscriptionPayload{
		SubscriptionID: subscriptionID,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal delete subscription payload: %w", err)
	}

	event := events.UserActionEvent{
		ActionID:       uuid.NewString(),
		TelegramUserID: telegramUserID,
		Type:           events.UserActionDeleteSubscription,
		Payload:        payloadBytes,
		CreatedAt:      time.Now(),
	}

	if err := u.producer.SendUserAction(ctx, event); err != nil {
		return "", fmt.Errorf("failed to send user action event: %w", err)
	}

	u.userStates[telegramUserID] = models.StateIdle
	return "Принял, удаляю", nil
}
