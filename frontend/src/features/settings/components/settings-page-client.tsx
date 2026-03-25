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
import { authClient } from "@/src/features/auth/services/authClient";

export default function SettingsPageClient() {
  const toast = useApiToasts();
  const meQ = useQuery({
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
              <KV k="Email" v={meQ.data?.email ?? "unknown"} />
              <KV k="Base currency" v={meQ.data?.baseCurrency ?? "USD"} />
              <KV k="User ID" v={meQ.data?.userId ?? "unknown"} />
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
