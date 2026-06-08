-- Seed: India company knowledge base
-- Run via: psql $DATABASE_URL -f seed/companies.sql

INSERT INTO companies (id, name, tier, india_bar_notes, website) VALUES
  (uuid_generate_v4(), 'Google', 'faang',
   'BLR SWE L3: LC Hard 3/5, system design mandatory from L4+. Bar raiser adds ~30% rejection rate. Hiring bar lower than MTV for same level.',
   'google.com'),

  (uuid_generate_v4(), 'Microsoft', 'faang',
   'Hyderabad SWE SDE2: LC Medium 4/5, system design and OOD. More lenient on DSA than Google. Strong culture-fit component.',
   'microsoft.com'),

  (uuid_generate_v4(), 'Amazon', 'faang',
   'Bangalore SDE2: LP principles (14 principles, heavy focus). DSA LC Medium-Hard. Bar raiser present. Coding bar slightly lower than Google.',
   'amazon.com'),

  (uuid_generate_v4(), 'Meta', 'faang',
   'Bangalore SWE E4: LC Hard expected. System design round from E4+. Coding assessed on optimality, not just correctness.',
   'meta.com'),

  (uuid_generate_v4(), 'Atlassian', 'global_product',
   'Bangalore: System design heavy, domain depth expected (dev tools). LC Medium sufficient. Strong values alignment check.',
   'atlassian.com'),

  (uuid_generate_v4(), 'Uber', 'global_product',
   'Bangalore: Backend systems focus, distributed systems depth. LC Medium-Hard. Practical coding over algorithms.',
   'uber.com'),

  (uuid_generate_v4(), 'Stripe', 'global_product',
   'Remote/Bangalore: API design, backend depth, payments domain a plus. LC Medium. Very high bar on code quality.',
   'stripe.com'),

  (uuid_generate_v4(), 'Razorpay', 'unicorn',
   'Bangalore: Practical backend, payments infra. LC Easy-Medium. Strong ownership culture. fintech domain signals valued.',
   'razorpay.com'),

  (uuid_generate_v4(), 'PhonePe', 'unicorn',
   'Bangalore: High-scale payments, Java/Go backend. LC Medium. System design at senior level. UPI/payments depth is differentiator.',
   'phonepe.com'),

  (uuid_generate_v4(), 'CRED', 'unicorn',
   'Bangalore: product-engineering culture, high bar on problem-solving. LC Medium-Hard. Full-stack awareness expected.',
   'cred.club'),

  (uuid_generate_v4(), 'Meesho', 'unicorn',
   'Bangalore: Scale (ecommerce), data pipelines. LC Medium. Practical backend, fast iterations.',
   'meesho.com'),

  (uuid_generate_v4(), 'Zepto', 'unicorn',
   'Mumbai/Bangalore: q-commerce, fast-paced. LC Medium. Backend + data focus. Ownership and speed valued.',
   'zeptonow.com'),

  (uuid_generate_v4(), 'Groww', 'unicorn',
   'Bangalore: Fintech, scale, reliability. LC Medium. Java/Go backend. Payments and trading infra depth is plus.',
   'groww.in'),

  (uuid_generate_v4(), 'Infosys', 'service',
   'Campus/lateral: Basic DSA, OOPS, SQL. LC Easy. Strong for fresh grads and <2 YOE.',
   'infosys.com'),

  (uuid_generate_v4(), 'TCS', 'service',
   'Campus: NQT assessment, basic reasoning + coding. Good entry point for 0-1 YOE.',
   'tcs.com')

ON CONFLICT (name) DO NOTHING;
