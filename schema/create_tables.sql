CREATE TYPE sex_type AS ENUM ('M', 'F');

CREATE TABLE customer (
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

CREATE TABLE password_hash (
       hash_id		   serial 	NOT NULL,
       customer_id	   integer	NOT NULL,
       hash		   varchar(72)	NOT NULL,
       -- handle the length of hashes
       CONSTRAINT password_hash_pk PRIMARY KEY(hash_id)
);

CREATE TABLE solo_savings_account (
       account_id	   serial	NOT NULL,
       customer_id	   integer	NOT NULL UNIQUE,
       balance_in_k	   integer	NOT NULL,
       -- the balance is stored in kobos
       CONSTRAINT solo_savings_account_pk PRIMARY KEY(account_id)
);

-- TODO: create the target savings table
-- TODO: create the family vault table

CREATE TABLE family_vault_plan_member (
       customer_id		      integer	NOT NULL,
       family_vault_plan_id	      integer	NOT NULL,
       date_added		      timestamp	NOT NULL,
       CONSTRAINT		      family_vault_plan_member_pk PRIMARY KEY(customer_id, family_vault_plan_id)
);

CREATE TABLE family_vault_plan_transaction (
       transaction_id			    serial	NOT NULL PRIMARY KEY,
       payment_processor_ref_no		    integer	NOT NULL UNIQUE,
       amount_transacted_in_k		    integer	NOT NULL
);

CREATE TABLE family_vault_plan (
       family_vault_plan_id    serial	NOT NULL,
       name		       varchar(32) NOT NULL,
       description	       varchar(128) ,
       balance_in_k	       integer	NOT NULL,
       is_active	       boolean	DEFAULT true,
       creator_id	       integer	NOT NULL
);

CREATE TABLE target_savings_plan (
       target_savings_plan_id	 serial NOT NULL,
       customer_id		 integer    NOT NULL,
       name		       varchar(32) NOT NULL,
       description	       varchar(128) ,
       balance_in_k	       integer	NOT NULL,
       goal_in_k	       integer	NOT NULL
);

CREATE TABLE target_savings_plan_transaction (
       transaction_id			    serial	NOT NULL PRIMARY KEY,
       payment_processor_ref_no		    integer	NOT NULL UNIQUE,
       amount_transacted_in_k		    integer	NOT NULL
);

CREATE TABLE loans_account (
       account_id	   serial	NOT NULL,
       customer_id 	   integer	UNIQUE NOT NULL,
       amount_owed_in_k	   integer	NOT NULL,
       CONSTRAINT loans_account_pk PRIMARY KEY(account_id)
);

CREATE TABLE loan_application (
       loans_account_id	      integer	NOT NULL,
       customer_id	      integer	NOT NULL,
       amount_requested_in_k  integer	NOT NULL,
       duration_requested_in_days	integer	 NOT NULL,
       CONSTRAINT loan_application_customer_fk FOREIGN KEY (customer_id) REFERENCES customer (customer_id),
       CONSTRAINT loan_application_loans_account_fk FOREIGN KEY (loans_account_id) REFERENCES loans_account (account_id)
);

CREATE TABLE investment_account (
       customer_id 		integer	NOT NULL,
       account_id		serial 	NOT NULL,
       balance_in_k		integer NOT NULL,
       CONSTRAINT investment_account_pk PRIMARY KEY(account_id)
);

CREATE TABLE thrift_plan (
       thrift_plan_id	 serial		PRIMARY KEY,
       name		 varchar(64)	NOT NULL,
       description	 varchar(72)	,
       amount_per_month_in_k		integer		NOT NULL,
       creator_id			integer		NOT NULL,
       number_of_members		integer		NOT NULL,
       number_of_rounds			integer		NOT NULL,
       current_round			integer		NOT NULL
);

CREATE TABLE thrift_transaction (
       thrift_plan_id		integer		NOT NULL,
       thrift_plan_round_id	integer 	NOT NULL,
       transaction_amount_in_k	 integer	NOT NULL,
       date_transacted		timestamp	NOT NULL,
       CONSTRAINT thrift_transaction_pk PRIMARY KEY(thrift_plan_id, thrift_plan_round_id)
);

CREATE TABLE thrift_plan_member (
       customer_id		      integer	NOT NULL,
       thrift_plan_id		      integer	NOT NULL,
       round_assigned		      integer	NOT NULL,
       date_added		      timestamp	NOT NULL
);

CREATE TABLE solo_savings_transaction (
       account_id		 integer	NOT NULL,
       -- the transaction amount should be depicted negative
       -- for debits and positive for credits
       transaction_amount_in_k	 integer	NOT NULL,
       date_transacted		 timestamp	NOT NULL       
);

CREATE TYPE payment_processor_transaction_status_type AS ENUM ('SUCCESSFUL', 'PENDING', 'FAILED');

CREATE TABLE payment_processor_transaction (
       -- this is to track paystack and monify transactions
       status  payment_processor_transaction_status_type NOT NULL,
       -- reference number maybe should not be varchar, or text, but char of length uuid
       reference_number				    text   NOT NULL
);

CREATE TYPE relationship_type AS ENUM ('sibling', 'spouse', 'parent', 'child', 'guardian');

CREATE TABLE next_of_kin (
       next_of_kin_id	serial		PRIMARY KEY,
       customer_id	integer		UNIQUE NOT NULL,
       fname	  	varchar(32) 	NOT NULL,
       lname	  	varchar(32) 	NOT NULL,
       email	  	varchar(320)	NOT NULL,
       phone_number	text		NOT NULL,
       kin_relationship	relationship	NOT NULL
);

-- CREATE TABLE account_suspension (
--        customer_id		   integer	NOT NULL,
--        suspension_status   boolean	NOT NULL,
-- );

CREATE TABLE admin_user (
       admin_user_id	serial		PRIMARY KEY UNIQUE,
       customer_id 	integer		UNIQUE NOT NULL,
       CONSTRAINT admin_user_customer_fk FOREIGN KEY (customer_id) REFERENCES customer (customer_id)
);

CREATE TABLE bvn (
       customer_id	integer		UNIQUE NOT NULL,
       bvn		integer		UNIQUE NOT NULL,
       is_verified	boolean		DEFAULT FALSE NOT NULL,
       CONSTRAINT bvn_customer_fk FOREIGN KEY (customer_id) REFERENCES customer (customer_id)
);
