"use client";

import { LoaderCircleIcon, Trash2Icon } from "lucide-react";
import { useState } from "react";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { useDeleteTransaction } from "@/features/transactions/hooks/use-delete-transaction";
import type { Transaction } from "@/features/transactions/types/transaction";
import { ApiError } from "@/lib/api/client";

function mutationErrorMessage(error: Error | null) {
  if (error instanceof ApiError && error.status === 404) {
    return "This transaction no longer exists or has already been deleted.";
  }

  return "We could not delete this transaction. Please try again.";
}

export function DeleteTransactionDialog({
  transaction,
}: {
  transaction: Transaction;
}) {
  const [open, setOpen] = useState(false);
  const deleteTransaction = useDeleteTransaction();
  const transactionLabel = transaction.description || `${transaction.type} transaction`;

  function handleOpenChange(nextOpen: boolean) {
    setOpen(nextOpen);
    if (!nextOpen) {
      deleteTransaction.reset();
    }
  }

  async function handleDelete() {
    try {
      await deleteTransaction.mutateAsync(transaction.id);
      handleOpenChange(false);
    } catch {
      // The mutation error is rendered in the confirmation dialog.
    }
  }

  return (
    <AlertDialog open={open} onOpenChange={handleOpenChange}>
      <AlertDialogTrigger asChild>
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          aria-label={`Delete ${transactionLabel}`}
          className="text-muted-foreground hover:text-destructive"
        >
          <Trash2Icon aria-hidden="true" />
        </Button>
      </AlertDialogTrigger>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete transaction?</AlertDialogTitle>
          <AlertDialogDescription>
            This permanently removes {transactionLabel}. This action cannot be
            undone.
          </AlertDialogDescription>
        </AlertDialogHeader>

        {deleteTransaction.isError ? (
          <p
            role="alert"
            className="mt-4 rounded-lg border border-destructive/20 bg-destructive/5 px-3 py-2.5 text-sm text-destructive"
          >
            {mutationErrorMessage(deleteTransaction.error)}
          </p>
        ) : null}

        <AlertDialogFooter>
          <AlertDialogCancel asChild>
            <Button variant="outline" disabled={deleteTransaction.isPending}>
              Cancel
            </Button>
          </AlertDialogCancel>
          <AlertDialogAction asChild>
            <Button
              variant="destructive"
              disabled={deleteTransaction.isPending}
              onClick={(event) => {
                event.preventDefault();
                void handleDelete();
              }}
            >
              {deleteTransaction.isPending ? (
                <LoaderCircleIcon
                  className="animate-spin motion-reduce:animate-none"
                  aria-hidden="true"
                />
              ) : null}
              {deleteTransaction.isPending ? "Deleting…" : "Delete transaction"}
            </Button>
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
