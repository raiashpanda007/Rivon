"use client"
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectTrigger,
  SelectValue,
} from "@workspace/ui/components/select"
import ApiCaller from "@workspace/api-caller";
import { RequestType } from "@workspace/api-caller";
import { useState, useEffect, useMemo } from "react";
import Loading from "../Loading"
import { motion, AnimatePresence } from "framer-motion";

interface AllCompetitions {
  id: string;
  code: string
  countryCode: string;
  countryEmblem: string;
  countryFootBallOrgCode: number;
  countryId: string;
  countryName: string;
  emblem: string;
  football_org_id: number;
  name: string
}

interface Season {
  id: string;
  period: { start: Date, end: Date }
  season: string;
  matchDay: number;
  winner: null | string;
  createdAt: Date;
  updatedAt: Date
}

interface LeagueSeason {
  leagueId: string;
  seasonId: string;
}

interface Standings {
  id: string;
  leagueId: string;
  seasonId: string;
  teamId: string;
  teamName: string;
  teamShortName: string;
  teamTLA: string;
  teamCode: string;
  teamEmblem: string;
  position: number;
  playedGames: number;
  lost: number;
  won: number;
  draw: number;
  goalDifference: number;
  goalsAgainst: number;
  goalsFor: number;
}

interface KnockoutLeg {
  homeTeamId: string;
  homeScore: number | null;
  awayScore: number | null;
  status: string;
}

interface KnockoutTeam {
  id: string;
  name: string;
  shortName: string;
  tla: string;
  emblem: string;
}

interface KnockoutMatchup {
  team1: KnockoutTeam;
  team2: KnockoutTeam;
  leg1: KnockoutLeg | null;
  leg2: KnockoutLeg | null;
  team1AggGoals: number | null;
  team2AggGoals: number | null;
  winnerTeamId: string | null;
}

interface KnockoutStageData {
  stage: string;
  matchups: KnockoutMatchup[];
}

const STAGE_LABELS: Record<string, string> = {
  KNOCKOUT_ROUND_PLAY_OFFS: "Playoffs",
  LAST_16: "Round of 16",
  QUARTER_FINALS: "Quarter Finals",
  SEMI_FINALS: "Semi Finals",
  FINAL: "Final",
}

const STAGE_ORDER = [
  "KNOCKOUT_ROUND_PLAY_OFFS",
  "LAST_16",
  "QUARTER_FINALS",
  "SEMI_FINALS",
  "FINAL",
]

function stageLabel(stage: string) {
  return STAGE_LABELS[stage] ?? stage.replace(/_/g, " ")
}

const getLeagueMessage = (c: AllCompetitions) => {
  const name = c.name.toLowerCase();
  const code = c.code;

  if (code === 'PD' || name.includes('la liga'))
    return "Favourite of the creator 🤍⚪ Real Madrid bias is real. Pure VAMOS league. Drama guaranteed."

  if (code === 'PL' || name.includes('premier'))
    return "Hardest league in the world 💥 Every match is war. Fav club: Man United (emotional damage included)."

  if (code === 'SA' || name.includes('serie'))
    return "Once the king 👑 Lost its aura, now slowly rising again. No fixed fav, but AC Milan has the heart."

  if (code === 'BL1' || name.includes('bundesliga'))
    return "Feels rigged 🤨 but watching Bayern destroy teams is oddly satisfying. Fav club: Bayern, obviously."

  if (code === 'FL1' || name.includes('ligue 1'))
    return "Barely watched 😴 PSG carries the league. Not gonna lie — their jerseys are fire 🔥"

  if (code === 'CL' || name.includes('champions'))
    return "Best nights in football 🌌 Champions League hits different. Pure legacy, pure madness."

  return "Enjoy the beautiful game of football!";
}

// ─── Knockout bracket components ─────────────────────────────────────────────

function ScoreCell({ score }: { score: number | null }) {
  return (
    <span className="w-6 text-center font-mono font-bold text-sm">
      {score !== null ? score : "–"}
    </span>
  )
}

