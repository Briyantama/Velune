import type { Metadata } from "next";
import { ThemeProvider } from "@/src/components/providers/theme-provider";
import { QueryProvider } from "@/src/components/providers/query-provider";
import { Toaster } from "@/src/components/ui/toaster";
import "./globals.css";

export const metadata: Metadata = {
  title: "Velune",
  description: "Expense tracker, money manager, budget planner."
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body>
        <ThemeProvider>
          <Toaster>
            <QueryProvider>{children}</QueryProvider>
          </Toaster>
        </ThemeProvider>
      </body>
    </html>
  );
}

