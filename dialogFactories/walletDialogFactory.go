package dialogFactories

import (
	"github.com/gameraccoon/telegram-bot-skeleton/dialog"
	"github.com/gameraccoon/telegram-bot-skeleton/dialogFactory"
	"github.com/gameraccoon/telegram-bot-skeleton/processing"
	"github.com/nicksnyder/go-i18n/i18n"
	"github.com/gameraccoon/telegram-accountant-bot/cryptoFunctions"
	"github.com/gameraccoon/telegram-accountant-bot/currencies"
	"github.com/gameraccoon/telegram-accountant-bot/serverData"
	"github.com/gameraccoon/telegram-accountant-bot/staticFunctions"
	"fmt"
	"math/big"
	"strconv"
)

type walletVariantPrototype struct {
	id string
	textId string
	process func(int64, *processing.ProcessData) bool
	// nil if the variant is always active
	isActiveFn func(int64, *processing.StaticProccessStructs) bool
	rowId int
}

type walletDialogFactory struct {
	variants []walletVariantPrototype
}

func MakeWalletDialogFactory() dialogFactory.DialogFactory {
	return &(walletDialogFactory{
		variants: []walletVariantPrototype{
			// walletVariantPrototype{
			// 	id: "send",
			// 	textId: "send",
			// 	process: sendFromWallet,
			// 	rowId:1,
			// },
			walletVariantPrototype{
				id: "get",
				textId: "receive",
				process: receiveToWallet,
				rowId:1,
			},
			walletVariantPrototype{
				id: "hist",
				textId: "history",
				process: showHistory,
				isActiveFn: isHistoryEnabled,
				rowId:1,
			},
			walletVariantPrototype{
				id: "set",
				textId: "settings",
				process: walletSettings,
				rowId:2,
			},
			walletVariantPrototype{
				id: "back",
				textId: "back_to_list",
				process: backToList,
				rowId:3,
			},
		},
	})
}

func isHistoryEnabled(walletId int64, staticData *processing.StaticProccessStructs) bool {
	walletAddress := staticFunctions.GetDb(staticData).GetWalletAddress(walletId)
	return currencies.IsHistoryEnabled(walletAddress.Currency)
}

func sendFromWallet(walletId int64, data *processing.ProcessData) bool {
	return false
}

func receiveToWallet(walletId int64, data *processing.ProcessData) bool {
	data.SubstitudeDialog(data.Static.MakeDialogFn("rc", walletId, data.Trans, data.Static))
	return true
}

func showHistory(walletId int64, data *processing.ProcessData) bool {
	data.SubstitudeDialog(data.Static.MakeDialogFn("hi", walletId, data.Trans, data.Static))
	return true
}

func walletSettings(walletId int64, data *processing.ProcessData) bool {
	data.SubstitudeDialog(data.Static.MakeDialogFn("ws", walletId, data.Trans, data.Static))
	return true
}

func backToList(walletId int64, data *processing.ProcessData) bool {
	data.SubstitudeDialog(data.Static.MakeDialogFn("wl", data.UserId, data.Trans, data.Static))
	return true
}

func (factory *walletDialogFactory) getDialogText(walletId int64, trans i18n.TranslateFunc, staticData *processing.StaticProccessStructs) (result string) {
	walletAddress := staticFunctions.GetDb(staticData).GetWalletAddress(walletId)

	serverData := serverData.GetServerData(staticData)

	if serverData == nil {
		return "Error"
	}

	balance := serverData.GetBalance(walletAddress)

	if balance == nil {
		return trans("no_data")
	}

	currencySymbol, currencyDecimals := staticFunctions.GetCurrencySymbolAndDecimals(serverData, walletAddress.Currency, walletAddress.ContractAddress)

	floatBalance := cryptoFunctions.GetFloatBalance(balance, currencyDecimals)

	if floatBalance == nil {
		return trans("no_data")
	}

	balanceText := cryptoFunctions.FormatFloatCurrencyAmount(floatBalance, currencyDecimals)

	result = fmt.Sprintf("<b>%s</b>\n%s %s",
		staticFunctions.GetDb(staticData).GetWalletName(walletId),
		balanceText,
		currencySymbol,
	)

	toUsdRate := serverData.GetRateToUsd(walletAddress.PriceId)

	if toUsdRate == nil {
		return
	}

	usdCost := new(big.Float).Mul(floatBalance, toUsdRate)

	if usdCost == nil {
		return
	}

	result = result + fmt.Sprintf("\n%s %s",
		usdCost.Text('f', 2),
		trans("usd"),
	)

	return
}

func (factory *walletDialogFactory) createVariants(walletId int64, trans i18n.TranslateFunc, staticData *processing.StaticProccessStructs) (variants []dialog.Variant) {
	variants = make([]dialog.Variant, 0)

	for _, variant := range factory.variants {
		if variant.isActiveFn == nil || variant.isActiveFn(walletId, staticData) {
			variants = append(variants, dialog.Variant{
				Id:   variant.id,
				Text: trans(variant.textId),
				AdditionalId: strconv.FormatInt(walletId, 10),
				RowId: variant.rowId,
			})
		}
	}
	return
}

func (factory *walletDialogFactory) MakeDialog(walletId int64, trans i18n.TranslateFunc, staticData *processing.StaticProccessStructs) *dialog.Dialog {
	return &dialog.Dialog{
		Text:     factory.getDialogText(walletId, trans, staticData),
		Variants: factory.createVariants(walletId, trans, staticData),
	}
}

func (factory *walletDialogFactory) ProcessVariant(variantId string, additionalId string, data *processing.ProcessData) bool {
	walletId, err := strconv.ParseInt(additionalId, 10, 64)

	if err != nil {
		return false
	}

	if !staticFunctions.GetDb(data.Static).IsWalletBelongsToUser(data.UserId, walletId) {
		return false
	}

	for _, variant := range factory.variants {
		if variant.id == variantId {
			return variant.process(walletId, data)
		}
	}
	return false
}