function MatchupCard({ matchup }: { matchup: KnockoutMatchup }) {
  const isTeam1Winner = matchup.winnerTeamId === matchup.team1.id
  const isTeam2Winner = matchup.winnerTeamId === matchup.team2.id
  const isFinal = matchup.leg2 === null

  return (
    <div className="rounded-xl border border-white/10 bg-background/40 backdrop-blur-sm overflow-hidden shadow-lg min-w-[220px]">
      {/* Team 1 row */}
      <div className={`flex items-center gap-2 px-3 py-2.5 transition-colors ${isTeam1Winner ? "bg-orange-500/15 border-l-2 border-orange-500" : ""}`}>
        <div className="w-7 h-7 flex-shrink-0 bg-white/10 rounded-full p-0.5 flex items-center justify-center">
          <img src={matchup.team1.emblem} alt={matchup.team1.name} className="w-full h-full object-contain" />
        </div>
        <span className={`flex-1 text-xs font-semibold truncate ${isTeam1Winner ? "text-orange-400" : "text-foreground"}`}>
          {matchup.team1.shortName || matchup.team1.name}
        </span>
        <div className="flex items-center gap-1">
          {!isFinal && (
            <>
              <ScoreCell score={matchup.leg1?.homeScore ?? null} />
              <ScoreCell score={matchup.leg2?.awayScore ?? null} />
            </>
          )}
          {isFinal && <ScoreCell score={matchup.leg1?.homeScore ?? null} />}
          {matchup.team1AggGoals !== null && (
            <span className={`ml-1 w-6 text-center text-xs font-black rounded ${isTeam1Winner ? "text-orange-400" : "text-muted-foreground"}`}>
              {matchup.team1AggGoals}
            </span>
          )}
        </div>
      </div>

      {/* Divider with leg labels */}
      <div className="flex items-center bg-background/20 border-y border-white/5 px-3 py-0.5">
        <span className="flex-1" />
        {!isFinal && (
          <div className="flex items-center gap-1 text-[9px] text-muted-foreground font-mono">
            <span className="w-6 text-center">L1</span>
            <span className="w-6 text-center">L2</span>
            <span className="ml-1 w-6 text-center">AGG</span>
          </div>
        )}
        {isFinal && (
          <span className="text-[9px] text-muted-foreground font-mono">FT · AGG</span>
        )}
      </div>

      {/* Team 2 row */}
      <div className={`flex items-center gap-2 px-3 py-2.5 transition-colors ${isTeam2Winner ? "bg-orange-500/15 border-l-2 border-orange-500" : ""}`}>
        <div className="w-7 h-7 flex-shrink-0 bg-white/10 rounded-full p-0.5 flex items-center justify-center">
          <img src={matchup.team2.emblem} alt={matchup.team2.name} className="w-full h-full object-contain" />
        </div>
        <span className={`flex-1 text-xs font-semibold truncate ${isTeam2Winner ? "text-orange-400" : "text-foreground"}`}>
          {matchup.team2.shortName || matchup.team2.name}
        </span>
        <div className="flex items-center gap-1">
          {!isFinal && (
            <>
              <ScoreCell score={matchup.leg1?.awayScore ?? null} />
              <ScoreCell score={matchup.leg2?.homeScore ?? null} />
            </>
          )}
          {isFinal && <ScoreCell score={matchup.leg1?.awayScore ?? null} />}
          {matchup.team2AggGoals !== null && (
            <span className={`ml-1 w-6 text-center text-xs font-black rounded ${isTeam2Winner ? "text-orange-400" : "text-muted-foreground"}`}>
              {matchup.team2AggGoals}
            </span>
          )}
        </div>
      </div>
    </div>
  )
}

