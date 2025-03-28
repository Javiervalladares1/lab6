CREATE TABLE IF NOT EXISTS matches (
    id SERIAL PRIMARY KEY,
    home_team TEXT NOT NULL,
    away_team TEXT NOT NULL,
    match_date DATE NOT NULL
);