package models

import (
	"time"

	"gorm.io/gorm"
)

// Instrument represents a financial instrument from DhanHQ with all attributes
type Instrument struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Basic symbol identification
	SecurityID  string `gorm:"uniqueIndex" json:"securityId"` // Primary identifier
	Symbol      string `gorm:"index" json:"symbol"`           // Trading symbol
	SymbolName  string `json:"symbolName"`                    // Display name (SYMBOL_NAME/SM_SYMBOL_NAME)
	DisplayName string `json:"displayName"`                   // Dhan display name (DISPLAY_NAME/SEM_CUSTOM_SYMBOL)

	// Exchange and market information
	ExchangeID     string `gorm:"index" json:"exchangeId"` // Exchange identifier (EXCH_ID/SEM_EXM_EXCH_ID)
	Segment        string `gorm:"index" json:"segment"`    // Market segment (SEGMENT/SEM_SEGMENT)
	SegmentType    string `json:"segmentType"`             // Parsed segment type (EQ, IDX, CUR, COMM)
	ISIN           string `gorm:"index" json:"isin"`       // ISIN code
	Instrument     string `json:"instrument"`              // Instrument type code (INSTRUMENT/SEM_INSTRUMENT_NAME)
	InstrumentType string `json:"instrumentType"`          // Additional instrument details (INSTRUMENT_TYPE/SEM_EXCH_INSTRUMENT_TYPE)
	Series         string `json:"series"`                  // Exchange series (SERIES/SEM_SERIES)

	// Contract specifications
	LotSize  int64   `json:"lotSize"`  // Trading lot size (LOT_SIZE/SEM_LOT_UNITS)
	TickSize float64 `json:"tickSize"` // Minimum price movement (TICK_SIZE/SEM_TICK_SIZE)

	// Derivative specific information
	ExpiryDate           *time.Time `json:"expiryDate"`           // For derivatives (SM_EXPIRY_DATE/SEM_EXPIRY_DATE)
	ExpiryCode           string     `json:"expiryCode"`           // Futures expiry code (SEM_EXPIRY_CODE)
	ExpiryFlag           string     `json:"expiryFlag"`           // Monthly or Weekly (EXPIRY_FLAG/SEM_EXPIRY_FLAG)
	StrikePrice          float64    `json:"strikePrice"`          // For options (STRIKE_PRICE/SEM_STRIKE_PRICE)
	OptionType           string     `json:"optionType"`           // CE or PE (OPTION_TYPE/SEM_OPTION_TYPE)
	UnderlyingSecurityID string     `json:"underlyingSecurityId"` // For derivatives (UNDERLYING_SECURITY_ID)
	UnderlyingSymbol     string     `json:"underlyingSymbol"`     // For derivatives (UNDERLYING_SYMBOL)

	// Trading restrictions and parameters
	BracketFlag      string `json:"bracketFlag"`      // Y or N (BRACKET_FLAG)
	CoverFlag        string `json:"coverFlag"`        // Y or N (COVER_FLAG)
	ASMGSMFlag       string `json:"asmGsmFlag"`       // Market surveillance flags (ASM_GSM_FLAG)
	ASMGSMCategory   string `json:"asmGsmCategory"`   // Category of surveillance (ASM_GSM_CATEGORY)
	BuySellIndicator string `json:"buySellIndicator"` // Trading permissions (BUY_SELL_INDICATOR)

	// Order margins and limits
	BuyCoMinMarginPer        float64 `json:"buyCoMinMarginPer"`        // (BUY_CO_MIN_MARGIN_PER)
	SellCoMinMarginPer       float64 `json:"sellCoMinMarginPer"`       // (SELL_CO_MIN_MARGIN_PER)
	BuyCoSlRangeMaxPerc      float64 `json:"buyCoSlRangeMaxPerc"`      // (BUY_CO_SL_RANGE_MAX_PERC)
	SellCoSlRangeMaxPerc     float64 `json:"sellCoSlRangeMaxPerc"`     // (SELL_CO_SL_RANGE_MAX_PERC)
	BuyCoSlRangeMinPerc      float64 `json:"buyCoSlRangeMinPerc"`      // (BUY_CO_SL_RANGE_MIN_PERC)
	SellCoSlRangeMinPerc     float64 `json:"sellCoSlRangeMinPerc"`     // (SELL_CO_SL_RANGE_MIN_PERC)
	BuyBoMinMarginPer        float64 `json:"buyBoMinMarginPer"`        // (BUY_BO_MIN_MARGIN_PER)
	SellBoMinMarginPer       float64 `json:"sellBoMinMarginPer"`       // (SELL_BO_MIN_MARGIN_PER)
	BuyBoSlRangeMaxPerc      float64 `json:"buyBoSlRangeMaxPerc"`      // (BUY_BO_SL_RANGE_MAX_PERC)
	SellBoSlRangeMaxPerc     float64 `json:"sellBoSlRangeMaxPerc"`     // (SELL_BO_SL_RANGE_MAX_PERC)
	BuyBoSlRangeMinPerc      float64 `json:"buyBoSlRangeMinPerc"`      // (BUY_BO_SL_RANGE_MIN_PERC)
	SellBoSlMinRange         float64 `json:"sellBoSlMinRange"`         // (SELL_BO_SL_MIN_RANGE)
	BuyBoProfitRangeMaxPerc  float64 `json:"buyBoProfitRangeMaxPerc"`  // (BUY_BO_PROFIT_RANGE_MAX_PERC)
	SellBoProfitRangeMaxPerc float64 `json:"sellBoProfitRangeMaxPerc"` // (SELL_BO_PROFIT_RANGE_MAX_PERC)
	BuyBoProfitRangeMinPerc  float64 `json:"buyBoProfitRangeMinPerc"`  // (BUY_BO_PROFIT_RANGE_MIN_PERC)
	SellBoProfitRangeMinPerc float64 `json:"sellBoProfitRangeMinPerc"` // (SELL_BO_PROFIT_RANGE_MIN_PERC)
	MTFLeverage              float64 `json:"mtfLeverage"`              // (MTF_LEVERAGE)

	// Additional metadata
	LastUpdated time.Time `json:"lastUpdated"` // Last time instrument was updated
	DataSource  string    `json:"dataSource"`  // Source of data (compact/detailed)
}

// BeforeCreate will set timestamps
func (i *Instrument) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	i.CreatedAt = now
	i.UpdatedAt = now
	i.LastUpdated = now
	return nil
}

// BeforeUpdate will update timestamps
func (i *Instrument) BeforeUpdate(tx *gorm.DB) error {
	i.UpdatedAt = time.Now()
	return nil
}
