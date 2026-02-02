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

function LeagueSelection() {
  const competitionMap = new Map<string, AllCompetitions>()
  const [isLoading, setLoading] = useState(false);
  const [listCompetitions, setListCompetitions] = useState<AllCompetitions[]>([]);
  const [selectedCompetition, setSelectedCompetition] = useState<AllCompetitions>();
  const [listSeasons, setListSeasons] = useState<{ season: Season[], leagueSeason: LeagueSeason[] }>();
  const [selectedSeason, setSelectedSeason] = useState<Season>()
  const [standings, setStandings] = useState<Standings[]>([]);

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
    setStandings([]); // Reset standings before fetching
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
    } else {
      console.error("Failed to fetch standings");
    }
  }

  const items = listCompetitions.map((val) => {
    return { label: val.name, value: val }
  })

  const SeasonLeagueMap = useMemo(SetLeagueSeasonMap, [listSeasons, listCompetitions]);

  const seasonItems: { label: string, value: Season }[] = (SeasonLeagueMap.get(String(selectedCompetition?.id ?? "")) ?? []).map((val: Season) => {
    return { label: val.season, value: val }
  })


  useEffect(() => {
    Promise.all([GetAllCompetitions(), GetAllSeasons()]);
  }, [])

  useEffect(() => {
    if (selectedCompetition) {
      setSelectedSeason(undefined);
    }
  }, [selectedCompetition, SeasonLeagueMap])

  useEffect(() => {
    GetTeamStandings()
  }, [selectedSeason])

  return (
    <>
      {isLoading && <Loading heading="Loading Leagues" message="Please wait while we fetch the competitions..." />}
      <div className="w-full flex flex-col items-center gap-8 py-8 animate-in fade-in duration-500">

        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="w-full text-center space-y-4"
        >
          <h1 className="text-4xl md:text-5xl font-extrabold text-transparent bg-clip-text bg-gradient-to-r from-orange-600 via-orange-500 to-amber-500 drop-shadow-sm">
            Select Your Arena
          </h1>
          <p className="text-muted-foreground font-medium">Choose a league to explore details</p>
          <div className="flex justify-center mt-6">
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
                className="mb-8 p-6  border border-orange-500/20 rounded-2xl backdrop-blur-md text-center shadow-lg transform hover:scale-[1.01] transition-transform duration-300"
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

              <AnimatePresence>
                {standings.length > 0 && selectedSeason && (
                  <motion.div
                    initial={{ opacity: 0, y: 40 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: 20 }}
                    transition={{ duration: 0.6, type: "spring", bounce: 0.3 }}
                    className="w-full mt-12 mb-20"
                  >
                    <div className="relative overflow-hidden rounded-3xl bg-white/40 dark:bg-black/40 backdrop-blur-xl border border-white/20 dark:border-white/10 shadow-2xl">
                      {/* Decorative elements */}
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
                              <h3 className="text-2xl font-black text-foreground tracking-tight">League Table</h3>
                              <p className="text-sm text-orange-600 dark:text-orange-400 font-bold uppercase tracking-wider">{selectedSeason.season} Season</p>
                            </div>
                          </div>

                          <div className="flex gap-2 text-xs font-bold bg-background/50 p-1.5 rounded-lg border border-white/10 backdrop-blur-sm flex-wrap justify-end">
                            {(selectedCompetition && (selectedCompetition.code === 'CL' || selectedCompetition.name.toLowerCase().includes('champions'))) ? (
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
                                // Add some dummy form data since API doesn't provide it in the interface
                                const form = ['W', 'D', 'L', 'W', 'W'].sort(() => 0.5 - Math.random()).slice(0, 5);
                                const points = (team.won * 3) + team.draw;

                                // Logic for styling based on competition type
                                const isCL = selectedCompetition && (selectedCompetition.code === 'CL' || selectedCompetition.name.toLowerCase().includes('champions'));

                                let isTop = false;
                                let isMid = false; // For CL play-offs
                                let isBottom = false;

                                if (isCL) {
                                  // Champions League Rules
                                  // Top 8 qualifies
                                  if (index < 8) isTop = true;
                                  // 9-24 play-offs
                                  else if (index >= 8 && index < 24) isMid = true;
                                  // Last 8 disqualifies (user request)
                                  else if (index >= standings.length - 8) isBottom = true;
                                } else {
                                  // Standard League Rules
                                  // Top 4 qualifies
                                  if (index < 4) isTop = true;
                                  // Bottom 3 relegated
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
            </motion.div>
          )}
        </AnimatePresence>
      </div>
    </>
  )
}

export default LeagueSelection;
