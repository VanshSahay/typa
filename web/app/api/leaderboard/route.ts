import { NextResponse } from "next/server";
import { loadScores } from "@/lib/scores";

export const runtime = "nodejs";

export async function GET() {
  const rows = await loadScores();
  const top = rows.slice(0, 100).map((r, i) => ({
    rank: i + 1,
    username: r.username,
    score: r.score,
    wpm: Math.round(r.wpm),
    updated: r.updated,
  }));
  return NextResponse.json({ scores: top });
}
