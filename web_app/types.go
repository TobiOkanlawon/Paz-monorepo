package web_app

import (
	"github.com/gorilla/sessions"
	"time"
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
	CreateLoanApplication(userID uint, amount uint64, termDuration uint64) (LoanApplicationInformation, error)
	GetThriftScreenInformation(userID uint) (ThriftScreenInformation, error)
	CreateSoloSavingsPendingTransaction(userID uint, amount int64, refNo string) (TransactionInformation, error)
	GetLoanScreenInformation(userID uint) (GetLoanScreenInformation, error)
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
	Balance      uint64
	Accounts     []DBUserBankAccount
	EmailAddress string
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
	ReferenceNumber string
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
