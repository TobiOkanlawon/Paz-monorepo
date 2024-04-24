package web_app

const GetHomeScreenInformationStatement = `SELECT customer.first_name, customer.last_name, solo_savings_account.balance_in_k, loans_account.amount_owed_in_k, investment_account.balance_in_k FROM customer, solo_savings_account, loans_account, investment_account WHERE customer.customer_id = $1;
`
const GetProfileScreenInformationStatement = `SELECT customer.first_name, customer.last_name, customer.postal_address, customer.email, customer.phone_number, customer.sex, customer.date_of_birth, next_of_kin.first_name, next_of_kin.last_name, next_of_kin.email, next_of_kin.phone_number, next_of_kin.kin_relationship FROM customer JOIN next_of_kin ON customer.customer_id = next_of_kin.customer_id WHERE customer.customer_id = $1;`

const GetSavingsScreenInformationStatement = `SELECT solo_savings_account.balance_in_k FROM solo_savings_account WHERE solo_savings_account.customer_id = $1;`

const GetFamilyVaultHomeScreenInformationStatement = `SELECT family_vault_plan.family_vault_plan_id, family_vault_plan.family_name, family_vault_plan.description, family_vault_plan.balance_in_k, family_vault_plan.creator_id FROM family_vault_plan, family_vault_plan_member WHERE family_vault_plan_member.customer_id = $1;`

const GetFamilyVaultPlanScreenInformationStatement = `SELECT family_vault_plan.family_vault_plan_id, family_vault_plan.family_name, family_vault_plan.description, family_vault_plan.balance_in_k, family_vault_plan.creator_id FROM family_vault_plan, family_vault_plan_member WHERE family_vault_plan.family_vault_plan_id = $1 AND family_vault_plan_member.customer_id = $2;`

const GetSoloSaverScreenInformationStatement = `SELECT ssa.balance_in_k,
       c.email,
       CASE
           WHEN EXISTS (
               SELECT 1
               FROM payment_processor_transaction AS p
               WHERE p.customer_id = $1
                 AND p.verification_status = 'PENDING'
           )
           THEN TRUE
           ELSE FALSE
       END AS has_pending_payment
FROM solo_savings_account ssa
JOIN customer c ON ssa.customer_id = c.customer_id
WHERE ssa.customer_id = $1;`

const GetTargetSavingsScreenInformationStatement = `SELECT tsp.target_savings_plan_id, tsp.name, tsp.description, tsp.balance_in_k, tsp.goal_in_k FROM target_savings_plan AS tsp where tsp.customer_id = $1;`

const GetLoansScreenInformationStatement = `SELECT amount_owed_in_k FROM loans_account WHERE customer_id = $1;`

const CreatePaymentProcessorPendingTransaction = `INSERT INTO payment_processor_transaction (customer_id, plan_id, reference_number, payment_originator, payment_amount_in_k) VALUES ($1, $2, $3, $4, $5);`

const GetPaystackVerificationInformation = `SELECT customer_id, plan_id, payment_originator FROM payment_processor_transaction WHERE reference_number = $1;`

const AuthenticateUserStatement = `SELECT c.customer_id,
       c.email,
       CASE WHEN a.customer_id IS NOT NULL THEN TRUE ELSE FALSE END AS is_admin,
       c.email_is_verified,
       ph.hash
FROM customer c
JOIN password_hash ph ON c.customer_id = ph.customer_id
LEFT JOIN admin_user a ON c.customer_id = a.customer_id
WHERE c.email = $1;`

const RegisterUserStatement = `WITH new_customer AS (
    INSERT INTO customer (first_name, last_name, email, email_is_verified)
    VALUES ($1, $2, $3, $4)
    RETURNING customer_id
),
solo_savings_insert AS (
    INSERT INTO solo_savings_account (customer_id, balance_in_k)
    SELECT customer_id, 0 FROM new_customer
),
loans_account_insert AS (
    INSERT INTO loans_account (customer_id, amount_owed_in_k)
    SELECT customer_id, 0 FROM new_customer
),
investment_account_insert AS (
    INSERT INTO investment_account (customer_id, balance_in_k)
    SELECT customer_id, 0 FROM new_customer
)
INSERT INTO password_hash (customer_id, hash)
SELECT customer_id, $5
FROM new_customer;
`

