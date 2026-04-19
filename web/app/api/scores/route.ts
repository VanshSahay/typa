import { NextRequest, NextResponse } from "next/server";
import {
  clampName,
  hashIP,
  submitScore,
  type ScoreRow,
} from "@/lib/scores";

export const runtime = "nodejs";

export async function POST(req: NextRequest) {
  const secret = process.env.TYPA_API_SECRET;
  if (!secret) {
    return NextResponse.json(
      { error: "TYPA_API_SECRET not configured" },
      { status: 503 },
    );
  }
  const auth = req.headers.get("authorization") || "";
  const tok = auth.startsWith("Bearer ") ? auth.slice(7) : "";
  if (tok !== secret) {
    return NextResponse.json({ error: "unauthorized" }, { status: 401 });
  }

  let body: { username?: string; score?: number; wpm?: number; ip?: string };
  try {
    body = await req.json();
  } catch {
    return NextResponse.json({ error: "invalid json" }, { status: 400 });
  }

  const username = clampName(String(body.username ?? ""));
  const score = Math.max(0, Math.floor(Number(body.score ?? 0)));
  const wpm = Math.max(0, Number(body.wpm ?? 0));
  const ip = String(body.ip ?? "0.0.0.0").slice(0, 64);

  if (!username || username.length < 1) {
    return NextResponse.json({ error: "username required" }, { status: 400 });
  }

  const row: ScoreRow = {
    ipHash: hashIP(ip),
    username,
    score,
    wpm,
    updated: new Date().toISOString(),
  };

  await submitScore(row);

  return NextResponse.json({ ok: true });
}
