import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Typa — leaderboard",
  description: "Public scores for the Typa terminal typing game",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
