package web_app

import (
	"errors"
	"log"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SqlStore struct {
	db *gorm.DB
}

func NewSqlStore(path string) (*SqlStore, error) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		panic("could not connect to the DB")
	}

	// Migrate the schema
	db.AutoMigrate(&User{}, &Activity{}, &PasswordHash{}, &Account{}, &Kin{})
	// Creating users with the right flow
	// user, err := RegisterUser("Tobi", "Okanlawon", "tobiinlondon34@gmail.com", "passphrase")

	if err != nil {
		panic(err)
	}

	return &SqlStore{
		db: db,
	}, nil
}

func (s *SqlStore) AuthenticateUser(email string, password string) (*User, error) {
	var user User
	var hash PasswordHash
	tx := s.db.First(&user, "email_address = ?", email)

	if tx.Error != nil {
		// TODO: create standard errors
		return nil, errors.New("Email doesn't exist")
	}

	// TODO: convert this to a preload GORM function
	tx = s.db.First(&hash, "user_id = ?", user.ID)

	if tx.Error != nil {
		// TODO: create standard errors
		return nil, errors.New("Error getting password from db ")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash.Hash), []byte(password))
	if err != nil {
		log.Println("doesn't match")
		return &User{}, err
	}

	return &user, nil
}

func (s *SqlStore) GetAccountInformation(userId uint) (Account, error) {
	var account Account
	tx := s.db.First(&account)

	if tx.Error != nil {
		// TODO: convert to custom error
		return account, tx.Error
	}

	return account, nil
}

func (s *SqlStore) GetInvestmentsAmount(userId uint) (float64, error) {
	var account Account
	tx := s.db.First(&account, "user_id = ?", userId)

	if tx.Error != nil {
		// TODO: convert to custom error
		return account.InvestmentsAmount, tx.Error
	}

	return account.InvestmentsAmount, nil
}

func (s *SqlStore) GetSavingsAmount(userId uint) (float64, error) {
	var account Account
	tx := s.db.First(&account, "user_id = ?", userId)

	if tx.Error != nil {
		// TODO: convert to custom error
		return account.SavingsAmount, tx.Error
	}

	return account.SavingsAmount, nil
}

func (s *SqlStore) GetLoansAmount(userId uint) (float64, error) {
	var account Account
	tx := s.db.First(&account, "user_id = ?", userId)

	if tx.Error != nil {
		// TODO: convert to custom error
		return account.LoansAmount, tx.Error
	}

	return account.LoansAmount, nil
}

func (s *SqlStore) GetActivities(userID uint) ([]Activity, error) {
	var activities []Activity

	tx := s.db.Find(&activities, "user_id = ?", userID)

	if tx.Error != nil {
		return activities, tx.Error
	}

	return activities, nil
}

func (s *SqlStore) RegisterUser(firstName, lastName, email, password string) (*User, error) {
	// TODO: change the errors to specific ones
	// TODO: check if the user already exists

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return &User{}, nil
	}

	hash := PasswordHash{
		Hash: string(passwordHash),
	}

	activity := Activity{
		Text: "Created account",
	}

	user := User{
		FirstName:      firstName,
		LastName:       lastName,
		EmailAddress:   email,
		EmailActivated: false,
		PasswordHash:   hash,
		Account:        Account{},
		Activity:       []*Activity{&activity},
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *SqlStore) GetKinDetails(userID uint) (Kin, error){
	// TODO: implement this
	return Kin{}, nil
}

func (s *SqlStore) GetUserInformation(userID uint) (User, error) {
	var user User
	tx := s.db.Find(&user, "id = ?", userID)

	if tx.Error != nil {
		return User{}, tx.Error
	}

	return user, nil
}

func (s *SqlStore) GetLoansInformation(userID uint) (LoansAccount, error) {
	return LoansAccount{}, nil
}

func (s *SqlStore) CreateNewThrift(userID uint, title string, description string, contributionAmount float64, numbersOfMembers int, frequency string) (ThriftGroup, error) {
	return ThriftGroup{}, nil
}
