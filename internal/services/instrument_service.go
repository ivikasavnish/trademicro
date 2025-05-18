package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vikasavnish/trademicro/internal/models"
	"gorm.io/gorm"
)

// InstrumentService handles instrument-related business logic
type InstrumentService struct {
	db *gorm.DB
}

// NewInstrumentService creates a new instrument service with database connection
func NewInstrumentService(db *gorm.DB) *InstrumentService {
	return &InstrumentService{
		db: db,
	}
}

// FetchMode describes the data mode to retrieve from DhanHQ
type FetchMode string

const (
	CompactMode  FetchMode = "compact"
	DetailedMode FetchMode = "detailed"
)

// FetchOptions contains options for fetching instruments
type FetchOptions struct {
	Mode            FetchMode
	ExchangeSegment string
	SaveToFile      bool
	OutputFile      string
	BatchSize       int
}

// FetchResult contains the results of an instrument fetch operation
type FetchResult struct {
	Created      int       `json:"created"`
	Updated      int       `json:"updated"`
	Skipped      int       `json:"skipped"`
	Total        int       `json:"total"`
	TotalBatches int       `json:"totalBatches"`
	BatchSize    int       `json:"batchSize"`
	Time         time.Time `json:"time"`
	Mode         string    `json:"mode"`
}

