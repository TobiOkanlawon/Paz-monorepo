package web_app

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/dustin/go-humanize"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

const (
	// TODO: move this to an env file
	host     = "localhost"
	port     = "5432"
	user     = "toby"
	password = "Afternoonglory1"
	dbname   = "pazsimple"
)

type DB struct {
	Conn *sql.DB
}

var (
	ErrAccountDoesNotExist = errors.New("account does not exist")
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrUserNotVerified = errors.New("user's email is not yet verified")
	ErrPasswordIncorrect = errors.New("password incorrect")
)

func HashPassword(password string) (string, error){
	// the second argument is the cost,
	// which I suppose is how many times the hash is rehashed
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 16)

	if err != nil {
		return "", err
	}
	
	return string(hash), nil
}

func (d *DB) Connect() {
	connectionStr := fmt.Sprintf("user=%s password=%s dbname=%s port=%s sslmode=disable", user, password, dbname, port)

	db, err := sql.Open("postgres", connectionStr)

	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	d.Conn = db
}

func (d *DB) GetHomeScreenInformation(userID uint) (HomeScreenInformation, error) {
	var information HomeScreenInformation

	var firstName sql.NullString
	var lastName sql.NullString
	var savingsBalance sql.NullInt64

	if err := d.Conn.QueryRow(GetHomeScreenInformationStatement, userID).Scan(
		&firstName,
		&lastName,
		&savingsBalance,
	); err != nil {
		log.Println(err);
		if err == sql.ErrNoRows {
			return information, fmt.Errorf("error %s returned from query", err.Error())
		}
	}

	information.FirstName = firstName.String
	information.LastName = lastName.String
	information.SavingsBalance = savingsBalance.Int64

	return information, nil
}

func (d *DB) AuthenticateUser(email string, password string) (LoginPostInformation, error) {
	var information LoginPostInformation
	var passwordHash string

	email = strings.ToLower(email)
	
	if err := d.Conn.QueryRow(AuthenticateUserStatement, email).Scan(
		&information.ID,
		&information.Email,
		&information.UserIsAdmin,
		&information.UserIsVerified,
		&passwordHash,
	); err != nil {
		if err == sql.ErrNoRows {
			return information, ErrAccountDoesNotExist
		}
		return information, err
	}

	if !information.UserIsVerified {
		return information, ErrUserNotVerified
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return information, ErrPasswordIncorrect
	}
	
	return information, nil
}

func (d *DB) GetProfileScreenInformation(userID uint) (ProfileScreenInformation, error) {
	var information ProfileScreenInformation

	var firstName sql.NullString
	var lastName sql.NullString
	var email sql.NullString
	var postalAddress sql.NullString
	var phoneNumber sql.NullString
	var sex sql.NullString
	var dateOfBirth sql.NullString
	var nextOfKinFirstName sql.NullString
	var nextOfKinLastName sql.NullString
	var nextOfKinEmail sql.NullString
	var nextOfKinPhoneNumber sql.NullString
	var relationship sql.NullString

	if err := d.Conn.QueryRow(
		GetProfileScreenInformationStatement, userID,
	).Scan(
		&firstName,
		&lastName,
		&postalAddress,
		&email,
		&phoneNumber,
		&sex,
		&dateOfBirth,
		&nextOfKinFirstName,
		&nextOfKinLastName,
		&nextOfKinEmail,
		&nextOfKinPhoneNumber,
		&relationship,
	); err != nil {
		if err == sql.ErrNoRows {
			return information, fmt.Errorf("error %s returned from query", err.Error())
		}
	}

	information.FirstName = firstName.String
	information.LastName = lastName.String
	information.EmailAddress = email.String
	information.PostalAddress = postalAddress.String
	information.Sex = sex.String
	// information.DateOfBirth = time.dateOfBirth.String
	information.NextOfKin.FirstName = nextOfKinFirstName.String
	information.NextOfKin.LastName = nextOfKinLastName.String
	information.NextOfKin.EmailAddress = nextOfKinEmail.String
	information.NextOfKin.PhoneNumber = nextOfKinPhoneNumber.String
	information.NextOfKin.Relationship = relationship.String

	return information, nil
}

func (d *DB) GetSavingsScreenInformation(userID uint) (SavingsScreenInformation, error) {
	var transitoryBalance sql.NullInt64
	var information SavingsScreenInformation

	if err := d.Conn.QueryRow(
		GetSavingsScreenInformationStatement, userID,
	).Scan(
		&transitoryBalance,
	); err != nil {
		return information, err
	}

	// the money data is stored as kobo. Change it to naira here
	if transitoryBalance.Valid {
		information.Balance = convertToNaira(transitoryBalance.Int64)		
	}

	return information, nil
}

func convertToNaira(kobo int64) uint64 {
	if kobo == 0 {
		return 0
	}
	// TODO: there should be proper testing here.
	// there shouldn't be any negative values passed in as kobo
	// this is a domain rule
	return uint64(kobo / 100)
}

func (d *DB) GetFamilyVaultScreenInformation(userID uint) (FamilyVaultScreenInformation, error) {
	var transitoryBalance int64
	var information FamilyVaultScreenInformation
	var plans []FamilyVaultBasicPlan

	rows, err := d.Conn.Query(
		GetFamilyVaultHomeScreenInformationStatement, userID,
	)

	if err != nil {
		return FamilyVaultScreenInformation{}, err
	}

	defer rows.Close();

	for rows.Next() {
		plan := new(FamilyVaultBasicPlan)
		var ownerID uint
		var balance int64

		err = rows.Scan(&plan.ID, &plan.Name, &plan.Description, &balance, &ownerID)

		if err != nil {
			return information, err
		}

		// highly inefficient
		plan.Balance = humanize.Comma(int64(convertToNaira(balance)))
		plan.IsCreator = (ownerID == userID)
		plans = append(plans, *plan)
	}

	information.FamilyVaultBasicPlans = plans
	// the money data is stored as kobo. Change it to naira here
	information.Balance = humanize.Comma(int64(convertToNaira(transitoryBalance)))
	return information, nil
}

