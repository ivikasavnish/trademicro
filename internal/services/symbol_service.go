package services

import (
	"encoding/csv"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/vikasavnish/trademicro/internal/models"
	"gorm.io/gorm"
)

type SymbolService struct {
	DB *gorm.DB
}

func NewSymbolService(db *gorm.DB) *SymbolService {
	return &SymbolService{DB: db}
}

func (s *SymbolService) GetAllSymbols(filter map[string]string, limit, offset int) ([]models.Symbol, int64, error) {
	var symbols []models.Symbol
	var count int64
	query := s.DB.Model(&models.Symbol{})

	// Apply filters
	if filter != nil {
		if exchange, ok := filter["exchange"]; ok && exchange != "" {
			query = query.Where("exchange_id = ?", exchange)
		}
		if segment, ok := filter["segment"]; ok && segment != "" {
			query = query.Where("segment = ?", segment)
		}
		if search, ok := filter["search"]; ok && search != "" {
			query = query.Where("symbol LIKE ? OR name LIKE ? OR trading_symbol LIKE ?",
				"%"+search+"%", "%"+search+"%", "%"+search+"%")
		}
		if active, ok := filter["active"]; ok {
			isActive := active == "true"
			query = query.Where("is_active = ?", isActive)
		}
	}

	// Count total before pagination
	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	// Get results
	err = query.Find(&symbols).Error
	return symbols, count, err
}

