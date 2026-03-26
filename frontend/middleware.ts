import { NextResponse, type NextRequest } from "next/server";
import { REFRESH_COOKIE } from "@/src/lib/auth/cookies";
import { CONSENT_MODE_COOKIE, SESSION_EXPIRES_AT_COOKIE } from "@/src/services/authStorage";

const PUBLIC_PATHS = ["/login", "/register"];

export function middleware(req: NextRequest) {
  const { pathname } = req.nextUrl;

  if (pathname.startsWith("/api/")) return NextResponse.next();
  if (pathname === "/" || PUBLIC_PATHS.some((p) => pathname === p || pathname.startsWith(p + "/"))) {
    return NextResponse.next();
  }

  // Everything else is protected (including /admin/* per chosen policy).
  const refresh = req.cookies.get(REFRESH_COOKIE)?.value ?? "";
  if (refresh) return NextResponse.next();

  const consentMode = req.cookies.get(CONSENT_MODE_COOKIE)?.value ?? "";
  if (consentMode === "localStorage") {
    const expiresAtRaw = req.cookies.get(SESSION_EXPIRES_AT_COOKIE)?.value ?? "";
    const expiresAt = Number(expiresAtRaw);
    if (expiresAt && Number.isFinite(expiresAt) && expiresAt > Date.now()) {
      return NextResponse.next();
    }
  }

  const url = req.nextUrl.clone();
  url.pathname = "/login";
  url.searchParams.set("next", pathname);
  return NextResponse.redirect(url);

}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"]
};

