ALTER TABLE "users"
    ADD COLUMN "sign_up_method" VARCHAR;
UPDATE "users"
SET "sign_up_method" = 'standard'
WHERE "sign_up_method" IS NULL OR "sign_up_method" = '';