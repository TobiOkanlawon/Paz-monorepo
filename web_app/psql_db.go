package web_app

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	Conn *sql.DB
}

var (
	ErrAccountDoesNotExist = errors.New("account does not exist")
	ErrAccountAlreadyExists = errors.New("account already exists")
	ErrUserNotVerified = errors.New("user's email is not yet verified")
	ErrPasswordIncorrect = errors.New("password incorrect")
	ErrReferenceNumberDoesNotExist = errors.New("this transaction's reference number does not exist in our records")
)

func (d *DB) Connect() {
	// host, status := os.LookupEnv("PAZ_WEB_DB_HOST")
	// if status == true {
	// 	log.Fatal("failed while trying to load db host from env")
	// }
	
	// At some point, I must have thought that I needed the host
	// env variable, but it seems like I might not
	user, status := os.LookupEnv("PAZ_WEB_DB_USER")
	if status != true {
		log.Fatal("failed while trying to load db user from env")
	}
	port, status := os.LookupEnv("PAZ_WEB_DB_PORT")
	if status != true {
		log.Fatal("failed while trying to load db port from env")
	}
	dbName, status := os.LookupEnv("PAZ_WEB_DB_NAME")
	if status != true {
		log.Fatal("failed while trying to load db name from env")
	}
	password, status := os.LookupEnv("PAZ_WEB_DB_PASSWORD")
	if status != true {
		log.Fatal("failed while trying to load db password from env")
	}
	
	connectionStr := fmt.Sprintf("user=%s password=%s dbname=%s port=%s sslmode=disable", user, password, dbName, port)

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
	var loansBalance int64
	var investmentBalance int64

	if err := d.Conn.QueryRow(GetHomeScreenInformationStatement, userID).Scan(
		&firstName,
		&lastName,
		&savingsBalance,
		&loansBalance,
		&investmentBalance,
	); err != nil {
		log.Println(err);
		if err == sql.ErrNoRows {
			return information, fmt.Errorf("error %s returned from query", err.Error())
		}
	}

	information.FirstName = firstName.String
	information.LastName = lastName.String
	information.SavingsBalance = int64(convertToNaira(savingsBalance.Int64))
	information.LoansBalance = int64(convertToNaira(loansBalance))
	information.InvestmentBalance = int64(convertToNaira(investmentBalance))
	// converted back into naira

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

	fmt.Println("Information from profile DB statement", information)

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
		&information.HasPendingPayment,
	); err != nil {
		if err == sql.ErrNoRows {
			return information, fmt.Errorf("error %s returned from query", err.Error())
		}
	}

	information.Balance = uint64(balance.Int64) / 100
	// the balance is converted back to normal naira
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

func (d *DB) RegisterUser(firstName, lastName, email, password string) (RegisterPostInformation, error) {

	var information RegisterPostInformation

	// For now, while we are still sorting out email verification,
	// we will set all emails to be verified on signup

	// to make sure that the statement is correct, we will depend
	// on variables passed from this function, which we will
	// hardcode as verified

	// then after verification has been implemented, we will use
	// the better flow

	email = strings.ToLower(email)

	passwordHash, err := HashPassword(password)
	isVerified := true

	if err != nil {
		return information, err
	}

	_, err = d.Conn.Exec(RegisterUserStatement, firstName, lastName, email, isVerified, passwordHash)

	if err != nil {
		return information, err
	}
	
	return information, nil
}