// FetchInstrumentsFromDhanHQ fetches instrument data from DhanHQ API and saves to database
func (s *InstrumentService) FetchInstrumentsFromDhanHQ(options FetchOptions) (*FetchResult, error) {
	// Validate options
	if options.BatchSize <= 0 {
		options.BatchSize = 500
	}

	var csvURL string
	switch options.Mode {
	case CompactMode:
		csvURL = "https://images.dhan.co/api-data/api-scrip-master.csv"
	case DetailedMode:
		csvURL = "https://images.dhan.co/api-data/api-scrip-master-detailed.csv"
	default:
		return nil, fmt.Errorf("invalid fetch mode: %s", options.Mode)
	}

	// Create output filename if not provided
	if options.OutputFile == "" && options.SaveToFile {
		options.OutputFile = fmt.Sprintf("dhan_instruments_%s_%s.csv",
			options.Mode,
			time.Now().Format("20060102_150405"))
	}

	// Fetch CSV data from DhanHQ
	log.Printf("Fetching instrument data from %s", csvURL)
	resp, err := http.Get(csvURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CSV: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Save file if requested
	var csvReader *csv.Reader
	if options.SaveToFile {
		file, err := os.Create(options.OutputFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()

		// Copy data to file
		if _, err := io.Copy(file, resp.Body); err != nil {
			return nil, fmt.Errorf("failed to save CSV data: %w", err)
		}

		// Reopen file for reading
		file, err = os.Open(options.OutputFile)
		if err != nil {
			return nil, fmt.Errorf("failed to open saved file: %w", err)
		}
		defer file.Close()

		csvReader = csv.NewReader(file)
	} else {
		csvReader = csv.NewReader(resp.Body)
	}

	// Parse CSV data
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV has insufficient data")
	}

	// Get header row to map column indices
	headers := records[0]
	columnMap := make(map[string]int)
	for i, header := range headers {
		columnMap[header] = i
	}

	// Calculate batches
	dataRows := len(records) - 1 // Exclude header
	totalBatches := (dataRows + options.BatchSize - 1) / options.BatchSize

	// Process in batches
	result := &FetchResult{
		BatchSize:    options.BatchSize,
		TotalBatches: totalBatches,
		Time:         time.Now(),
		Mode:         string(options.Mode),
	}

	log.Printf("Processing %d instruments in %d batches of size %d",
		dataRows, totalBatches, options.BatchSize)

	for batchNum := 0; batchNum < totalBatches; batchNum++ {
		startIdx := batchNum*options.BatchSize + 1 // +1 to skip header
		endIdx := min((batchNum+1)*options.BatchSize+1, len(records))

		log.Printf("Processing batch %d/%d (rows %d-%d)",
			batchNum+1, totalBatches, startIdx, endIdx-1)

		// Begin transaction for this batch
		tx := s.db.Begin()
		if tx.Error != nil {
			return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
		}

		batchCreated := 0
		batchUpdated := 0
		batchSkipped := 0

		// Process each row in batch
		for i := startIdx; i < endIdx; i++ {
			row := records[i]
			if len(row) < len(headers) {
				batchSkipped++
				continue
			}

			// Extract instrument data
			instrument, err := s.extractInstrumentFromCSV(row, columnMap, string(options.Mode))
			if err != nil {
				batchSkipped++
				continue
			}

			// Filter by exchange segment if specified
			if options.ExchangeSegment != "" {
				fullSegment := instrument.ExchangeID + "_" + instrument.SegmentType
				if !strings.Contains(strings.ToLower(fullSegment),
					strings.ToLower(options.ExchangeSegment)) {
					batchSkipped++
					continue
				}
			}

			// Check if instrument already exists
			var existingInstrument models.Instrument
			result := tx.Where("security_id = ?", instrument.SecurityID).First(&existingInstrument)

			if result.Error == nil {
				// Update existing instrument
				instrument.ID = existingInstrument.ID
				instrument.CreatedAt = existingInstrument.CreatedAt

				if err := tx.Save(instrument).Error; err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("failed to update instrument %s: %w",
						instrument.SecurityID, err)
				}
				batchUpdated++
			} else {
				// Create new instrument
				if err := tx.Create(instrument).Error; err != nil {
					tx.Rollback()
					return nil, fmt.Errorf("failed to create instrument %s: %w",
						instrument.SecurityID, err)
				}
				batchCreated++
			}
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			return nil, fmt.Errorf("failed to commit batch %d: %w", batchNum+1, err)
		}

		// Update results
		result.Created += batchCreated
		result.Updated += batchUpdated
		result.Skipped += batchSkipped

		log.Printf("Batch %d/%d completed: created=%d, updated=%d, skipped=%d",
			batchNum+1, totalBatches, batchCreated, batchUpdated, batchSkipped)
	}

	result.Total = result.Created + result.Updated + result.Skipped
	log.Printf("Instrument fetch completed: created=%d, updated=%d, skipped=%d, total=%d",
		result.Created, result.Updated, result.Skipped, result.Total)

	// Clean up temporary file if not keeping it
	if !options.SaveToFile && options.OutputFile != "" {
		os.Remove(options.OutputFile)
	}

	return result, nil
}

