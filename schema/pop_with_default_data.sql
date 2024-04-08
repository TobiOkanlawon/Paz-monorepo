INSERT INTO customer (first_name, last_name, email, email_is_verified, date_joined) VALUES ('Tester', 'Xavi', 'tester@xavi.com', true, Now());
INSERT INTO next_of_kin (customer_id, first_name, last_name, email, phone_number, kin_relationship) VALUES ('1', 'Tolu', 'Okanlawon', 'tolu@okanlawon.com', '08024584851', 'sibling');
INSERT INTO solo_savings_account(customer_id, balance_in_k) VALUES ('1', '000');
INSERT INTO family_vault_plan(name, description, balance_in_k, creator_id) VALUES ('Olowo Family', 'Monthly food savings quota', '0', '1');
INSERT INTO family_vault_plan_member(customer_id, family_vault_plan_id, date_added) VALUES ('1', '2', Now());

-- second customer

INSERT INTO customer (first_name, last_name, email, email_is_verified, date_joined) VALUES ('Tester', 'Alonso', 'tester@alonso.com', true, Now());
INSERT INTO admin (customer_id) VALUES(1);
