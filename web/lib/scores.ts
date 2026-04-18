import { createHash } from "crypto";
import { mkdir, readFile, writeFile } from "fs/promises";
import path from "path";

export type ScoreRow = {
  ipHash: string;
  username: string;
  score: number;
  wpm: number;
  updated: string;
};

async function dataFile(): Promise<string> {
  if (process.env.VERCEL) {
    return "/tmp/typa-scores.json";
  }
  const dir = path.join(process.cwd(), "data");
  await mkdir(dir, { recursive: true });
  return path.join(dir, "scores.json");
}

export async function loadScores(): Promise<ScoreRow[]> {
  try {
    const buf = await readFile(await dataFile(), "utf-8");
    const parsed = JSON.parse(buf) as ScoreRow[];
    return Array.isArray(parsed) ? parsed : [];
  } catch {
    return [];
  }
}

export async function saveScores(rows: ScoreRow[]): Promise<void> {
  const f = await dataFile();
  await writeFile(f, JSON.stringify(rows, null, 2), "utf-8");
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
