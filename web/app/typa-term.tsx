"use client";

type Row = {
  rank: number;
  username: string;
  score: number;
  wpm: number;
};

/** Matches terminal title art (TYPA, not TUPA). */
const TYPA_BANNER = [
  "████████╗ ██╗   ██╗ ██████╗   █████╗ ",
  "╚══██╔══╝ ╚██╗ ██╔╝ ██╔══██╗ ██╔══██╗",
  "   ██║     ╚████╔╝  ██████╔╝ ███████║",
  "   ██║      ╚██╔╝   ██╔═══╝  ██╔══██║",
  "   ██║       ██║    ██║      ██║  ██║",
  "   ╚═╝       ╚═╝    ╚═╝      ╚═╝  ╚═╝",
].join("\n");

export default function TypaTerm({ scores }: { scores: Row[] }) {
  return (
    <main
      style={{
        minHeight: "100vh",
        padding: "2rem 1.25rem 3rem",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        gap: "1.5rem",
        color: "#e4e4e7",
        boxSizing: "border-box",
        width: "100%",
      }}
    >
      <pre
        style={{
          color: "#a3e635",
          textAlign: "center",
          width: "100%",
          maxWidth: "52rem",
        }}
      >
        {`user@typa:~$ ssh play@typa.example` +
          "\n" +
          `Welcome to Typa. Good luck.`}
      </pre>

      <div className="typa-banner-wrap">
        <pre className="typa-banner">{TYPA_BANNER}</pre>
      </div>

      <p
        style={{
          margin: 0,
          maxWidth: "52rem",
          width: "100%",
          textAlign: "center",
          color: "#71717a",
          fontSize: "12px",
        }}
      >
        <span style={{ color: "#a3e635" }}>TYPA</span>
        {" — global leaderboard"}
      </p>

      <p
        style={{
          margin: 0,
          maxWidth: "52rem",
          width: "100%",
          textAlign: "center",
          color: "#52525b",
          fontSize: "12px",
        }}
      >
        $ curl -s typa.leaderboard | sort -nr -k3{" "}
        <span style={{ fontStyle: "italic" }}># fiction</span>
      </p>

      <div className="typa-table-wrap">
        <table className="typa-table" aria-label="Global leaderboard">
          <thead>
            <tr>
              <th scope="col">Rank</th>
              <th scope="col">Pilot</th>
              <th scope="col" className="num">
                Score
              </th>
              <th scope="col" className="num">
                WPM
              </th>
            </tr>
          </thead>
          <tbody>
            {scores.length === 0 ? (
              <tr>
                <td className="typa-empty" colSpan={4}>
                  No runs yet — deploy with{" "}
                  <code style={{ color: "#a3e635" }}>TYPA_API_SECRET</code>.
                </td>
              </tr>
            ) : (
              scores.map((s) => (
                <tr key={`${s.rank}-${s.username}`}>
                  <td data-label="Rank">{s.rank}</td>
                  <td className="pilot" data-label="Pilot">
                    {s.username}
                  </td>
                  <td className="num" data-label="Score">
                    {s.score}
                  </td>
                  <td className="num" data-label="WPM">
                    {s.wpm}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      <p
        style={{
          margin: 0,
          maxWidth: "52rem",
          width: "100%",
          textAlign: "center",
          color: "#71717a",
          fontSize: "12px",
          lineHeight: 1.5,
        }}
      >
        Scores: best run per player IP (hashed). SSH hosts send{" "}
        <code style={{ color: "#a3e635" }}>TYPA_CLIENT_IP</code>.
      </p>

      <footer
        style={{
          marginTop: "auto",
          paddingTop: "0.5rem",
          color: "#71717a",
          fontSize: "12px",
          textAlign: "center",
          width: "100%",
          maxWidth: "52rem",
        }}
      >
        typa · terminal typing chase
      </footer>
    </main>
  );
}
