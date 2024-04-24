CREATE TYPE sex_type AS ENUM ('M', 'F');
CREATE TYPE frequency_type AS ENUM ('D', 'W', 'M', 'Y');
CREATE TYPE status_type AS ENUM ('SUCCESSFUL', 'PENDING', 'FAILED');
CREATE TYPE employment_status_type AS ENUM ('SALARIED', 'SELF-EMPLOYED', 'RETIRED', 'UNEMPLOYED');

CREATE TABLE IF NOT EXISTS customer (
       -- all money is stored as kobos which is the minimum denomination of Naira
       customer_id	serial		PRIMARY KEY UNIQUE,
       first_name	  	varchar(32) 	NOT NULL,
       last_name	  	varchar(32) 	NOT NULL,
       email	  	varchar(320)	NOT NULL UNIQUE,
       email_is_verified boolean	NOT NULL DEFAULT FALSE,
       date_joined	timestamp	NOT NULL DEFAULT CURRENT_TIMESTAMP,
       phone_number	varchar(14)	,
       sex		sex_type		,
       date_of_birth	date		,
       postal_address	varchar(128)	
);

CREATE TABLE IF NOT EXISTS password_hash (
       hash_id		   serial 	NOT NULL,
       customer_id	   integer	NOT NULL,
       hash		   varchar(72)	NOT NULL,
       -- handle the length of hashes
       CONSTRAINT password_hash_pk PRIMARY KEY(hash_id)
);

CREATE TABLE IF NOT EXISTS solo_savings_account (
       account_id	   serial	NOT NULL,
       customer_id	   integer	NOT NULL UNIQUE,
       balance_in_k	   integer	NOT NULL CHECK(balance_in_k > 0),
       -- the balance is stored in kobos
       CONSTRAINT solo_savings_account_pk PRIMARY KEY(account_id)
);

-- TODO: create the target savings table

-- CREATE TABLE IF NOT EXISTS family_vault_plan_invitation (
--  -- this table holds invitation records for people that have been invited to family_vault plans but do not have PAZ accounts yet

-- we can implement this when we do emails
 
-- )

CREATE TABLE IF NOT EXISTS family_vault_plan (
       family_vault_plan_id    serial	UNIQUE NOT NULL,
       family_name	       varchar(64) NOT NULL,
       description	       varchar(128) ,
       contribution_amount_in_k		    integer NOT NULL,
       savings_duration_in_d		    integer NOT NULL,
       savings_frequency		    frequency_type	NOT NULL,
       balance_in_k	       integer	NOT NULL CHECK(balance_in_k > 0),
       is_active	       boolean	DEFAULT true,
       creator_id	       integer	NOT NULL,
       CONSTRAINT	       family_vault_plan_fk FOREIGN KEY (creator_id) REFERENCES customer (customer_id)
);

CREATE TABLE IF NOT EXISTS family_vault_plan_member (
       customer_id		      integer	NOT NULL,
       family_vault_plan_id	      integer	NOT NULL,
       date_added		      timestamp	DEFAULT CURRENT_TIMESTAMP,
       CONSTRAINT		      family_vault_plan_member_pk PRIMARY KEY(customer_id, family_vault_plan_id)
       -- CONSTRAINT		      family_vault_plan_member_fk FOREIGN KEY (family_vault_plan_id) REFERENCES family_vault_plan (family_vault_plan_id)

-- TODO: I have removed it because, for some reason, it has refused to work and I cannot spend too much time here
);

CREATE TABLE withdrawal_application (       
       customer_id		    integer	NOT NULL,
       amount_in_k		    integer	NOT NULL CHECK(amount_in_k > 0),
       status			    status_type	NOT NULL DEFAULT 'PENDING',
       failure_reason		    text,
       date_created		    timestamp	DEFAULT CURRENT_TIMESTAMP,
       CONSTRAINT		    withdrawal_application_customer_fk FOREIGN KEY (customer_id) REFERENCES customer (customer_id)
);

CREATE TABLE IF NOT EXISTS target_savings_plan (
       target_savings_plan_id	 serial NOT NULL,
       customer_id		 integer    NOT NULL,
       name		       varchar(32) NOT NULL,
       description	       varchar(128) ,
       balance_in_k	       integer	NOT NULL CHECK(balance_in_k > 0),
       goal_in_k	       integer	NOT NULL,
       CONSTRAINT	       target_savings_plan_pk PRIMARY KEY (target_savings_plan)
);

CREATE TABLE IF NOT EXISTS loans_account (
       account_id	   serial	NOT NULL,
       customer_id 	   integer	UNIQUE NOT NULL,
       amount_owed_in_k	   integer	NOT NULL,
       CONSTRAINT loans_account_pk PRIMARY KEY(account_id)
);

CREATE TABLE IF NOT EXISTS loan_application (
       loans_account_id	      integer	NOT NULL,
       amount_requested_in_k  integer	NOT NULL,
       duration_requested_in_days	integer	 NOT NULL,
       status				status_type NOT NULL DEFAULT 'PENDING',
       CONSTRAINT loan_application_loans_account_fk FOREIGN KEY (loans_account_id) REFERENCES loans_account (account_id)
);

