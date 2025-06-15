package alphavantage

import "fmt"

// AlphaVantageResponse represents common response structure from Alpha Vantage API
type AlphaVantageResponse struct {
	Information  string `json:"Information,omitempty"`
	ErrorMessage string `json:"Error Message,omitempty"`
	Note         string `json:"Note,omitempty"`
}

// TimeSeriesDailyResponse represents daily historical data response
type TimeSeriesDailyResponse struct {
	AlphaVantageResponse
	MetaData   TimeSeriesMetaData        `json:"Meta Data"`
	TimeSeries map[string]DailyStockData `json:"Time Series (Daily)"`
}

// TimeSeriesWeeklyResponse represents weekly historical data response
type TimeSeriesWeeklyResponse struct {
	AlphaVantageResponse
	MetaData   TimeSeriesMetaData         `json:"Meta Data"`
	TimeSeries map[string]WeeklyStockData `json:"Weekly Time Series"`
}

// TimeSeriesMonthlyResponse represents monthly historical data response
type TimeSeriesMonthlyResponse struct {
	AlphaVantageResponse
	MetaData   TimeSeriesMetaData          `json:"Meta Data"`
	TimeSeries map[string]MonthlyStockData `json:"Monthly Time Series"`
}

// TimeSeriesMetaData represents metadata for time series data
type TimeSeriesMetaData struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	OutputSize    string `json:"4. Output Size"`
	TimeZone      string `json:"5. Time Zone"`
}

// DailyStockData represents daily OHLCV data
type DailyStockData struct {
	Open             string `json:"1. open"`
	High             string `json:"2. high"`
	Low              string `json:"3. low"`
	Close            string `json:"4. close"`
	AdjustedClose    string `json:"5. adjusted close"`
	Volume           string `json:"6. volume"`
	DividendAmount   string `json:"7. dividend amount"`
	SplitCoefficient string `json:"8. split coefficient"`
}

// WeeklyStockData represents weekly OHLCV data
type WeeklyStockData struct {
	Open          string `json:"1. open"`
	High          string `json:"2. high"`
	Low           string `json:"3. low"`
	Close         string `json:"4. close"`
	AdjustedClose string `json:"5. adjusted close"`
	Volume        string `json:"6. volume"`
}

// MonthlyStockData represents monthly OHLCV data
type MonthlyStockData struct {
	Open          string `json:"1. open"`
	High          string `json:"2. high"`
	Low           string `json:"3. low"`
	Close         string `json:"4. close"`
	AdjustedClose string `json:"5. adjusted close"`
	Volume        string `json:"6. volume"`
}

// CompanyOverviewResponse represents fundamental data response
type CompanyOverviewResponse struct {
	AlphaVantageResponse
	Symbol                     string `json:"Symbol"`
	AssetType                  string `json:"AssetType"`
	Name                       string `json:"Name"`
	Description                string `json:"Description"`
	CIK                        string `json:"CIK"`
	Exchange                   string `json:"Exchange"`
	Currency                   string `json:"Currency"`
	Country                    string `json:"Country"`
	Sector                     string `json:"Sector"`
	Industry                   string `json:"Industry"`
	Address                    string `json:"Address"`
	FiscalYearEnd              string `json:"FiscalYearEnd"`
	LatestQuarter              string `json:"LatestQuarter"`
	MarketCapitalization       string `json:"MarketCapitalization"`
	EBITDA                     string `json:"EBITDA"`
	PERatio                    string `json:"PERatio"`
	PEGRatio                   string `json:"PEGRatio"`
	BookValue                  string `json:"BookValue"`
	DividendPerShare           string `json:"DividendPerShare"`
	DividendYield              string `json:"DividendYield"`
	EPS                        string `json:"EPS"`
	RevenuePerShareTTM         string `json:"RevenuePerShareTTM"`
	ProfitMargin               string `json:"ProfitMargin"`
	OperatingMarginTTM         string `json:"OperatingMarginTTM"`
	ReturnOnAssetsTTM          string `json:"ReturnOnAssetsTTM"`
	ReturnOnEquityTTM          string `json:"ReturnOnEquityTTM"`
	RevenueTTM                 string `json:"RevenueTTM"`
	GrossProfitTTM             string `json:"GrossProfitTTM"`
	DilutedEPSTTM              string `json:"DilutedEPSTTM"`
	QuarterlyEarningsGrowthYOY string `json:"QuarterlyEarningsGrowthYOY"`
	QuarterlyRevenueGrowthYOY  string `json:"QuarterlyRevenueGrowthYOY"`
	AnalystTargetPrice         string `json:"AnalystTargetPrice"`
	TrailingPE                 string `json:"TrailingPE"`
	ForwardPE                  string `json:"ForwardPE"`
	PriceToSalesRatioTTM       string `json:"PriceToSalesRatioTTM"`
	PriceToBookRatio           string `json:"PriceToBookRatio"`
	EVToRevenue                string `json:"EVToRevenue"`
	EVToEBITDA                 string `json:"EVToEBITDA"`
	Beta                       string `json:"Beta"`
	WeekHigh52                 string `json:"52WeekHigh"`
	WeekLow52                  string `json:"52WeekLow"`
	DayMovingAverage50         string `json:"50DayMovingAverage"`
	DayMovingAverage200        string `json:"200DayMovingAverage"`
	SharesOutstanding          string `json:"SharesOutstanding"`
	DividendDate               string `json:"DividendDate"`
	ExDividendDate             string `json:"ExDividendDate"`
}

