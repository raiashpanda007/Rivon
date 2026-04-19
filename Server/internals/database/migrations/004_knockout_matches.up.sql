BEGIN;
CREATE TABLE knockout_matches (
  id UUID PRIMARY KEY,
  league_id UUID NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
  season_id UUID NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  football_org_match_id INT NOT NULL UNIQUE,
  stage TEXT NOT NULL,
  home_team_id UUID NOT NULL REFERENCES teams(id),
  away_team_id UUID NOT NULL REFERENCES teams(id),
  home_score INT,
  away_score INT,
  status TEXT NOT NULL DEFAULT 'SCHEDULED',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
COMMIT;
