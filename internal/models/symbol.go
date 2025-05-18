package models

import (
	"time"
)

type Symbol struct {
	ID                       uint      `gorm:"primaryKey" json:"id"`
	ExchangeID               string    `json:"exchangeId" gorm:"column:EXCH_ID;index"`
	Segment                  string    `json:"segment" gorm:"column:SEGMENT;index"`
	Symbol                   string    `json:"symbol" gorm:"column:SYMBOL_NAME;index"`
	Name                     string    `json:"name" gorm:"column:DISPLAY_NAME"`
	ISIN                     string    `json:"isin" gorm:"column:ISIN"`
	Instrument               string    `json:"instrument" gorm:"column:INSTRUMENT"`
	SecurityID               int64     `json:"securityId" gorm:"column:SECURITY_ID;index"`
	UnderlyingSecurityID     float64   `json:"underlyingSecurityId" gorm:"column:UNDERLYING_SECURITY_ID"`
	UnderlyingSymbol         string    `json:"underlyingSymbol" gorm:"column:UNDERLYING_SYMBOL"`
	DisplayName              string    `json:"displayName" gorm:"column:DISPLAY_NAME"`
	InstrumentType           string    `json:"instrumentType" gorm:"column:INSTRUMENT_TYPE"`
	Series                   string    `json:"series" gorm:"column:SERIES"`
	LotSize                  float64   `json:"lotSize" gorm:"column:LOT_SIZE"`
	TickSize                 float64   `json:"tickSize" gorm:"column:TICK_SIZE"`
	ExpiryDate               float64   `json:"expiryDate" gorm:"column:SM_EXPIRY_DATE"`
	StrikePrice              float64   `json:"strikePrice" gorm:"column:STRIKE_PRICE"`
	OptionType               float64   `json:"optionType" gorm:"column:OPTION_TYPE"`
	ExpiryFlag               float64   `json:"expiryFlag" gorm:"column:EXPIRY_FLAG"`
	BuySellIndicator         string    `json:"buySellIndicator" gorm:"column:BUY_SELL_INDICATOR"`
	BracketFlag              string    `json:"bracketFlag" gorm:"column:BRACKET_FLAG"`
	CoverFlag                string    `json:"coverFlag" gorm:"column:COVER_FLAG"`
	AsmGsmFlag               string    `json:"asmGsmFlag" gorm:"column:ASM_GSM_FLAG"`
	AsmGsmCategory           float64   `json:"asmGsmCategory" gorm:"column:ASM_GSM_CATEGORY"`
	MTFLeverage              float64   `json:"mtfLeverage" gorm:"column:MTF_LEVERAGE"`
	BuyCoMinMarginPer        float64   `json:"buyCoMinMarginPer" gorm:"column:BUY_CO_MIN_MARGIN_PER"`
	SellCoMinMarginPer       float64   `json:"sellCoMinMarginPer" gorm:"column:SELL_CO_MIN_MARGIN_PER"`
	BuyCoSlRangeMaxPerc      float64   `json:"buyCoSlRangeMaxPerc" gorm:"column:BUY_CO_SL_RANGE_MAX_PERC"`
	SellCoSlRangeMaxPerc     float64   `json:"sellCoSlRangeMaxPerc" gorm:"column:SELL_CO_SL_RANGE_MAX_PERC"`
	BuyCoSlRangeMinPerc      float64   `json:"buyCoSlRangeMinPerc" gorm:"column:BUY_CO_SL_RANGE_MIN_PERC"`
	SellCoSlRangeMinPerc     float64   `json:"sellCoSlRangeMinPerc" gorm:"column:SELL_CO_SL_RANGE_MIN_PERC"`
	BuyBoMinMarginPer        float64   `json:"buyBoMinMarginPer" gorm:"column:BUY_BO_MIN_MARGIN_PER"`
	SellBoMinMarginPer       float64   `json:"sellBoMinMarginPer" gorm:"column:SELL_BO_MIN_MARGIN_PER"`
	BuyBoSlRangeMaxPerc      float64   `json:"buyBoSlRangeMaxPerc" gorm:"column:BUY_BO_SL_RANGE_MAX_PERC"`
	SellBoSlRangeMaxPerc     float64   `json:"sellBoSlRangeMaxPerc" gorm:"column:SELL_BO_SL_RANGE_MAX_PERC"`
	BuyBoSlRangeMinPerc      float64   `json:"buyBoSlRangeMinPerc" gorm:"column:BUY_BO_SL_RANGE_MIN_PERC"`
	SellBoMinRange           float64   `json:"sellBoMinRange" gorm:"column:SELL_BO_SL_MIN_RANGE"`
	BuyBoProfitRangeMaxPerc  float64   `json:"buyBoProfitRangeMaxPerc" gorm:"column:BUY_BO_PROFIT_RANGE_MAX_PERC"`
	SellBoProfitRangeMaxPerc float64   `json:"sellBoProfitRangeMaxPerc" gorm:"column:SELL_BO_PROFIT_RANGE_MAX_PERC"`
	BuyBoProfitRangeMinPerc  float64   `json:"buyBoProfitRangeMinPerc" gorm:"column:BUY_BO_PROFIT_RANGE_MIN_PERC"`
	SellBoProfitRangeMinPerc float64   `json:"sellBoProfitRangeMinPerc" gorm:"column:SELL_BO_PROFIT_RANGE_MIN_PERC"`
	IsActive                 bool      `json:"isActive" gorm:"default:true"`
	LastUpdated              time.Time `json:"lastUpdated"`
	CreatedAt                time.Time `json:"createdAt"`
	UpdatedAt                time.Time `json:"updatedAt"`
}

func (s *Symbol) TableName() string {
	return "symboltbl"
}