// TechnicalIndicatorResponse represents technical indicator response structure
type TechnicalIndicatorResponse struct {
	AlphaVantageResponse
	MetaData      TechnicalIndicatorMetaData         `json:"Meta Data"`
	TechnicalData map[string]TechnicalIndicatorValue `json:"Technical Analysis: RSI,omitempty"`
}

// MACDResponse represents MACD indicator response
type MACDResponse struct {
	AlphaVantageResponse
	MetaData TechnicalIndicatorMetaData `json:"Meta Data"`
	MACD     map[string]MACDValue       `json:"Technical Analysis: MACD"`
}

// BollingerBandsResponse represents Bollinger Bands response
type BollingerBandsResponse struct {
	AlphaVantageResponse
	MetaData TechnicalIndicatorMetaData     `json:"Meta Data"`
	Bands    map[string]BollingerBandsValue `json:"Technical Analysis: BBANDS"`
}

// SMAResponse represents Simple Moving Average response
type SMAResponse struct {
	AlphaVantageResponse
	MetaData TechnicalIndicatorMetaData `json:"Meta Data"`
	SMA      map[string]SMAValue        `json:"Technical Analysis: SMA"`
}

// EMAResponse represents Exponential Moving Average response
type EMAResponse struct {
	AlphaVantageResponse
	MetaData TechnicalIndicatorMetaData `json:"Meta Data"`
	EMA      map[string]EMAValue        `json:"Technical Analysis: EMA"`
}

// RSIResponse represents RSI indicator response
type RSIResponse struct {
	AlphaVantageResponse
	MetaData TechnicalIndicatorMetaData `json:"Meta Data"`
	RSI      map[string]RSIValue        `json:"Technical Analysis: RSI"`
}

// STOCHResponse represents Stochastic Oscillator response
type STOCHResponse struct {
	AlphaVantageResponse
	MetaData TechnicalIndicatorMetaData `json:"Meta Data"`
	STOCH    map[string]STOCHValue      `json:"Technical Analysis: STOCH"`
}

// ADXResponse represents ADX indicator response
type ADXResponse struct {
	AlphaVantageResponse
	MetaData TechnicalIndicatorMetaData `json:"Meta Data"`
	ADX      map[string]ADXValue        `json:"Technical Analysis: ADX"`
}

// CCIResponse represents CCI indicator response
type CCIResponse struct {
	AlphaVantageResponse
	MetaData TechnicalIndicatorMetaData `json:"Meta Data"`
	CCI      map[string]CCIValue        `json:"Technical Analysis: CCI"`
}

// AROONResponse represents AROON indicator response
type AROONResponse struct {
	AlphaVantageResponse
	MetaData TechnicalIndicatorMetaData `json:"Meta Data"`
	AROON    map[string]AROONValue      `json:"Technical Analysis: AROON"`
}