// extractInstrumentFromCSV extracts instrument data from a CSV row
func (s *InstrumentService) extractInstrumentFromCSV(row []string, columnMap map[string]int, mode string) (*models.Instrument, error) {
	instrument := &models.Instrument{
		DataSource: mode,
	}

	// Helper to get string value safely
	getString := func(names ...string) string {
		for _, name := range names {
			if idx, ok := columnMap[name]; ok && idx < len(row) {
				return row[idx]
			}
		}
		return ""
	}

	// Helper to get float value safely
	getFloat := func(names ...string) float64 {
		for _, name := range names {
			if idx, ok := columnMap[name]; ok && idx < len(row) {
				val, err := strconv.ParseFloat(row[idx], 64)
				if err == nil {
					return val
				}
			}
		}
		return 0
	}

	// Helper to get int value safely
	getInt := func(names ...string) int64 {
		for _, name := range names {
			if idx, ok := columnMap[name]; ok && idx < len(row) {
				val, err := strconv.ParseInt(row[idx], 10, 64)
				if err == nil {
					return val
				}
			}
		}
		return 0
	}

	// Get SecurityID (primary key)
	securityID := getString("SECURITY_ID", "SEM_SCRIP_ID")
	if securityID == "" {
		return nil, fmt.Errorf("missing security ID")
	}
	instrument.SecurityID = securityID

	// Basic symbol identification
	instrument.Symbol = getString("SYMBOL", "SEM_TRADING_SYMBOL")
	instrument.SymbolName = getString("SYMBOL_NAME", "SM_SYMBOL_NAME")
	instrument.DisplayName = getString("DISPLAY_NAME", "SEM_CUSTOM_SYMBOL")

	// Exchange and market information
	instrument.ExchangeID = getString("EXCH_ID", "SEM_EXM_EXCH_ID")
	instrument.Segment = getString("SEGMENT", "SEM_SEGMENT")

	// Map segment to segment type
	segment := instrument.Segment
	switch segment {
	case "E":
		instrument.SegmentType = "EQ" // Equity
	case "D":
		instrument.SegmentType = "IDX" // Derivatives
	case "C":
		instrument.SegmentType = "CUR" // Currency
	case "M":
		instrument.SegmentType = "COMM" // Commodity
	default:
		instrument.SegmentType = "OTHER"
	}

	instrument.ISIN = getString("ISIN")
	instrument.Instrument = getString("INSTRUMENT", "SEM_INSTRUMENT_NAME")
	instrument.InstrumentType = getString("INSTRUMENT_TYPE", "SEM_EXCH_INSTRUMENT_TYPE")
	instrument.Series = getString("SERIES", "SEM_SERIES")

	// Contract specifications
	instrument.LotSize = getInt("LOT_SIZE", "SEM_LOT_UNITS")
	instrument.TickSize = getFloat("TICK_SIZE", "SEM_TICK_SIZE")

	// Derivative specific information
	expDateStr := getString("SM_EXPIRY_DATE", "SEM_EXPIRY_DATE")
	if expDateStr != "" {
		// Try to parse date in multiple formats
		layouts := []string{"2006-01-02", "02-01-2006", "2006/01/02", "02/01/2006"}
		for _, layout := range layouts {
			expDate, err := time.Parse(layout, expDateStr)
			if err == nil {
				instrument.ExpiryDate = &expDate
				break
			}
		}
	}

	instrument.ExpiryCode = getString("SEM_EXPIRY_CODE")
	instrument.ExpiryFlag = getString("EXPIRY_FLAG", "SEM_EXPIRY_FLAG")
	instrument.StrikePrice = getFloat("STRIKE_PRICE", "SEM_STRIKE_PRICE")
	instrument.OptionType = getString("OPTION_TYPE", "SEM_OPTION_TYPE")
	instrument.UnderlyingSecurityID = getString("UNDERLYING_SECURITY_ID")
	instrument.UnderlyingSymbol = getString("UNDERLYING_SYMBOL")

	// Trading restrictions and parameters
	instrument.BracketFlag = getString("BRACKET_FLAG")
	instrument.CoverFlag = getString("COVER_FLAG")
	instrument.ASMGSMFlag = getString("ASM_GSM_FLAG")
	instrument.ASMGSMCategory = getString("ASM_GSM_CATEGORY")
	instrument.BuySellIndicator = getString("BUY_SELL_INDICATOR")

	// Order margins and limits
	instrument.BuyCoMinMarginPer = getFloat("BUY_CO_MIN_MARGIN_PER")
	instrument.SellCoMinMarginPer = getFloat("SELL_CO_MIN_MARGIN_PER")
	instrument.BuyCoSlRangeMaxPerc = getFloat("BUY_CO_SL_RANGE_MAX_PERC")
	instrument.SellCoSlRangeMaxPerc = getFloat("SELL_CO_SL_RANGE_MAX_PERC")
	instrument.BuyCoSlRangeMinPerc = getFloat("BUY_CO_SL_RANGE_MIN_PERC")
	instrument.SellCoSlRangeMinPerc = getFloat("SELL_CO_SL_RANGE_MIN_PERC")
	instrument.BuyBoMinMarginPer = getFloat("BUY_BO_MIN_MARGIN_PER")
	instrument.SellBoMinMarginPer = getFloat("SELL_BO_MIN_MARGIN_PER")
	instrument.BuyBoSlRangeMaxPerc = getFloat("BUY_BO_SL_RANGE_MAX_PERC")
	instrument.SellBoSlRangeMaxPerc = getFloat("SELL_BO_SL_RANGE_MAX_PERC")
	instrument.BuyBoSlRangeMinPerc = getFloat("BUY_BO_SL_RANGE_MIN_PERC")
	instrument.SellBoSlMinRange = getFloat("SELL_BO_SL_MIN_RANGE")
	instrument.BuyBoProfitRangeMaxPerc = getFloat("BUY_BO_PROFIT_RANGE_MAX_PERC")
	instrument.SellBoProfitRangeMaxPerc = getFloat("SELL_BO_PROFIT_RANGE_MAX_PERC")
	instrument.BuyBoProfitRangeMinPerc = getFloat("BUY_BO_PROFIT_RANGE_MIN_PERC")
	instrument.SellBoProfitRangeMinPerc = getFloat("SELL_BO_PROFIT_RANGE_MIN_PERC")
	instrument.MTFLeverage = getFloat("MTF_LEVERAGE")

	return instrument, nil
}