const CreateLoanApplicationStatement = `INSERT INTO loan_application (loans_account_id, amount_requested_in_k, duration_requested_in_days) (SELECT account_id, $2, $3 FROM loans_account WHERE customer_id = $1);
`

// do the one for the loans and investments accounts too

// const GetThriftScreenInformationStatement = `SELECT thrift_id, `

// TODO: a better implementation is to count the rows returned.
const GetLoanScreenInformationStatement = `SELECT bvn FROM bvn WHERE customer_id = $1;`

const CreateFamilyVaultStatement = `WITH new_family_vault AS (
    INSERT INTO family_vault_plan (creator_id, family_name, description, contribution_amount_in_k, balance_in_k, savings_duration_in_d, savings_frequency)
    VALUES ($1, $2, $3, $4, $5, $6, $7)
    RETURNING family_vault_plan_id
),
family_vault_plan_creator_insert AS (
    INSERT INTO family_vault_plan_member (customer_id, family_vault_plan_id)
    SELECT $1, family_vault_plan_id FROM new_family_vault
)
INSERT INTO family_vault_plan_member (customer_id, family_vault_plan_id)
SELECT c.customer_id, nfv.family_vault_plan_id
FROM new_family_vault nfv
JOIN customer c ON c.email = $8
RETURNING nfv.family_vault_plan_id;`

const UpdateSoloSaverPaymentInformationStatement = `WITH transaction_update AS (
  UPDATE payment_processor_transaction
  SET verification_status = 'SUCCESSFUL',
  fulfillment_status = 'SUCCESSFUL',
  payment_amount_in_k = $1
  WHERE reference_number = $2
  AND verification_status = 'PENDING'
  RETURNING payment_amount_in_k, customer_id
)

-- The pending check makes sure that we aren't updating a previously successful payment (that could happen in a replay attack)

UPDATE solo_savings_account
SET balance_in_k = balance_in_k + transaction_update.payment_amount_in_k
FROM transaction_update WHERE solo_savings_account.customer_id = transaction_update.customer_id;`

const UpdateSoloSaverPaymentFailureStatement = `UPDATE payment_processor_transaction SET verification_status = 'FAILED', fulfillment_status = 'FAILED' WHERE reference_number = $1 AND verification_status = 'PENDING';`

const GetInvestmentsScreenInformationStatement = `SELECT balance_in_k FROM investment_account WHERE customer_id = $1;`

const GetAdminHomeScreenInformationStatement = `SELECT count(loan_application.status = 'PENDING') AS ls, count(investment_application.status = 'PENDING') AS inv, count(withdrawal_application.status = 'PENDING') AS wa FROM loan_application, investment_application, withdrawal_application;
`

const CreateInvestmentApplicationStatement = `WITH check_pending_applications AS (
    SELECT count(investment_account.account_id) as pending
    FROM investment_application
    JOIN investment_account ON investment_application.investment_account_id = investment_account.account_id
    WHERE investment_account.customer_id = $1
    AND status = 'PENDING'
), get_customer_account_id AS (
   SELECT investment_account.account_id AS account_id FROM investment_account, investment_application WHERE investment_application.investment_account_id = investment_account.account_id AND investment_account.customer_id = $1 LIMIT 1
)
INSERT INTO investment_application (investment_account_id, employment_status, date_of_employment, employer_name, tenure, tin, bank_account_name, bank_account_number, amount_in_k)
SELECT account_id, $2, $3, $4, $5, $6, $7, $8, $9
FROM get_customer_account_id WHERE (SELECT pending FROM check_pending_applications) = 0;`
