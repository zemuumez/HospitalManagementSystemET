CREATE TABLE hospital_general_setting (
  key varchar(128) PRIMARY KEY,
  value text NOT NULL DEFAULT '',
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE hospital_schedule_day (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  day_of_week smallint NOT NULL UNIQUE CHECK(day_of_week BETWEEN 1 AND 7),
  start_time varchar(8) NOT NULL DEFAULT '08:00',
  end_time varchar(8) NOT NULL DEFAULT '17:00',
  is_closed boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE front_cms_setting (
  key varchar(128) PRIMARY KEY,
  value text NOT NULL DEFAULT '',
  type varchar(32) NOT NULL DEFAULT 'general',
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

CREATE TABLE cms_testimonial (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar(191) NOT NULL,
  description text NOT NULL,
  position varchar(191) NOT NULL DEFAULT 'Patient',
  rating integer NOT NULL DEFAULT 5 CHECK(rating BETWEEN 1 AND 5),
  status smallint NOT NULL DEFAULT 1 CHECK(status IN (0, 1)),
  created_at timestamptz NOT NULL DEFAULT clock_timestamp(),
  updated_at timestamptz NOT NULL DEFAULT clock_timestamp()
);

-- Seed general settings
INSERT INTO hospital_general_setting (key, value)
VALUES
  ('app_name', 'Hospital Management System'),
  ('hospital_name', 'Addis Ababa Central Hospital'),
  ('hospital_email', 'info@hospital.et'),
  ('hospital_phone', '+251911000000'),
  ('hospital_address', 'Bole Sub-City, Addis Ababa, Ethiopia'),
  ('current_currency', 'ETB'),
  ('default_lang', 'en'),
  ('logo_url', '/images/logo.png'),
  ('favicon_url', '/favicon.ico'),
  ('facebook_url', 'https://facebook.com/hms.et'),
  ('twitter_url', 'https://twitter.com/hms_et'),
  ('instagram_url', 'https://instagram.com/hms.et'),
  ('linkedin_url', 'https://linkedin.com/company/hms-et'),
  ('queue_theme', 'modern-blue')
ON CONFLICT (key) DO NOTHING;

-- Seed hospital schedules (1=Mon ... 7=Sun)
INSERT INTO hospital_schedule_day (day_of_week, start_time, end_time, is_closed)
VALUES
  (1, '08:00', '17:00', false),
  (2, '08:00', '17:00', false),
  (3, '08:00', '17:00', false),
  (4, '08:00', '17:00', false),
  (5, '08:00', '17:00', false),
  (6, '08:00', '13:00', false),
  (7, '08:00', '12:00', true)
ON CONFLICT (day_of_week) DO NOTHING;

-- Seed CMS settings
INSERT INTO front_cms_setting (key, value, type)
VALUES
  ('home_title', 'Comprehensive Compassionate Healthcare in Addis Ababa', 'home'),
  ('home_description', 'Delivering premier specialized clinical care, diagnostics, and 24/7 emergency response.', 'home'),
  ('home_experience_years', '15', 'home'),
  ('home_certified_doctor_title', 'Certified and Compassionate Specialists', 'home'),
  ('home_certified_doctor_description', 'Board-certified physicians and dedicated nurses committed to patient wellness and clinical excellence.', 'home'),
  ('about_title', 'About Addis Ababa Central Hospital', 'about'),
  ('about_description', 'Founded with a dedication to accessible, evidence-based medicine, our hospital provides complete inpatient and outpatient medical services.', 'about'),
  ('about_mission', 'To heal, comfort, and advance the well-being of every patient through compassionate care and medical innovation.', 'about'),
  ('terms_and_conditions', 'Standard hospital terms and conditions of medical service.', 'terms'),
  ('privacy_policy', 'Commitment to strict patient medical record privacy and data confidentiality.', 'privacy'),
  ('contact_address', 'Bole Sub-City, Addis Ababa, Ethiopia', 'contact'),
  ('contact_phone', '+251911000000', 'contact'),
  ('contact_email', 'contact@hospital.et', 'contact'),
  ('contact_map_embed_url', 'https://maps.google.com/?q=Addis+Ababa+Ethiopia', 'contact')
ON CONFLICT (key) DO NOTHING;

-- Seed sample published testimonials
INSERT INTO cms_testimonial (id, name, description, position, rating, status)
VALUES
  ('fa111111-1111-1111-1111-111111111111', 'Almaz Tadesse', 'The surgical and nursing team provided outstanding care during my recovery.', 'Patient', 5, 1),
  ('fa222222-2222-2222-2222-222222222222', 'Dawit Haile', 'Fast emergency intake and very professional doctors. Highly recommended.', 'Patient', 5, 1)
ON CONFLICT (id) DO NOTHING;