// Query filters for instruments
type InstrumentQuery struct {
	SecurityID       string    // Exact match
	Symbol           string    // Partial match
	ExchangeID       string    // Exact match
	Segment          string    // Exact match
	ISIN             string    // Exact match
	InstrumentType   string    // Partial match
	StrikePriceMin   float64   // Range filter
	StrikePriceMax   float64   // Range filter
	ExpiryDateStart  time.Time // Range filter
	ExpiryDateEnd    time.Time // Range filter
	OptionType       string    // Exact match
	UnderlyingSymbol string    // Exact match
	Series           string    // Exact match
	Limit            int       // Result limit
	Offset           int       // Result offset
}

// GetInstruments fetches instruments based on query filters
func (s *InstrumentService) GetInstruments(query InstrumentQuery) ([]models.Instrument, int64, error) {
	// Start building query
	dbQuery := s.db.Model(&models.Instrument{})

	// Apply filters
	if query.SecurityID != "" {
		dbQuery = dbQuery.Where("security_id = ?", query.SecurityID)
	}

	if query.Symbol != "" {
		dbQuery = dbQuery.Where("symbol LIKE ?", "%"+query.Symbol+"%")
	}

	if query.ExchangeID != "" {
		dbQuery = dbQuery.Where("exchange_id = ?", query.ExchangeID)
	}

	if query.Segment != "" {
		dbQuery = dbQuery.Where("segment = ?", query.Segment)
	}

	if query.ISIN != "" {
		dbQuery = dbQuery.Where("isin = ?", query.ISIN)
	}

	if query.InstrumentType != "" {
		dbQuery = dbQuery.Where("instrument_type LIKE ?", "%"+query.InstrumentType+"%")
	}

	if query.StrikePriceMin > 0 {
		dbQuery = dbQuery.Where("strike_price >= ?", query.StrikePriceMin)
	}

	if query.StrikePriceMax > 0 {
		dbQuery = dbQuery.Where("strike_price <= ?", query.StrikePriceMax)
	}

	if !query.ExpiryDateStart.IsZero() {
		dbQuery = dbQuery.Where("expiry_date >= ?", query.ExpiryDateStart)
	}

	if !query.ExpiryDateEnd.IsZero() {
		dbQuery = dbQuery.Where("expiry_date <= ?", query.ExpiryDateEnd)
	}

	if query.OptionType != "" {
		dbQuery = dbQuery.Where("option_type = ?", query.OptionType)
	}

	if query.UnderlyingSymbol != "" {
		dbQuery = dbQuery.Where("underlying_symbol = ?", query.UnderlyingSymbol)
	}

	if query.Series != "" {
		dbQuery = dbQuery.Where("series = ?", query.Series)
	}

	// Count total records for pagination
	var count int64
	if err := dbQuery.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count instruments: %w", err)
	}

	// Apply limit and offset for pagination
	if query.Limit > 0 {
		dbQuery = dbQuery.Limit(query.Limit)
	}

	if query.Offset > 0 {
		dbQuery = dbQuery.Offset(query.Offset)
	}

	// Execute the query
	var instruments []models.Instrument
	if err := dbQuery.Find(&instruments).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to query instruments: %w", err)
	}

	return instruments, count, nil
}

