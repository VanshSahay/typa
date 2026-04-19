import { createHash } from "crypto";
import { mkdir, readFile, writeFile } from "fs/promises";
import { neon } from "@neondatabase/serverless";
import path from "path";

export type ScoreRow = {
  ipHash: string;
  username: string;
  score: number;
  wpm: number;
  updated: string;
};

type Sql = ReturnType<typeof neon>;

function getNeon(): Sql | null {
  const url = process.env.DATABASE_URL?.trim();
  if (!url) {
    return null;
  }
  return neon(url);
}

let schemaPromise: Promise<void> | null = null;

function ensureSchema(sql: Sql): Promise<void> {
  if (!schemaPromise) {
    schemaPromise = (async () => {
      await sql`
        CREATE TABLE IF NOT EXISTS leaderboard_scores (
          ip_hash VARCHAR(32) PRIMARY KEY,
          username VARCHAR(24) NOT NULL,
          score INTEGER NOT NULL CHECK (score >= 0),
          wpm DOUBLE PRECISION NOT NULL CHECK (wpm >= 0),
          updated_at TIMESTAMPTZ NOT NULL
        )
      `;
      await sql`
        CREATE INDEX IF NOT EXISTS leaderboard_scores_score_idx
        ON leaderboard_scores (score DESC)
      `;
    })();
  }
  return schemaPromise;
}

async function dataFile(): Promise<string> {
  if (process.env.VERCEL) {
    return "/tmp/typa-scores.json";
  }
  const dir = path.join(process.cwd(), "data");
  await mkdir(dir, { recursive: true });
  return path.join(dir, "scores.json");
}

async function loadScoresFromFile(): Promise<ScoreRow[]> {
  try {
    const buf = await readFile(await dataFile(), "utf-8");
    const parsed = JSON.parse(buf) as ScoreRow[];
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

function rowFromDb(r: {
  ip_hash: string;
  username: string;
  score: number;
  wpm: number;
  updated_at: Date | string;
}): ScoreRow {
  const updated =
    r.updated_at instanceof Date
      ? r.updated_at.toISOString()
      : String(r.updated_at);
  return {
    ipHash: r.ip_hash,
    username: r.username,
    score: r.score,
    wpm: r.wpm,
    updated,
  };
}

type DbScoreRow = {
  ip_hash: string;
  username: string;
  score: number;
  wpm: number;
  updated_at: Date | string;
};

async function loadScoresFromDatabase(sql: Sql): Promise<ScoreRow[]> {
  await ensureSchema(sql);
  const result = await sql`
    SELECT ip_hash, username, score, wpm, updated_at
    FROM leaderboard_scores
    ORDER BY score DESC
    LIMIT 500
  `;
  const rows = result as unknown as DbScoreRow[];
  return rows.map((r) => rowFromDb(r));
}

export async function loadScores(): Promise<ScoreRow[]> {
  const sql = getNeon();
  if (sql) {
    return loadScoresFromDatabase(sql);
  }
  return loadScoresFromFile();
}

async function saveScoresToFile(rows: ScoreRow[]): Promise<void> {
  const f = await dataFile();
  await writeFile(f, JSON.stringify(rows, null, 2), "utf-8");
}

export async function saveScores(rows: ScoreRow[]): Promise<void> {
  if (getNeon()) {
    throw new Error("saveScores is not used when DATABASE_URL is set");
  }
  await saveScoresToFile(rows);
}

async function submitScoreToDatabase(sql: Sql, row: ScoreRow): Promise<void> {
  await ensureSchema(sql);
  await sql`
    INSERT INTO leaderboard_scores (ip_hash, username, score, wpm, updated_at)
    VALUES (
      ${row.ipHash},
      ${row.username},
      ${row.score},
      ${row.wpm},
      ${row.updated}::timestamptz
    )
    ON CONFLICT (ip_hash) DO UPDATE SET
      username = EXCLUDED.username,
      score = EXCLUDED.score,
      wpm = EXCLUDED.wpm,
      updated_at = EXCLUDED.updated_at
    WHERE leaderboard_scores.score < EXCLUDED.score
  `;
}

async function submitScoreToFile(row: ScoreRow): Promise<void> {
  const rows = await loadScoresFromFile();
  const next = upsert(rows, row);
  await saveScoresToFile(next);
}

/** Persists one run: best score per ipHash (same rules as file upsert). */
export async function submitScore(row: ScoreRow): Promise<void> {
  const sql = getNeon();
  if (sql) {
    await submitScoreToDatabase(sql, row);
    return;
  }
  await submitScoreToFile(row);
}

export function hashIP(ip: string): string {
  const salt = process.env.TYPA_IP_SALT || "typa-dev-salt";
  return createHash("sha256")
    .update(`${salt}:${ip}`)
    .digest("hex")
    .slice(0, 32);
}

/** Keeps the best score per ipHash; replaces row when new score is higher. */
export function upsert(rows: ScoreRow[], incoming: ScoreRow): ScoreRow[] {
  const map = new Map(rows.map((r) => [r.ipHash, r]));
  const prev = map.get(incoming.ipHash);
  if (!prev || incoming.score > prev.score) {
    map.set(incoming.ipHash, incoming);
  }
  return [...map.values()].sort((a, b) => b.score - a.score);
}

export function clampName(s: string): string {
  const t = s.trim();
  return t.slice(0, 24);
}