CREATE TABLE IF NOT EXISTS investment_application (
       investment_application_id		  serial	NOT NULL UNIQUE,
       investment_account_id			  integer	NOT NULL,
       employment_status			  employment_status_type	 NOT NULL,
       date_of_employment			  date				 NOT NULL,
       employer_name				  varchar(128)			 NOT NULL,
       tenure					  integer			 NOT NULL,
       -- this tracks whether the admin has accepted the investment request or not
       status					  status_type			 NOT NULL DEFAULT 'PENDING',
       -- tax identification number
       tin    bigint NOT NULL,
       bank_account_name  varchar(64)	NOT NULL,
       bank_account_number		bigint	NOT NULL,
       amount_in_k			bigint NOT NULL DEFAULT 0,
       -- TODO: Remove that default when we fix the migraitions
       CONSTRAINT investment_application_pk PRIMARY KEY(investment_application_id),
       CONSTRAINT investment_account_fk	FOREIGN KEY (investment_account_id) REFERENCES investment_account (account_id)
);     

CREATE TABLE IF NOT EXISTS investment_account (
       customer_id 		integer	NOT NULL,
       account_id		serial 	NOT NULL,
       balance_in_k		integer NOT NULL CHECK(balance_in_k > 0),
       CONSTRAINT investment_account_pk PRIMARY KEY(account_id)
);

CREATE TABLE IF NOT EXISTS thrift_plan (
       thrift_plan_id	 serial		PRIMARY KEY,
       name		 varchar(64)	NOT NULL,
       description	 varchar(72)	,
       amount_per_month_in_k		integer		NOT NULL,
       creator_id			integer		NOT NULL,
       number_of_members		integer		NOT NULL,
       number_of_rounds			integer		NOT NULL,
       current_round			integer		NOT NULL
);

CREATE TABLE IF NOT EXISTS thrift_plan_member (
       customer_id		      integer	NOT NULL,
       thrift_plan_id		      integer	NOT NULL,
       round_assigned		      integer	NOT NULL,
       date_added		      timestamp	NOT NULL
);

CREATE TYPE payment_originator_type AS ENUM ('SOLO_SAVINGS', 'TARGET_SAVINGS', 'FAMILY_SAVINGS', 'LOAN_REPAYMENT', 'INVESTMENTS');

CREATE TABLE payment_processor_transaction (
    payment_id				   serial	PRIMARY KEY,
    customer_id				   integer	NOT NULL,
    plan_id				   integer	NULL,
    -- the plan_id refers to the plan which the payer intends to pay for
    -- this will be used in conjunction witht the payment_originator_type to fulfill the payment
    reference_number 			   UUID UNIQUE NOT NULL,
    payment_originator 			   payment_originator_type NOT NULL,
    payment_amount_in_k 		   integer 		   NOT NULL,
    -- we have this f_st.. field because of potential failures trying to fulfill the payment. We can then use this to implement refunds
    fulfillment_status			   status_type NOT NULL DEFAULT 'PENDING',
    fulfillment_failure_reason		   	       text,
    verification_status 		   status_type NOT NULL DEFAULT 'PENDING',
    verification_failure_reason			       text,
    created_at 				   timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
    verified_at 			   timestamp DEFAULT NULL,
    CONSTRAINT payment_processor_transaction_fk FOREIGN KEY (customer_id) REFERENCES customer (customer_id)
);

CREATE TYPE relationship_type AS ENUM ('sibling', 'spouse', 'parent', 'child', 'guardian');

CREATE TABLE IF NOT EXISTS next_of_kin (
       next_of_kin_id	serial		PRIMARY KEY,
       customer_id	integer		UNIQUE NOT NULL,
       fname	  	varchar(32) 	NOT NULL,
       lname	  	varchar(32) 	NOT NULL,
       email	  	varchar(320)	NOT NULL,
       phone_number	text		NOT NULL,
       kin_relationship	relationship	NOT NULL,
       CONSTRAINT next_of_kin_fk FOREIGN KEY (customer_id) REFERENCES customer (customer_id)
);

-- CREATE TABLE IF NOT EXISTS account_suspension (
--        customer_id		   integer	NOT NULL,
--        suspension_status   boolean	NOT NULL,
-- );

CREATE TABLE IF NOT EXISTS admin_user (
       admin_user_id	serial		PRIMARY KEY UNIQUE,
       customer_id 	integer		UNIQUE NOT NULL,
       CONSTRAINT admin_user_customer_fk FOREIGN KEY (customer_id) REFERENCES customer (customer_id)
);

CREATE TABLE IF NOT EXISTS bvn (
       customer_id	integer		UNIQUE NOT NULL,
       bvn		integer		UNIQUE NOT NULL,
       is_verified	boolean		DEFAULT FALSE NOT NULL,
       CONSTRAINT bvn_customer_fk FOREIGN KEY (customer_id) REFERENCES customer (customer_id)
);