function KnockoutBracket({ data }: { data: KnockoutStageData[] }) {
  // Sort stages by canonical order
  const sorted = [...data].sort((a, b) => {
    const oi = STAGE_ORDER.indexOf(a.stage)
    const oj = STAGE_ORDER.indexOf(b.stage)
    return (oi === -1 ? 99 : oi) - (oj === -1 ? 99 : oj)
  })

  if (sorted.length === 0) {
    return (
      <motion.div
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        className="flex flex-col items-center justify-center py-20 gap-4"
      >
        <div className="px-2 py-1 rounded-sm bg-orange-500/10 border border-orange-500/20 flex items-center gap-2">
          <span className="w-1.5 h-1.5 rounded-full bg-orange-500/50" />
          <span className="font-mono text-[10px] font-bold text-orange-500/70 tracking-widest">AWAITING_KICKOFF</span>
        </div>
        <p className="text-muted-foreground text-sm font-mono">Knockout stage has not started yet.</p>
        <p className="text-muted-foreground/50 text-xs font-mono">Check back after the league phase concludes.</p>
      </motion.div>
    )
  }

  return (
    <div className="w-full overflow-x-auto pb-4">
      {/* Desktop: horizontal bracket flow */}
      <div className="hidden md:flex items-start gap-0 min-w-max">
        {sorted.map((stage, stageIdx) => (
          <div key={stage.stage} className="flex items-start">
            {/* Stage column */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: stageIdx * 0.12, duration: 0.4 }}
              className="flex flex-col gap-2"
            >
              {/* Stage label */}
              <div className="flex items-center justify-center mb-3">
                <div className="px-2 py-1 rounded-sm bg-orange-500/10 border border-orange-500/20 flex items-center gap-1.5">
                  <span className="w-1.5 h-1.5 rounded-full bg-orange-500 animate-pulse" />
                  <span className="font-mono text-[10px] font-bold text-orange-500 tracking-widest uppercase">
                    {stageLabel(stage.stage)}
                  </span>
                </div>
              </div>

              {/* Matchup cards spaced to align with next round */}
              <div className="flex flex-col gap-6">
                {stage.matchups.map((matchup, mIdx) => (
                  <motion.div
                    key={`${matchup.team1.id}-${matchup.team2.id}`}
                    initial={{ opacity: 0, x: -10 }}
                    animate={{ opacity: 1, x: 0 }}
                    transition={{ delay: stageIdx * 0.12 + mIdx * 0.05 }}
                  >
                    <MatchupCard matchup={matchup} />
                  </motion.div>
                ))}
              </div>
            </motion.div>

            {/* Connector lines between rounds */}
            {stageIdx < sorted.length - 1 && (
              <div className="flex items-center self-stretch mx-1 mt-14">
                <div className="flex items-center gap-0">
                  <div className="w-4 h-px bg-orange-500/30" />
                  <svg width="8" height="8" viewBox="0 0 8 8" className="text-orange-500/40">
                    <path d="M0 4 L8 4 M5 1 L8 4 L5 7" stroke="currentColor" strokeWidth="1.5" fill="none" strokeLinecap="round" strokeLinejoin="round" />
                  </svg>
                </div>
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Mobile: vertical stacked stages */}
      <div className="flex md:hidden flex-col gap-8">
        {sorted.map((stage, stageIdx) => (
          <motion.div
            key={stage.stage}
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: stageIdx * 0.1 }}
            className="flex flex-col gap-3"
          >
            <div className="flex items-center gap-2">
              <div className="px-2 py-1 rounded-sm bg-orange-500/10 border border-orange-500/20 flex items-center gap-1.5">
                <span className="w-1.5 h-1.5 rounded-full bg-orange-500 animate-pulse" />
                <span className="font-mono text-[10px] font-bold text-orange-500 tracking-widest uppercase">
                  {stageLabel(stage.stage)}
                </span>
              </div>
              {stageIdx < sorted.length - 1 && (
                <div className="flex-1 flex items-center gap-1 text-orange-500/30">
                  <div className="flex-1 h-px bg-orange-500/20" />
                  <span className="text-xs">↓</span>
                </div>
              )}
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
              {stage.matchups.map((matchup) => (
                <MatchupCard key={`${matchup.team1.id}-${matchup.team2.id}`} matchup={matchup} />
              ))}
            </div>
          </motion.div>
        ))}
      </div>
    </div>
  )
}

// ─── Main component ───────────────────────────────────────────────────────────

function LeagueSelection() {
  const competitionMap = new Map<string, AllCompetitions>()
  const [isLoading, setLoading] = useState(false);
  const [listCompetitions, setListCompetitions] = useState<AllCompetitions[]>([]);
  const [selectedCompetition, setSelectedCompetition] = useState<AllCompetitions>();
  const [listSeasons, setListSeasons] = useState<{ season: Season[], leagueSeason: LeagueSeason[] }>();
  const [selectedSeason, setSelectedSeason] = useState<Season>()
  const [standings, setStandings] = useState<Standings[]>([]);
  const [viewMode, setViewMode] = useState<"league" | "knockout">("league");
  const [knockoutData, setKnockoutData] = useState<KnockoutStageData[]>([]);

  const isCL = selectedCompetition != null &&
    (selectedCompetition.code === 'CL' || selectedCompetition.name.toLowerCase().includes('champions'))

  async function GetAllCompetitions() {
    setLoading(true)
    const x = await ApiCaller<{}, AllCompetitions[]>({
      requestType: RequestType.GET,
      paths: ["api", "rivon", "football-meta", "competitions"],
      body: {},
    });
    if (x.ok) {
      setListCompetitions(x.response.data);
      x.response.data.forEach((data) => {
        competitionMap.set(data.id, data)
      })
    }
    setLoading(false)
  }

  async function GetAllSeasons() {
    const seasons = await ApiCaller<{}, { season: Season[], leagueSeason: LeagueSeason[] }>({
      requestType: RequestType.GET,
      paths: ["api", "rivon", "football-meta", "seasons"],
      body: {},
    })
    if (seasons.ok) {
      setListSeasons(seasons.response.data);
    }
  }

  function SetLeagueSeasonMap() {
    if (!listSeasons || !listCompetitions) return new Map<string, Season[]>();
    const seasonLeagueMatch = new Map<string, Season[]>();

    listCompetitions.forEach((competition) => {
      const compId = String(competition.id);

      const targetSeasonIds = new Set(
        listSeasons.leagueSeason
          .filter((lSeason) => String(lSeason.leagueId) === compId)
          .map((val) => String(val.seasonId))
      );

      const seasons = listSeasons.season.filter((season) => targetSeasonIds.has(String(season.id)));
      seasonLeagueMatch.set(String(competition.id), seasons);
    })

    return seasonLeagueMatch;
  }

  async function GetTeamStandings() {
    setStandings([]);
    if (!selectedCompetition || !selectedSeason) return;

    const stands = await ApiCaller<{}, Standings[]>({
      paths: ["api", "rivon", "football-meta", "standings"],
      requestType: RequestType.GET,
      queryParams: {
        leagueId: selectedCompetition?.id,
        seasonId: selectedSeason?.id
      }
    });

    if (stands.ok) {
      setStandings(stands.response.data);
    }
  }

  async function GetKnockoutData() {
    setKnockoutData([]);
    if (!selectedCompetition || !selectedSeason) return;

    const result = await ApiCaller<{}, KnockoutStageData[]>({
      paths: ["api", "rivon", "football-meta", "knockout"],
      requestType: RequestType.GET,
      queryParams: {
        leagueId: selectedCompetition.id,
        seasonId: selectedSeason.id,
      }
    });

    if (result.ok && result.response.data) {
      setKnockoutData(result.response.data);
    }
  }

  const items = listCompetitions.map((val) => ({ label: val.name, value: val }))

  const SeasonLeagueMap = useMemo(SetLeagueSeasonMap, [listSeasons, listCompetitions]);

  const seasonItems: { label: string, value: Season }[] =
    (SeasonLeagueMap.get(String(selectedCompetition?.id ?? "")) ?? []).map((val: Season) => ({
      label: val.season, value: val
    }))

  useEffect(() => {
    Promise.all([GetAllCompetitions(), GetAllSeasons()]);
  }, [])

  useEffect(() => {
    if (selectedCompetition) {
      setSelectedSeason(undefined);
      setViewMode("league");
      setKnockoutData([]);
    }
  }, [selectedCompetition, SeasonLeagueMap])

  useEffect(() => {
    GetTeamStandings();
    if (isCL) GetKnockoutData();
    else setKnockoutData([]);
  }, [selectedSeason])

  return (
    <>
      {isLoading && <Loading heading="Loading Leagues" message="Please wait while we fetch the competitions..." />}
      <div className="w-full flex flex-col items-center gap-8 py-8 animate-in fade-in duration-500">

        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="w-full flex flex-col items-center text-center space-y-6 mb-4"
        >
          <div className="px-2 py-1 rounded-sm bg-orange-500/10 border border-orange-500/20 flex items-center gap-2">
            <span className="w-1.5 h-1.5 rounded-full bg-orange-500 animate-pulse shadow-[0_0_8px_rgba(249,115,22,0.8)]" />
            <span className="font-mono text-[10px] font-bold text-orange-500 tracking-widest">LEAGUE_DATA</span>
          </div>
          <h1 className="text-5xl sm:text-6xl md:text-7xl font-black tracking-tighter text-foreground leading-none">
            League <span className="text-transparent bg-clip-text bg-gradient-to-r from-orange-400 to-orange-600">Standings</span>
          </h1>
          <p className="font-mono text-xs text-muted-foreground max-w-xl">
            Access global football data directly from the Meta API. Select an arena to fetch the latest tables and form.
          </p>
          <div className="flex justify-center mt-6 w-full">
            <div className="w-full max-w-sm relative">
              <Select onValueChange={(e) => setSelectedCompetition(JSON.parse(e))}>
                <SelectTrigger className="w-full bg-background/80 backdrop-blur-sm border-input h-12 text-lg relative shadow-sm hover:bg-accent/50 transition-colors">
                  <SelectValue placeholder="Select a League" />
                </SelectTrigger>
                <SelectContent className="max-h-[300px]">
                  <SelectGroup>
                    <SelectLabel className="text-orange-600 font-bold px-4 py-2">Leagues</SelectLabel>
                    {items.map((item) => (
                      <SelectItem
                        key={item.value.id}
                        value={JSON.stringify(item.value)}
                        className="cursor-pointer focus:bg-orange-50 focus:text-orange-700 dark:focus:bg-orange-900/20 dark:focus:text-orange-400 font-medium">
                        {item.label}
                      </SelectItem>
                    ))}
                  </SelectGroup>
                </SelectContent>
              </Select>
            </div>
          </div>
        </motion.div>

        <AnimatePresence mode="wait">
          {selectedCompetition && (
            <motion.div
              key={selectedCompetition.id}
              initial={{ opacity: 0, scale: 0.95, y: 20 }}
              animate={{ opacity: 1, scale: 1, y: 0 }}
              exit={{ opacity: 0, scale: 0.95, y: 10 }}
              transition={{ duration: 0.4, type: "spring" }}
              className="w-full max-w-5xl"
            >

              <motion.div
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.2 }}
                className="mb-8 p-6 border border-orange-500/20 rounded-2xl backdrop-blur-md text-center shadow-lg transform hover:scale-[1.01] transition-transform duration-300"
              >
                <p className="text-lg md:text-xl text-orange-600 dark:text-orange-400 italic font-serif">
                  "{getLeagueMessage(selectedCompetition)}"
                </p>
              </motion.div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">

                <div className="group relative overflow-hidden rounded-3xl bg-white/30 dark:bg-black/20 backdrop-blur-md border border-white/20 dark:border-white/5 shadow-2xl p-8 hover:border-orange-500/30 transition-all duration-300">
                  <div className="absolute top-0 right-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                    <svg className="w-32 h-32 text-orange-500" fill="currentColor" viewBox="0 0 24 24"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z" /></svg>
                  </div>
                  <div className="relative z-10 flex flex-col items-center">
                    <div className="h-32 w-32 rounded-full p-4 bg-white shadow-lg flex items-center justify-center mb-6 ring-4 ring-orange-100 dark:ring-orange-900/30">
                      <img className="w-full h-full object-contain" src={selectedCompetition.emblem} alt={selectedCompetition.name} />
                    </div>
                    <h2 className="text-3xl font-bold text-foreground mb-1 text-center">{selectedCompetition.name}</h2>
                    <span className="inline-block px-3 py-1 bg-orange-100 dark:bg-orange-900/30 text-orange-600 dark:text-orange-400 rounded-full text-sm font-bold mb-6">
                      {selectedCompetition.code}
                    </span>
                    <div className="w-full grid grid-cols-2 gap-4 mt-2">
                      <div className="flex flex-col items-center p-3 bg-background/40 rounded-xl border border-white/10">
                        <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">Org ID</span>
                        <span className="text-lg font-bold font-mono">{selectedCompetition.football_org_id}</span>
                      </div>
                      <div className="flex flex-col items-center p-3 bg-background/40 rounded-xl border border-white/10">
                        <span className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">FootBall ORG ID</span>
                        <span className="text-lg font-bold">{selectedCompetition.football_org_id}</span>
                      </div>
                    </div>
                  </div>
                </div>

                <div className="group relative overflow-hidden rounded-3xl bg-white/30 dark:bg-black/20 backdrop-blur-md border border-white/20 dark:border-white/5 shadow-2xl p-8 hover:border-orange-500/30 transition-all duration-300 delay-75">
                  <div className="absolute top-0 right-0 p-3 opacity-10 group-hover:opacity-20 transition-opacity">
                    <svg className="w-32 h-32 text-orange-500" fill="currentColor" viewBox="0 0 24 24"><path d="M12 2C6.47 2 2 6.47 2 12s4.47 10 10 10 10-4.47 10-10S17.53 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm-5.59-2.12L6 17l6.5-6.5L6 4l.41-.88C9.59 4.25 12 7.12 12 12c-2.88 0-5.29-1.93-6.41-4.88z" /></svg>
                  </div>
                  <div className="relative z-10 flex flex-col items-center">
                    <h3 className="text-xl font-bold text-orange-600 dark:text-orange-400 mb-6 uppercase tracking-widest border-b-2 border-orange-500/20 pb-2">Location</h3>
                    <div className="h-24 w-24 rounded-full p-2 bg-white shadow-md flex items-center justify-center mb-4 ring-2 ring-orange-50 dark:ring-orange-900/20">
                      <img className="w-full h-full object-contain" src={selectedCompetition.countryEmblem} alt={selectedCompetition.countryName} />
                    </div>
                    <h2 className="text-3xl font-bold text-foreground mb-4 text-center">{selectedCompetition.countryName}</h2>
                    <div className="w-full p-4 bg-background/40 rounded-2xl border border-white/10 flex items-center justify-between">
                      <div className="flex items-center space-x-3">
                        <div className="h-10 w-10 rounded-full bg-orange-100 dark:bg-orange-900/50 flex items-center justify-center text-orange-600">
                          <span className="font-bold text-sm">CC</span>
                        </div>
                        <div className="flex flex-col">
                          <span className="text-xs text-muted-foreground font-semibold">Country Code</span>
                          <span className="font-bold">{selectedCompetition.countryCode}</span>
                        </div>
                      </div>
                      <div className="h-8 w-[1px] bg-border mx-2"></div>
                      <div className="flex items-center space-x-3">
                        <div className="h-10 w-10 rounded-full bg-orange-100 dark:bg-orange-900/50 flex items-center justify-center text-orange-600">
                          <span className="font-bold text-sm">ID</span>
                        </div>
                        <div className="flex flex-col">
                          <span className="text-xs text-muted-foreground font-semibold">Country ID</span>
                          <span className="font-bold truncate max-w-[80px]">{selectedCompetition.countryFootBallOrgCode}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Season selector */}
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.3, duration: 0.5 }}
                className="w-full text-center space-y-6 mt-12 mb-8 pt-8 border-t border-orange-500/20">
                <div className="space-y-2">
                  <h2 className="text-3xl font-bold bg-clip-text text-transparent bg-gradient-to-r from-orange-600 via-orange-500 to-amber-500">
                    Select Season
                  </h2>
                  <p className="text-muted-foreground font-medium text-lg">Choose a season to view matches and stats</p>
                </div>
                <div className="flex justify-center">
                  <div className="w-full max-w-sm">
                    <Select onValueChange={(e) => setSelectedSeason(JSON.parse(e))}>
                      <SelectTrigger className="w-full h-14 bg-background/90 backdrop-blur-md border border-input text-xl font-medium shadow-sm transition-all hover:bg-accent/50">
                        <SelectValue placeholder="Select a Season" />
                      </SelectTrigger>
                      <SelectContent className="max-h-[300px]">
                        <SelectGroup>
                          <SelectLabel className="text-orange-600 font-bold px-4 py-2">Available Seasons</SelectLabel>
                          {seasonItems.map((item) => (
                            <SelectItem
                              key={item.value.id}
                              value={JSON.stringify(item.value)}
                              className="cursor-pointer py-3 text-base focus:bg-orange-50 focus:text-orange-700 dark:focus:bg-orange-900/20 dark:focus:text-orange-400">
                              {item.label}
                            </SelectItem>
                          ))}
                        </SelectGroup>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
              </motion.div>

              {/* View mode toggle — only for CL once a season is picked */}
              <AnimatePresence>
                {isCL && selectedSeason && (
                  <motion.div
                    key="cl-toggle"
                    initial={{ opacity: 0, y: -8 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -8 }}
                    transition={{ duration: 0.25 }}
                    className="flex justify-center mb-8"
                  >
                    <div className="flex items-center gap-1 p-1 rounded-xl bg-background/60 backdrop-blur-sm border border-orange-500/20 shadow-md">
                      <button
                        onClick={() => setViewMode("league")}
                        className={`
                          flex items-center gap-2 px-5 py-2 rounded-lg text-sm font-bold transition-all duration-200
                          ${viewMode === "league"
                            ? "bg-orange-500 text-white shadow-[0_0_12px_rgba(249,115,22,0.4)]"
                            : "text-muted-foreground hover:text-foreground hover:bg-accent/50"
                          }
                        `}
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 10h16M4 14h16M4 18h16" />
                        </svg>
                        League Phase
                      </button>
                      <button
                        onClick={() => setViewMode("knockout")}
                        className={`
                          flex items-center gap-2 px-5 py-2 rounded-lg text-sm font-bold transition-all duration-200
                          ${viewMode === "knockout"
                            ? "bg-orange-500 text-white shadow-[0_0_12px_rgba(249,115,22,0.4)]"
                            : "text-muted-foreground hover:text-foreground hover:bg-accent/50"
                          }
                        `}
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                        </svg>
                        Knockout
                      </button>
                    </div>
                  </motion.div>
                )}
              </AnimatePresence>

              {/* ── League Phase table ── */}
              <AnimatePresence>
                {standings.length > 0 && selectedSeason && viewMode === "league" && (
                  <motion.div
                    initial={{ opacity: 0, y: 40 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: 20 }}
                    transition={{ duration: 0.6, type: "spring", bounce: 0.3 }}
                    className="w-full mt-4 mb-20"
                  >
                    <div className="relative overflow-hidden rounded-3xl bg-white/40 dark:bg-black/40 backdrop-blur-xl border border-white/20 dark:border-white/10 shadow-2xl">
                      <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-transparent via-orange-500 to-transparent opacity-50"></div>
                      <div className="absolute -top-24 -right-24 w-64 h-64 bg-orange-500/10 rounded-full blur-3xl p-3"></div>
                      <div className="absolute -bottom-24 -left-24 w-64 h-64 bg-amber-500/10 rounded-full blur-3xl p-3"></div>

                      <div className="p-6 md:p-8 relative z-10">
                        <div className="flex flex-col md:flex-row justify-between items-center mb-8 gap-4">
                          <div className="flex items-center gap-4">
                            <div className="h-14 w-14 p-2 bg-white rounded-xl shadow-lg flex items-center justify-center ring-2 ring-orange-100 dark:ring-orange-900/30">
                              <img src={selectedCompetition.emblem} alt="logo" className="w-full h-full object-contain" />
                            </div>
                            <div>
                              <h3 className="text-2xl font-black text-foreground tracking-tight">
                                {isCL ? "League Phase Table" : "League Table"}
                              </h3>
                              <p className="text-sm text-orange-600 dark:text-orange-400 font-bold uppercase tracking-wider">{selectedSeason.season} Season</p>
                            </div>
                          </div>

                          <div className="flex gap-2 text-xs font-bold bg-background/50 p-1.5 rounded-lg border border-white/10 backdrop-blur-sm flex-wrap justify-end">
                            {isCL ? (
                              <>
                                <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-green-500/10 text-green-600 dark:text-green-400 border border-green-500/20">
                                  <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse"></span>
                                  <span>Qualifies</span>
                                </div>
                                <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-yellow-500/10 text-yellow-600 dark:text-yellow-400 border border-yellow-500/20">
                                  <span className="w-2 h-2 rounded-full bg-yellow-500 animate-pulse"></span>
                                  <span>Play-off</span>
                                </div>
                                <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-red-500/10 text-red-600 dark:text-red-400 border border-red-500/20">
                                  <span className="w-2 h-2 rounded-full bg-red-500 animate-pulse"></span>
                                  <span>Eliminated</span>
                                </div>
                              </>
                            ) : (
                              <>
                                <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-green-500/10 text-green-600 dark:text-green-400 border border-green-500/20">
                                  <span className="w-2 h-2 rounded-full bg-green-500 animate-pulse"></span>
                                  <span>Qualification</span>
                                </div>
                                <div className="flex items-center gap-1.5 px-3 py-1.5 rounded-md bg-red-500/10 text-red-600 dark:text-red-400 border border-red-500/20">
                                  <span className="w-2 h-2 rounded-full bg-red-500 animate-pulse"></span>
                                  <span>Relegation</span>
                                </div>
                              </>
                            )}
                          </div>
                        </div>

                        <div className="overflow-x-auto rounded-xl border border-white/10 bg-background/20">
                          <table className="w-full text-sm text-left border-collapse">
                            <thead className="bg-background/60 backdrop-blur-md text-xs uppercase font-bold text-muted-foreground sticky top-0 md:static">
                              <tr>
                                <th className="px-4 py-4 text-center w-16">Pos</th>
                                <th className="px-4 py-4">Club</th>
                                <th className="px-4 py-4 text-center hidden md:table-cell" title="Played">MP</th>
                                <th className="px-4 py-4 text-center hidden md:table-cell" title="Won">W</th>
                                <th className="px-4 py-4 text-center hidden md:table-cell" title="Drawn">D</th>
                                <th className="px-4 py-4 text-center hidden md:table-cell" title="Lost">L</th>
                                <th className="px-4 py-4 text-center hidden lg:table-cell" title="Goals For">GF</th>
                                <th className="px-4 py-4 text-center hidden lg:table-cell" title="Goals Against">GA</th>
                                <th className="px-4 py-4 text-center" title="Goal Difference">GD</th>
                                <th className="px-4 py-4 text-center font-black text-foreground text-base">Pts</th>
                                <th className="px-4 py-4 text-center w-24">Form</th>
                              </tr>
                            </thead>
                            <tbody className="divide-y divide-border/40">
                              {standings.sort((a, b) => a.position - b.position).map((team, index) => {
                                const form = ['W', 'D', 'L', 'W', 'W'].sort(() => 0.5 - Math.random()).slice(0, 5);
                                const points = (team.won * 3) + team.draw;

                                let isTop = false;
                                let isMid = false;
                                let isBottom = false;

                                if (isCL) {
                                  if (index < 8) isTop = true;
                                  else if (index >= 8 && index < 24) isMid = true;
                                  else if (index >= standings.length - 8) isBottom = true;
                                } else {
                                  if (index < 4) isTop = true;
                                  if (index >= standings.length - 3) isBottom = true;
                                }

                                return (
                                  <motion.tr
                                    key={team.teamId}
                                    initial={{ opacity: 0, x: -20 }}
                                    animate={{ opacity: 1, x: 0 }}
                                    transition={{ delay: index * 0.03 }}
                                    className={`
                                      group transition-colors hover:bg-orange-500/5
                                      ${isTop ? 'bg-gradient-to-r from-green-500/5 to-transparent' : ''}
                                      ${isMid ? 'bg-gradient-to-r from-yellow-500/5 to-transparent' : ''}
                                      ${isBottom ? 'bg-gradient-to-r from-red-500/5 to-transparent' : ''}
                                    `}
                                  >
                                    <td className="px-4 py-3 text-center">
                                      <span className={`
                                        inline-flex items-center justify-center w-8 h-8 rounded-full font-bold text-xs
                                        ${isTop ? 'bg-green-500 text-white shadow-green-500/20 shadow-lg' : ''}
                                        ${isMid ? 'bg-yellow-500 text-white shadow-yellow-500/20 shadow-lg' : ''}
                                        ${isBottom ? 'bg-red-500 text-white shadow-red-500/20 shadow-lg' : ''}
                                        ${!isTop && !isMid && !isBottom ? 'bg-accent/50 text-muted-foreground' : ''}
                                      `}>
                                        {team.position}
                                      </span>
                                    </td>
                                    <td className="px-4 py-3">
                                      <div className="flex items-center space-x-4">
                                        <div className="relative w-10 h-10 transition-transform group-hover:scale-110 duration-200">
                                          <div className="absolute inset-0 bg-white dark:bg-black/20 rounded-full blur-sm opacity-50"></div>
                                          <img className="relative w-full h-full object-contain drop-shadow-sm" src={team.teamEmblem} alt={team.teamName} />
                                        </div>
                                        <div className="flex flex-col">
                                          <span className="font-bold text-foreground text-base truncate max-w-[120px] md:max-w-none group-hover:text-orange-600 dark:group-hover:text-orange-400 transition-colors">
                                            {team.teamShortName || team.teamName}
                                          </span>
                                          <span className="text-[10px] text-muted-foreground hidden md:inline-block">
                                            {team.teamTLA}
                                          </span>
                                        </div>
                                      </div>
                                    </td>
                                    <td className="px-4 py-3 text-center font-medium text-muted-foreground hidden md:table-cell">{team.playedGames}</td>
                                    <td className="px-4 py-3 text-center text-green-600 dark:text-green-400 font-medium hidden md:table-cell">{team.won}</td>
                                    <td className="px-4 py-3 text-center text-amber-600 dark:text-amber-400 font-medium hidden md:table-cell">{team.draw}</td>
                                    <td className="px-4 py-3 text-center text-red-600 dark:text-red-400 font-medium hidden md:table-cell">{team.lost}</td>
                                    <td className="px-4 py-3 text-center text-muted-foreground hidden lg:table-cell">{team.goalsFor}</td>
                                    <td className="px-4 py-3 text-center text-muted-foreground hidden lg:table-cell">{team.goalsAgainst}</td>
                                    <td className="px-4 py-3 text-center font-bold">
                                      <span className={team.goalDifference > 0 ? "text-green-500" : team.goalDifference < 0 ? "text-red-500" : "text-muted-foreground"}>
                                        {team.goalDifference > 0 ? `+${team.goalDifference}` : team.goalDifference}
                                      </span>
                                    </td>
                                    <td className="px-4 py-3 text-center">
                                      <span className="text-lg font-black text-transparent bg-clip-text bg-gradient-to-br from-foreground to-foreground/70">
                                        {points}
                                      </span>
                                    </td>
                                    <td className="px-4 py-3 text-center">
                                      <div className="flex items-center justify-center gap-1">
                                        {form.map((res, i) => (
                                          <span key={i} className={`
                                            w-1.5 h-1.5 rounded-full
                                            ${res === 'W' ? 'bg-green-500' : res === 'D' ? 'bg-gray-400' : 'bg-red-500'}
                                          `}></span>
                                        ))}
                                      </div>
                                    </td>
                                  </motion.tr>
                                )
                              })}
                            </tbody>
                          </table>
                        </div>

                        <div className="mt-6 flex justify-between items-center text-xs text-muted-foreground px-2">
                          <p>Last updated: {new Date().toLocaleDateString()}</p>
                          <p className="hidden md:block">Stats provided by Football-Meta API</p>
                        </div>
                      </div>
                    </div>
                  </motion.div>
                )}
              </AnimatePresence>

              {/* ── Knockout bracket ── */}
              <AnimatePresence>
                {isCL && selectedSeason && viewMode === "knockout" && (
                  <motion.div
                    key="knockout-bracket"
                    initial={{ opacity: 0, y: 30 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: 20 }}
                    transition={{ duration: 0.5, type: "spring", bounce: 0.2 }}
                    className="w-full mt-4 mb-20"
                  >
                    <div className="relative overflow-hidden rounded-3xl bg-white/40 dark:bg-black/40 backdrop-blur-xl border border-white/20 dark:border-white/10 shadow-2xl">
                      <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-transparent via-orange-500 to-transparent opacity-50" />
                      <div className="absolute -top-24 -right-24 w-64 h-64 bg-orange-500/10 rounded-full blur-3xl" />
                      <div className="absolute -bottom-24 -left-24 w-64 h-64 bg-amber-500/10 rounded-full blur-3xl" />

                      <div className="p-6 md:p-8 relative z-10">
                        <div className="flex items-center gap-4 mb-8">
                          <div className="h-14 w-14 p-2 bg-white rounded-xl shadow-lg flex items-center justify-center ring-2 ring-orange-100 dark:ring-orange-900/30">
                            <img src={selectedCompetition.emblem} alt="logo" className="w-full h-full object-contain" />
                          </div>
                          <div>
                            <h3 className="text-2xl font-black text-foreground tracking-tight">Knockout Bracket</h3>
                            <p className="text-sm text-orange-600 dark:text-orange-400 font-bold uppercase tracking-wider">{selectedSeason.season} Season</p>
                          </div>
                        </div>

                        <KnockoutBracket data={knockoutData} />

                        <div className="mt-6 flex justify-between items-center text-xs text-muted-foreground px-2">
                          <p>Last updated: {new Date().toLocaleDateString()}</p>
                          <div className="hidden md:flex items-center gap-3">
                            <div className="flex items-center gap-1.5">
                              <span className="w-3 h-3 rounded-sm bg-orange-500/20 border border-orange-500/40" />
                              <span>Winner</span>
                            </div>
                            <span>·</span>
                            <span>L1 = Leg 1 · L2 = Leg 2 · AGG = Aggregate</span>
                          </div>
                        </div>
                      </div>
                    </div>
                  </motion.div>
                )}
              </AnimatePresence>

            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </>
  )
}

export default LeagueSelection;
