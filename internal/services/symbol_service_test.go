package services

import (
	"testing"
	"time"

	"github.com/vikasavnish/trademicro/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSymbolService(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Migrate schema
	err = db.AutoMigrate(&models.Symbol{})
	if err != nil {
		t.Fatalf("Failed to migrate schema: %v", err)
	}

	// Create service
	service := NewSymbolService(db)

	// Create test symbol
	symbol := models.Symbol{
		ExchangeID:              "NSE",
		Segment:                 "E",
		Symbol:                  "RELIANCE",
		Name:                    "Reliance Industries Ltd.",
		ISIN:                    "INE002A01018",
		Instrument:              "EQ",
		SecurityID:              500325,
		DisplayName:             "RELIANCE",
		InstrumentType:          "EQ",
		Series:                  "EQ",
		LotSize:                 1.0,
		TickSize:                0.05,
		IsActive:                true,
		LastUpdated:             time.Now(),
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
		UnderlyingSecurityID:    0,
		ExpiryDate:              0,
		StrikePrice:             0,
		OptionType:              0,
		ExpiryFlag:              0,
		BuySellIndicator:        "",
		BracketFlag:             "",
		CoverFlag:               "",
		AsmGsmFlag:              "",
		AsmGsmCategory:          0,
		MTFLeverage:             0,
		BuyCoMinMarginPer:       0,
		SellCoMinMarginPer:      0,
		BuyCoSlRangeMaxPerc:     0,
		SellCoSlRangeMaxPerc:    0,
		BuyCoSlRangeMinPerc:     0,
		SellCoSlRangeMinPerc:    0,
		BuyBoMinMarginPer:       0,
		SellBoMinMarginPer:      0,
		BuyBoSlRangeMaxPerc:     0,
		SellBoSlRangeMaxPerc:    0,
		BuyBoSlRangeMinPerc:     0,
		SellBoMinRange:          0,
		BuyBoProfitRangeMaxPerc: 0,
		SellBoProfitRangeMaxPerc:0,
		BuyBoProfitRangeMinPerc: 0,
		SellBoProfitRangeMinPerc:0,
	}

	// Test CreateSymbol
	err = service.CreateSymbol(&symbol)
	if err != nil {
		t.Fatalf("Failed to create symbol: %v", err)
	}

	// Test GetSymbolByID
	retrieved, err := service.GetSymbolByID(symbol.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve symbol: %v", err)
	}
	if retrieved.Symbol != "RELIANCE" {
		t.Errorf("Expected symbol name to be RELIANCE, got %s", retrieved.Symbol)
	}

	// Test GetSymbolByCode
	byCode, err := service.GetSymbolByCode("RELIANCE")
	if err != nil {
		t.Fatalf("Failed to retrieve symbol by code: %v", err)
	}
	if byCode.SecurityID != 500325 {
		t.Errorf("Expected security ID to be 500325, got %d", byCode.SecurityID)
	}

	// Test GetAllSymbols
	symbols, count, err := service.GetAllSymbols(nil, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get all symbols: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count to be 1, got %d", count)
	}
	if len(symbols) != 1 {
		t.Errorf("Expected to get 1 symbol, got %d", len(symbols))
	}

	// Test filtering
	filter := map[string]string{
		"exchange": "NSE",
		"segment":  "E",
	}
	filtered, filteredCount, err := service.GetAllSymbols(filter, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get filtered symbols: %v", err)
	}
	if filteredCount != 1 {
		t.Errorf("Expected filtered count to be 1, got %d", filteredCount)
	}
	if len(filtered) != 1 {
		t.Errorf("Expected to get 1 filtered symbol, got %d", len(filtered))
	}

	// Test UpdateSymbol
	symbol.Name = "Reliance Industries Limited"
	err = service.UpdateSymbol(&symbol)
	if err != nil {
		t.Fatalf("Failed to update symbol: %v", err)
	}
	updated, _ := service.GetSymbolByID(symbol.ID)
	if updated.Name != "Reliance Industries Limited" {
		t.Errorf("Expected updated name to be 'Reliance Industries Limited', got '%s'", updated.Name)
	}

	// Test DeleteSymbol
	err = service.DeleteSymbol(symbol.ID)
	if err != nil {
		t.Fatalf("Failed to delete symbol: %v", err)
	}
	_, err = service.GetSymbolByID(symbol.ID)
	if err == nil {
		t.Errorf("Expected error after deleting symbol, but got nil")
	}

	t.Log("All tests passed!")
}
