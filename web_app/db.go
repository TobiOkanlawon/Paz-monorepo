package web_app

import (
	"gorm.io/gorm"
	"time"
)

type IStore interface {
	AuthenticateUser(email string, password string) (*User, error)
	RegisterUser(firstName, lastName, email, password string) (*User, error)
	GetAccountInformation(userId uint) (Account, error)
	GetSavingsAmount(userId uint) (float64, error)
	GetLoansInformation(userId uint) (LoansAccount, error)
	GetInvestmentsAmount(userId uint) (float64, error)
	GetActivities(userID uint) ([]Activity, error)
	GetUserInformation(userID uint) (User, error)
	GetKinDetails(userID uint) (Kin, error)
	CreateNewThrift(userID uint, title string, description string, contributionAmount float64, numbersOfMembers int, frequency string) (ThriftGroup, error)
}

// NOTE: should I make DBUsers include Admin users
// TODO I should add three states for email activation
// Not activated, pending, and activated
// TODO: make sex an enum type
type User struct {
	gorm.Model
	FirstName        string
	LastName         string
	EmailAddress     string `gorm:"unique"`
	PhoneNumber      int64
	ProfileImage     string
	DateOfBirth      string
	Sex              string
	EmailActivated   bool
	EmailActivatedAt *time.Time
	Activity         []*Activity
	PasswordHash     PasswordHash
	Account          Account
	KinID            uint
	BVN              int64
}

// TODO: Make relationship an enum type
type Kin struct {
	gorm.Model
	FirstName    string
	LastName     string
	EmailAddress string `gorm:"unique"`
	PhoneNumber  int64
	Relationship string
	Users        []User
}

type Account struct {
	gorm.Model
	SavingsAmount     float64
	LoansAmount       float64
	InvestmentsAmount float64
	UserID            uint
}

type ThriftAccount struct {
	gorm.Model
	Balance float64
}

type LoansAccount struct {
	Balance float64
}

type Loan struct {
}

// We will probably need to remove the accountToDomicileLoan field so that we can use the user's verified account.
// ApplicationStatus is an ENUM type between 'PENDING', 'FAILURE', 'SUCCESS'
type LoanApplication struct {
	LoanAmount            float64
	LoanPurpose           string
	LoanDuration          int32
	UserID                uint
	AccountToDomicileLoan int64
	ApplicationStatus     string
}

type PasswordHash struct {
	gorm.Model
	Hash   string
	UserID uint
}

// TODO: how do we model all the Activity.
// Each section of the app has its own activity, then the home page summarises all the activities on the app in all the sections. It is also time-sorted (FIFO)
// How do we model this in the DB
// The ProcessOwner is the process that generates the activity.
// The ProcessOwner is how we filter the activities when we want to use one on a particular screen
type Activity struct {
	gorm.Model
	UserID       uint
	Text         string
	Time         *time.Time
	ProcessOwner string
}

type ThriftGroup struct {
	gorm.Model
	Name string
	Description string
	DefaultingMembers []User
	// positive integer. Should be >= 2
	Rounds int
	// Creator doubles as admin under this assumed permissions model
	Creator User
	CreationDate *time.Time
	// endingDate is the date when the thrift group is supposed to end. This is estimated based on the rounds and the frequency.
	// We'll be skipping this field for now
	// EndingDate *time.Time
	IsActive bool
	Account ThriftAccount
	Members []User	
}