// TechnicalIndicatorMetaData represents metadata for technical indicators
type TechnicalIndicatorMetaData struct {
	Symbol             string      `json:"1: Symbol"`
	Indicator          string      `json:"2: Indicator"`
	LastRefreshed      string      `json:"3: Last Refreshed"`
	Interval           string      `json:"4: Interval"`
	TimePeriod         interface{} `json:"5: Time Period,omitempty"`
	SeriesType         string      `json:"6: Series Type,omitempty"`
	TimeZone           string      `json:"7: Time Zone"`
	FastPeriod         string      `json:"5.1: Fast Period,omitempty"`
	SlowPeriod         string      `json:"5.2: Slow Period,omitempty"`
	SignalPeriod       string      `json:"5.3: Signal Period,omitempty"`
	FastKPeriod        string      `json:"5.1: FastK Period,omitempty"`
	SlowKPeriod        string      `json:"5.2: SlowK Period,omitempty"`
	SlowDPeriod        string      `json:"5.3: SlowD Period,omitempty"`
	SlowKMAType        string      `json:"5.4: SlowK MA Type,omitempty"`
	SlowDMAType        string      `json:"5.5: SlowD MA Type,omitempty"`
	StandardDeviations string      `json:"5.1: Standard Deviations,omitempty"`
	MAType             string      `json:"5.2: MA Type,omitempty"`
}

