package entity

import "time"

type MasterData struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Customer struct {
	ID           int64     `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	CustomerType string    `json:"customer_type"`
	Phone        string    `json:"phone"`
	Address      string    `json:"address"`
	CreatedAt    time.Time `json:"created_at"`
}

type TransactionLine struct {
	ItemID    int64  `json:"item_id" validate:"required,gt=0"`
	ItemName  string `json:"item_name,omitempty"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
	UnitPrice int64  `json:"unit_price" validate:"gte=0"`
	Subtotal  int64  `json:"subtotal,omitempty"`
}

type CreateTransactionRequest struct {
	TransactionType string            `json:"transaction_type" validate:"required,oneof=SALE PURCHASE"`
	CustomerID      *int64            `json:"customer_id"`
	SupplierID      *int64            `json:"supplier_id"`
	PaymentMethodID *int64            `json:"payment_method_id"`
	PaidAmount      int64             `json:"paid_amount" validate:"gte=0"`
	Notes           string            `json:"notes" validate:"max=500"`
	Items           []TransactionLine `json:"items" validate:"required,min=1,dive"`
}

type Transaction struct {
	ID                int64             `json:"id"`
	InvoiceNo         string            `json:"invoice_no"`
	TransactionType   string            `json:"transaction_type"`
	TransactionDate   time.Time         `json:"transaction_date"`
	CustomerID        *int64            `json:"customer_id"`
	CustomerName      *string           `json:"customer_name"`
	SupplierID        *int64            `json:"supplier_id"`
	SupplierName      *string           `json:"supplier_name"`
	PaymentMethodID   *int64            `json:"payment_method_id"`
	PaymentMethodName *string           `json:"payment_method_name"`
	PaymentStatus     string            `json:"payment_status"`
	GrandTotal        int64             `json:"grand_total"`
	PaidAmount        int64             `json:"paid_amount"`
	AmountReceived    int64             `json:"amount_received"`
	ChangeAmount      int64             `json:"change_amount"`
	Status            string            `json:"status"`
	VoidReason        string            `json:"void_reason"`
	Notes             string            `json:"notes"`
	Items             []TransactionLine `json:"items,omitempty"`
}

type Debt struct {
	ID              int64     `json:"id"`
	TransactionID   int64     `json:"transaction_id"`
	InvoiceNo       string    `json:"invoice_no"`
	CustomerID      int64     `json:"customer_id"`
	CustomerName    string    `json:"customer_name"`
	OriginalAmount  int64     `json:"original_amount"`
	RemainingAmount int64     `json:"remaining_amount"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type Dashboard struct {
	TodaySales     int64          `json:"today_sales"`
	TodayPurchases int64          `json:"today_purchases"`
	OpenDebt       int64          `json:"open_debt"`
	LowStockItems  int64          `json:"low_stock_items"`
	TotalItems     int64          `json:"total_items"`
	TotalCustomers int64          `json:"total_customers"`
	TotalSuppliers int64          `json:"total_suppliers"`
	Year           int            `json:"year"`
	MonthlySales   []int64        `json:"monthly_sales"`
	Month          int            `json:"month"`
	Today          PeriodSummary  `json:"today"`
	Yesterday      PeriodSummary  `json:"yesterday"`
	SelectedMonth  PeriodSummary  `json:"selected_month"`
	SelectedYear   PeriodSummary  `json:"selected_year"`
	Daily          []DailySummary `json:"daily"`
}

type PeriodSummary struct {
	Income  int64 `json:"income"`
	Expense int64 `json:"expense"`
	Debt    int64 `json:"debt"`
}

type DailySummary struct {
	Date      string `json:"date"`
	Income    int64  `json:"income"`
	Expense   int64  `json:"expense"`
	Debt      int64  `json:"debt"`
	NetIncome int64  `json:"net_income"`
}
