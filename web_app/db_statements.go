package web_app

const GetHomeScreenInformationStatement = `SELECT customer.first_name, customer.last_name, solo_savings_account.balance_in_k AS "Savings Balance" FROM customer, solo_savings_account WHERE customer.customer_id = $1;
`
const GetProfileScreenInformationStatement = `SELECT customer.first_name, customer.last_name, customer.postal_address, customer.email, customer.phone_number, customer.sex, customer.date_of_birth, next_of_kin.first_name, next_of_kin.last_name, next_of_kin.email, next_of_kin.phone_number, next_of_kin.kin_relationship FROM customer JOIN next_of_kin ON customer.customer_id = next_of_kin.customer_id WHERE customer.customer_id = $1;`

const GetSavingsScreenInformationStatement = `SELECT solo_savings_account.balance_in_k FROM solo_savings_account WHERE solo_savings_account.customer_id = $1;`

const GetFamilyVaultHomeScreenInformationStatement = `SELECT family_vault_plan.family_vault_plan_id, family_vault_plan.name, family_vault_plan.description, family_vault_plan.balance_in_k, family_vault_plan.creator_id FROM family_vault_plan, family_vault_plan_member WHERE family_vault_plan_member.customer_id = $1;`

const GetFamilyVaultPlanScreenInformationStatement = `SELECT family_vault_plan.family_vault_plan_id, family_vault_plan.name, family_vault_plan.description, family_vault_plan.balance_in_k, family_vault_plan.creator_id FROM family_vault_plan, family_vault_plan_member WHERE family_vault_plan.family_vault_plan_id = $1 AND family_vault_plan_member.customer_id = $2;`

const GetSoloSaverScreenInformationStatement = `SELECT solo_savings_account.balance_in_k, customer.email FROM solo_savings_account, customer WHERE solo_savings_account.customer_id = $1;`;

const GetTargetSavingsScreenInformationStatement = `SELECT tsp.target_savings_plan_id, tsp.name, tsp.description, tsp.balance_in_k, tsp.goal_in_k FROM target_savings_plan AS tsp where tsp.customer_id = $1;`

const GetLoansScreenInformationStatement = `SELECT amount_owed_in_k FROM loans_account WHERE customer_id = $1;`;

const CreateSoloSavingsPendingTransactionStatement = ``;

const AuthenticateUserStatement = `SELECT customer.customer_id, customer.email, CAST((admin.admin_id = customer.customer_id) AS BOOLEAN), customer.email_is_verified, password_hash.hash FROM customer, admin, password_hash WHERE customer.email = $1 AND customer.customer_id = password_hash.customer_id;`

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
`;

const CreateLoanApplicationStatement = `INSERT INTO loan_application (loans_account_id, customer_id, amount_requested_in_k, duration_requested_in_days) VALUES (SELECT $1, $2, $3, $4);`
// do the one for the loans and investments accounts too

// const GetThriftScreenInformationStatement = `SELECT thrift_id, `

// TODO: a better implementation is to count the rows returned.
// 
const GetLoanScreenInformationStatement = `SELECT bvn FROM bvn WHERE customer_id = $1;`