func (d *DB) GetFamilyVaultPlanScreenInformation(userID uint, planID int) (FamilyVaultPlanScreenInformation, error){
	var information FamilyVaultPlanScreenInformation
	var description sql.NullString
	var creatorID int
	var balance int64

	if err := d.Conn.QueryRow(GetFamilyVaultPlanScreenInformationStatement, planID, userID).Scan(
		&information.ID,
		&information.Name,
		&description,
		&balance,
		&creatorID,
	); err != nil {
		if err == sql.ErrNoRows {
			return information, fmt.Errorf("error %s returned from query", err.Error())
		}
	}

	information.Description = description.String
	information.Balance = humanize.Comma(int64(convertToNaira(balance)))
	return information, nil
}


func (d *DB) GetSoloSaverScreenInformation(userID uint) (SoloSaverScreenInformation, error){
	var information SoloSaverScreenInformation
	var balance sql.NullInt64
	var email sql.NullString

	if err := d.Conn.QueryRow(GetSoloSaverScreenInformationStatement, userID).Scan(
		&balance,
		&email,
	); err != nil {
		if err == sql.ErrNoRows {
			return information, fmt.Errorf("error %s returned from query", err.Error())
		}
	}

	information.Balance = uint64(balance.Int64)
	information.EmailAddress = email.String
	return information, nil
}

func (d *DB) GetTargetSavingsPlanScreenInformation(userID uint, planID int) (TargetSavingsPlanScreenInformation, error) {
	return TargetSavingsPlanScreenInformation{}, nil
}

func (d *DB) GetTargetSavingsScreenInformation(userID uint) (TargetSavingsScreenInformation, error) {
	var information TargetSavingsScreenInformation
	var balance sql.NullInt64
	var plans []TargetSavingsBasicPlan
	
	rows, err := d.Conn.Query(GetTargetSavingsScreenInformationStatement, userID)
	if err == sql.ErrNoRows {
		return information, err
	}

	defer rows.Close();

	for rows.Next() {
		plan := new(TargetSavingsBasicPlan)

		err = rows.Scan(&plan.ID, &plan.Name, &plan.Description, &plan.Balance, &plan.Goal)

		balance.Int64 += int64(plan.Balance)

		if err != nil {
			return information, err
		}
		plan.CompletionPercentage = uint((plan.Balance/plan.Goal) * 100)
		plans = append(plans, *plan)
	}

	information.Plans = plans
	// the money data is stored as kobo. Change it to naira here
	return information, nil
}

func (d *DB) GetLoansScreenInformation(userID uint) (LoansScreenInformation, error) {
	var information LoansScreenInformation
	var balance sql.NullInt64

	if err := d.Conn.QueryRow(GetLoansScreenInformationStatement, userID).Scan(
		&balance,
	); err != nil {
		if err == sql.ErrNoRows {
			return information, fmt.Errorf("error %s returned from query", err.Error())
		}
	}

	information.Balance = balance.Int64
	return information, nil
}

func (d *DB) GetThriftScreenInformation(userID uint) (ThriftScreenInformation, error) {
	return ThriftScreenInformation{}, nil
}

func (d *DB) CreateSoloSavingsPendingTransaction(userID uint, amount int64, refNo string) (TransactionInformation, error) {
	var information TransactionInformation
	_, err := d.Conn.Exec(CreateSoloSavingsPendingTransactionStatement, userID, amount, refNo)

	if err != nil {
		return information, fmt.Errorf("Error while executing statement: %s", err)
	}

	// Get the new album's generated ID for the client.
	return information, nil	
}

func (d *DB) RegisterUser(firstName, lastName, email, password string) (RegisterPostInformation, error) {

	var information RegisterPostInformation

	// For now, while we are still sorting out email verification,
	// we will set all emails to be verified on signup

	// to make sure that the statement is correct, we will depend
	// on variables passed from this function, which we will
	// hardcode as verified

	// then after verification has been implemented, we will use
	// the better flow

	// TODO: test that this works properly, i.e it doesn't scramble the @ sign or possible full stops
	email = strings.ToLower(email)

	passwordHash, err := HashPassword(password)
	isVerified := true

	if err != nil {
		return RegisterPostInformation{}, err
	}

	_, err = d.Conn.Exec(RegisterUserStatement, firstName, lastName, email, isVerified, passwordHash)

	if err != nil {
		return information, err
	}
	
	return RegisterPostInformation{}, nil
}

func (d *DB) CreateLoanApplication(userID uint, amount, termDuration uint64) (LoanApplicationInformation, error) {
	var information LoanApplicationInformation
	amount_in_k := amount * 100
	_, err := d.Conn.Exec(CreateLoanApplicationStatement, userID, amount_in_k, termDuration)

	if err != nil {
		return information, fmt.Errorf("Error while executing statement: %s", err)
	}

	return information, nil	
}

func (d *DB) GetLoanScreenInformation(userID uint) (GetLoanScreenInformation, error) {
	// TODO: the name of this function is too alike to another. The other function breaks the convention of this entire interface. FIX IT.

	var information GetLoanScreenInformation
	_, err := d.Conn.Exec(GetLoanScreenInformationStatement, userID)
	
	// TODO: after BVN validation step, this should change

	if err != nil {
		if err == sql.ErrNoRows {
			information.HasValidBVN = false
			return information, nil
		}
		return information, err
	}

	information.HasValidBVN = true
	return information, nil
}
