-- ZZ ("Global") is a sentinel row, not a real country — ZZ is the ISO
-- 3166-1 user-assigned code libphonenumber itself uses for "unknown
-- region". It's the resolution target for any user/request whose real
-- country can't be determined or isn't yet supported (see
-- middleware.AccountCountryResolverFunc's fallback and Config.DefaultCountryCode),
-- so display code always has *something* to render instead of leaving a
-- caller to handle a third "nothing resolved" state. It ships active,
-- since it's the safety net other resolution falls back to — a host app
-- deactivating it would break that fallback chain, so admin surfaces
-- should treat it as non-deactivatable by convention.
INSERT INTO countries (code, name, currency_code, currency_minor_unit_factor, default_timezone, phone_dial_code, phone_region, is_active)
VALUES ('ZZ', 'Global', 'USD', 100, 'UTC', '', 'ZZ', true)
ON CONFLICT (code) DO NOTHING;