// GetTimePeriod returns the time period as a string, handling both string and numeric values
func (m *TechnicalIndicatorMetaData) GetTimePeriod() string {
	if m.TimePeriod == nil {
		return ""
	}

	switch v := m.TimePeriod.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// TechnicalIndicatorValue represents a generic technical indicator value
type TechnicalIndicatorValue struct {
	Value string `json:"RSI,omitempty"`
}

// MACDValue represents MACD indicator values
type MACDValue struct {
	MACD       string `json:"MACD"`
	MACDHist   string `json:"MACD_Hist"`
	MACDSignal string `json:"MACD_Signal"`
}

// BollingerBandsValue represents Bollinger Bands values
type BollingerBandsValue struct {
	RealMiddleBand string `json:"Real Middle Band"`
	RealUpperBand  string `json:"Real Upper Band"`
	RealLowerBand  string `json:"Real Lower Band"`
}

// SMAValue represents Simple Moving Average value
type SMAValue struct {
	SMA string `json:"SMA"`
}

// EMAValue represents Exponential Moving Average value
type EMAValue struct {
	EMA string `json:"EMA"`
}

// RSIValue represents RSI value
type RSIValue struct {
	RSI string `json:"RSI"`
}

// STOCHValue represents Stochastic Oscillator values
type STOCHValue struct {
	SlowK string `json:"SlowK"`
	SlowD string `json:"SlowD"`
}

// ADXValue represents ADX value
type ADXValue struct {
	ADX string `json:"ADX"`
}

// CCIValue represents CCI value
type CCIValue struct {
	CCI string `json:"CCI"`
}

// AROONValue represents AROON values
type AROONValue struct {
	AroonDown string `json:"Aroon Down"`
	AroonUp   string `json:"Aroon Up"`
}

// EarningsResponse represents earnings data response
type EarningsResponse struct {
	AlphaVantageResponse
	Symbol            string             `json:"symbol"`
	AnnualEarnings    []AnnualEarning    `json:"annualEarnings"`
	QuarterlyEarnings []QuarterlyEarning `json:"quarterlyEarnings"`
}

// AnnualEarning represents annual earnings data
type AnnualEarning struct {
	FiscalDateEnding string `json:"fiscalDateEnding"`
	ReportedEPS      string `json:"reportedEPS"`
}

// QuarterlyEarning represents quarterly earnings data
type QuarterlyEarning struct {
	FiscalDateEnding   string `json:"fiscalDateEnding"`
	ReportedDate       string `json:"reportedDate"`
	ReportedEPS        string `json:"reportedEPS"`
	EstimatedEPS       string `json:"estimatedEPS"`
	Surprise           string `json:"surprise"`
	SurprisePercentage string `json:"surprisePercentage"`
}

// IncomeStatementResponse represents income statement response
type IncomeStatementResponse struct {
	AlphaVantageResponse
	Symbol           string            `json:"symbol"`
	AnnualReports    []AnnualReport    `json:"annualReports"`
	QuarterlyReports []QuarterlyReport `json:"quarterlyReports"`
}

// AnnualReport represents annual financial report
type AnnualReport struct {
	FiscalDateEnding                  string `json:"fiscalDateEnding"`
	ReportedCurrency                  string `json:"reportedCurrency"`
	GrossProfit                       string `json:"grossProfit"`
	TotalRevenue                      string `json:"totalRevenue"`
	CostOfRevenue                     string `json:"costOfRevenue"`
	CostofGoodsAndServicesSold        string `json:"costofGoodsAndServicesSold"`
	OperatingIncome                   string `json:"operatingIncome"`
	SellingGeneralAndAdministrative   string `json:"sellingGeneralAndAdministrative"`
	ResearchAndDevelopment            string `json:"researchAndDevelopment"`
	OperatingExpenses                 string `json:"operatingExpenses"`
	InvestmentIncomeNet               string `json:"investmentIncomeNet"`
	NetInterestIncome                 string `json:"netInterestIncome"`
	InterestIncome                    string `json:"interestIncome"`
	InterestExpense                   string `json:"interestExpense"`
	NonInterestIncome                 string `json:"nonInterestIncome"`
	OtherNonOperatingIncome           string `json:"otherNonOperatingIncome"`
	Depreciation                      string `json:"depreciation"`
	DepreciationAndAmortization       string `json:"depreciationAndAmortization"`
	IncomeBeforeTax                   string `json:"incomeBeforeTax"`
	IncomeTaxExpense                  string `json:"incomeTaxExpense"`
	InterestAndDebtExpense            string `json:"interestAndDebtExpense"`
	NetIncomeFromContinuingOperations string `json:"netIncomeFromContinuingOperations"`
	ComprehensiveIncomeNetOfTax       string `json:"comprehensiveIncomeNetOfTax"`
	EBIT                              string `json:"ebit"`
	EBITDA                            string `json:"ebitda"`
	NetIncome                         string `json:"netIncome"`
}

// QuarterlyReport represents quarterly financial report
type QuarterlyReport struct {
	FiscalDateEnding                  string `json:"fiscalDateEnding"`
	ReportedCurrency                  string `json:"reportedCurrency"`
	GrossProfit                       string `json:"grossProfit"`
	TotalRevenue                      string `json:"totalRevenue"`
	CostOfRevenue                     string `json:"costOfRevenue"`
	CostofGoodsAndServicesSold        string `json:"costofGoodsAndServicesSold"`
	OperatingIncome                   string `json:"operatingIncome"`
	SellingGeneralAndAdministrative   string `json:"sellingGeneralAndAdministrative"`
	ResearchAndDevelopment            string `json:"researchAndDevelopment"`
	OperatingExpenses                 string `json:"operatingExpenses"`
	InvestmentIncomeNet               string `json:"investmentIncomeNet"`
	NetInterestIncome                 string `json:"netInterestIncome"`
	InterestIncome                    string `json:"interestIncome"`
	InterestExpense                   string `json:"interestExpense"`
	NonInterestIncome                 string `json:"nonInterestIncome"`
	OtherNonOperatingIncome           string `json:"otherNonOperatingIncome"`
	Depreciation                      string `json:"depreciation"`
	DepreciationAndAmortization       string `json:"depreciationAndAmortization"`
	IncomeBeforeTax                   string `json:"incomeBeforeTax"`
	IncomeTaxExpense                  string `json:"incomeTaxExpense"`
	InterestAndDebtExpense            string `json:"interestAndDebtExpense"`
	NetIncomeFromContinuingOperations string `json:"netIncomeFromContinuingOperations"`
	ComprehensiveIncomeNetOfTax       string `json:"comprehensiveIncomeNetOfTax"`
	EBIT                              string `json:"ebit"`
	EBITDA                            string `json:"ebitda"`
	NetIncome                         string `json:"netIncome"`
}

// BalanceSheetResponse represents balance sheet response
type BalanceSheetResponse struct {
	AlphaVantageResponse
	Symbol           string                  `json:"symbol"`
	AnnualReports    []AnnualBalanceSheet    `json:"annualReports"`
	QuarterlyReports []QuarterlyBalanceSheet `json:"quarterlyReports"`
}

// AnnualBalanceSheet represents annual balance sheet data
type AnnualBalanceSheet struct {
	FiscalDateEnding                       string `json:"fiscalDateEnding"`
	ReportedCurrency                       string `json:"reportedCurrency"`
	TotalAssets                            string `json:"totalAssets"`
	TotalCurrentAssets                     string `json:"totalCurrentAssets"`
	CashAndCashEquivalentsAtCarryingValue  string `json:"cashAndCashEquivalentsAtCarryingValue"`
	CashAndShortTermInvestments            string `json:"cashAndShortTermInvestments"`
	Inventory                              string `json:"inventory"`
	CurrentNetReceivables                  string `json:"currentNetReceivables"`
	TotalNonCurrentAssets                  string `json:"totalNonCurrentAssets"`
	PropertyPlantEquipment                 string `json:"propertyPlantEquipment"`
	AccumulatedDepreciationAmortizationPPE string `json:"accumulatedDepreciationAmortizationPPE"`
	IntangibleAssets                       string `json:"intangibleAssets"`
	IntangibleAssetsExcludingGoodwill      string `json:"intangibleAssetsExcludingGoodwill"`
	Goodwill                               string `json:"goodwill"`
	Investments                            string `json:"investments"`
	LongTermInvestments                    string `json:"longTermInvestments"`
	ShortTermInvestments                   string `json:"shortTermInvestments"`
	OtherCurrentAssets                     string `json:"otherCurrentAssets"`
	OtherNonCurrentAssets                  string `json:"otherNonCurrentAssets"`
	TotalLiabilities                       string `json:"totalLiabilities"`
	TotalCurrentLiabilities                string `json:"totalCurrentLiabilities"`
	CurrentAccountsPayable                 string `json:"currentAccountsPayable"`
	DeferredRevenue                        string `json:"deferredRevenue"`
	CurrentDebt                            string `json:"currentDebt"`
	ShortTermDebt                          string `json:"shortTermDebt"`
	TotalNonCurrentLiabilities             string `json:"totalNonCurrentLiabilities"`
	CapitalLeaseObligations                string `json:"capitalLeaseObligations"`
	LongTermDebt                           string `json:"longTermDebt"`
	CurrentLongTermDebt                    string `json:"currentLongTermDebt"`
	LongTermDebtNoncurrent                 string `json:"longTermDebtNoncurrent"`
	ShortLongTermDebtTotal                 string `json:"shortLongTermDebtTotal"`
	OtherCurrentLiabilities                string `json:"otherCurrentLiabilities"`
	OtherNonCurrentLiabilities             string `json:"otherNonCurrentLiabilities"`
	TotalShareholderEquity                 string `json:"totalShareholderEquity"`
	TreasuryStock                          string `json:"treasuryStock"`
	RetainedEarnings                       string `json:"retainedEarnings"`
	CommonStock                            string `json:"commonStock"`
	CommonStockSharesOutstanding           string `json:"commonStockSharesOutstanding"`
}

// QuarterlyBalanceSheet represents quarterly balance sheet data
type QuarterlyBalanceSheet struct {
	FiscalDateEnding                       string `json:"fiscalDateEnding"`
	ReportedCurrency                       string `json:"reportedCurrency"`
	TotalAssets                            string `json:"totalAssets"`
	TotalCurrentAssets                     string `json:"totalCurrentAssets"`
	CashAndCashEquivalentsAtCarryingValue  string `json:"cashAndCashEquivalentsAtCarryingValue"`
	CashAndShortTermInvestments            string `json:"cashAndShortTermInvestments"`
	Inventory                              string `json:"inventory"`
	CurrentNetReceivables                  string `json:"currentNetReceivables"`
	TotalNonCurrentAssets                  string `json:"totalNonCurrentAssets"`
	PropertyPlantEquipment                 string `json:"propertyPlantEquipment"`
	AccumulatedDepreciationAmortizationPPE string `json:"accumulatedDepreciationAmortizationPPE"`
	IntangibleAssets                       string `json:"intangibleAssets"`
	IntangibleAssetsExcludingGoodwill      string `json:"intangibleAssetsExcludingGoodwill"`
	Goodwill                               string `json:"goodwill"`
	Investments                            string `json:"investments"`
	LongTermInvestments                    string `json:"longTermInvestments"`
	ShortTermInvestments                   string `json:"shortTermInvestments"`
	OtherCurrentAssets                     string `json:"otherCurrentAssets"`
	OtherNonCurrentAssets                  string `json:"otherNonCurrentAssets"`
	TotalLiabilities                       string `json:"totalLiabilities"`
	TotalCurrentLiabilities                string `json:"totalCurrentLiabilities"`
	CurrentAccountsPayable                 string `json:"currentAccountsPayable"`
	DeferredRevenue                        string `json:"deferredRevenue"`
	CurrentDebt                            string `json:"currentDebt"`
	ShortTermDebt                          string `json:"shortTermDebt"`
	TotalNonCurrentLiabilities             string `json:"totalNonCurrentLiabilities"`
	CapitalLeaseObligations                string `json:"capitalLeaseObligations"`
	LongTermDebt                           string `json:"longTermDebt"`
	CurrentLongTermDebt                    string `json:"currentLongTermDebt"`
	LongTermDebtNoncurrent                 string `json:"longTermDebtNoncurrent"`
	ShortLongTermDebtTotal                 string `json:"shortLongTermDebtTotal"`
	OtherCurrentLiabilities                string `json:"otherCurrentLiabilities"`
	OtherNonCurrentLiabilities             string `json:"otherNonCurrentLiabilities"`
	TotalShareholderEquity                 string `json:"totalShareholderEquity"`
	TreasuryStock                          string `json:"treasuryStock"`
	RetainedEarnings                       string `json:"retainedEarnings"`
	CommonStock                            string `json:"commonStock"`
	CommonStockSharesOutstanding           string `json:"commonStockSharesOutstanding"`
}

// CashFlowResponse represents cash flow statement response
type CashFlowResponse struct {
	AlphaVantageResponse
	Symbol           string              `json:"symbol"`
	AnnualReports    []AnnualCashFlow    `json:"annualReports"`
	QuarterlyReports []QuarterlyCashFlow `json:"quarterlyReports"`
}

// AnnualCashFlow represents annual cash flow data
type AnnualCashFlow struct {
	FiscalDateEnding                                          string `json:"fiscalDateEnding"`
	ReportedCurrency                                          string `json:"reportedCurrency"`
	OperatingCashflow                                         string `json:"operatingCashflow"`
	PaymentsForOperatingActivities                            string `json:"paymentsForOperatingActivities"`
	ProceedsFromOperatingActivities                           string `json:"proceedsFromOperatingActivities"`
	ChangeInOperatingLiabilities                              string `json:"changeInOperatingLiabilities"`
	ChangeInOperatingAssets                                   string `json:"changeInOperatingAssets"`
	DepreciationDepletionAndAmortization                      string `json:"depreciationDepletionAndAmortization"`
	CapitalExpenditures                                       string `json:"capitalExpenditures"`
	ChangeInReceivables                                       string `json:"changeInReceivables"`
	ChangeInInventory                                         string `json:"changeInInventory"`
	ProfitLoss                                                string `json:"profitLoss"`
	CashflowFromInvestment                                    string `json:"cashflowFromInvestment"`
	CashflowFromFinancing                                     string `json:"cashflowFromFinancing"`
	ProceedsFromRepaymentsOfShortTermDebt                     string `json:"proceedsFromRepaymentsOfShortTermDebt"`
	PaymentsForRepurchaseOfCommonStock                        string `json:"paymentsForRepurchaseOfCommonStock"`
	PaymentsForRepurchaseOfEquity                             string `json:"paymentsForRepurchaseOfEquity"`
	PaymentsForRepurchaseOfPreferredStock                     string `json:"paymentsForRepurchaseOfPreferredStock"`
	DividendPayout                                            string `json:"dividendPayout"`
	DividendPayoutCommonStock                                 string `json:"dividendPayoutCommonStock"`
	DividendPayoutPreferredStock                              string `json:"dividendPayoutPreferredStock"`
	ProceedsFromIssuanceOfCommonStock                         string `json:"proceedsFromIssuanceOfCommonStock"`
	ProceedsFromIssuanceOfLongTermDebtAndCapitalSecuritiesNet string `json:"proceedsFromIssuanceOfLongTermDebtAndCapitalSecuritiesNet"`
	ProceedsFromIssuanceOfPreferredStock                      string `json:"proceedsFromIssuanceOfPreferredStock"`
	ProceedsFromRepurchaseOfEquity                            string `json:"proceedsFromRepurchaseOfEquity"`
	ProceedsFromSaleOfTreasuryStock                           string `json:"proceedsFromSaleOfTreasuryStock"`
	ChangeInCashAndCashEquivalents                            string `json:"changeInCashAndCashEquivalents"`
	ChangeInExchangeRate                                      string `json:"changeInExchangeRate"`
	NetIncome                                                 string `json:"netIncome"`
}

// QuarterlyCashFlow represents quarterly cash flow data
type QuarterlyCashFlow struct {
	FiscalDateEnding                                          string `json:"fiscalDateEnding"`
	ReportedCurrency                                          string `json:"reportedCurrency"`
	OperatingCashflow                                         string `json:"operatingCashflow"`
	PaymentsForOperatingActivities                            string `json:"paymentsForOperatingActivities"`
	ProceedsFromOperatingActivities                           string `json:"proceedsFromOperatingActivities"`
	ChangeInOperatingLiabilities                              string `json:"changeInOperatingLiabilities"`
	ChangeInOperatingAssets                                   string `json:"changeInOperatingAssets"`
	DepreciationDepletionAndAmortization                      string `json:"depreciationDepletionAndAmortization"`
	CapitalExpenditures                                       string `json:"capitalExpenditures"`
	ChangeInReceivables                                       string `json:"changeInReceivables"`
	ChangeInInventory                                         string `json:"changeInInventory"`
	ProfitLoss                                                string `json:"profitLoss"`
	CashflowFromInvestment                                    string `json:"cashflowFromInvestment"`
	CashflowFromFinancing                                     string `json:"cashflowFromFinancing"`
	ProceedsFromRepaymentsOfShortTermDebt                     string `json:"proceedsFromRepaymentsOfShortTermDebt"`
	PaymentsForRepurchaseOfCommonStock                        string `json:"paymentsForRepurchaseOfCommonStock"`
	PaymentsForRepurchaseOfEquity                             string `json:"paymentsForRepurchaseOfEquity"`
	PaymentsForRepurchaseOfPreferredStock                     string `json:"paymentsForRepurchaseOfPreferredStock"`
	DividendPayout                                            string `json:"dividendPayout"`
	DividendPayoutCommonStock                                 string `json:"dividendPayoutCommonStock"`
	DividendPayoutPreferredStock                              string `json:"dividendPayoutPreferredStock"`
	ProceedsFromIssuanceOfCommonStock                         string `json:"proceedsFromIssuanceOfCommonStock"`
	ProceedsFromIssuanceOfLongTermDebtAndCapitalSecuritiesNet string `json:"proceedsFromIssuanceOfLongTermDebtAndCapitalSecuritiesNet"`
	ProceedsFromIssuanceOfPreferredStock                      string `json:"proceedsFromIssuanceOfPreferredStock"`
	ProceedsFromRepurchaseOfEquity                            string `json:"proceedsFromRepurchaseOfEquity"`
	ProceedsFromSaleOfTreasuryStock                           string `json:"proceedsFromSaleOfTreasuryStock"`
	ChangeInCashAndCashEquivalents                            string `json:"changeInCashAndCashEquivalents"`
	ChangeInExchangeRate                                      string `json:"changeInExchangeRate"`
	NetIncome                                                 string `json:"netIncome"`
}
