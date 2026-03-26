"use client";

import { useQuery } from "@tanstack/react-query";
import { useTheme } from "next-themes";
import { PageHeader } from "@/src/components/common/page-header";
import { EmptyState } from "@/src/components/common/empty-state";
import { LoadingSkeleton } from "@/src/components/common/loading-skeleton";
import { ConfirmDialog } from "@/src/components/common/confirm-dialog";
import { Button } from "@/src/components/ui/button";
import { Card, CardContent } from "@/src/components/ui/card";
import { queryKeys } from "@/src/lib/query/keys";
import { useApiToasts } from "@/src/lib/api/toast";
import { authClient, Me } from "@/src/features/auth/services/authClient";
import { useAppDispatch } from "@/src/store/hooks";
import { clearAuthState } from "@/src/store/slices/authSlice";
import { clearMe } from "@/src/store/slices/userSessionSlice";
import { useAppSelector } from "@/src/store/hooks";
import { setConsent, type StorageMode } from "@/src/store/slices/consentSlice";
import { clearAllLocalSessionArtifacts, setConsentModeCookie } from "@/src/services/authStorage";

export default function SettingsPageClient() {
  const toast = useApiToasts();
  const dispatch = useAppDispatch();
  const storageMode = useAppSelector((s) => s.consent.storageMode);
  const meQ = useQuery<Me>({
    queryKey: queryKeys.me(),
    queryFn: () => authClient.me(),
  });
  const { theme, setTheme } = useTheme();

  if (meQ.isLoading) {
    return (
      <div>
        <PageHeader
          title="Settings"
          description="Preferences and session controls."
        />
        <LoadingSkeleton />
      </div>
    );
  }

  if (meQ.isError) {
    return (
      <div>
        <PageHeader
          title="Settings"
          description="Preferences and session controls."
        />
        <EmptyState
          title="Couldn’t load profile"
          description="Check auth and try again."
          actionLabel="Retry"
          onAction={() => meQ.refetch()}
        />
      </div>
    );
  }

  return (
    <div>
      <PageHeader
        title="Settings"
        description="Preferences and session controls."
      />

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        <Card>
          <CardContent className="p-6">
            <div className="text-base font-semibold">Profile</div>
            <dl className="mt-4 grid grid-cols-1 gap-3 text-sm">
              <KV k="Email" v={meQ.data?.email ?? "N/A"} />
              <KV k="Base currency" v={meQ.data?.baseCurrency ?? "IDR"} />
              <KV k="User ID" v={meQ.data?.userId ?? "N/A"} />
            </dl>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <div className="text-base font-semibold">Appearance</div>
            <div className="mt-4 flex items-center gap-2">
              <Button
                variant={theme === "light" ? "default" : "outline"}
                onClick={() => setTheme("light")}
              >
                Light
              </Button>
              <Button
                variant={theme === "dark" ? "default" : "outline"}
                onClick={() => setTheme("dark")}
              >
                Dark
              </Button>
              <Button
                variant={theme === "system" ? "default" : "outline"}
                onClick={() => setTheme("system")}
              >
                System
              </Button>
            </div>
            <div className="mt-2 text-sm text-muted-foreground">
              Theme preference is stored client-side.
            </div>
          </CardContent>
        </Card>

        <Card className="lg:col-span-2">
          <CardContent className="p-6">
            <div className="text-base font-semibold">Notifications</div>
            <div className="mt-2 text-sm text-muted-foreground">
              Notification preferences API isn’t exposed yet. When available,
              this section will manage delivery channels and thresholds.
            </div>
          </CardContent>
        </Card>

        <Card className="lg:col-span-2">
          <CardContent className="p-6">
            <div className="text-base font-semibold">Session</div>
            <div className="mt-4">
              <ConfirmDialog
                title="Log out?"
                description="This clears your session cookies on this device."
                confirmLabel="Log out"
                variant="destructive"
                onConfirm={async () => {
                  try {
                    await authClient.logout();
                    dispatch(clearAuthState());
                    dispatch(clearMe());
                    toast.showSuccess("Logged out");
                    window.location.href = "/login";
                  } catch (e) {
                    toast.showError(e, "Logout failed");
                  }
                }}
              >
                <Button variant="destructive">Log out</Button>
              </ConfirmDialog>
            </div>

            <div className="mt-6 rounded-2xl border bg-card p-4">
              <div className="text-sm font-semibold">Storage mode</div>
              <div className="mt-1 text-sm text-muted-foreground">
                {storageMode === "localStorage"
                  ? "Auth is stored in your browser (local storage) instead of cookies."
                  : "Auth is stored in secure cookies (recommended)."}
              </div>

              <div className="mt-4 grid grid-cols-1 gap-2 sm:grid-cols-2">
                <ConfirmDialog
                  title="Use cookies?"
                  description="Switch auth storage to cookies and sign in again."
                  confirmLabel="Use cookies"
                  variant="default"
                  onConfirm={async () => {
                    try {
                      await authClient.logout();
                      clearAllLocalSessionArtifacts();
                      setConsentModeCookie("cookie");
                      dispatch(setConsent({ storageMode: "cookie" }));
                      dispatch(clearAuthState());
                      dispatch(clearMe());
                      toast.showSuccess("Storage updated");
                      window.location.href = "/login";
                    } catch (e) {
                      toast.showError(e, "Update failed");
                    }
                  }}
                >
                  <Button variant={storageMode === "cookie" ? "default" : "outline"}>
                    Use cookies
                  </Button>
                </ConfirmDialog>

                <ConfirmDialog
                  title="Use local storage?"
                  description="Switch auth storage to local storage and sign in again."
                  confirmLabel="Use local storage"
                  variant="default"
                  onConfirm={async () => {
                    try {
                      await authClient.logout();
                      clearAllLocalSessionArtifacts();
                      setConsentModeCookie("localStorage");
                      dispatch(setConsent({ storageMode: "localStorage" }));
                      dispatch(clearAuthState());
                      dispatch(clearMe());
                      toast.showSuccess("Storage updated");
                      window.location.href = "/login";
                    } catch (e) {
                      toast.showError(e, "Update failed");
                    }
                  }}
                >
                  <Button variant={storageMode === "localStorage" ? "default" : "outline"}>
                    Use local storage
                  </Button>
                </ConfirmDialog>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

function KV({ k, v }: { k: string; v: string }) {
  return (
    <div>
      <dt className="text-xs font-semibold text-muted-foreground">{k}</dt>
      <dd className="mt-1 break-words">{v}</dd>
    </div>
  );
}
