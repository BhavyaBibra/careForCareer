ALTER TABLE candidate_target_companies DROP CONSTRAINT IF EXISTS fk_target_company;
ALTER TABLE job_descriptions DROP CONSTRAINT IF EXISTS fk_jd_company;
DROP TABLE IF EXISTS company_patterns CASCADE;
DROP TABLE IF EXISTS companies CASCADE;
DROP TYPE IF EXISTS company_tier;
