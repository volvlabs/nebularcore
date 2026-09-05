-- Generic country/locale reference data. Seeded broadly so any host app
-- built on nebularcore can activate the countries it operates in without a
-- schema change; everything starts inactive except nothing — activation is
-- entirely up to the host app's own data/admin action.
CREATE TABLE IF NOT EXISTS countries (
    code TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    currency_code TEXT NOT NULL,
    currency_minor_unit_factor INTEGER NOT NULL DEFAULT 100,
    default_timezone TEXT NOT NULL,
    phone_dial_code TEXT NOT NULL,
    phone_region TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_countries_is_active ON countries(is_active);

CREATE OR REPLACE FUNCTION update_countries_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_countries_updated_at
    BEFORE UPDATE ON countries
    FOR EACH ROW
    EXECUTE FUNCTION update_countries_updated_at_column();

INSERT INTO countries (code, name, currency_code, currency_minor_unit_factor, default_timezone, phone_dial_code, phone_region, is_active) VALUES
    ('NG', 'Nigeria',       'NGN', 100, 'Africa/Lagos',      '+234', 'NG', false),
    ('GH', 'Ghana',         'GHS', 100, 'Africa/Accra',      '+233', 'GH', false),
    ('KE', 'Kenya',         'KES', 100, 'Africa/Nairobi',    '+254', 'KE', false),
    ('ZA', 'South Africa',  'ZAR', 100, 'Africa/Johannesburg','+27', 'ZA', false),
    ('EG', 'Egypt',         'EGP', 100, 'Africa/Cairo',      '+20',  'EG', false),
    ('TZ', 'Tanzania',      'TZS', 100, 'Africa/Dar_es_Salaam','+255','TZ', false),
    ('UG', 'Uganda',        'UGX', 100, 'Africa/Kampala',    '+256', 'UG', false),
    ('RW', 'Rwanda',        'RWF', 100, 'Africa/Kigali',     '+250', 'RW', false),
    ('CI', 'Ivory Coast',   'XOF', 100, 'Africa/Abidjan',    '+225', 'CI', false),
    ('SN', 'Senegal',       'XOF', 100, 'Africa/Dakar',      '+221', 'SN', false),
    ('CM', 'Cameroon',      'XAF', 100, 'Africa/Douala',     '+237', 'CM', false),
    ('ET', 'Ethiopia',      'ETB', 100, 'Africa/Addis_Ababa','+251', 'ET', false),
    ('MA', 'Morocco',       'MAD', 100, 'Africa/Casablanca', '+212', 'MA', false),
    ('DZ', 'Algeria',       'DZD', 100, 'Africa/Algiers',    '+213', 'DZ', false),
    ('TN', 'Tunisia',       'TND', 1000,'Africa/Tunis',      '+216', 'TN', false),
    ('ZM', 'Zambia',        'ZMW', 100, 'Africa/Lusaka',     '+260', 'ZM', false),
    ('ZW', 'Zimbabwe',      'ZWL', 100, 'Africa/Harare',     '+263', 'ZW', false),
    ('MZ', 'Mozambique',    'MZN', 100, 'Africa/Maputo',     '+258', 'MZ', false),
    ('BW', 'Botswana',      'BWP', 100, 'Africa/Gaborone',   '+267', 'BW', false),
    ('NA', 'Namibia',       'NAD', 100, 'Africa/Windhoek',   '+264', 'NA', false),
    ('US', 'United States', 'USD', 100, 'America/New_York',  '+1',   'US', false),
    ('GB', 'United Kingdom','GBP', 100, 'Europe/London',     '+44',  'GB', false)
ON CONFLICT (code) DO NOTHING;
