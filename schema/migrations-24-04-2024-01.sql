ALTER TABLE investment_application ALTER COLUMN bank_account_number TYPE BIGINT;
ALTER TABLE investment_application ALTER COLUMN tin TYPE BIGINT;
BEGIN;
ALTER TABLE investment_application DROP CONSTRAINT investment_application_pk;
ALTER TABLE investment_application ADD CONSTRAINT investment_application_pk PRIMARY KEY(investment_application_id);
COMMIT;
ALTER TABLE investment_application ADD amount_in_k bigint NOT NULL DEFAULT 0;
-- TODO: I'll remove the default on the next push, so that there won't be that error;
-- The default is just an expedient to help migrations
