import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "NetInspector",
  description: "Local network discovery and topology dashboard"
};

export default function RootLayout({
  children
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}

