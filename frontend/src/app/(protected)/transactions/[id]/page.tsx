import TransactionDetailClient from "@/src/features/transactions/components/transaction-detail-client";

export default function TransactionDetailPage({ params }: { params: { id: string } }) {
  return <TransactionDetailClient id={params.id} />;
}