// GetInstrumentByID fetches a single instrument by ID
func (s *InstrumentService) GetInstrumentByID(id uint) (*models.Instrument, error) {
	var instrument models.Instrument
	if err := s.db.First(&instrument, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get instrument with ID %d: %w", id, err)
	}
	return &instrument, nil
}

// GetInstrumentBySecurityID fetches a single instrument by security ID
func (s *InstrumentService) GetInstrumentBySecurityID(securityID string) (*models.Instrument, error) {
	var instrument models.Instrument
	if err := s.db.Where("security_id = ?", securityID).First(&instrument).Error; err != nil {
		return nil, fmt.Errorf("failed to get instrument with security ID %s: %w", securityID, err)
	}
	return &instrument, nil
}

// ScheduledUpdateOptions configures how scheduled updates work
type ScheduledUpdateOptions struct {
	Mode         FetchMode
	BatchSize    int
	HourOfDay    int
	MinuteOfHour int
}

// StartScheduledUpdates begins a scheduled job to update instruments
func (s *InstrumentService) StartScheduledUpdates(options ScheduledUpdateOptions) chan struct{} {
	stopCh := make(chan struct{})

	// Set default values if not provided
	if options.BatchSize <= 0 {
		options.BatchSize = 500
	}
	if options.HourOfDay < 0 || options.HourOfDay > 23 {
		options.HourOfDay = 3 // Default to 3 AM
	}
	if options.MinuteOfHour < 0 || options.MinuteOfHour > 59 {
		options.MinuteOfHour = 0
	}

	go func() {
		for {
			// Calculate time until next run
			now := time.Now()
			nextRun := time.Date(
				now.Year(), now.Month(), now.Day(),
				options.HourOfDay, options.MinuteOfHour, 0, 0,
				now.Location(),
			)

			// If already past the scheduled time, schedule for tomorrow
			if now.After(nextRun) {
				nextRun = nextRun.Add(24 * time.Hour)
			}

			waitDuration := nextRun.Sub(now)
			log.Printf("Scheduled instrument update will run at %02d:%02d, waiting %v",
				options.HourOfDay, options.MinuteOfHour, waitDuration)

			// Wait until next run time or stop signal
			select {
			case <-time.After(waitDuration):
				// Run the update
				log.Println("Starting scheduled instrument update")
				fetchOptions := FetchOptions{
					Mode:       options.Mode,
					BatchSize:  options.BatchSize,
					SaveToFile: true,
					OutputFile: fmt.Sprintf("dhan_instruments_scheduled_%s.csv",
						time.Now().Format("20060102")),
				}

				result, err := s.FetchInstrumentsFromDhanHQ(fetchOptions)
				if err != nil {
					log.Printf("Scheduled update error: %v", err)
				} else {
					log.Printf("Scheduled update completed: created=%d, updated=%d, skipped=%d",
						result.Created, result.Updated, result.Skipped)
				}

				// Clean up temporary file
				if fetchOptions.SaveToFile {
					os.Remove(fetchOptions.OutputFile)
				}

			case <-stopCh:
				log.Println("Scheduled instrument updates stopped")
				return
			}
		}
	}()

	return stopCh
}

// Helper function for min of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
