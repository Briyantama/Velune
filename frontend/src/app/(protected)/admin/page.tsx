import Link from "next/link";
import { PageHeader } from "@/src/components/common/page-header";
import { Card, CardContent } from "@/src/components/ui/card";

export default function AdminHomePage() {
  return (
    <div>
      <PageHeader
        title="Admin"
        description="Operational tooling (DLQ, outbox, reconciliation, health)."
      />
      <Card>
        <CardContent className="p-6 text-sm">
          <div className="grid gap-2">
            <Link className="underline underline-offset-4" href="/admin/health">
              Health overview
            </Link>
            <Link className="underline underline-offset-4" href="/admin/dlq">
              DLQ
            </Link>
            <Link className="underline underline-offset-4" href="/admin/outbox">
              Outbox
            </Link>
            <Link
              className="underline underline-offset-4"
              href="/admin/reconciliation"
            >
              Reconciliation
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