func (d *DB) CreateLoanApplication(userID uint, amount, termDuration uint64, bvn uint64) (LoanApplicationInformation, error) {
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
	statement, err := d.Conn.Prepare(GetLoanScreenInformationStatement)

	if err != nil {
		return information, err
	}
	
	err = statement.QueryRow(userID).Scan(&information.HasValidBVN)
	
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

func (d *DB) GetPaystackVerificationInformation(referenceNumber string) (PaystackTransactionInformation, error) {
	var information PaystackTransactionInformation

	err := d.Conn.QueryRow(GetPaystackVerificationInformation, referenceNumber).Scan(&information.CustomerID, &information.PlanID, &information.PaymentOriginator)

	if err != nil {
		if err == sql.ErrNoRows {
			return information, ErrReferenceNumberDoesNotExist
		}
		return information, err
	}

	return information, nil
}

func (d *DB) UpdateSoloSaverPaymentFailure(referenceNumber uuid.UUID) (SoloSaverPaymentInformation, error) {
	var information SoloSaverPaymentInformation
	if _, err := d.Conn.Exec(UpdateSoloSaverPaymentFailureStatement, referenceNumber); err != nil {
		return information, err
	}
	return information, nil
}

// For solo savers payments, planID doesn't matter, but you'll still need to provide something, by convention, we can make that 99909990
func (d *DB) CreatePayment(userID, planID uint, referenceNumber uuid.UUID, paymentOriginator string, amountInK int64) (PaymentInformation, error) {
	var information PaymentInformation

	_, err := d.Conn.Exec(CreatePaymentProcessorPendingTransaction, userID, planID, referenceNumber, paymentOriginator, amountInK)

	if err != nil {
		log.Printf("An error occured while trying to create a pending payment %s", err)
		return information, nil
	}
	
	return information, nil
}

// Takes a paystack payment, saves the paystack information then also for this function at least, it updates the user's solo saver account
func (d *DB) UpdateSoloSaverPaymentInformation(amountInK uint64, customerID uint, referenceNumber uuid.UUID) (SoloSaverPaymentInformation, error) {
	var information SoloSaverPaymentInformation

	if _, err := d.Conn.Exec(UpdateSoloSaverPaymentInformationStatement, amountInK, customerID, referenceNumber); err != nil {
		return information, nil
	}
	
	return information, nil
}

func (d *DB) CreateNewFamilyVault(userID uint, familyName, familyMemberEmail string, amount float64, frequency string, duration int64) (FamilyVaultInformation, error) {
	var information FamilyVaultInformation
	description := ""
	balance_in_k := 0
	amount_in_k := amount * 100

	var planID uint

	transformedFrequency, err := convertFrequency(frequency)

	if err != nil {
		return information, err
	}

	statement, err := d.Conn.Prepare(CreateFamilyVaultStatement)
	err = statement.QueryRow(userID, familyName, description, amount_in_k, balance_in_k, duration, transformedFrequency, familyMemberEmail).Scan(&planID)

	if err != nil {
		return information, err
	}

	information.PlanID = planID
	return information, nil
}

func (d *DB) GetInvestmentsScreenInformation(userID uint) (InvestmentsScreenInformation, error) {
	var information InvestmentsScreenInformation
	var balance int64
	err := d.Conn.QueryRow(GetInvestmentsScreenInformationStatement, userID).Scan(&balance)

	if err != nil {
		return information, err
	}

	information.Balance = convertToNaira(balance)

	return information, nil
}

func (d *DB) CreateInvestmentApplication(userID uint, employmentInformation string, yearOfEmployment time.Time, employerName string, investmentAmount uint64, investmentTenure uint64, taxIdentificationNumber uint64, bankAccountName string, bankAccountNumber uint64) (InvestmentApplicationInformation, error) {
	return InvestmentApplicationInformation{}, nil
}

func (d *DB) GetAdminHomeScreenInformation(userID uint) (AdminHomeScreenInformation, error) {
	// TODO: Limit this to only admins
	// that's what the userID is for
	var information AdminHomeScreenInformation

	err := d.Conn.QueryRow(GetAdminHomeScreenInformationStatement).Scan(&information.LoanRequests, &information.InvestmentsRequests, &information.WithdrawalRequests)

	if err != nil {
		return information, err
	}

	return information, nil
}

func convertFrequency(frequency string) (string, error) {
	if frequency == "Daily" {
		return "D", nil
	}
	if frequency == "Weekly" {
		return "W", nil
	}
	if frequency == "Montly" {
		return "M", nil
	}
	if frequency == "Yearly" {
		return "Y", nil
	}
	return "", errors.New(fmt.Sprintf("Unknown frequency specified %s", frequency))
}

func HashPassword(password string) (string, error){
	// the second argument is the cost,
	// which I suppose is how many times the hash is rehashed
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 16)

	if err != nil {
		return "", err
	}
	
	return string(hash), nil
}