func (s *SymbolService) GetSymbolByID(id uint) (*models.Symbol, error) {
	var symbol models.Symbol
	result := s.DB.First(&symbol, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &symbol, nil
}

func (s *SymbolService) GetSymbolByCode(code string) (*models.Symbol, error) {
	var symbol models.Symbol
	result := s.DB.Where("symbol = ? OR trading_symbol = ?", code, code).First(&symbol)
	if result.Error != nil {
		return nil, result.Error
	}
	return &symbol, nil
}

func (s *SymbolService) CreateSymbol(symbol *models.Symbol) error {
	symbol.CreatedAt = time.Now()
	symbol.UpdatedAt = time.Now()
	symbol.LastUpdated = time.Now()
	result := s.DB.Create(symbol)
	return result.Error
}

func (s *SymbolService) UpdateSymbol(symbol *models.Symbol) error {
	symbol.UpdatedAt = time.Now()
	symbol.LastUpdated = time.Now()
	result := s.DB.Save(symbol)
	return result.Error
}

func (s *SymbolService) DeleteSymbol(id uint) error {
	result := s.DB.Delete(&models.Symbol{}, id)
	return result.Error
}

func (s *SymbolService) ImportSymbolsFromDhanAPI(compact bool) (int, error) {
	var url string
	if compact {
		url = "https://images.dhan.co/api-data/api-scrip-master.csv"
	} else {
		url = "https://images.dhan.co/api-data/api-scrip-master-detailed.csv"
	}

	// Download the CSV file
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("failed to download symbol list: " + resp.Status)
	}

	// Parse the CSV
	reader := csv.NewReader(resp.Body)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return 0, err
	}

	// Map column indices
	columnMap := make(map[string]int)
	for i, col := range header {
		columnMap[col] = i
	}

	// Begin transaction
	tx := s.DB.Begin()

	// Process rows
	count := 0
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			tx.Rollback()
			return count, err
		}

		// Create symbol from CSV row
		symbol := models.Symbol{
			IsActive:    true,
			LastUpdated: time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Map fields based on which CSV format we're using
		if compact {
			// Compact CSV mapping
			if idx, ok := columnMap["SEM_EXM_EXCH_ID"]; ok && idx < len(row) {
				symbol.ExchangeID = row[idx]
			}
			if idx, ok := columnMap["SEM_SEGMENT"]; ok && idx < len(row) {
				symbol.Segment = row[idx]
			}
			if idx, ok := columnMap["SEM_TRADING_SYMBOL"]; ok && idx < len(row) {
				symbol.TradingSymbol = row[idx]
				symbol.Symbol = row[idx] // Use trading symbol as symbol too
			}
			if idx, ok := columnMap["SEM_CUSTOM_SYMBOL"]; ok && idx < len(row) {
				symbol.DisplayName = row[idx]
			}
			if idx, ok := columnMap["SM_SYMBOL_NAME"]; ok && idx < len(row) {
				symbol.Name = row[idx]
			}
			if idx, ok := columnMap["SEM_INSTRUMENT_NAME"]; ok && idx < len(row) {
				symbol.Instrument = row[idx]
			}
			if idx, ok := columnMap["SEM_SMST_SECURITY_ID"]; ok && idx < len(row) {
				symbol.SecurityID = row[idx]
			}
			if idx, ok := columnMap["SEM_EXCH_INSTRUMENT_TYPE"]; ok && idx < len(row) {
				symbol.InstrumentType = row[idx]
			}
			if idx, ok := columnMap["SEM_SERIES"]; ok && idx < len(row) {
				symbol.Series = row[idx]
			}
			if idx, ok := columnMap["SEM_LOT_UNITS"]; ok && idx < len(row) {
				lotSize, _ := strconv.Atoi(row[idx])
				symbol.LotSize = lotSize
			}
			if idx, ok := columnMap["SEM_TICK_SIZE"]; ok && idx < len(row) {
				tickSize, _ := strconv.ParseFloat(row[idx], 64)
				symbol.TickSize = tickSize
			}
		} else {
			// Detailed CSV mapping
			if idx, ok := columnMap["EXCH_ID"]; ok && idx < len(row) {
				symbol.ExchangeID = row[idx]
			}
			if idx, ok := columnMap["SEGMENT"]; ok && idx < len(row) {
				symbol.Segment = row[idx]
			}
			if idx, ok := columnMap["SYMBOL_NAME"]; ok && idx < len(row) {
				symbol.Symbol = row[idx]
				symbol.Name = row[idx]
			}
			if idx, ok := columnMap["DISPLAY_NAME"]; ok && idx < len(row) {
				symbol.DisplayName = row[idx]
			}
			if idx, ok := columnMap["ISIN"]; ok && idx < len(row) {
				symbol.ISIN = row[idx]
			}
			if idx, ok := columnMap["INSTRUMENT"]; ok && idx < len(row) {
				symbol.Instrument = row[idx]
			}
			if idx, ok := columnMap["INSTRUMENT_TYPE"]; ok && idx < len(row) {
				symbol.InstrumentType = row[idx]
			}
			if idx, ok := columnMap["SERIES"]; ok && idx < len(row) {
				symbol.Series = row[idx]
			}
			if idx, ok := columnMap["LOT_SIZE"]; ok && idx < len(row) {
				lotSize, _ := strconv.Atoi(row[idx])
				symbol.LotSize = lotSize
			}
			if idx, ok := columnMap["TICK_SIZE"]; ok && idx < len(row) {
				tickSize, _ := strconv.ParseFloat(row[idx], 64)
				symbol.TickSize = tickSize
			}
			if idx, ok := columnMap["STRIKE_PRICE"]; ok && idx < len(row) {
				strikePrice, _ := strconv.ParseFloat(row[idx], 64)
				symbol.StrikePrice = strikePrice
			}
			if idx, ok := columnMap["OPTION_TYPE"]; ok && idx < len(row) {
				symbol.OptionType = row[idx]
			}
			if idx, ok := columnMap["EXPIRY_FLAG"]; ok && idx < len(row) {
				symbol.ExpiryFlag = row[idx]
			}
			if idx, ok := columnMap["SM_EXPIRY_DATE"]; ok && idx < len(row) {
				symbol.ExpiryDate = row[idx]
			}
		}

		// Skip if required fields are missing
		if symbol.Symbol == "" || symbol.SecurityID == "" {
			continue
		}

		// Check for existing symbol and update or create
		var existingSymbol models.Symbol
		result := tx.Where("security_id = ?", symbol.SecurityID).First(&existingSymbol)
		if result.Error == nil {
			// Update existing symbol
			existingSymbol.ExchangeID = symbol.ExchangeID
			existingSymbol.Segment = symbol.Segment
			existingSymbol.Symbol = symbol.Symbol
			existingSymbol.Name = symbol.Name
			existingSymbol.ISIN = symbol.ISIN
			existingSymbol.Instrument = symbol.Instrument
			existingSymbol.TradingSymbol = symbol.TradingSymbol
			existingSymbol.DisplayName = symbol.DisplayName
			existingSymbol.InstrumentType = symbol.InstrumentType
			existingSymbol.Series = symbol.Series
			existingSymbol.LotSize = symbol.LotSize
			existingSymbol.TickSize = symbol.TickSize
			existingSymbol.ExpiryDate = symbol.ExpiryDate
			existingSymbol.StrikePrice = symbol.StrikePrice
			existingSymbol.OptionType = symbol.OptionType
			existingSymbol.ExpiryFlag = symbol.ExpiryFlag
			existingSymbol.UpdatedAt = time.Now()
			existingSymbol.LastUpdated = time.Now()
			tx.Save(&existingSymbol)
		} else {
			// Create new symbol
			tx.Create(&symbol)
		}
		count++
	}

	// Commit transaction
	err = tx.Commit().Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *SymbolService) ImportSymbolsFromCSV(reader io.Reader) (int, error) {
	csvReader := csv.NewReader(reader)

	// Read header
	header, err := csvReader.Read()
	if err != nil {
		return 0, err
	}

	// Map column indices
	columnMap := make(map[string]int)
	for i, col := range header {
		columnMap[strings.TrimSpace(col)] = i
	}

	// Begin transaction
	tx := s.DB.Begin()

	// Process rows
	count := 0
	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			tx.Rollback()
			return count, err
		}

		symbol := models.Symbol{
			IsActive:    true,
			LastUpdated: time.Now(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Map CSV fields to symbol struct
		mapCSVRowToSymbol(&symbol, row, columnMap)

		// Skip if symbol is empty
		if symbol.Symbol == "" {
			continue
		}

		// Check for existing symbol and update or create
		var existingSymbol models.Symbol
		result := tx.Where("symbol = ? OR trading_symbol = ?", symbol.Symbol, symbol.Symbol).First(&existingSymbol)
		if result.Error == nil {
			// Update existing
			symbol.ID = existingSymbol.ID
			symbol.CreatedAt = existingSymbol.CreatedAt
			tx.Save(&symbol)
		} else {
			// Create new
			tx.Create(&symbol)
		}
		count++
	}

	// Commit transaction
	err = tx.Commit().Error
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Helper function to map CSV row to Symbol struct
func mapCSVRowToSymbol(symbol *models.Symbol, row []string, columnMap map[string]int) {
	// Try to map common field names
	if idx, ok := columnMap["Symbol"]; ok && idx < len(row) {
		symbol.Symbol = row[idx]
	}
	if idx, ok := columnMap["Name"]; ok && idx < len(row) {
		symbol.Name = row[idx]
	}
	if idx, ok := columnMap["TradingSymbol"]; ok && idx < len(row) {
		symbol.TradingSymbol = row[idx]
	}
	if idx, ok := columnMap["SecurityID"]; ok && idx < len(row) {
		symbol.SecurityID = row[idx]
	}
	if idx, ok := columnMap["ExchangeID"]; ok && idx < len(row) {
		symbol.ExchangeID = row[idx]
	}
	if idx, ok := columnMap["Segment"]; ok && idx < len(row) {
		symbol.Segment = row[idx]
	}

	// Handle Dhan API specific fields
	if idx, ok := columnMap["SEM_TRADING_SYMBOL"]; ok && idx < len(row) {
		symbol.TradingSymbol = row[idx]
		if symbol.Symbol == "" {
			symbol.Symbol = row[idx]
		}
	}
	if idx, ok := columnMap["SEM_SMST_SECURITY_ID"]; ok && idx < len(row) {
		symbol.SecurityID = row[idx]
	}
	if idx, ok := columnMap["SEM_EXM_EXCH_ID"]; ok && idx < len(row) {
		symbol.ExchangeID = row[idx]
	}
	if idx, ok := columnMap["SEM_SEGMENT"]; ok && idx < len(row) {
		symbol.Segment = row[idx]
	}
}
