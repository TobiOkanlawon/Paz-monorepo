package web_app

import (
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

type IStore interface {
	AuthenticateUser(email string, password string) (LoginPostInformation, error)
	RegisterUser(firstName, lastName, email, password string) (RegisterPostInformation, error)
	GetHomeScreenInformation(userID uint) (HomeScreenInformation, error)
	GetProfileScreenInformation(userID uint) (ProfileScreenInformation, error)
	GetSavingsScreenInformation(userID uint) (SavingsScreenInformation, error)
	GetFamilyVaultScreenInformation(userID uint) (FamilyVaultScreenInformation, error)
	GetFamilyVaultPlanScreenInformation(userID uint, planID int) (FamilyVaultPlanScreenInformation, error)
	GetSoloSaverScreenInformation(userID uint) (SoloSaverScreenInformation, error)
	GetTargetSavingsScreenInformation(userID uint) (TargetSavingsScreenInformation, error)
	GetTargetSavingsPlanScreenInformation(userID uint, planID int) (TargetSavingsPlanScreenInformation, error)
	GetLoansScreenInformation(userID uint) (LoansScreenInformation, error)
	CreateLoanApplication(userID uint, amount uint64, termDuration uint64, bvn uint64) (LoanApplicationInformation, error)
	GetThriftScreenInformation(userID uint) (ThriftScreenInformation, error)
	CreatePayment(userID, planID uint, referenceNumber uuid.UUID, paymentoriginator string, amountInK int64) (PaymentInformation, error)
	GetLoanScreenInformation(userID uint) (GetLoanScreenInformation, error)
	CreateNewFamilyVault(userID uint, familyName, familyMemberEmail string, amount float64, frequency string, duration int64) (FamilyVaultInformation, error)
	GetPaystackVerificationInformation(referenceNumber string) (PaystackTransactionInformation, error)
	UpdateSoloSaverPaymentInformation(amountInK uint64, customerID uint, referenceNumber uuid.UUID) (SoloSaverPaymentInformation, error)
	UpdateSoloSaverPaymentFailure(referenceNumber uuid.UUID) (SoloSaverPaymentInformation, error)
	CreateInvestmentApplication(userID uint, employmentInformation string, yearOfEmployment time.Time, employerName string, investmentAmount uint64, investmentTenure uint64, taxIdentificationNumber uint64, bankAccountName string, bankAccountNumber uint64) (InvestmentApplicationInformation, error)
	GetInvestmentsScreenInformation(userID uint) (InvestmentsScreenInformation, error)
}

type User struct {
	ID        uint
	FirstName string
	LastName  string
	Email     string
}

type Activity struct {
	PrimaryInformation   string
	SecondaryInformation string
	Time                 time.Time
}

type HomeScreenInformation struct {
	FirstName         string
	LastName          string
	SavingsBalance    int64
	LoansBalance      int64
	InvestmentBalance int64
	Activities        []Activity
	isBVNAdded        bool
	isDebitCardAdded  bool
	ShowModal         bool
}

type ProfileScreenInformation struct {
	FirstName       string
	LastName        string
	EmailAddress    string
	PostalAddress   string
	PhoneNumber     string
	Sex             string
	DateOfBirth     time.Time
	NextOfKin       NextOfKin
	CompletionCount int
}

type FamilyVaultScreenInformation struct {
	FamilyVaultBasicPlans []FamilyVaultBasicPlan
	Balance               string
}

type FamilyVaultPlanScreenInformation = FamilyVaultBasicPlan

type NextOfKin struct {
	FirstName    string
	LastName     string
	EmailAddress string
	PhoneNumber  string
	Relationship string
}

type SavingsScreenInformation struct {
	Balance uint64
}

type BasicSavingsPlan struct {
	PlanID          uint
	Name            string
	Description     string
	Amount          uint64
	OwnerStatus     bool
	NumberOfMembers uint
}

type FamilyVaultBasicPlan struct {
	ID              string
	Name            string
	Description     string
	Balance         string
	IsCreator       bool
	NumberOfMembers uint
}

type SoloSaverScreenInformation struct {
	Balance           uint64
	Accounts          []DBUserBankAccount
	EmailAddress      string
	HasPendingPayment bool
}

type TargetSavingsScreenInformation struct {
	Balance uint64
	Plans   []TargetSavingsBasicPlan
}

type TargetSavingsBasicPlan struct {
	ID                   string
	Name                 string
	Description          string
	Balance              uint64
	CompletionPercentage uint
	Goal                 uint64
}

type TargetSavingsPlanScreenInformation struct {
}

type LoansScreenInformation struct {
	Balance int64
}

type ThriftScreenInformation struct {
	Plans ThriftBasicPlan
}

type ThriftBasicPlan struct {
	ID              string
	Name            string
	Description     string
	Amount          string
	IsCreator       bool
	NumberOfMembers uint
}

type DBUserBankAccount struct {
	// this is the representation of the account in the DB
	// the views should send the accounts to the frontend through hash values
	ID   string
	Name string
}

type HandlerManager struct {
	partialsManager   IPartialsManager
	store             IStore
	cookieStore       *sessions.CookieStore
	paystackPublicKey string
	paystackSecretKey string
}

type LoginData struct {
	Errors    map[string]string
	csrfField string
}

type VerificationData struct {
	Message string
}

type SoloSaverAddFundsRequestType struct {
	Amount          int64
	Account         int64
	ReferenceNumber uuid.UUID
	// SessionToken string
}

type TransactionInformation struct {
	// TODO: iota is a better representation
	Status          string
	ReferenceNumber string
}

type LoginPostInformation struct {
	UserIsVerified bool
	UserIsAdmin    bool
	Email          string
	ID             uint
}

type RegisterPostInformation struct {
}

type LoanApplicationInformation struct {
}

type GetLoanScreenInformation struct {
	HasValidBVN bool
}

type FamilyVaultInformation struct {
	PlanID uint
}

type PaystackTransactionInformation struct {
	// PaymentOriginator should be an enum type
	// as well as verification status
	CustomerID                uint
	PlanID                    uint
	ReferenceNumber           uuid.UUID
	PaymentOriginator         string
	PaymentAmountInKobo       int64
	FulfillmentStatus         string
	FulfillmentFailureReason  string
	VerificationStatus        string
	VerificationFailureReason string
	CreatedAt                 time.Time
	VerifiedAt                time.Time
}

type SoloSaverPaymentInformation struct {
}

type PaymentInformation struct {
}

type InvestmentApplicationInformation struct {
	
}

type InvestmentsScreenInformation struct {
	Balance uint64
}
