BEGIN;

CREATE TABLE countries (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  code TEXT NOT NULL UNIQUE,
  emblem TEXT NOT NULL UNIQUE,
  football_org_id INT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE leagues (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  code TEXT NOT NULL UNIQUE,
  emblem TEXT NOT NULL UNIQUE,
  football_org_id INT NOT NULL UNIQUE,
  country_id UUID NOT NULL REFERENCES countries(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE teams (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  short_name TEXT NOT NULL,
  code TEXT NOT NULL UNIQUE,
  tla TEXT NOT NULL UNIQUE,
  emblem TEXT NOT NULL UNIQUE,
  football_org_id INT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE seasons (
  id UUID PRIMARY KEY,
  football_org_id INT NOT NULL UNIQUE,
  season TEXT NOT NULL UNIQUE,
  period DATERANGE NOT NULL,
  match_day INT NOT NULL DEFAULT 0,
  winner_team_id UUID REFERENCES teams(id),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE league_seasons (
  league_id UUID NOT NULL REFERENCES leagues(id) ON DELETE CASCADE,
  season_id UUID NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
  PRIMARY KEY (league_id, season_id)
);


CREATE TABLE teams_leagues (
  team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
  league_id UUID NOT NULL,
  season_id UUID NOT NULL,
  PRIMARY KEY (team_id, league_id, season_id),
  FOREIGN KEY (league_id, season_id)
    REFERENCES league_seasons(league_id, season_id)
    ON DELETE CASCADE
);




CREATE TABLE standings (
  id UUID PRIMARY KEY,
  team_id UUID NOT NULL, 
  league_id UUID NOT NULL,
  season_id UUID NOT NULL,
  played_games INT NOT NULL DEFAULT 0,
  won INT NOT NULL DEFAULT 0,
  lost INT NOT NULL DEFAULT 0,
  draw INT NOT NULL DEFAULT 0,
  points INT NOT NULL DEFAULT 0,
  goals_for INT NOT NULL DEFAULT 0,
  goals_against INT NOT NULL DEFAULT 0,
  goal_difference INT NOT NULL DEFAULT 0,
  position INT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (team_id, league_id, season_id),

  FOREIGN KEY (team_id, league_id, season_id)
  REFERENCES teams_leagues(team_id, league_id, season_id)
  ON DELETE CASCADE,

  CHECK (played_games >= 0),
  CHECK (won >= 0),
  CHECK (draw >= 0),
  CHECK (lost >= 0),
  CHECK (points >= 0),
  CHECK (position > 0),
  CHECK (won + draw + lost = played_games)
);

COMMIT;
