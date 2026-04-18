import { loadScores } from "@/lib/scores";
import TypaTerm from "./typa-term";

export const dynamic = "force-dynamic";

export default async function Home() {
  const rows = await loadScores();
  const scores = rows.slice(0, 100).map((r, i) => ({
    rank: i + 1,
    username: r.username,
    score: r.score,
    wpm: Math.round(r.wpm),
    updated: r.updated,
  }));

  return <TypaTerm scores={scores} />;
}
