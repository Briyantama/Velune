import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { ACCESS_COOKIE, REFRESH_COOKIE, cookieOptions } from "@/src/lib/auth/cookies";

export async function POST() {
  const jar = cookies();
  jar.set(ACCESS_COOKIE, "", { ...cookieOptions(), maxAge: 0 });
  jar.set(REFRESH_COOKIE, "", { ...cookieOptions(), maxAge: 0 });
  return NextResponse.json({ status: "ok" });
}

